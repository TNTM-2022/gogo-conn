package front_end

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go-connector/config"
	"go-connector/filters"
	"go-connector/global"
	"go-connector/libs/package_coder"
	libPomeloCoder "go-connector/libs/pomelo_coder"
	libProtobufCoder "go-connector/libs/protobuf_coder"
	"go-connector/logger"
	"go.uber.org/zap"
	"net"
	"net/http/pprof"
	"strings"
)

var (
	upgrader = websocket.Upgrader{}
)

func handleReq(pomeloCoder *libPomeloCoder.Coder, buf []byte, sid uint32) package_coder.BackendMsg {
	// pomelo decode
	dMsg := pomeloCoder.HandleData(buf)

	// protobuf decode
	buf, err := libProtobufCoder.PbToJson(dMsg.Route, dMsg.Body)
	if err != nil {
		logger.ERROR.Println("empty return", zap.Error(err))
		return package_coder.BackendMsg{}
	}

	dMsg.Body = buf
	serverType := strings.SplitN(dMsg.Route, ".", 2)[0]

	// package encode 跟 mqttclient 绑定
	return package_coder.BackendMsg{
		Route:         dMsg.Route,
		ServerType:    serverType,
		Payload:       dMsg.Body,
		PkgID:         dMsg.ID,
		Sid:           sid,
		FrontServerId: *config.ServerID,

		MType:         dMsg.Type, // todo  没有实现剩下这几项
		CompressGzip:  dMsg.CompressGzip,
		CompressRoute: dMsg.CompressRoute,
	}
}

func handleSend(bMsg package_coder.BackendMsg) []byte {
	if bMsg.Route == "" {
		return nil
	}
	var (
		pkgType       int
		buf           []byte
		clientId      uint64 = bMsg.PkgID
		mType         byte
		compressRoute int
		compressGzip  bool

		isPush bool
	)
	switch bMsg.MType {
	case libPomeloCoder.Message["TYPE_PUSH"]:
		pkgType = libPomeloCoder.Package["TYPE_DATA"]
		mType = libPomeloCoder.Message["TYPE_PUSH"]
		isPush = true
	case libPomeloCoder.Message["TYPE_RESPONSE"]:
		pkgType = libPomeloCoder.Package["TYPE_DATA"]
		mType = libPomeloCoder.Message["TYPE_RESPONSE"]
		isPush = false
	default:
		logger.ERROR.Println("reserved mType", zap.Uint8("mType", bMsg.MType), zap.String("route", bMsg.Route))
		return nil
	}

	// package decode 在 mqtt server 进行处理
	// protobuf encode
	if b, e := libProtobufCoder.JsonToPb(bMsg.Route, bMsg.Payload, isPush); b != nil && e == nil {
		logger.DEBUG.Println(" json2bt转换成功-", bMsg.Route)
		bMsg.Payload = b
	} else {
		logger.DEBUG.Println(" json2pb转换失败", bMsg.Route)
	}

	// pomelo encode
	buf = libPomeloCoder.MessageEncode(clientId, mType, compressRoute, bMsg.Route, bMsg.Payload, compressGzip)
	buf = libPomeloCoder.PackageEncode(pkgType, buf)

	return buf
}

func judgeWsConnError(err error) {
	if websocket.IsUnexpectedCloseError(err) {
		logger.DEBUG.Println("front_end,ws,client", "client unexpected close error", zap.Error(err))
		return
	}
	if websocket.IsCloseError(err) {
		logger.DEBUG.Println("front_end,ws,client", "client close error", zap.Error(err))
		return
	}
	if v, ok := err.(net.Error); ok {
		logger.DEBUG.Println("front_end,ws,client", "network error", zap.Bool("isTimeout", v.Timeout()), zap.Bool("isTemporary", v.Temporary()), zap.String("error", v.Error()))
		return
	}
}

