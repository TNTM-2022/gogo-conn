package test

import (
	"fmt"
	concurrentMap "github.com/orcaman/concurrent-map"
	"sync"
	"testing"
)

var m1 = sync.Map{}

func BenchmarkTestSyncMapPerformanceRW(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N * 2)

	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()

			m1.Store(fmt.Sprintf("%v", i), i)
		}()
	}
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			m1.Load(fmt.Sprintf("%v", i))
		}()
	}

	wg.Wait()
}

var m2 = concurrentMap.New()

func BenchmarkTestConCurrencyMapPerformanceRW(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N * 2)

	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			m2.Set(fmt.Sprintf("%v", i), i)
		}()
	}

	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			m2.Get(fmt.Sprintf("%v", i))
		}()
	}

	wg.Wait()
}

var m3 = sync.Map{}

func init() {
	for i := 0; i < 100; i++ {
		m3.Store(fmt.Sprintf("%v", i), i)
	}
}
func BenchmarkTestSyncMapPerformanceR(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			m3.Load(fmt.Sprintf("%v", i % 2))
		}()
	}

	wg.Wait()
}

var m4 = concurrentMap.New()

func init() {
	for i := 0; i < 100; i++ {
		go func() {
			m4.Get(fmt.Sprintf("%v", i))
		}()
	}

}
func BenchmarkTestConCurrencyMapPerformanceR(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			m4.Get(fmt.Sprintf("%v", i %2))
		}()
	}

	wg.Wait()
}
