package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/slankdev/vtyang/pkg/vtyang"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := vtyang.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
