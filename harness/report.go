package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Result struct {
	Lang    string
	Version string
	Status  string // ok | FAIL | skipped
	Note    string
	Samples []Sample
}

type Stats struct{ First, Min, Median, Mean, P95, Max float64 }

func stats(v []float64) Stats {
	if len(v) == 0 {
		return Stats{}
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	var sum float64
	for _, x := range s {
		sum += x
	}
	p95 := s[min(len(s)-1, int(float64(len(s))*0.95))]
	med := s[len(s)/2]
	if len(s)%2 == 0 {
		med = (s[len(s)/2-1] + s[len(s)/2]) / 2
	}
	return Stats{First: v[0], Min: s[0], Median: med, Mean: sum / float64(len(s)), P95: p95, Max: s[len(s)-1]}
}

func col(r Result, f func(Sample) float64) Stats {
	v := make([]float64, len(r.Samples))
	for i, s := range r.Samples {
		v[i] = f(s)
	}
	return stats(v)
}

func cpuModel() string {
	b, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "?"
	}
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "model name") {
			if i := strings.Index(l, ":"); i >= 0 {
				return strings.TrimSpace(l[i+1:])
			}
		}
	}
	return "?"
}

func writeReport(path, bench, phase string, iter int, res []Result) error {
	var ok, bad []Result
	for _, r := range res {
		if r.Status == "ok" && len(r.Samples) > 0 {
			ok = append(ok, r)
		} else {
			bad = append(bad, r)
		}
	}
	main := func(s Sample) float64 { return s.Main }
	ini := func(s Sample) float64 { return s.Init }
	sort.SliceStable(ok, func(i, j int) bool { return col(ok[i], main).Median < col(ok[j], main).Median })

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", bench)
	fmt.Fprintf(&b, "- date: %s\n- cpu: %s (%d threads)\n- os: %s/%s\n- iterations: %d\n- times in ms; sorted by %s median\n\n",
		time.Now().Format("2006-01-02 15:04"), cpuModel(), runtime.NumCPU(), runtime.GOOS, runtime.GOARCH, iter, phase)
	fmt.Fprintf(&b, "| # | lang | version | init median | %[1]s first | %[1]s min | %[1]s median | %[1]s mean | %[1]s p95 | x best | status |\n", phase)
	b.WriteString("|--:|------|---------|------------:|------:|------:|------:|------:|------:|------:|--------|\n")
	var best float64
	if len(ok) > 0 {
		best = col(ok[0], main).Median
	}
	for i, r := range ok {
		m, in := col(r, main), col(r, ini)
		x := 0.0
		if best > 0 {
			x = m.Median / best
		}
		fmt.Fprintf(&b, "| %d | %s | %s | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2fx | ok |\n",
			i+1, r.Lang, r.Version, in.Median, m.First, m.Min, m.Median, m.Mean, m.P95, x)
	}
	if len(bad) > 0 {
		b.WriteString("\n## not measured\n\n| lang | status | reason |\n|------|--------|--------|\n")
		for _, r := range bad {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", r.Lang, r.Status, strings.ReplaceAll(r.Note, "\n", " "))
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
