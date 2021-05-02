package global

import (
	"fmt"
	"sync"
	"testing"
)

const testC = 99

var sid = [testC]uint32{}

func TestGetSid(t *testing.T) {
	blen := CountSid()
	var wg sync.WaitGroup
	wg.Add(testC)
	for i := 0; i < testC; i++ {
		go func(i int) {
			defer wg.Done()
			_sid, ok := GetSid()
			sid[i] = _sid
			if !ok {
				panic("not get")
			}
		}(i)
	}
	wg.Wait()
	fmt.Println(sid, CountSid(), sidsHead)
	if CountSid() != blen-testC {
		panic("get error")
	}
}

func TestBackSid(t *testing.T) {
	blen := CountSid()
	var wg sync.WaitGroup
	wg.Add(len(sid))
	for _, id := range sid {
		go func(id uint32) {
			defer wg.Done()
			BackSid(id)
		}(id)
	}
	wg.Wait()
	fmt.Println(sid, true, CountSid(), sidsHead)
	if CountSid() != blen+testC {
		panic("get error")
	}
	s := sidsHead.start
	for {
		fmt.Println(s.sid)
		s = s.next
		if s == nil {
			break
		}
	}
}

func BenchmarkGetSid(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			sid, ok := GetSid()
			if !ok {
				return
			}
			defer BackSid(sid)
		}()
	}
	wg.Wait()
	//fmt.Println(sidsHead, CountSid())
}

func BenchmarkGetSid2(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	var m uint32 = 0
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			m++

			defer func() { m-- }()
		}()
	}
	wg.Wait()
	//fmt.Println(sidsHead, CountSid())
}
