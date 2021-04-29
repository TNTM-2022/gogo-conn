package mointor_watcher

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/libs/mqtt_client"
	"go-connector/libs/package_coder"
)

func Request(m *mqtt_client.MQTT, topic, moduleId string, msg []byte, cb interface{}) {
	reqId := m.GetReqId()
	m.Callbacks.Set(fmt.Sprintf("%v", reqId), cb)
	m.Publish(topic, msg, 0, true)
}

func OnPublishHandler(m *mqtt_client.MQTT, client paho.Client, msg paho.Message) {
	pkgId, dpkg := package_coder.DecodeResp(msg.Topic(), msg.MessageID(), msg.Payload())
	m.Callbacks.RemoveCb(fmt.Sprintf("%v", pkgId), func(k string, v interface{}, exists bool) bool {
		if !exists {
			return false
		}
		if vv, ok := v.(*PkgBelong); ok {
			fmt.Println(vv.SID, "receive a letter.", vv.Route, string(dpkg.Payload))
			// todo 分发至个人
		}
		return true
	})
}
