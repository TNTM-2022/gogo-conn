package monitor_watcher

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/global"
	"go-connector/libs/mqtt_client"
	"go-connector/libs/package_coder"
	"strconv"
)

func Request(m *mqtt_client.MQTT, topic, moduleId string, msg []byte, cb interface{}) {
	reqId := m.GetReqId()
	m.Callbacks.Set(fmt.Sprintf("%v", reqId), cb)
	m.Publish(topic, msg, 0, true)
}

func OnPublishHandler(m *mqtt_client.MQTT, client paho.Client, msg paho.Message) {
	pkgId, dpkg := package_coder.DecodeResp(msg.Topic(), msg.MessageID(), msg.Payload())
	m.Callbacks.RemoveCb(fmt.Sprintf("%v", pkgId), func(k string, v interface{}, exists bool) bool {
		// k: pkgId; v 存储的这个包相关信息
		if !exists {
			return false
		}
		if pkgBelong, ok := v.(*PkgBelong); ok {
			fmt.Println(pkgBelong.SID, "receive a letter.", pkgBelong.Route, string(dpkg.Payload))
			dpkg.Sid = pkgBelong.SID
			dpkg.Route = pkgBelong.Route
			dpkg.PkgID = pkgBelong.ClientPkgID
			if v, ok := global.SidFrontChanStore.Get(strconv.FormatUint(uint64(pkgBelong.SID), 10)); ok {
				if ch, ok := v.(chan package_coder.BackendMsg); ok {
					ch <- *dpkg // todo 使用指针会好一些？
				}
			}
			// todo 分发至个人
		}
		return true
	})
}
