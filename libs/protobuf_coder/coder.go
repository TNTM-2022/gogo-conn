package proto_coder

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"log"
	"strings"
	"sync"
)

// todo 使用 sync.map 进行改造
var globalReqProtoMap map[string]*desc.FileDescriptor
var globalRespProtoMap map[string]*desc.FileDescriptor
var globalPushProtoMap map[string]*desc.FileDescriptor
var lk sync.RWMutex

func init() {
	globalReqProtoMap = make(map[string]*desc.FileDescriptor, 200)
	globalRespProtoMap = make(map[string]*desc.FileDescriptor, 200)
	globalPushProtoMap = make(map[string]*desc.FileDescriptor, 200)
}

func UpdateProto(path string) {
	p := protoparse.Parser{}
	fds, err := p.ParseFiles(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fd := fds[0]

	lk.Lock()
	defer lk.Unlock()

	namespace := fd.GetPackage()

	var m map[string]*desc.FileDescriptor
	if strings.HasSuffix(path, "req.proto") {
		m = globalReqProtoMap

	} else if strings.HasSuffix(path, "res.proto") {
		m = globalRespProtoMap

	} else if strings.HasSuffix(path, "push.proto") {
		m = globalPushProtoMap

	} else {
		fmt.Println("wrong proto file")
		return
	}

	m[namespace] = fd
	log.Printf("proto 文件初始化 namespace=> %s; isNil=>%s", namespace, fd.GetName())
}

func getNamesace(p string) string {
	s := strings.Split(p, ".")
	//if len(s) !=3 {
	//	return ""
	//}
	fmt.Println(s[0 : len(s)-1])
	return strings.Join(s[0:len(s)-1], ".")
}

func JsonToPb(messageName string, jsonStr []byte, isPush bool) ([]byte, error) {
	lk.RLock()
	defer lk.RUnlock()

	namespace := getNamesace(messageName)
	var fd *desc.FileDescriptor
	if isPush {
		fd = globalPushProtoMap[namespace]
	} else {
		fd = globalRespProtoMap[namespace]
	}

	if fd == nil {
		fmt.Println("no fd", namespace)
		return nil, nil
	}
	log.Printf("json2pb; namespace=> %s; messageName=> %s", namespace, messageName)
	fmt.Println(" jsonStr>> ", string(jsonStr))
	msg := fd.FindMessage(messageName)
	if msg == nil {
		fmt.Println("no msg", messageName)
		return nil, nil
	}
	fmt.Println(1, msg)
	dymsg := dynamic.NewMessage(msg)

	var err error
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	fmt.Println(3)

	err = dymsg.UnmarshalJSONPB(&unmarshaler, jsonStr)
	fmt.Println(4, err)
	if err != nil {
		fmt.Println("UnmarshalJSONPB error", err)

		return nil, err
	}

	fmt.Println(5, dymsg)

	any, err := ptypes.MarshalAny(dymsg)
	if err != nil {
		fmt.Println(err, "err")
		return nil, err
	}
	fmt.Println(6)
	fmt.Println(any.Value)

	return any.Value, nil
}

func PbToJson(messageName string, protoData []byte) ([]byte, error) {
	lk.RLock()
	defer lk.RUnlock()

	namespace := getNamesace(messageName)
	fd := globalReqProtoMap[namespace]
	fmt.Println(namespace, "--==1", messageName)
	if fd == nil {
		return nil, nil
	}

	msg := fd.FindMessage(messageName)
	dymsg := dynamic.NewMessage(msg)

	err := proto.Unmarshal(protoData, dymsg)

	jsonByte, err := dymsg.MarshalJSON()
	fmt.Println(namespace, "--==1", messageName, dymsg)
	return jsonByte, err
}
