package front_server

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go-connector/global"
	"go-connector/interfaces"
	"runtime"

	//servers "go-connector/libs/monitor"
	coder "go-connector/libs/pomelo_coder"
	"log"
	"math"
	"net"
	"reflect"
	"strconv"
	//"strings"
	"sync"
	"time"
)

// user conn 状态
const (
	StateInited  = 0
	StateWaitAck = 1
	StateWorking = 2
	StateClosed  = 3
)

// 握手状态
const (
	CODE_OK         = 200
	CODE_USE_ERROR  = 500
	CODE_OLD_CLIENT = 501
)

func StartWsServer(mc interfaces.MainControl, l net.Listener) {
	defer func() {
		time.Sleep(time.Duration(time.Second * 1))
		mc.Wg.Done()
	}()
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
	e.Listener = l
	go func() {
		<-mc.Ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Error(err)
		}
	}()
	go func () {
		for {
			time.Sleep(time.Duration(time.Second * 5))
			fmt.Printf("users: %d, sids: %d, goroutine: %d\n", global.Users.Count(), global.Sids.Count(), runtime.NumGoroutine())
		}
	}()
	e.Logger.Error(e.Start(""))
}

var upgrader = websocket.Upgrader{}
func ws(c echo.Context) error {
	//strToken := c.Request().Header.Get("Sec-Websocket-Protocol")
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Print(err)
		return err
	}
	defer conn.Close()

	var sid interfaces.Sid
	var uid interfaces.UserId
	ctx, cancel := context.WithCancel(context.Background())
	//uid = interfaces.Sid(getSid())

	user := interfaces.CreateUserConn(uid, sid, ctx, cancel, StateInited)
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
		//i := 100
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			messageType, p, err := conn.ReadMessage()
			fmt.Println(messageType, p, err)
			if err != nil {
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
			//i++
			if websocket.BinaryMessage == messageType && p != nil {
				m := coder.PackageDecode(p)
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

var sidCounterMutex sync.Mutex
var sidCounter uint64

func getSid() uint64 {
	sidCounterMutex.Lock()
	defer sidCounterMutex.Unlock()
	sidCounter++
	if sidCounter > math.MaxUint64-10 {
		sidCounter = 1
	}
	return sidCounter
}

//===================== handle pomelo protocol ===========================

func packageTypeHandler(t byte, b []byte, user *interfaces.UserConn) {
	switch int(t) {
	case coder.Package["TYPE_HANDSHAKE"]:
		handleHandshake(user)

	case coder.Package["TYPE_HANDSHAKE_ACK"]:
		handleHandshakeAck(user)

	case coder.Package["TYPE_HEARTBEAT"]:
		fmt.Println("TYPE_HEARTBEAT")

	case coder.Package["TYPE_DATA"]:
		fmt.Println("TYPE_DATA")
		handleData(user, b)

	case coder.Package["TYPE_KICK"]:
		fmt.Println("TYPE_KICK")

	default:
	}
	return
}

// 处理客户端handshake
// 检查整体状态 ST_INITED
// todo checkClient
// var opts = { heartbeat : setupHeartbeat(this) };
//  opts.useProto = true;
// 返回 TYPE_HANDSHAKE 报文， 携带上述的对象 packageEncode
func handleHandshake(user *interfaces.UserConn) {
	if user.State != StateInited {
		user.Cancel()
		return
	}
	s := handshake{
		Code: CODE_OK,
		Sys: sys{
			Heartbeat:   60,
			Dict:        dict{},
			RouteToCode: routeToCode{},
			CodeToRoute: codeToRoute{},
			DictVersion: genDictVersion(),
			UseDict:     true,
			UseProto:    true,
		},
	}
	j, _ := json.Marshal(s)
	p := coder.PackageEncode(coder.Package["TYPE_HANDSHAKE"], []byte(string(j)))
	user.MsgPush <- p
	user.State = StateWaitAck
	// fmt.Println("handshacke handler = 0")
}

//
func handleHandshakeAck(user *interfaces.UserConn) {
	if user.State != StateWaitAck {
		user.Cancel()
		return
	}
	user.State = StateWorking
}

func genDictVersion() string {
	m := md5.Sum([]byte("{}"))
	return base64.StdEncoding.EncodeToString(m[:])
}

func handleData(user *interfaces.UserConn, b []byte) {
	if user.State != StateWorking {
		user.Cancel()
		return
	}

	c := coder.MessageDecode(b)
	fmt.Println(c.Route, string(c.Body))
	//server := strings.SplitN(c.Route, ".", 2)[0]
	//fmt.Println("servertype", server, "server.len", len(servers.ServerTypeMap[server]))
	//if server != "" {
	//	servers.ServerOptLocker.RLock()
	//	ch := servers.ServerTypeChMap[server]
	//	servers.ServerOptLocker.RUnlock()
	//	if ch != nil {
	//		fmt.Println("将要写入消息", server)
	//		//ch <- userReqType.UserReq{
	//		//	UID:        user.UID,
	//		//	Route:      c.Route,
	//		//	ServerType: server,
	//		//	Payload:    c.Body,
	//		//	PkgID:      c.ID,
	//		//	Sid:        user.Sid,
	//		//}
	//		m := interfaces.UserReq{
	//			UID:        user.UID,
	//			Route:      c.Route,
	//			ServerType: server,
	//			Payload:    c.Body,
	//			PkgID:      c.ID,
	//			Sid:        user.Sid,
	//		}
	//		select {
	//		case ch <- m:
	//		default:
	//			{
	//				fmt.Println("写入失败，队列堵塞", server)
	//			}
	//		}
	//		fmt.Println("将要写入消息 done")
	//
	//	}
	//}
}

type KickMsg struct {
	Reason string `json:"reason"`
}

var defaultKick = []byte(`{"reson":"kick"}`)

func handleKick(code int32, msg string, conn *websocket.Conn) {
	r, err := json.Marshal(KickMsg{Reason: msg})
	if err != nil {
		r = defaultKick
	}
	b := coder.PackageEncode(coder.Package["TYPE_KICK"], r)
	if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		fmt.Println(err)
		return
	}
}

// 推送消息 全局只需要一个/负载几个
// 客户端发送消息 全局也只需要几个
// 客户端接收消息 全局只需要几个

// m := coder.MessageEncode(0, coder.Message["TYPE_PUSH"], 0, "_test.testHandler.test", []byte("ok"), false)
// p := coder.PackageEncode(coder.Package["TYPE_DATA"], m)
