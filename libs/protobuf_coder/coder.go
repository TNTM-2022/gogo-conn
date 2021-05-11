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
	logger.DEBUG.Printf("json2pb; namespace=> %s; messageName=> %s", namespace, messageName, string(jsonStr))
	msg := fd.FindMessage(messageName)
	if msg == nil {
		fmt.Println("no msg", messageName)
		return nil, nil
	}
	dymsg := dynamic.NewMessage(msg)

	var err error
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}

	err = dymsg.UnmarshalJSONPB(&unmarshaler, jsonStr)
	if err != nil {
		fmt.Println("UnmarshalJSONPB error", err)
		return nil, err
	}

	//dymsg.ConvertTo(a)
	//any, err := googleAnyPb.New(a)
	//any, err := anypb.New(*dymsg)
	any, err := dymsg.Marshal()

	//googleProto.Marshal(dymsg.ProtoMessage())
	//any, err := googleAnypb.New(dymsg)
	//any, err := ptypes.MarshalAny(dymsg)
	if err != nil {
		fmt.Println(err, "err")
		return nil, err
	}
	//fmt.Println(any.Value)

	//return any.Value, nil
	return any, nil
}

func PbToJson(messageName string, protoData []byte) ([]byte, error) {
	namespace := getNamesace(messageName)
	_fd, ok := globalReqProtoMap.Load(namespace)
	if !ok {
		return protoData, nil
	}
	fd, ok := _fd.(*desc.FileDescriptor)
	if fd == nil || !ok {
		return protoData, nil
	}

	msg := fd.FindMessage(messageName)
	if msg == nil {
		return protoData, nil
	}
	dymsg := dynamic.NewMessage(msg)

	err := githubProto.Unmarshal(protoData, dymsg)
	if err != nil {
		log.Println(err)
		return protoData, nil
	}
	jsonByte, err := dymsg.MarshalJSON()
	if err != nil {
		log.Println(err)
		return protoData, nil
	}
	return jsonByte, err
}
