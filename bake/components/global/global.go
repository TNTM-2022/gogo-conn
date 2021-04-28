package global

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	concurrentMap "github.com/orcaman/concurrent-map"
	"gogo-connector/libs/mqtt"
	"log"
	"sync"
	"time"
)

const UserCap = 10

type SessionID uint32
type UserID uint32
type ServerID string

type UserChannel struct {
	ServerId  ServerID  `json:"sv"`
	SessionId SessionID `json:"sn"`
}

var Users = concurrentMap.New() // make(map[uint64]int32)
var Sids = concurrentMap.New()
var BlackList = concurrentMap.New()

func init() {
	for i := 0; i < UserCap; i++ {
		//fmt.Println(i, cap(sidPool.pool))
		sidPool.pool = append(sidPool.pool, SessionID(i))
	}
}

// session/sid 缓冲池，预防sid一直增大直到溢出.
var sidPool = &struct {
	pool   []SessionID
	locker sync.RWMutex
}{
	make([]SessionID, UserCap, UserCap+1),
	sync.RWMutex{},
}

func GetSid() (sid SessionID, ok bool) {
	sidPool.locker.Lock()
	defer sidPool.locker.Unlock()
	if len(sidPool.pool) == 0 {
		return 0, false
	}
	sid, sidPool.pool = sidPool.pool[0], sidPool.pool[1:]
	return sid, true
}
func BackSid(sid SessionID) { // 内存可能会出现问题， 最好用 栈 方式
	sidPool.locker.Lock()
	defer sidPool.locker.Unlock()
	sidPool.pool = append(sidPool.pool, sid)
	return
}
func GetOnlineUserNum() int {
	sidPool.locker.RLock()
	defer sidPool.locker.RUnlock()
	return len(sidPool.pool)
}

type UserReq struct {
	UID        UserID
	Route      string
	ServerType string
	Payload    []byte
	PkgID      int64
	Sid        SessionID
}

type PkgBelong struct {
	UID         UserID
	StartAt     time.Time
	ClientPkgID int64
	Route       string
}

var RemoteStore = concurrentMap.New()     // serverId > registerInfo + mqttClient
var RemoteTypeStore = concurrentMap.New() // serverType > []serverIdStore + forwardChannel

type RemoteTypeStoreType struct {
	Servers   []*RemoteConnect
	ForwardCh chan UserReq
}

// RequestRemote
type RemoteConnect struct {
	//pkgid      int64
	//countMutex sync.Mutex

	Host       string
	Port       int32
	ServerID   string
	ServerType string

	MqttClient *mqtt.MQTT
}

func (conn *RemoteConnect) AddPublishCb(f func(*mqtt.MQTT, paho.Message)) {
	conn.MqttClient.OnPublishCb = f
}
func (conn *RemoteConnect) Start(connCb func(mqtt *mqtt.MQTT), publishCb func(*mqtt.MQTT, paho.Message)) error {
	log.Println("mqtt client connect to ", conn.Host, conn.Port, conn.ServerID)

	// todo pomelo bug  如果不修改connect事件返回， 这里将会一直堵着 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 中的这个  socket.on('connect', function (pkg) { 代码块内部
	mqttClient := mqtt.CreateMQTTClient(&mqtt.MQTT{
		Host:            conn.Host,                    // "127.0.0.1",
		Port:            fmt.Sprintf("%v", conn.Port), //"6050",
		ClientID:        conn.ServerID,                //"dfgdfg",
		SubscriptionQos: 1,                            // 1
		Persistent:      false,
		Order:           false,
		KeepAliveSec:    5,
		PingTimeoutSec:  10,

		OnConnectCb: connCb,
		OnPublishCb: publishCb,
	})

	conn.MqttClient = mqttClient
	mqttClient.Start()
	fmt.Println("is connected", mqttClient.IsConnected())

	//go func() {
	//	ss, ok := RemoteTypeStore.Get(conn.serverType)
	//	if !ok {
	//		fmt.Println("no found server in store.")
	//	}
	//	serverTypeInfo, _ := ss.(remoteTypeStoreType)
	//
	//	for msg := range serverTypeInfo.forwardCh {
	//		fmt.Println(">>>forward rpc to backend <<<", host, port, serverId, msg.ServerType)
	//		pkg, err := proto_coder.PbToJson(msg.Route, msg.Payload)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		fmt.Println(">><>><", string(pkg))
	//		msg.Payload = pkg
	//		p := conn.encode(&msg) // 后端 wrap
	//		if p == nil {
	//			continue
	//		}
	//		mqttClient.Publish("rpc", p, 0, true)
	//		fmt.Println("rpc send ok")
	//	}
	//}()

	return nil
}
