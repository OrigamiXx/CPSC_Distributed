package raft

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

// Debugging
var Debug = flag.Bool("debug", false, "whether to print debug logs")

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func DPrintf(format string, a ...interface{}) {
	if *Debug {
		_, path, line, _ := runtime.Caller(1 /*skip*/)
		file := filepath.Base(path)
		msg := fmt.Sprintf(format, a...)
		log.Printf("%s (%v:%v)", msg, file, line)
	}
}

func DPrintfFromNode(me int, format string, a ...interface{}) {
	if *Debug {
		_, path, line, _ := runtime.Caller(1 /*skip*/)
		file := filepath.Base(path)
		msg := fmt.Sprintf(format, a...)
		log.Printf("[Node %d] %s (%v:%v)", me, msg, file, line)
	}
}
