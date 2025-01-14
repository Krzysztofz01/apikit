package log

import (
	"fmt"
	"os"
)

func FatalErr(err interface{}) {
	var message string
	switch e := err.(type) {
	case string:
		message = e
	case error:
		message = e.Error()
	default:
		message = "cmd: unknown runtime failure"
	}

	fmt.Fprintf(os.Stderr, "%s\n", message)
	os.Exit(1)
}
