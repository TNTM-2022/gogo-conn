package mqtt_client

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
	"go-connector/logger"
	"log"
	"math"
	"sync"
	"time"
)

type CallBack func(err string, data []byte)
type MessageHandler paho.MessageHandler

type MQTT struct {
	Host            string
	Port            string
	ClientID        string
	username        string
	password        string
	KeepAliveSec    int
	PingTimeoutSec  int
	verbose         bool
	connecting      bool
	disconnected    bool
	Order           bool
	client          paho.Client
	connectToken    *paho.ConnectToken
	SubscriptionQos byte
	Persistent      bool

	OnConnectCb func(mqtt *MQTT)                           //paho.OnConnectHandler
	OnPublishCb func(client paho.Client, msg paho.Message) // paho.MessageHandler

	reqId       int64
	reqIdLocker sync.Mutex
	Callbacks   cmap.ConcurrentMap

	connectedNum int
}

func initMQTTClientOps(client *MQTT) (*paho.ClientOptions, error) {
	opts := paho.NewClientOptions()

	if client.username != "" {
		opts.SetUsername(client.username)
	}
	if client.password != "" {
		opts.SetPassword(client.password)
	}

	opts.AddBroker(fmt.Sprintf("tcp://%s:%v", client.Host, client.Port))
	opts.SetConnectTimeout(time.Second * 5)
	opts.SetClientID(client.ClientID)
	opts.SetCleanSession(!client.Persistent)
	opts.SetOrderMatters(client.Order)
	opts.SetKeepAlive(time.Duration(client.KeepAliveSec) * time.Second)
	opts.SetPingTimeout(time.Duration(client.PingTimeoutSec) * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 2)

	opts.SetDefaultPublishHandler(client.publishHandler)
	opts.SetOnConnectHandler(client.connectHandler)
	opts.SetReconnectingHandler(client.reconnectHandler)

	return opts, nil
}
// CreateMQTTClient creates a new MQTT client
func CreateMQTTClient(mqttClient *MQTT) *MQTT {
	mqttClient.Callbacks = cmap.New()

	opts, err := initMQTTClientOps(mqttClient)
	if err != nil {
		log.Fatalf("unable to configure MQTT client: %s", err)
	}

	pahoClient := paho.NewClient(opts)
	mqttClient.client = pahoClient
	return mqttClient
}
func (m *MQTT) Stop() {
	m.client.Disconnect(500)
}
func (m *MQTT) Start() {
	log.Printf("Starting MQTT client on tcp://%s:%v with Prefix:%v, Persistence:%v, OrderMatters:%v, KeepAlive:%v, PingTimeout:%v, QOS:%v", m.Host, m.Port, "", true, m.Order, m.KeepAliveSec, m.PingTimeoutSec, 1)
	t := m.client.Connect()
	t.Wait() //Timeout(time.Second * 2) // pinus 问题， 没有 connack确认。 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 44L。 client.connack({ returnCode: 0 });
	logger.ERROR.Println("error ", t.Error())
	if t.Error() != nil {
		log.Panicf("cannot connect to %v server;  mqtt monitor timeout. %v:%v", m.ClientID, m.Host, m.Port)
	}
	logger.DEBUG.Println("mqtt started??")
}

func (m *MQTT) reconnectHandler(_ paho.Client, opts *paho.ClientOptions) {
	fmt.Printf("reconnecting to %v server ....\n", opts.ClientID)
}
func (m *MQTT) connectionLostHandler(_ paho.Client, err error) {
	log.Printf("MQTT client lost connection: %v", err)
	m.disconnected = true
}
func (m *MQTT) IsReconnect() bool {
	return m.connectedNum > 1
}
func (m *MQTT) IsConnected() bool {
	return m.client.IsConnected()
}
func (m *MQTT) connectHandler(_ paho.Client) {
	logger.DEBUG.Printf("MQTT client connected")
	m.connectedNum++
	if m.OnConnectCb != nil {
		m.OnConnectCb(m)
	}
	m.disconnected = false
}
func (m *MQTT) Publish(topic string, message []byte, qos byte, _async bool) {
	token := m.client.Publish(topic, qos, false, message)
	if token.Error() != nil {
		logger.ERROR.Println("publish error ", token.Error(), m.client.IsConnected())
	}
	if !_async {
		token.Wait()
	}
}


func (m *MQTT) GetReqId() int64 {
	m.reqIdLocker.Lock()
	defer m.reqIdLocker.Unlock()
	m.reqId++
	if m.reqId > math.MaxInt64-2 {
		m.reqId = 2
	}
	return m.reqId
}

func (m *MQTT) publishHandler(client paho.Client, msg paho.Message) {
	if m.OnPublishCb == nil {
		return
	}
	m.OnPublishCb(client, msg)
}

//func (m *MQTT) publishHandler(client paho.Client, msg paho.Message) {
//	if msg.Topic() == "monitor" {
//		var mm struct {
//			RespId int64           `json:"respId"`
//			Body   json.RawMessage `json:"body"`
//			Error  string          `json:"error"`
//		}
//		var ss string
//		d := msg.Payload()
//		if e := json.Unmarshal(d, &ss); e == nil {
//			d = []byte(ss)
//		}
//		if e := json.Unmarshal(d, &mm); e != nil {
//			logger.ERROR.Println(mm, e)
//		}
//		if mm.RespId > 0 {
//			respId := fmt.Sprintf("%v", mm.RespId)
//			if m.Callbacks.RemoveCb(respId, func(key string, v interface{}, exists bool) bool {
//				if !exists {
//					return false
//				}
//				if fn, ok := v.(CallBack); ok {
//					fn(mm.Error, mm.Body)
//				} else {
//					log.Println("callback fn error")
//				}
//				return true
//			}) {
//				return
//			} else {
//				logger.ERROR.Printf("unknown respId = %v", mm.RespId)
//			}
//		}
//	}
//
//	logger.INFO.Println(msg)
//	if m.OnPublishCb != nil {
//		go m.OnPublishCb(m, msg)
//	}
//}