func ws(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = ws.Close()
	}()

	pomeloCoder := libPomeloCoder.InitCoder()
	sid := global.GetSid()

	_sid := fmt.Sprintf("%v", sid)

	session := global.CreateSession(sid)
	defer session.Destroy()
	if !session.Bind(sid) {
		logger.INFO.Println("userId bind error", zap.Uint32("sessionId", sid))
		return nil
	}

	MsgFront := make(chan package_coder.BackendMsg, 100)

	global.SidFrontChanStore.Set(_sid, MsgFront)
	defer global.SidFrontChanStore.Remove(_sid)

	go func() {
		defer func() {
			fmt.Println("close 1")
		}()
		for {
			// Write
			select {
			//case <-ctx.Done():
			//	return
			case bb := <-MsgFront:
				{
					b := handleSend(bb)
					if b == nil {
						return
					}
					fmt.Println("write..")
					if err := ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
						logger.ERROR.Println("msg push channel closed", zap.Error(err))
						return
					}
				}
			}
		}
	}()

	func() {
		defer func() {
			close(MsgFront)
			for range MsgFront {
				logger.DEBUG.Println("front_end,ws,send_msg,req,res,push", "msg lost for cleaning")
			}
		}()
		for {
			// Read
			messageType, p, err := ws.ReadMessage()
			if err != nil {
				judgeWsConnError(err)
				break
			}
			if websocket.BinaryMessage == messageType && p != nil {
				for _, mm := range libPomeloCoder.PackageDecode(p) {
					switch int(mm.Type) {
					case libPomeloCoder.Package["TYPE_HANDSHAKE"]:
						b := pomeloCoder.HandleHandshake()
						if err := ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
							logger.ERROR.Println("ws handshake error", zap.Error(err))
							return
						}

					case libPomeloCoder.Package["TYPE_HANDSHAKE_ACK"]:
						if !pomeloCoder.HandleHandshakeAck() {
							return
						}

					case libPomeloCoder.Package["TYPE_HEARTBEAT"]:
						// todo 要怎么处理
						logger.DEBUG.Println("front_end,ws_heartbeat,ws", "heartbeat")

					case libPomeloCoder.Package["TYPE_DATA"]:
						{
							backendMsg := handleReq(pomeloCoder, mm.Body, sid)
							logger.DEBUG.Println("front_end,ws", "TYPE_DATA", zap.String("route", backendMsg.Route), zap.String("payload", string(backendMsg.Payload)))
							serverType := strings.SplitN(backendMsg.Route, ".", 2)[0]
							if v, ok := global.RemoteBackendTypeForwardChan.Get(serverType); ok {
								if ch, ok := v.(chan package_coder.BackendMsg); ok {
									select {
									case ch <- backendMsg: // 会出现
									default:
										logger.ERROR.Println("cannot write in backend forward channel")
									}
								}
							} else {
								// todo 请路径不存在, 增加全局 路径不存在拦截 已经写了
								MsgFront <- filters.NoRouteFilter(backendMsg.Route, backendMsg.PkgID, backendMsg.CompressRoute, backendMsg.CompressGzip)
								logger.ERROR.Println("cannot find route", zap.String("route", backendMsg.Route), zap.String("serverType", serverType))
							}
						}
					case libPomeloCoder.Package["TYPE_KICK"]:
						fmt.Println("TYPE_KICK")
						// todo 完善
						return

					default:
					}
				}
			}
		}
	}()
	return nil
}

func StartFrontServer() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/ws", ws)

	e.GET("/debug/pprof/", func(ctx echo.Context) error {
		pprof.Index(ctx.Response().Writer, ctx.Request())
		return nil
	})
	e.Any("/debug/pprof/cmdline", func(ctx echo.Context) error {
		pprof.Cmdline(ctx.Response().Writer, ctx.Request())
		return nil
	})
	e.Any("/debug/pprof/profile", func(ctx echo.Context) error {
		pprof.Profile(ctx.Response().Writer, ctx.Request())
		return nil
	})
	e.Any("/debug/pprof/goroutine", func(ctx echo.Context) error {
		pprof.Handler("goroutine").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	})
	e.Any("/debug/pprof/symbol", func(ctx echo.Context) error {
		pprof.Symbol(ctx.Response().Writer, ctx.Request())
		return nil
	})

	e.Any("/debug/pprof/trace", func(ctx echo.Context) error {
		pprof.Trace(ctx.Response().Writer, ctx.Request())
		return nil
	})

	go func() {
		e.Logger.Fatal(e.Start("127.0.0.1:23456"))
	}()
}
