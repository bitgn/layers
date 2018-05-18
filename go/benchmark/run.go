package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bitgn/layers/go/benchmark/bench"
)

var (
	pendingRequests int32
)

func runBenchmark(ms chan metrics, hz int, l bench.Launcher) {

	period := time.Duration(float64(time.Second) / float64(hz))

	fmt.Println("Period is", period)

	xor := NewXorShift()

	for range time.Tick(period) {
		begin := time.Now()
		x := xor.Next()
		atomic.AddInt32(&pendingRequests, 1)
		go func() {
			err := l.Exec(x)
			total := time.Since(begin)

			result := metrics{
				error:       err != nil,
				nanoseconds: total.Nanoseconds(),
			}
			ms <- result
			atomic.AddInt32(&pendingRequests, -1)
		}()

	}
}

type metrics struct {
	nanoseconds int64
	error       bool
}
