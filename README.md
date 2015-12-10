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
       ?

    GLOBAL OPTIONS:
       --color, -c		Force colored output
       --stderr, -e		Only annotate standard error
       --no-color, -n	No colored output
       --stdout, -o		Only annotate standard output
       --prefix, -p 	Override the default prefix
       --version, -v	print the version
