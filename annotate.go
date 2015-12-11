package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/codegangsta/cli"

	"golang.org/x/crypto/ssh/terminal"
)

var version = "?"
var defaultFormat = "%0 "

func main() {
	app := cli.NewApp()
	app.Name = "annotate"
	app.Usage = "Annotate a command's standard output and standard error."
	app.Version = version
	app.Action = actionMain
	// Hide help otherwise `help` as the second paramater is not interpreted as
	// first argument but as subcommand.
	app.HideHelp = true
	// The order of the version flag's short- and long-form are swapped.
	app.HideVersion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "p, prefix",
			Usage: "     Override the default prefix",
		},
		cli.BoolFlag{
			Name:  "o, stdout",
			Usage: "     Only annotate standard output",
		},
		cli.BoolFlag{
			Name:  "e, stderr",
			Usage: "     Only annotate standard error",
		},
		cli.BoolFlag{
			Name:  "c, color",
			Usage: "      Force colored output",
		},
		cli.BoolFlag{
			Name:  "n, no-color",
			Usage: "    No colored output",
		},
		cli.BoolFlag{
			Name:  "v, version",
			Usage: "     Print the version",
		},
	}

	app.RunAndExitOnError()
}

func actionMain(c *cli.Context) {
	if c.Bool("stdout") && c.Bool("stderr") {
		fmt.Fprintln(os.Stderr, "Conflicting flags: -o|--stdout and -e|--stderr")
		os.Exit(1)
	}

	if c.Bool("color") && c.Bool("no-color") {
		fmt.Fprintln(os.Stderr, "Conflicting flags: -c|--color and -n|--no-color")
		os.Exit(1)
	}

	if c.Bool("version") {
		cli.ShowVersion(c)
		os.Exit(0)
	}

	if len(c.Args()) > 0 {
		annotateCommand(c)
	} else if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		annotatePipe(c)
	} else {
		cli.ShowAppHelp(c)
	}
}

func annotatePipe(c *cli.Context) {
	name := "?"
	prefix, color := getPrefixAndColor(c, name, defaultFormat)
	stdoutPrefix, _ := preparePrefix(prefix, color, c.Bool("color"))
	stdoutFormatter := func() string { return formatPrefix(stdoutPrefix, os.Stdout, name) }

	r := bufio.NewReader(os.Stdin)
	annotate(r, os.Stdout, stdoutFormatter)
}

func annotateCommand(c *cli.Context) {
	name, args := splitArgs(c.Args())
	prefix, color := getPrefixAndColor(c, name, defaultFormat)
	stdoutPrefix, stderrPrefix := preparePrefix(prefix, color, c.Bool("color"))
	stdoutFormatter := func() string { return formatPrefix(stdoutPrefix, os.Stdout, name) }
	stderrFormatter := func() string { return formatPrefix(stderrPrefix, os.Stderr, name) }
	stdoutAnnotator := func(r io.Reader, w io.Writer) { annotate(r, w, stdoutFormatter) }
	stderrAnnotator := func(r io.Reader, w io.Writer) { annotate(r, w, stderrFormatter) }

	cmd := exec.Command(name, args...)

	if c.Bool("stdout") {
		cmd.Stdout = pipe(os.Stdout, stdoutAnnotator)
		cmd.Stderr = os.Stderr
	} else if c.Bool("stderr") {
		cmd.Stdout = os.Stdout
		cmd.Stderr = pipe(os.Stderr, stderrAnnotator)
	} else {
		cmd.Stdout = pipe(os.Stdout, stdoutAnnotator)
		cmd.Stderr = pipe(os.Stderr, stderrAnnotator)
	}

	// Pass-through environment variables
	for _, env := range os.Environ() {
		cmd.Env = append(cmd.Env, env)
	}
	err := cmd.Run()

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			// FIXME: Not sure if we should print the error text
			//fmt.Fprintln(os.Stderr, err)
			os.Exit(status.ExitStatus())
		} else {
			// FIXME: Not sure how this error could happen
			halt(err)
		}
	} else if err != nil {
		halt(err)
	}
}

func splitArgs(a []string) (cmd string, args []string) {
	if len(a) > 0 {
		cmd = a[0]
	}
	if len(a) > 1 {
		args = a[1:]
	}
	return
}

func getPrefixAndColor(c *cli.Context, name string, prefixDefault string) (prefix string, color uint32) {
	prefix = prefixDefault
	color = hashedColor(name)

	if c.IsSet("prefix") {
		prefix = c.String("prefix")
	}

	if c.Bool("no-color") {
		color = 0
	}

	return
}

func preparePrefix(p string, color uint32, force bool) (stdoutPrefix string, stderrPrefix string) {
	hasStdout := terminal.IsTerminal(int(os.Stdout.Fd()))
	hasStderr := terminal.IsTerminal(int(os.Stderr.Fd()))

	if color > 0 && (force || hasStdout) {
		stdoutPrefix = fmt.Sprintf("\x1b[3%dm%s\x1b[0m", color, p)
	} else {
		stdoutPrefix = p
	}

	if color > 0 && (force || hasStderr) {
		stderrPrefix = fmt.Sprintf("\x1b[3%d;1m%s\x1b[0m", color, p)
	} else {
		stderrPrefix = p
	}

	return
}

func formatPrefix(format string, w *os.File, prog string) string {
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
	col := hash(name) % 6
	return col + 1
}

// hash generates a consistent integer has from a string.
func hash(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 1
	}

	return h.Sum32()
}

// pipe reads from r, applies f, and writes to w.
func pipe(w io.Writer, f func(io.Reader, io.Writer)) io.Writer {
	r, wOut := io.Pipe()
	go f(r, w)
	return wOut
}

// annotate each line read from "r" with prefix and write to "w".
func annotate(r io.Reader, w io.Writer, f func() string) {
	s := bufio.NewScanner(r)

	for s.Scan() {
		fmt.Fprintf(w, "%s%s\n", f(), s.Bytes())
	}

	if err := s.Err(); err != nil {
		halt(err)
	}
}

func halt(err error) {
	fmt.Fprintf(os.Stderr, "annotate: error: %v", err)
	os.Exit(1)
}
