package proto_coder

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	githubProto "github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"go-connector/logger"
	"log"
	"strings"
	"sync"
)

var globalReqProtoMap = sync.Map{}
var globalRespProtoMap = sync.Map{}
var globalPushProtoMap = sync.Map{}

func UpdateProto(path string) {
	p := protoparse.Parser{}
	fds, err := p.ParseFiles(path)
	if err != nil {
		fmt.Println("err32", err)
		return
	}
	fd := fds[0]

	namespace := fd.GetPackage()

	if strings.HasSuffix(path, "req.proto") {
		globalReqProtoMap.Store(namespace, fd)

	} else if strings.HasSuffix(path, "res.proto") {
		globalRespProtoMap.Store(namespace, fd)

	} else if strings.HasSuffix(path, "push.proto") {
		globalPushProtoMap.Store(namespace, fd)

	} else {
		fmt.Println("wrong proto file")
		return
	}

	log.Printf("proto 文件初始化 namespace=> %s; isNil=>%s", namespace, fd.GetName())
}

func getNamesace(p string) string {
	s := strings.Split(p, ".")
	//log.Println("getNamespace", s[0:len(s)-1])
	return strings.Join(s[0:len(s)-1], ".")
}

func JsonToPb(messageName string, jsonStr []byte, isPush bool) ([]byte, error) {
	namespace := getNamesace(messageName)
	var fd *desc.FileDescriptor
	if isPush {

		_fd, ok := globalPushProtoMap.Load(namespace)
		if !ok {
			return nil, nil
		}
		fd, ok = _fd.(*desc.FileDescriptor)
	} else {
		_fd, ok := globalRespProtoMap.Load(namespace)
		if !ok {
			return nil, nil
		}
		fd, ok = _fd.(*desc.FileDescriptor)
	}

	if fd == nil {
		logger.DEBUG.Println("protobuf no fd", namespace, isPush)
		return nil, nil
	}
	log.Printf("json2pb; namespace=> %s; messageName=> %s", namespace, messageName, string(jsonStr))
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

	//dymsg.ConvertTo(a)
	//any, err := googleAnyPb.New(a)
	//any, err := anypb.New(*dymsg)
	any, _ := dymsg.Marshal()

	//googleProto.Marshal(dymsg.ProtoMessage())
	//any, err := googleAnypb.New(dymsg)
	//any, err := ptypes.MarshalAny(dymsg)
	if err != nil {
		fmt.Println(err, "err")
		return nil, err
	}
	fmt.Println(6)
	//fmt.Println(any.Value)

	//return any.Value, nil
	return any, nil
}

func PbToJson(messageName string, protoData []byte) ([]byte, error) {
	namespace := getNamesace(messageName)
	_fd, ok := globalReqProtoMap.Load(namespace)
	if !ok {
		fmt.Println(">>>>>>>>1 ", namespace)

		return protoData, nil
	}
	fd, ok := _fd.(*desc.FileDescriptor)
	if fd == nil || !ok {
		fmt.Println(">>>>>>>>2 ", namespace)

		return protoData, nil
	}

	msg := fd.FindMessage(messageName)
	if msg == nil {
		fmt.Println(">>>>>>>>3 ", namespace)
		return protoData, nil
	}
	dymsg := dynamic.NewMessage(msg)

	err := githubProto.Unmarshal(protoData, dymsg)

	jsonByte, err := dymsg.MarshalJSON()
	return jsonByte, err
}
