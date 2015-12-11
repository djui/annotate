# Annotate

A command-line tool that annotates command output (standard output and standard
error). By default it prefixes the output with the command name in color.

The purpose is to separate the output of a sequence of commands e.g. executed by
a Makefile.


## Install

    $ go get github.com/djui/annotate

or:

    $ make


## Usage

    $ annotate
    NAME:
       annotate - Annotate a command's standard output and standard error.

    USAGE:
       ./annotate [global options] [arguments...]

    VERSION:
       3.0

    GLOBAL OPTIONS:
       -a, --print-args             Print command with arguments before output
       -s, --print-separator "="    Print separator before and after output
       -p, --prefix "%0 "           Override the default prefix [$ANNOTATE_PREFIX]
       -o, --stdout                 Only annotate standard output
       -e, --stderr                 Only annotate standard error
       -c, --color                  Force colored output. Default on. If set, force colorize.
       -v, --version                Print the version
       -h, --help                   Show help
