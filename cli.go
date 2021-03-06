package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"

	"golang.org/x/crypto/ssh/terminal"
)

var version = "?"

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
		cli.BoolFlag{
			Name:  "a, print-args",
			Usage: "Print command with arguments before output",
		},
		cli.StringFlag{
			Name:  "s, print-separator",
			Value: "=",
			Usage: "Print separator before and after output",
		},
		cli.StringFlag{
			Name:   "p, prefix",
			Value:  "%0 ",
			EnvVar: "ANNOTATE_PREFIX",
			Usage:  "Override the default prefix",
		},
		cli.BoolFlag{
			Name:  "o, stdout",
			Usage: "Only annotate standard output",
		},
		cli.BoolFlag{
			Name:  "e, stderr",
			Usage: "Only annotate standard error",
		},
		cli.BoolTFlag{
			Name:  "c, color",
			Usage: "Colorize prefix. Default on. If set, force colorize.",
		},
		cli.BoolFlag{
			Name:  "v, version",
			Usage: "Print the version",
		},
		cli.BoolFlag{
			Name:  "h, help",
			Usage: "Show help",
		},
	}

	app.RunAndExitOnError()
}

func actionMain(c *cli.Context) {
	if err := validateOptions(c); err != nil {
		halt(err)
	}

	if c.Bool("help") {
		cli.ShowAppHelp(c)
		return
	}

	if c.Bool("version") {
		cli.ShowVersion(c)
		return
	}

	if len(c.Args()) > 0 {
		annotateCommand(c)
	} else if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		annotatePipe(c)
	} else {
		cli.ShowAppHelp(c)
	}
}

func validateOptions(c *cli.Context) error {
	if c.Bool("stdout") && c.Bool("stderr") {
		return errors.New("Conflicting flags: -o|--stdout and -e|--stderr")
	}
	return nil
}

func annotatePipe(c *cli.Context) {
	name := ">"
	color := hashedColor(name)
	stdoutPrefix, _ := preparePrefix(c.String("prefix"), color, c.IsSet("color"), c.Bool("color"))
	stdoutFormatter := func() string { return formatPrefix(name, stdoutPrefix, os.Stdout) }

	rw := newReadWriter(bufio.NewReader(os.Stdin), os.Stdout)

	if c.IsSet("print-separator") {
		printSeparator(c.String("print-separator"), stdoutFormatter)
	}
	err := annotate(rw, stdoutFormatter)
	if c.IsSet("print-separator") {
		printSeparator(c.String("print-separator"), stdoutFormatter)
	}

	if err != nil {
		halt(err)
	}
}

func annotateCommand(c *cli.Context) {
	name := guessCommand(c.Args())
	color := hashedColor(name)
	stdoutPrefix, stderrPrefix := preparePrefix(c.String("prefix"), color, c.IsSet("color"), c.Bool("color"))
	stdoutFormatter := func() string { return formatPrefix(name, stdoutPrefix, os.Stdout) }
	stderrFormatter := func() string { return formatPrefix(name, stderrPrefix, os.Stderr) }

	stdoutAnnotator := func(rw *readWriter) { annotate(rw, stdoutFormatter) }
	stderrAnnotator := func(rw *readWriter) { annotate(rw, stderrFormatter) }

	prog, args := c.Args()[0], c.Args()[1:]
	cmd := exec.Command(prog, args...)

	// Pass-throughs
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin

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

	if c.Bool("print-args") {
		printArguments(c.Args(), stdoutFormatter)
	}
	if c.IsSet("print-separator") {
		printSeparator(c.String("print-separator"), stdoutFormatter)
	}
	err := cmd.Run()
	if c.IsSet("print-separator") {
		printSeparator(c.String("print-separator"), stdoutFormatter)
	}

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

func preparePrefix(prefix string, color uint32, coloredSet bool, colored bool) (string, string) {
	stdoutPrefix := prefix
	stderrPrefix := prefix
	hasStdout := terminal.IsTerminal(int(os.Stdout.Fd()))
	hasStderr := terminal.IsTerminal(int(os.Stderr.Fd()))

	if colored && (coloredSet || hasStdout) {
		stdoutPrefix = fmt.Sprintf("\x1b[3%dm%s\x1b[0m", color, prefix)
	}
	if colored && (coloredSet || hasStderr) {
		stderrPrefix = fmt.Sprintf("\x1b[3%d;1m%s\x1b[0m", color, prefix)
	}

	return stdoutPrefix, stderrPrefix
}

func guessCommand(args []string) string {
	firstFlagIndex := -1

	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			firstFlagIndex = i
			break
		}
	}

	if firstFlagIndex > 1 {
		return args[0] + " " + args[1]
	}

	return args[0]
}
