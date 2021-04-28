package logger

type Log interface {
	Println(...interface{})
	Printf(...interface{})
}

type log struct{}

func (l log) Println(...interface{})        {}
func (l log) Printf(string, ...interface{}) {}

var (
	DEBUG = log{}
	ERROR = log{}
	INFO  = log{}
)
