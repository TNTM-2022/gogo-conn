package logger

import "fmt"

type Log interface {
	Println(...interface{})
	Printf(...interface{})
}

type log struct{}

func (l log) Println(v ...interface{}) {
	fmt.Println(v...)
}
func (l log) Printf(s string, v ...interface{}) {
	fmt.Printf(s, v...)
}

var (
	DEBUG = log{}
	ERROR = log{}
	INFO  = log{}
)
