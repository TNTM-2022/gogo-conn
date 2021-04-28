package front_end

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
		//ServerId:   serverId, // todo 不清楚

		MType:         dMsg.Type, // todo  没有实现剩下这几项
		CompressGzip:  dMsg.CompressGzip,
		CompressRoute: dMsg.CompressRoute,
	}
}

func handleSend(bmsg package_coder.BackendMsg) []byte {
	var (
		pkgType       int
		buf           []byte
		clientId      uint64
		mType         byte
		compressRoute int
		compressGzip  bool
		isPush        bool
	)

	switch bmsg.MType {
	case libPomeloCoder.Message["TYPE_PUSH"]:
		pkgType = libPomeloCoder.Package["TYPE_DATA"]
		mType = libPomeloCoder.Message["TYPE_PUSH"]
		isPush = true
	case libPomeloCoder.Message["TYPE_RESPONSE"]:
		pkgType = libPomeloCoder.Package["TYPE_DATA"]
		mType = libPomeloCoder.Message["TYPE_RESPONSE"]
		isPush = false
	}

	// package decode 在 mqtt server 进行处理
	// protobuf encode
	if b, e := libProtobufCoder.JsonToPb(bmsg.Route, bmsg.Payload, isPush); b != nil && e == nil {
		logger.DEBUG.Println(" json2bt转换成功-", bmsg.Route)
		bmsg.Payload = b
	} else {
		logger.DEBUG.Println(" json2pb转换失败", bmsg.Route)
	}

	// pomelo encode
	buf = libPomeloCoder.MessageEncode(clientId, mType, compressRoute, bmsg.Route, bmsg.Payload, compressGzip)
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
		fmt.Println("error net work")
		fmt.Println(v.Timeout(), v.Temporary(), v.Error())
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
	defer global.BackSid(sid)
	if !ok {
		return nil
	}

	MsgSend, MsgReceive := make(chan []byte, 100), make(chan package_coder.BackendMsg, 100)

	go func() {
		for {
			// Write
			select {
			//case <-ctx.Done():
			//	return
			case bb := <-MsgReceive:
				{
					b := handleSend(bb)
					fmt.Println("msg push", b)
					if err := ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
						fmt.Println("msg push closed>>2", err)
						return
					}
				}
			}
		}
	}()

	go func() {
		for {
			// Read
			messageType, p, err := ws.ReadMessage()
			fmt.Println("read message ** ", messageType, p, err)
			if err != nil {
				judgeWsConnError(err)
				break
			}
			if websocket.BinaryMessage == messageType && p != nil {
				for _, mm := range libPomeloCoder.PackageDecode(p) {
					switch int(mm.Type) {
					case libPomeloCoder.Package["TYPE_HANDSHAKE"]:
						MsgSend <- pomeloCoder.HandleHandshake()

					case libPomeloCoder.Package["TYPE_HANDSHAKE_ACK"]:
						if !pomeloCoder.HandleHandshakeAck() {
							return
						}

					case libPomeloCoder.Package["TYPE_HEARTBEAT"]:
						// todo 要怎么处理
						fmt.Println("TYPE_HEARTBEAT")

					case libPomeloCoder.Package["TYPE_DATA"]:
						{
							fmt.Println("TYPE_DATA")
							backendMsg := handleReq(pomeloCoder, mm.Body, sid)
							logger.DEBUG.Println(backendMsg.Route)
							// todo 进行转发
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

func main() {
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
	e.Any("/debug/pprof/symbol", func(ctx echo.Context) error {
		pprof.Symbol(ctx.Response().Writer, ctx.Request())
		return nil
	})

	e.Any("/debug/pprof/trace", func(ctx echo.Context) error {
		pprof.Trace(ctx.Response().Writer, ctx.Request())
		return nil
	})

	e.Logger.Fatal(e.Start(":1323"))
}
