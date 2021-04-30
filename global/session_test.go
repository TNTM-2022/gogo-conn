package global

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSession_MarshalJSONest(t *testing.T) {
	s := sessionType{
		Uid: 123,
		Sid: 123,
	}
	buf, err := json.Marshal(s)
	fmt.Println(string(buf), err)
}
