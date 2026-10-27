package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const N = 4096
const EXPECTED = 6647453

var start = time.Now()

func now() float64 {
	return float64(time.Since(start).Nanoseconds()) / 1e6
}

func Init() ([]uint8, []uint8, []int32, []int32) {
	a := make([]uint8, N)
	b := make([]uint8, N)
	s := uint32(12345)
	for i := 0; i < N; i++ {
		s = s*1664525 + 1013904223
		a[i] = uint8((s >> 16) & 3)
	}
	for i := 0; i < N; i++ {
		s = s*1664525 + 1013904223
		b[i] = uint8((s >> 16) & 3)
	}
	return a, b, make([]int32, N+1), make([]int32, N+1)
}

func Lcs(a, b []uint8, prev, cur []int32) []int32 {
	for i := 1; i <= N; i++ {
		cur[0] = 0
		for j := 1; j <= N; j++ {
			if a[i-1] == b[j-1] {
				cur[j] = prev[j-1] + 1
			} else {
				cur[j] = max(prev[j], cur[j-1])
			}
		}
		prev, cur = cur, prev
	}
	return prev
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
	a, b, prev, cur := Init()
	init := now() - watch

	watch = now()
	res := Lcs(a, b, prev, cur)
	lcs := now() - watch

	if sum := Checksum(res); sum != EXPECTED {
		fmt.Fprintf(os.Stderr, "Checksum mismatch: %d != %d\n", sum, EXPECTED)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Init: %.2f ms, Lcs: %.2f ms\n", init, lcs)
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
