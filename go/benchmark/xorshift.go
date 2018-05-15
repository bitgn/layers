package main

import "time"

type XorShift struct {
	state uint64
}

func NewXorShift() *XorShift {
	seed := uint64(time.Now().Nanosecond())
	return &XorShift{seed}
}

func (x *XorShift) Next() uint64 {
	x.state ^= (x.state >> 12)
	x.state ^= (x.state << 25)
	x.state ^= (x.state >> 27)
	return x.state * 2685821657736338717
}
