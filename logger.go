package main

import "fmt"

// Log initial logging methods
// just a combination of println + fmt.Sprintf
func Log(msg string, args ...interface{}) {
	println(fmt.Sprintf(msg, args...))
}
