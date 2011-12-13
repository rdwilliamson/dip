package dip

import (
	"runtime"
)

var goRoutines int

func init() {
	goRoutines = runtime.GOMAXPROCS(0)
}

func SetGoRoutines(c int) {
	goRoutines = c
}

func goRountineRanges(min, max, i int) (int, int) {
	r := (max - min) / goRoutines
	if i == goRoutines-1 {
		return min + i*r, max
	}
	b := min + i*r
	return b, b + r
}
