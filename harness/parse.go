package main

import (
	"regexp"
	"strconv"
)

var lineRe = regexp.MustCompile(`Init:\s*([\d.]+)\s*ms,\s*(\w+):\s*([\d.]+)\s*ms`)

func itoa(n int) string { return strconv.Itoa(n) }

// Sample is one benchmark iteration.
type Sample struct{ Init, Main float64 }

// parseLine returns the sample and main phase name, or ok=false.
func parseLine(s string) (smp Sample, phase string, ok bool) {
	m := lineRe.FindStringSubmatch(s)
	if m == nil {
		return
	}
	smp.Init, _ = strconv.ParseFloat(m[1], 64)
	smp.Main, _ = strconv.ParseFloat(m[3], 64)
	return smp, m[2], true
}
