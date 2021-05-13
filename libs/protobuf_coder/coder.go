package proto_coder

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	githubProto "github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"go-connector/logger"
	"go.uber.org/zap"
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
		logger.ERROR.Println("parse proto file when updating", zap.Error(err))
		return
	}
	fd := fds[0] // 会不会有多个文件？
	if len(fds) > 1 {
		//todo 疑似存在多个文件
		fmt.Println("proto fds len:", len(fds))
	}

	namespace := fd.GetPackage()

	if strings.HasSuffix(path, "req.proto") {
		globalReqProtoMap.Store(namespace, fd)

	} else if strings.HasSuffix(path, "res.proto") {
		globalRespProtoMap.Store(namespace, fd)

	} else if strings.HasSuffix(path, "push.proto") {
		globalPushProtoMap.Store(namespace, fd)

	} else {
		logger.ERROR.Println("wrong proto file name, should ends with .res.proto/.req.proto/.push.proto", zap.String("current name", path))
		return
	}

	logger.INFO.Println("proto file init done", zap.String("nameSpace", namespace))
}

func getNamesace(p string) string {
	s := strings.Split(p, ".")
	ss := strings.Join(s[0:len(s)-1], ".")
	logger.DEBUG.Println("protobuf,coder,getNamespace", "get namespace", zap.String("namespace", ss))
	return ss
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
		logger.DEBUG.Println("protobuf,coder,push,resp", "protobuf no fd", zap.String("namespace", namespace), zap.Bool("isPush", isPush))
		return nil, nil
	}
	logger.DEBUG.Println("protobuf,coder,push,resp", "json2pb", zap.String("namespace", namespace), zap.String("route", messageName), zap.String("json str", string(jsonStr)))
	msg := fd.FindMessage(messageName)
	if msg == nil {
		logger.ERROR.Println("cannot find namespace in pb", zap.String("namespace", messageName))
		return nil, nil
	}
	dymsg := dynamic.NewMessage(msg)

	var err error
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}

	err = dymsg.UnmarshalJSONPB(&unmarshaler, jsonStr)
	if err != nil {
		logger.ERROR.Println("UnmarshalJSONPB failed", zap.Error(err))
		return nil, err
	}

	any, err := dymsg.Marshal()
	if err != nil {
		logger.ERROR.Println("dymsg marshal failed", zap.Error(err))
		return nil, err
	}
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
		logger.ERROR.Println("protobuf to dynamic msg unmarshal failed", zap.Error(err))
		return protoData, nil
	}
	jsonByte, err := dymsg.MarshalJSON()
	if err != nil {
		logger.ERROR.Println("dynamic msg marshal to json failed", zap.Error(err))
		return protoData, nil
	}
	return jsonByte, err
}
