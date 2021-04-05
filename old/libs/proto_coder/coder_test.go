package proto_coder

import (
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	UpdateProto("/Users/mac/go/src/github.com/jhump/protoreflect/internal/testprotos/desc_test_proto3.proto");

	js := []byte(`{"foo":["VALUE1"],"bar":"bedazzle", "sddss": 34}`)

	ll1, err1 := JsonToPb("testprotos.TestRequest", js, true)
	fmt.Println("-dd1-", ll1, err1)



	UpdateProto("/Users/mac/Codes/golang/go-connector/test_proto/user.proto");

	js1 := []byte(`{"foo":["VALUE1"],"bar":"bedazzle", "code": 34}`)
	ll, err := JsonToPb("user.Entry", js1)
	fmt.Println("-dd-", ll, err)

}
