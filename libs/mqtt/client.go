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

	opts.SetClientID(client.ClientID)
	opts.SetCleanSession(!client.Persistent)
	opts.SetOrderMatters(client.Order)
	opts.SetKeepAlive(time.Duration(client.KeepAliveSec) * time.Second)
	opts.SetPingTimeout(time.Duration(client.PingTimeoutSec) * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(client.connectionLostHandler)
	//opts.SetDefaultPublishHandler(client.OnPublishCb)
	opts.SetDefaultPublishHandler(client.publishHandler)
	opts.SetOnConnectHandler(client.connectHandler)
	return opts, nil
}

func (m *MQTT) connectionLostHandler(_ paho.Client, err error) {
	log.Printf("MQTT client lost connection: %v", err)
	m.disconnected = true
	m.retryConnect()
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

			}
		}

	}
	m.OnPublishCb(m, msg)
}
func (m *MQTT) connectHandler(c paho.Client) {
	log.Printf("MQTT client connected")
	hasSession := m.connectToken.SessionPresent()
	log.Printf("MQTT Session present: %v", hasSession)

	// on first connect, connection lost and persistance is off or no previous session found
	//if !m.disconnected || (m.disconnected && !m.persistent) || !hasSession {
	//	m.subscribe()
	//}
	m.OnConnectCb(m)
	m.disconnected = false
}

func (m *MQTT) connect() {
	m.connectToken = m.client.Connect().(*paho.ConnectToken)
	// todo 看看后期咋么修改
	// todo pomelo bug  如果不修改connect事件返回， 这里将会一直堵着 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 中的这个  socket.on('connect', function (pkg) { 代码块内部
	//if !m.connectToken.WaitTimeout(time.Second * 5) {
	//	log.Println("mqtt monitor timeout", m.Host, m.Port)
	//	return
	//}
	if m.connectToken.Error() != nil {
		if !m.connecting {
			log.Fatalf("MQTT client %s", m.connectToken.Error())
			m.retryConnect()
		}
	}
}

func (m *MQTT) retryConnect() {
	log.Printf("MQTT client starting reconnect procedure in background")
	m.connecting = true
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			m.connect()
			if m.client.IsConnected() {
				ticker.Stop()
				m.connecting = false
			}
		}
	}()
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

/**
  request(moduleId: string, msg: any, cb: Callback) {
      if (this.state !== ST_REGISTERED) {
          logger.error('agent can not request now, state:' + this.state);
          return;
      }
      let reqId = this.reqId++;
      this.callbacks[reqId] = cb;
      this.doSend('monitor', protocol.composeRequest(reqId, moduleId, msg));
      // this.socket.emit('monitor', protocol.composeRequest(reqId, moduleId, msg));
  }

*/
/**
  if (id) {
      // request message
      return JSON.stringify({
          reqId: id,
          moduleId: moduleId,
          body: body
      });
  } else {
      // notify message
      return {
          moduleId: moduleId,
          body: body
      };
  }
*/
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
	//func (m *MQTT) Publish(topic string, message []byte, qos byte) {
	token := m.client.Publish(topic, qos, false, message)
	if _async {
		//go token.Wait()
	} else {
		token.Wait()
	}
}

func (m *MQTT) Stop() {
	m.client.Disconnect(500)
}

func (m *MQTT) Start() {
	log.Printf("Starting MQTT client on tcp://%s:%v with Prefix:%v, Persistence:%v, OrderMatters:%v, KeepAlive:%v, PingTimeout:%v, QOS:%v", m.Host, m.Port, "", true, m.Order, m.KeepAliveSec, m.PingTimeoutSec, 1)
	m.connect()
}

// CreateMQTTClient creates a new MQTT client
func CreateMQTTClient(mqttClient *MQTT) *MQTT {
	//mqttClient := &MQTT{
	//	host:            "127.0.0.1",
	//	port:            "3005",
	//	clientID:        "clientId-1",
	//	subscriptionQos: 1,
	//	persistent:      true,
	//	order:           false,
	//	keepAliveSec:    5,
	//	pingTimeoutSec:  10,
	//}
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
