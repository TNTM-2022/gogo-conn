package mointor_watcher

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/libs/mqtt_client"
)

type MqttClient struct {
	*mqtt_client.MQTT
}

func (m *MqttClient) Request(topic, moduleId string, msg []byte, cb mqtt_client.CallBack) {
	reqId := m.GetReqId()
	m.Callbacks.Set(fmt.Sprintf("%v", reqId), cb)
	m.Publish(topic, msg, 0, true)
}

func (m *MqttClient) OnPublishHandler(client paho.Client, msg paho.Message) {

}

