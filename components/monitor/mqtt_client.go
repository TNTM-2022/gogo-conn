package monitor

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/libs/mqtt_client"
	"go-connector/logger"
	"log"
)

func Request(m *mqtt_client.MQTT, topic, moduleId string, msg []byte, cb interface{}) {
	reqId := m.GetReqId()
	rr := ComposeRequest(reqId, moduleId, msg)
	m.Publish(topic, rr, 0, true)
	m.Callbacks.Set(fmt.Sprintf("%v", reqId), cb)
}

func Notify(m *mqtt_client.MQTT, topic, moduleId string, msg []byte) {
	rr := ComposeRequest(0, moduleId, msg)
	m.Publish(topic, rr, 0, true)
}

func Response(m *mqtt_client.MQTT, topic string, reqId int64, err, data []byte) {
	rr := ComposeResponse(reqId, err, data)

	m.Publish(topic, rr, 0, true)
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

func OnPublishHandler(m *mqtt_client.MQTT, client paho.Client, msg paho.Message) bool {
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
			logger.ERROR.Println(mm, e)
		}
		if mm.RespId > 0 {
			respId := fmt.Sprintf("%v", mm.RespId)
			if m.Callbacks.RemoveCb(respId, func(key string, v interface{}, exists bool) bool {
				if !exists {
					return false
				}
				if fn, ok := v.(mqtt_client.CallBack); ok {
					fn(mm.Error, mm.Body)
				} else {
					log.Println("callback fn error")
				}
				return true
			}) {
				return true
			} else {
				logger.ERROR.Printf("unknown respId = %v", mm.RespId)
			}
		}
	}
	return false
}
