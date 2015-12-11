package main

import (
	"bufio"
	"fmt"
	"io"
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
func annotate(r *readWriter, f func() string) {
	s := bufio.NewScanner(r)

	for s.Scan() {
		fmt.Fprintf(r, "%s%s\n", f(), s.Bytes())
	}

	if err := s.Err(); err != nil {
		halt(err)
	}
}
