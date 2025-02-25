package main

import (
	"fmt"
)

// log logs an informational message as JSON.
func log(fields ...interface{}) {
	fmt.Println(fields...)
}
