package monitor_watcher

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/global"
	"go-connector/libs/mqtt_client"
	"go-connector/libs/package_coder"
	"go-connector/libs/pomelo_coder"
	"go-connector/logger"
	"strconv"
)

func Request(m *mqtt_client.MQTT, topic, _ string, reqId int64, msg []byte, cb interface{}) bool {
	m.Callbacks.Set(fmt.Sprintf("%v", reqId), cb)
	if !m.IsConnected() {
		return false
	}
	m.Publish(topic, msg, 0, true)
	return true
}

func OnPublishHandler(m *mqtt_client.MQTT, _ paho.Client, msg paho.Message) {
	pkgId, dpkg := package_coder.DecodeResp(msg.Topic(), msg.MessageID(), msg.Payload())
	// todo 使用 pop 代替
	m.Callbacks.RemoveCb(fmt.Sprintf("%v", pkgId), func(k string, v interface{}, exists bool) bool {
		// k: pkgId; v 存储的这个包相关信息
		if !exists {
			fmt.Println("no exists", fmt.Sprintf("%v", pkgId), k)
			return false
		}
		if pkgBelong, ok := v.(*PkgBelong); ok {
			dpkg.Sid = pkgBelong.SID
			dpkg.Route = pkgBelong.Route
			dpkg.PkgID = pkgBelong.ClientPkgID
			dpkg.MType = pomelo_coder.Message["TYPE_RESPONSE"]
			//dpkg.CompressRoute = pkgBelong.
			if v, ok := global.SidFrontChanStore.Get(strconv.FormatUint(uint64(pkgBelong.SID), 10)); ok {
				if ch, ok := v.(chan package_coder.BackendMsg); ok {
					ch <- *dpkg // todo 使用指针会好一些？
				}
			} else {
				fmt.Println("no sid ch")
			}
			logger.DEBUG.Printf("resp router:>> %v; jsonStr>> %v", dpkg.Route, string(dpkg.Payload))
		} else {
			fmt.Println("no pkg belong")
		}
		return true
	})
}
