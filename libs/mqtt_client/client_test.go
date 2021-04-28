package mqtt_client

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	mqttClient := CreateMQTTClient(&MQTT{
		Host:            "127.0.0.1",
		Port:            "6050",
		ClientID:        "MQTT_RPC_2617886369579",
		SubscriptionQos: 1,
		Persistent:      true,
		Order:           false,
		KeepAliveSec:    5,
		PingTimeoutSec:  10,
	})
	mqttClient.Start()
	time.Sleep(time.Second)
	fmt.Println(mqttClient.IsConnected())
}
