package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"time"
)

func formatPrefix(prog string, format string, w *os.File) string {
	t := time.Now()

	var fd string
	switch w {
	case os.Stdout:
		fd = "O"
	case os.Stderr:
		fd = "E"
	case os.Stdin:
		fd = "I"
	default:
		fd = "?"
	}

	var escaped bool
	var out string
	for _, c := range format {
		if escaped {
			switch c {
			case '0':
				out += path.Base(prog)
			case '>':
				out += fd
			case 'd':
				out += fmt.Sprintf("%02d", t.Day())
			case 'F':
				out += fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
			case 'H':
				out += fmt.Sprintf("%02d", t.Hour())
			case 'M':
				out += fmt.Sprintf("%02d", t.Minute())
			case 'm':
				out += fmt.Sprintf("%02d", t.Month())
			case 'N':
				out += fmt.Sprintf("%02d", t.Nanosecond())
			case 'S':
				out += fmt.Sprintf("%02d", t.Second())
			case 's':
				out += fmt.Sprintf("%02d", t.Unix())
			case 'T':
				out += fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
			case 'Y':
				out += fmt.Sprintf("%04d", t.Year())
			case '%':
				out += "%"
			default:
				out += "%" + string(c)
			}
			escaped = false
		} else if c == '%' {
			escaped = true
		} else {
			out += string(c)
		}
	}

	if escaped {
		out += "%"
	}

	return out
}

// hashedColor consistently generates a number between 1..6 for a given
// string. The color values represent red, green, yellow, blue, magenta, cyan.
func hashedColor(name string) uint32 {
	return hashStr(name)%6 + 1
}

// hash generates a consistent integer has from a string.
func hashStr(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 1
	}

	return h.Sum32()
}
