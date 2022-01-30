package util

import (
	"fmt"
	"runtime"
	"strings"
)

func FUNC() string {
	pt, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("UNSUPPORTED")
	}
	fn := runtime.FuncForPC(pt).Name()
	fna := strings.Split(fn, "/")
	return fna[len(fna)-1]
}

func FULLFUNC() string {
	pt, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("UNSUPPORTED")
	}
	return runtime.FuncForPC(pt).Name()
}

func LINE() string {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("UNSUPPORTED")
	}
	return fmt.Sprintf("%s:%d", file, line)
}
