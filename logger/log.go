package logger

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

var logger *zap.Logger

var debugFlags = make(map[string]bool)

func init() {
	var err error
	cfg := zap.NewProductionConfig()
	logger, err = cfg.Build(zap.IncreaseLevel(zap.DebugLevel), zap.AddCallerSkip(1))
	if err != nil {
		log.Panicln(err)
	}

	s := os.Getenv("DEBUG")
	ss := strings.SplitN(s, ",", -1)
	fmt.Println(ss)
	for _, s := range ss {
		if s == "" {
			continue
		}
		cfg.Level.SetLevel(zap.DebugLevel)
		debugFlags[s] = true
	}
}

type Log interface {
	Println(string, ...zap.Field)
}

type LogDebug interface {
	Println(string, string, ...zap.Field)
}

var (
	DEBUG LogDebug = logDebug{}
	ERROR Log      = logError{}
	INFO  Log      = logInfo{}
)

type logInfo struct{}

func (l logInfo) Println(k string, v ...zap.Field) {
	logger.Info(k, v...)
}

type logError struct{}

func (l logError) Println(k string, v ...zap.Field) {
	logger.Error(k, v...)
}

type logDebug struct{}

// Println namespace 所在的包名，文件名，功能模块名
func (l logDebug) Println(namespace string, s string, v ...zap.Field) {
	if len(debugFlags) == 0 {
		return
	}
	nss := strings.SplitN(namespace, ",", -1)
	cc := false
	if debugFlags["*"] {
		cc = true
	}
	for _, ns := range nss {
		if debugFlags[ns] || cc {
			cc = true
			break
		}
	}
	if !cc {
		return
	}
	v = append(v, zap.String("ns", namespace))
	logger.Debug(s, v...)
}
