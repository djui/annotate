package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type readWriter struct {
	io.Reader
	io.Writer
}

func newReadWriter(r io.Reader, w io.Writer) *readWriter {
	return &readWriter{r, w}
}

// pipe reads from r, applies f, and writes to w.
func pipe(w io.Writer, f func(*readWriter)) io.Writer {
	r, wOut := io.Pipe()
	rw := newReadWriter(r, w)
	go f(rw)
	return wOut
}

// annotate each line read from "r" with prefix and write to "w".
func annotate(rw *readWriter, prefix func() string) error {
	s := bufio.NewScanner(rw)
	for s.Scan() {
		printAnnotated(rw, prefix, s.Text())
	}
	return s.Err()
}

func printArguments(args []string, prefix func() string) {
	printAnnotated(os.Stdout, prefix, strings.Join(args, " "))
}

func printSeparator(sep string, prefix func() string) {
	printAnnotated(os.Stdout, prefix, strings.Repeat(sep, 80))
}

func printAnnotated(w io.Writer, prefix func() string, s string) {
	fmt.Fprintf(w, "%s%s\n", prefix(), s)
}
