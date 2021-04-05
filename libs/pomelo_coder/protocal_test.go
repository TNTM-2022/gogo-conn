package coder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	mb := MessageEncode(10, 2, 0, "test.testHandler.test", []byte("this is a test"), false)
	fmt.Println(base64.StdEncoding.EncodeToString(mb))

	pb := PackageEncode(1, mb)
	fmt.Println(base64.StdEncoding.EncodeToString(pb))

	dpb := PackageDecode(pb)
	dmb := MessageDecode(dpb[0].Body)
	j, _ := json.Marshal(dmb)
	fmt.Println(string(j))

}
