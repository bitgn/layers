package main

import (
	"sync/atomic"
	"time"

	"github.com/bitgn/layers/go/benchmark/bench"
)

var (
	pendingRequests int64
	waitingRequests int64
)

func runBenchmark(ms chan metrics, hz int, l bench.Launcher) {
	if hz > 0 {
		runFixedThroughput(ms, hz, l)
	} else {
		runAdaptiveThroughput(ms, -hz, l)
	}
}

func runAdaptiveThroughput(ms chan metrics, concurrency int, l bench.Launcher) {
	for i := 0; i < concurrency; i++ {
		xor := NewXorShift()

		go func() {
			for {
				begin := time.Now()
				x := xor.Next()

				atomic.AddInt64(&pendingRequests, 1)

				err := l.Exec(x)
				total := time.Since(begin)

				result := metrics{
					error:       err != nil,
					nanoseconds: total.Nanoseconds(),
				}
				ms <- result
				atomic.AddInt64(&pendingRequests, -1)
			}
		}()
	}
}

func runFixedThroughput(ms chan metrics, hz int, l bench.Launcher) {

	var sent int
	xor := NewXorShift()
	started := time.Now()

	period := time.Duration(float64(time.Second) / float64(hz))

	for range time.Tick(period) {
		begin := time.Now()
		x := xor.Next()

		// ticker might be slow or lagging,
		// so we want to track how many requests we should've sent by now
		elapsed := begin.Sub(started)
		planned := int(elapsed.Seconds() * float64(hz))
		missing := planned - sent

		waitingRequests = int64(missing)

		// don't trust the ticker to catch up
		// just sent all missing requests
		for i := 0; i < missing; i++ {

			atomic.AddInt64(&pendingRequests, 1)
			go func() {
				err := l.Exec(x)
				total := time.Since(begin)

				result := metrics{
					error:       err != nil,
					nanoseconds: total.Nanoseconds(),
				}
				ms <- result
				atomic.AddInt64(&pendingRequests, -1)
			}()
			sent++
		}

	}
}

type metrics struct {
	nanoseconds int64
	error       bool
}
