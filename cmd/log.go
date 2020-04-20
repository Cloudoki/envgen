package cmd

import (
	"fmt"
	"os"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m\n"
	ErrorColor   = "\033[1;31m%s\033[0m\n"
)

func printLog(color string, msg interface{}) {
	fmt.Printf(color, msg)
}

func logInfo(msg interface{}) {
	printLog(InfoColor, msg)
}

type errorWriter struct {}
func (w errorWriter) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(os.Stdout, ErrorColor, p)
}