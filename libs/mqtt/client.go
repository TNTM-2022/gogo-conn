// https://github.com/gost/server/tree/master/mqtt
package mqtt

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

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

	OnConnectCb func (mqtt *MQTT) //paho.OnConnectHandler
	OnPublishCb paho.MessageHandler
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
	opts.SetOnConnectHandler(client.connectHandler)
	opts.SetDefaultPublishHandler(client.OnPublishCb)
	return opts, nil
}

func (m *MQTT) connectionLostHandler(_ paho.Client, err error) {
	log.Printf("MQTT client lost connection: %v", err)
	m.disconnected = true
	m.retryConnect()
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
	if !m.connectToken.WaitTimeout(time.Second * 5) {
		log.Println("mqtt monitor timeout")
		return
	}
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

//func (m *MQTT) Publish(topic string, message []byte, qos byte, _async bool) {
func (m *MQTT) Publish(topic string, message []byte, qos byte) {
	token := m.client.Publish(topic, qos, false, message)
	//if _async {
	go token.Wait()
	//} else {
	//	token.Wait()
	//}
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
	opts, err := initMQTTClientOps(mqttClient)
	if err != nil {
		log.Fatalf("unable to configure MQTT client: %s", err)
	}

	pahoClient := paho.NewClient(opts)
	mqttClient.client = pahoClient

	return mqttClient
}
