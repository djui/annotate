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
       2.1

    GLOBAL OPTIONS:
       -a, --print-args             Print command with arguments before output
       -s, --print-separator "="    Print separator before and after output
       -p, --prefix                 Override the default prefix
       -o, --stdout                 Only annotate standard output
       -e, --stderr                 Only annotate standard error
       -c, --color                  Force colored output
       -n, --no-color               No colored output
       -v, --version                Print the version
