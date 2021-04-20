// https://github.com/gost/server/tree/master/mqtt
package mqtt

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
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

	OnConnectCb func(mqtt *MQTT)          //paho.OnConnectHandler
	OnPublishCb func(*MQTT, paho.Message) // paho.MessageHandler

	reqId       int64
	reqIdLocker sync.Mutex
	callbacks   cmap.ConcurrentMap

	connectedNum int
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
	//opts.SetConnectionLostHandler(client.connectionLostHandler)
	//opts.SetDefaultPublishHandler(client.OnPublishCb)
	opts.SetDefaultPublishHandler(client.publishHandler)
	opts.SetOnConnectHandler(client.connectHandler)
	opts.SetReconnectingHandler(client.reconnectHandler)
	return opts, nil
}

func (m *MQTT) reconnectHandler(_ paho.Client, opts *paho.ClientOptions) {
	fmt.Printf("reconnecting to %v server ....\n", opts.ClientID)
}
func (m *MQTT) connectionLostHandler(_ paho.Client, err error) {
	log.Printf("MQTT client lost connection: %v", err)
	m.disconnected = true
	//m.retryConnect()
}

func (m *MQTT) publishHandler(client paho.Client, msg paho.Message) {
	if msg.Topic() == "monitor" {
		var mm struct {
			RespId int64           `json:"respId"`
			Body   json.RawMessage `json:"body"`
			Error  string          `json:"error"`
		}
		var ss string
		d := msg.Payload()
		if e := json.Unmarshal(d, &ss); e == nil {
			d = []byte(ss)
		}
		if e := json.Unmarshal(d, &mm); e != nil {
			fmt.Println(mm, e)
		}
		respId := fmt.Sprintf("%v", mm.RespId)
		if mm.RespId > 0 {
			if f, ok := m.callbacks.Get(respId); ok {
				m.callbacks.Remove(respId)
				if fn, ok := f.(CallBack); ok {
					fn(mm.Error, mm.Body)
					return
				}

			} else {
				log.Println("unknown respId =", respId)
			}
		}

	}
	fmt.Println(msg)
	if m.OnPublishCb != nil {
		go m.OnPublishCb(m, msg)
	}
}

func (m *MQTT) IsReconnect() bool {
	return m.connectedNum > 1
}
func (m *MQTT) IsConnected() bool {
	return m.client.IsConnected()
}
func (m *MQTT) connectHandler(_ paho.Client) {
	log.Printf("MQTT client connected")
	//hasSession := m.connectToken.SessionPresent()
	//log.Printf("MQTT Session present: %v", hasSession)
	m.connectedNum++
	// on first connect, connection lost and persistance is off or no previous session found
	//if !m.disconnected || (m.disconnected && !m.persistent) || !hasSession {
	//	m.subscribe()
	//}
	if m.OnConnectCb != nil {
		go m.OnConnectCb(m)
	}
	m.disconnected = false
}

//func (m *MQTT) subscribe() {
//	//a := *m.api
//	//topics := *a.GetTopics(m.prefix)
//	//
//	//for _, t := range topics {
//	//	topic := t
//	//	log.Printf("MQTT client subscribing to %s", topic.Path)
//	//
//	//	if token := m.client.Subscribe(topic.Path, m.subscriptionQos, func(client paho.Client, msg paho.Message) {
//	//		go topic.Handler(m.api, m.prefix, msg.Topic(), msg.Payload())
//	//	}); token.Wait() && token.Error() != nil {
//	//		log.Fatal(token.Error())
//	//	}
//	//}
//}

func (m *MQTT) Request(topic, moduleId string, msg []byte, cb CallBack) {
	reqId := m.GetReqId()
	m.callbacks.Set(fmt.Sprintf("%v", reqId), cb)
	rr := ComposeRequest(reqId, moduleId, msg)
	m.Publish(topic, rr, 0, true)
}

func (m *MQTT) Notify(topic, moduleId string, msg []byte) {
	rr := ComposeRequest(0, moduleId, msg)
	m.Publish(topic, rr, 0, true)
}

func (m *MQTT) Response(topic string, reqId int64, err, data []byte) {
	rr := ComposeResponse(reqId, err, data)

	m.Publish(topic, rr, 0, true)
}

func (m *MQTT) Publish(topic string, message []byte, qos byte, _async bool) {
	token := m.client.Publish(topic, qos, false, message)
	if token.Error() != nil {
		fmt.Println("publish error ", token.Error(), m.client.IsConnected())
	}
	if !_async {
		token.Wait()
	}
}

func (m *MQTT) Stop() {
	m.client.Disconnect(500)
}

func (m *MQTT) Start() {
	log.Printf("Starting MQTT client on tcp://%s:%v with Prefix:%v, Persistence:%v, OrderMatters:%v, KeepAlive:%v, PingTimeout:%v, QOS:%v", m.Host, m.Port, "", true, m.Order, m.KeepAliveSec, m.PingTimeoutSec, 1)
	t := m.client.Connect()
	t.Wait() //Timeout(time.Second * 2) // pinus 问题， 没有 connack确认。 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 44L。 client.connack({ returnCode: 0 });
	fmt.Println("error ", t.Error())
	if t.Error() != nil {
		log.Println("mqtt monitor timeout", m.Host, m.Port)
		log.Panicf("cannot connect to %v server", m.ClientID)
	}
	fmt.Println("mqtt started??")
	//m.connectToken = m.client.Connect().(*paho.ConnectToken)
	//if m.connectToken.Error() != nil {
	//	log.Println("mqtt monitor timeout", m.Host, m.Port)
	//	log.Panic("cannot connect to master server")
	//}
}

type L interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// CreateMQTTClient creates a new MQTT client
func CreateMQTTClient(mqttClient *MQTT) *MQTT {
	mqttClient.callbacks = cmap.New()

	opts, err := initMQTTClientOps(mqttClient)
	if err != nil {
		log.Fatalf("unable to configure MQTT client: %s", err)
	}

	pahoClient := paho.NewClient(opts)
	mqttClient.client = pahoClient

	return mqttClient
}

func ComposeRequest(reqId int64, moduleId string, body []byte) (rr []byte) {
	if reqId > 0 { // request
		rr, _ = json.Marshal(struct {
			ReqId    int64           `json:"reqId"`
			ModuleId string          `json:"moduleId"`
			Body     json.RawMessage `json:"body"`
		}{
			ReqId:    reqId,
			ModuleId: moduleId,
			Body:     body,
		})
	} else { // notify
		rr, _ = json.Marshal(struct {
			ModuleId string          `json:"moduleId"`
			Body     json.RawMessage `json:"body"` // pomelo 自身错误导致的， error first 忘记第一个是error了
		}{
			ModuleId: moduleId,
			Body:     body,
		})
	}
	return
}
func ComposeResponse(reqId int64, err, res []byte) (rr []byte) { // req: {reqId: number}, err: string | Error, res: any
	rr, _ = json.Marshal(struct {
		RespID int64           `json:"respId"`
		Error  json.RawMessage `json:"error"` // pomelo 自身错误导致的， error first 忘记第一个是error了
		Body   json.RawMessage `json:"body"`  // pomelo 自身错误导致的， error first 忘记第一个是error了
	}{
		RespID: reqId,
		Error:  err,
		Body:   res,
	})
	return
}
