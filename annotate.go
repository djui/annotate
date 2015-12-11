package main

import (
	"bufio"
	"bytes"
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
	s.Split(scanLines)
	for s.Scan() {
		printAnnotated(rw, prefix, s.Text())
	}
	return s.Err()
}

// ScanLines is a split function for a Scanner that returns each line of text,
// stripped of any trailing end-of-line marker. The returned line may be
// empty. The end-of-line marker is either one carriage return not followed by a
// newline or one optional carriage return followed by one mandatory newline or
// one carriage return. In regular expression notation, it is `\r[^\n]|\r?\n`.
// The last non-empty line of input will be returned even if it has no newline.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	r := bytes.IndexByte(data, '\r')
	n := bytes.IndexByte(data, '\n')
	// We have a full \r\n-terminated line.
	if r >= 0 && n == r+1 {
		return r + 2, data[0:r], nil
	}
	// We have a full \r-terminated line.
	if r >= 0 && (n < 0 || r < n) {
		return r + 1, data[0:r], nil
	}
	// We have a full \n-terminated line.
	if n >= 0 && (r < 0 || n < r) {
		return n + 1, data[0:n], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
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
