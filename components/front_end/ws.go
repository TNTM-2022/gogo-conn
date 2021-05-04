package front_end

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go-connector/config"
	"go-connector/global"
	"go-connector/libs/package_coder"
	libPomeloCoder "go-connector/libs/pomelo_coder"
	libProtobufCoder "go-connector/libs/protobuf_coder"
	"go-connector/logger"
	"net"
	"net/http/pprof"
	"reflect"
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
		logger.ERROR.Println(err)
	}
	dMsg.Body = buf

	serverType := strings.SplitN(dMsg.Route, ".", 2)[0]

	// package encode 跟 mqttclient 绑定
	return package_coder.BackendMsg{
		Route:      dMsg.Route,
		ServerType: serverType,
		Payload:    buf,
		PkgID:      dMsg.ID,
		Sid:        sid,
		//ServerId:   serverId, // todo 不清楚 应该不是 frontendserverid
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
		fmt.Println("多出来一个 mType", bMsg.MType, bMsg)
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
	fmt.Println("type: ", reflect.TypeOf(err))
	if websocket.IsUnexpectedCloseError(err) {
		fmt.Println("IsUnexpectedCloseError")
		return
	}
	if websocket.IsCloseError(err) {
		fmt.Println("IsCloseError")
		return
	}
	if v, ok := err.(net.Error); ok {
		fmt.Println("error net work;", v.Timeout(), v.Temporary(), v.Error())
		return
	}
}

func ws(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	pomeloCoder := libPomeloCoder.InitCoder()
	sid, ok := global.GetSid()
	if !ok {
		return nil
	}
	defer global.BackSid(sid)
	_sid := fmt.Sprintf("%v", sid)

	session := global.CreateSession(sid)
	if !session.Bind(sid) {
		fmt.Println("Bind userid error")
		return nil
	}
	defer session.Destroy()

	MsgFront := make(chan package_coder.BackendMsg, 100)
	defer func() {
		close(MsgFront)
		for m := range MsgFront {
			fmt.Println("msg lost;", m)
		}
	}()
	global.SidFrontChanStore.Set(_sid, MsgFront)
	defer global.SidFrontChanStore.Remove(_sid)
	_running := true
	go func() {
		defer func() {
			_running = false
		}()
		for _running {
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
					//fmt.Println("msg push", string(b))
					if err := ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
						fmt.Println("msg push closed>>2", err)
						return
					}
				}
			}
		}
	}()

	func() {
		defer func() {
			_running = false
		}()
		for _running {
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
							fmt.Println("msg push closed>>handshake", err)
							return
						}

					case libPomeloCoder.Package["TYPE_HANDSHAKE_ACK"]:
						if !pomeloCoder.HandleHandshakeAck() {
							return
						}

					case libPomeloCoder.Package["TYPE_HEARTBEAT"]:
						// todo 要怎么处理
						fmt.Println("TYPE_HEARTBEAT")

					case libPomeloCoder.Package["TYPE_DATA"]:
						{
							backendMsg := handleReq(pomeloCoder, mm.Body, sid)
							logger.DEBUG.Println("TYPE_DATA >> ", backendMsg.Route)
							serverType := strings.SplitN(backendMsg.Route, ".", 2)[0]
							if v, ok := global.RemoteBackendTypeForwardChan.Get(serverType); ok {
								if ch, ok := v.(chan package_coder.BackendMsg); ok {
									select {
									case ch <- backendMsg: // 会出现
									default:
										fmt.Println("写不进去")
									}
								}
							} else {
								// todo 请路径不存在, 增加全局 路径不存在拦截
								fmt.Println("路径不存在")
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

	e.Any("/debug/pprof/", func(ctx echo.Context) error {
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

	e.Logger.Fatal(e.Start("127.0.0.1:23456"))
}
