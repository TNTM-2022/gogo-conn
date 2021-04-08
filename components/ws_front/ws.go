package ws_front

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gogo-connector/components/global"
	pomelo_coder "gogo-connector/libs/pomelo_coder"
	"log"
	"net"
	"reflect"
	"runtime"

	"strconv"
	"sync"
	"time"
)

func StartWsServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// Echo instance
	e := echo.New()

	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/ws", ws)

	// Start server
	l, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal(err)
	}
	e.Listener = l
	host, port, err := net.SplitHostPort(l.Addr().String())
	log.Println("ws listen on port:", host, port)

	go func() { // graceful shutdown
		<-ctx.Done()

		ctxT, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := e.Shutdown(ctxT); err != nil {
			e.Logger.Error(err)
		}
	}()
	go func() {
		for { // interval report
			return
			time.Sleep(time.Duration(time.Second * 5))
			fmt.Printf("users: %d, sids: %d, goroutine: %d\n", global.Users.Count(), global.Sids.Count(), runtime.NumGoroutine())
		}
	}()
	e.Logger.Error(e.Start(""))
}

var upgrader = websocket.Upgrader{}

func ws(c echo.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var uid global.UserID // 校验过程已经获取了
	// todo 获取头部 进行校验， 如果ok 继续放行。 同时设置倒计时，防止攻击
	//strToken := c.Request().Header.Get("Sec-Websocket-Protocol")
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Print(err)
		return err
	}
	defer conn.Close()

	sid, ok := global.GetSid()
	if !ok {
		log.Fatal("no sid can use")
	}
	defer global.BackSid(sid)

	user := CreateUserConn(uid, sid, ctx, cancel, pomelo_coder.StateInited)
	global.Users.Set(strconv.FormatUint(uint64(uid), 10), &user)
	global.Sids.Set(strconv.FormatUint(uint64(sid), 10), uid)
	defer func() {
		global.Users.Remove(strconv.FormatUint(uint64(uid), 10))
		global.Sids.Remove(strconv.FormatUint(uint64(sid), 10))
		close(user.MsgPush)
		close(user.MsgResp)
		// todo 还没有处理接收chan
	}()

	go func() { // 处理读取
		defer func() {
			cancel()
			fmt.Println("close --==read==--")
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			messageType, p, err := conn.ReadMessage()
			fmt.Println("read message ** ", messageType, p, err)
			if err != nil {
				judgeWsConnError(err)
				return
			}
			if websocket.BinaryMessage == messageType && p != nil {
				m := pomelo_coder.PackageDecode(p)
				for _, mm := range m {
					packageTypeHandler(mm.Type, mm.Body, user)
				}
			}
		}
	}()

	go func() {
		defer func() {
			cancel()
			fmt.Println("close --==read==--")
		}()
		for { // 处理写入
			select {
			case <-ctx.Done():
				return
			case b := <-user.MsgPush:
				{
					fmt.Println("msg push", b)
					if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
						fmt.Println("msg push closed>>2", err)
						return
					}
				}
			case b := <-user.MsgResp:
				{
					fmt.Println("send >>2; err =>", err)
					fmt.Println("msg  res start")
					if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
						fmt.Println("closed>>2", err)
						return
					}
					fmt.Println("msg  res flush")

				}
			}
		}
	}()

	<-ctx.Done()
	return nil
}

//========================================================================

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
func CreateUserConn(uid global.UserID, sid global.SessionID, ctx context.Context, cancel context.CancelFunc, state int) *pomelo_coder.UserConn {
	return &pomelo_coder.UserConn{
		//MsgReq:  make(chan []byte, 1000),
		MsgResp: make(chan []byte, 1000),
		MsgPush: make(chan []byte, 1000),
		Kick:    make(chan []byte),
		//MsgSend: make(chan []byte, 1000),

		Tick:   time.Now(),
		UID:    uid,
		Ctx:    ctx,
		Cancel: cancel,
		State:  state,
		Sid:    sid,
	}
}

func packageTypeHandler(t byte, b []byte, user *pomelo_coder.UserConn) {
	switch int(t) {
	case pomelo_coder.Package["TYPE_HANDSHAKE"]:
		pomelo_coder.HandleHandshake(user)

	case pomelo_coder.Package["TYPE_HANDSHAKE_ACK"]:
		pomelo_coder.HandleHandshakeAck(user)

	case pomelo_coder.Package["TYPE_HEARTBEAT"]:
		fmt.Println("TYPE_HEARTBEAT")

	case pomelo_coder.Package["TYPE_DATA"]:
		fmt.Println("TYPE_DATA")
		pomelo_coder.HandleData(user, b)

	case pomelo_coder.Package["TYPE_KICK"]:
		fmt.Println("TYPE_KICK")

	default:
	}
	return
}

// 推送消息 全局只需要一个/负载几个
// 客户端发送消息 全局也只需要几个
// 客户端接收消息 全局只需要几个

// m := pomelo_coder.MessageEncode(0, pomelo_coder.Message["TYPE_PUSH"], 0, "_test.testHandler.test", []byte("ok"), false)
// p := pomelo_coder.PackageEncode(pomelo_coder.Package["TYPE_DATA"], m)
