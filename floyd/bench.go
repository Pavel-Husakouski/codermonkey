package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const N = 512
const INF = 1000000000
const EXPECTED = 12970375

var start = time.Now()

func now() float64 {
	return float64(time.Since(start).Nanoseconds()) / 1e6
}

func Init() []int32 {
	d := make([]int32, N*N)
	s := uint32(12345)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			s = s*1664525 + 1013904223
			if i == j {
				d[i*N+j] = 0
			} else if (s>>8)%10 < 7 {
				d[i*N+j] = INF
			} else {
				d[i*N+j] = int32((s>>16)&1023) + 1
			}
		}
	}
	return d
}

func Floyd(d []int32) {
	for k := 0; k < N; k++ {
		for i := 0; i < N; i++ {
			dik := d[i*N+k]
			for j := 0; j < N; j++ {
				t := dik + d[k*N+j]
				if t < d[i*N+j] {
					d[i*N+j] = t
				}
			}
		}
	}
}

func Checksum(d []int32) int64 {
	var sum int64
	for _, v := range d {
		sum += int64(v)
	}
	return sum
}

func Main() {
	watch := now()
	d := Init()
	init := now() - watch

	watch = now()
	Floyd(d)
	floyd := now() - watch

	if sum := Checksum(d); sum != EXPECTED {
		fmt.Fprintf(os.Stderr, "Checksum mismatch: %d != %d\n", sum, EXPECTED)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Init: %.2f ms, Floyd: %.2f ms\n", init, floyd)
}

func delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func main() {
	iterations := 100
	// ITER env var overrides the iteration count (smoke testing only)
	if s := os.Getenv("ITER"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			iterations = n
		}
	}
	for i := 0; i < iterations; i++ {
		Main()
		delay(1000)
	}
}
