package logger

import (
	"go.uber.org/zap"
	"testing"
)

func TestLogInfo_Println(t *testing.T) {
	INFO.Println("info---info", zap.String("url", "http://ww.baidu.com"))
}

func TestLogError_Println(t *testing.T) {
	ERROR.Println("error---error")
}

func TestLogDebug_Println(t *testing.T) {
	DEBUG.Println("test", "info", zap.String("url", "http://ww.google.com"))
	DEBUG.Println("test11,test", "info", zap.String("url", "http://ww.google.com"))
}
