package monitor_watcher

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/filters"
	"go-connector/global"
	"go-connector/libs/mqtt_client"
	"go-connector/libs/package_coder"
	"go-connector/libs/pomelo_coder"
	"go-connector/logger"
	"go.uber.org/zap"
	"strconv"
)

func Request(m *mqtt_client.MQTT, topic, _ string, reqId int64, msg []byte, cb interface{}) bool {
	if reqId > 0 {
		m.Callbacks.Set(fmt.Sprintf("%v", reqId), cb)
	}
	if !m.IsConnectionOpen() {
		return false
	}
	m.Publish(topic, msg, 0, true)
	return true
}

func OnPublishHandler(m *mqtt_client.MQTT, _ paho.Client, msg paho.Message) {
	var rec package_coder.RawRecv
	if e := json.Unmarshal(msg.Payload(), &rec); e != nil {
		logger.ERROR.Println("json.unmarshal push payload failed", zap.Error(e))
		return
	}

	logger.DEBUG.Println("res,ws", "decodeResp", zap.String("payload", string(msg.Payload())))

	pkgId, dpkg, err := decodeResp(msg.Topic(), msg.MessageID(), &rec)
	if v, exists := m.Callbacks.Pop(fmt.Sprintf("%v", pkgId)); exists {
		if pkgBelong, ok := v.(*PkgBelong); ok {
			if err != "null" {
				logger.ERROR.Println("wrong router requested", zap.String("route", pkgBelong.Route), zap.String("error", err))
				dpkg = filters.NoRouteFilter(pkgBelong.Route, pkgBelong.ClientPkgID, pkgBelong.CompressRoute, pkgBelong.CompressGzip)
			}

			dpkg.Sid = pkgBelong.SID
			dpkg.Route = pkgBelong.Route
			dpkg.PkgID = pkgBelong.ClientPkgID
			dpkg.MType = pomelo_coder.Message["TYPE_RESPONSE"]
			dpkg.CompressGzip = pkgBelong.CompressGzip
			dpkg.CompressRoute = pkgBelong.CompressRoute

			if v, ok := global.SidFrontChanStore.Get(strconv.FormatUint(uint64(pkgBelong.SID), 10)); ok {
				if ch, ok := v.(chan package_coder.BackendMsg); ok {
					fmt.Println("mqtt_write")
					ch <- dpkg
				}
			} else {
				logger.ERROR.Println("cannot find user channel by sessionId", zap.Uint32("sessionId", pkgBelong.SID), zap.String("payload", string(msg.Payload())))
			}
		} else {
			logger.ERROR.Println("cannot find pkgBelong by pkgId")
		}
		return
	}
}

func decodeResp(_ string, _ uint16, rec *package_coder.RawRecv) (pkgId uint64, u BackendMsg, err string) {
	u = BackendMsg{}
	if rec.Resp != nil {
		pkgId = rec.Id
		err = string(rec.Resp[0])
		if len(rec.Resp) >= 2 && rec.Resp[1] != nil {
			u.Payload = rec.Resp[1]
		}
	}
	return
}
