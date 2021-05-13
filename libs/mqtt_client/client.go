package mqtt_client

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
	"go-connector/logger"
	"go.uber.org/zap"
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
	clientOpts      *paho.ClientOptions
	SubscriptionQos byte
	Persistent      bool

	OnConnectCb func(mqtt *MQTT) //paho.OnConnectHandler
	//OnPublishCb func(client paho.Client, msg paho.Message) // paho.MessageHandler
	onReconnectCb func(mqtt *MQTT)

	reqId       int64
	reqIdLocker sync.Mutex
	Callbacks   cmap.ConcurrentMap

	connectedNum int
	//Closing      bool
	ServerType string
	Quit       chan bool

	t int // start time
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
	opts.SetMaxReconnectInterval(time.Second * 3)

	//opts.SetDefaultPublishHandler(client.publishHandler)
	opts.SetOnConnectHandler(client.connectHandler)
	opts.SetReconnectingHandler(client.reconnectHandler)

	return opts, nil
}

// CreateMQTTClient creates a new MQTT client
func CreateMQTTClient(mqttClient *MQTT) *MQTT {
	mqttClient.Callbacks = cmap.New()
	mqttClient.Quit = make(chan bool, 0)
	mqttClient.t = time.Now().Second()
	opts, err := initMQTTClientOps(mqttClient)
	if err != nil {
		logger.ERROR.Println("unable to configure MQTT client", zap.Error(err), zap.String("host", mqttClient.Host), zap.String("port", mqttClient.Port), zap.String("serverId", mqttClient.ClientID))
	}
	mqttClient.clientOpts = opts

	return mqttClient
}

func (m *MQTT) SetCallbacks(OnConnectCb func(mqtt *MQTT), OnPublishCb func(client paho.Client, msg paho.Message)) {
	m.clientOpts.SetDefaultPublishHandler(OnPublishCb)
	if OnConnectCb != nil {
		m.OnConnectCb = OnConnectCb
	}
}
func (m *MQTT) SetReconnectCb(f func(mqtt *MQTT)) {
	if m.onReconnectCb == nil {
		m.onReconnectCb = f
	}
}
func (m *MQTT) Stop(t time.Duration) {
	defer logger.DEBUG.Println("mqtt", "disconnect", zap.String("serverId", m.ClientID))
	close(m.Quit)
	time.Sleep(t)
	m.client.Disconnect(1)
}
func (m *MQTT) Start() {
	pahoClient := paho.NewClient(m.clientOpts)
	m.client = pahoClient

	log.Printf("Starting MQTT client on tcp://%s:%v with Prefix:%v, Persistence:%v, OrderMatters:%v, KeepAlive:%v, PingTimeout:%v, QOS:%v", m.Host, m.Port, "", true, m.Order, m.KeepAliveSec, m.PingTimeoutSec, 1)
	t := m.client.Connect()
	t.Wait() //Timeout(time.Second * 2) // pinus 问题， 没有 connack确认。 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 44L。 client.connack({ returnCode: 0 });
	if t.Error() != nil {
		logger.ERROR.Println("cannot connect to %v server", zap.String("host", m.Host), zap.String("port", m.Port), zap.String("serverId", m.ClientID))
		m.Stop(1)
	}
	logger.DEBUG.Println("mqtt_client,monitor_client,servers", "mqtt client connected!!")
}

func (m *MQTT) reconnectHandler(_ paho.Client, opts *paho.ClientOptions) {
	logger.INFO.Println("reconnecting to server", zap.String("serverId", opts.ClientID), zap.Int("second", time.Now().Second()))
	if m.onReconnectCb != nil {
		m.onReconnectCb(m)
	}
}
func (m *MQTT) connectionLostHandler(_ paho.Client, err error) {
	logger.ERROR.Println("MQTT client lost connection", zap.Error(err))

	m.disconnected = true
}
func (m *MQTT) IsReconnect() bool {
	return m.connectedNum > 1
}
func (m *MQTT) IsConnectionOpen() bool {
	return m.client != nil && m.client.IsConnectionOpen()
}
func (m *MQTT) IsConnected() bool {
	return m.client != nil && m.client.IsConnected()
}
func (m *MQTT) connectHandler(_ paho.Client) {
	logger.DEBUG.Println("mqtt,mqtt_client", "MQTT client connected.", zap.String("serverId", m.ClientID), zap.Int("start tick", m.t))
	m.connectedNum++
	if m.OnConnectCb != nil {
		m.OnConnectCb(m)
	}
	m.disconnected = false
}
func (m *MQTT) Publish(topic string, message []byte, qos byte, _async bool) {
	token := m.client.Publish(topic, qos, false, message)
	if token.Error() != nil {
		logger.ERROR.Println("publish error", zap.Error(token.Error()), zap.Bool("mqtt_client is connected", m.client.IsConnectionOpen()))
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
