package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/codegangsta/cli"

	"golang.org/x/crypto/ssh/terminal"
)

var version = "?"
var defaultFormat = "%0 "

func halt(err error) {
	fmt.Fprintf(os.Stderr, "annotate: error: %v", err)
	os.Exit(1)
}

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
			Usage: "Override the default prefix",
		},
		cli.BoolFlag{
			Name:  "o, stdout",
			Usage: "Only annotate standard output",
		},
		cli.BoolFlag{
			Name:  "e, stderr",
			Usage: "Only annotate standard error",
		},
		cli.BoolFlag{
			Name:  "c, color",
			Usage: "Force colored output",
		},
		cli.BoolFlag{
			Name:  "n, no-color",
			Usage: "No colored output",
		},
		cli.BoolFlag{
			Name:  "v, version",
			Usage: "Print the version",
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
	stdoutFormatter := func() string { return formatPrefix(name, stdoutPrefix, os.Stdout) }

	rw := newReadWriter(bufio.NewReader(os.Stdin), os.Stdout)
	annotate(rw, stdoutFormatter)
}

func annotateCommand(c *cli.Context) {
	name, args := splitArgs(c.Args())
	prefix, color := getPrefixAndColor(c, name, defaultFormat)
	stdoutPrefix, stderrPrefix := preparePrefix(prefix, color, c.Bool("color"))
	stdoutFormatter := func() string { return formatPrefix(name, stdoutPrefix, os.Stdout) }
	stderrFormatter := func() string { return formatPrefix(name, stderrPrefix, os.Stderr) }
	stdoutAnnotator := func(rw *readWriter) { annotate(rw, stdoutFormatter) }
	stderrAnnotator := func(rw *readWriter) { annotate(rw, stderrFormatter) }

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

func splitArgs(a []string) (cmd string, args []string) {
	if len(a) > 0 {
		cmd = a[0]
	}
	if len(a) > 1 {
		args = a[1:]
	}
	return
}
