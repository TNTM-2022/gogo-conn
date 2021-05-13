package monitor_watcher

import (
	"go-connector/logger"
	"sync"
)

type taskType struct {
	next *taskType
	v    interface{}
	c    chan interface{}
}
type taskHead struct {
	head *taskType
	last *taskType

	sync.Mutex
	C chan bool
}

type TaskLoop interface {
	Push(interface{}) chan interface{}
	Run(func(interface{}) interface{})
}

func CreateTaskLoop() TaskLoop {
	return &taskHead{
		C: make(chan bool, 1),
	}
}

func (t *taskHead) add(tk *taskType) {
	t.Lock()
	defer t.Unlock()

	if t.head == nil {
		t.head = tk
		t.last = tk
	} else {
		t.last.next = tk
		t.last = tk
	}
	return
}
func (t *taskHead) get() (v interface{}, exists bool, res chan interface{}) {
	t.Lock()
	defer t.Unlock()

	vv := t.head
	if vv == nil {
		t.head = nil
		t.last = nil
	} else {
		t.head = vv.next
		res = vv.c
		v = vv.v
		exists = true
	}
	return
}

func (t *taskHead) Run(f func(interface{}) interface{}) {
	for {
		v, ok, ch := t.get()
		if !ok {
			<-t.C
			continue
		}
		ch <- f(v)
		close(ch)
	}
}

func (t *taskHead) Push(v interface{}) (c chan interface{}) {
	c = make(chan interface{}, 1)
	t.add(&taskType{
		v: v,
		c: c,
	})
	select {
	case t.C <- true:
	default:
		logger.DEBUG.Println("looper", "not push into looper")
	}
	return
}
