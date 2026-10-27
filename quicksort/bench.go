package main

import (
	"fmt"
	"iter"
	"math"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const N = 1000 * 1000

var start = time.Now()

func now() float64 {
	return float64(time.Since(start).Nanoseconds()) / 1e6
}

var byId = func(x, y *TestClass) int {
	return strings.Compare(x.Id, y.Id)
}

func Init(number int) []*TestClass {
	return ToArray(Select(Range(0, number), func(x int) *TestClass { return NewTestClass() }))
}

func Sort(vals []*TestClass) {
	QuickSort(vals, byId)
}

func Main(number int) {
	watch := now()
	vals := Init(number)
	init := now() - watch

	// runtime.GC()
	watch = now()
	Sort(vals)
	sort := now() - watch

	fmt.Fprintf(os.Stderr, "Init: %.2f ms, Sort: %.2f ms\n", init, sort)
}

func QuickSort(vals []*TestClass, compare func(x, y *TestClass) int) {
	QuickSortImpl(vals, compare, 0, len(vals)-1)
}

func QuickSortImpl(vals []*TestClass, compare func(x, y *TestClass) int, left, right int) {
	if right <= left {
		return
	}

	i := left
	j := right
	mid := (left + right) >> 1
	midVal := vals[mid]

	for i <= j {
		for compare(vals[i], midVal) > 0 {
			i++
		}

		for compare(vals[j], midVal) < 0 {
			j--
		}

		if i <= j {
			tmp := vals[i]
			vals[i] = vals[j]
			vals[j] = tmp

			i++
			j--
		}
	}

	if left < i-1 {
		QuickSortImpl(vals, compare, left, i-1)
	}

	if right > i {
		QuickSortImpl(vals, compare, i, right)
	}
}

type TestClass struct {
	Id    string
	Value string
}

func NewTestClass() *TestClass {
	t := &TestClass{}
	t.Id = CreateGuid()
	t.Value = CreateGuid()
	return t
}

func (t *TestClass) String() string {
	return t.Id
}

func ToArray[T any](iterator iter.Seq[T]) []T {
	return slices.Collect(iterator)
}

func Range(start, end int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := start; i < end; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func Select[T, U any](iterator iter.Seq[T], projector func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for x := range iterator {
			if !yield(projector(x)) {
				return
			}
		}
	}
}

func CreateGuid() string {
	return strconv.FormatFloat(math.Floor(rand.Float64()*1e16), 'f', -1, 64)
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
		Main(N)
		delay(1000)
	}
}
