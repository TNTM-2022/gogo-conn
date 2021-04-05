package mqtt_connect

import (
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	fmt.Println("a ...interface{}0")

	g := Gate{}
	e := g.StartGate("127.0.0.1:4616")
	fmt.Println(e)

}
