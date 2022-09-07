# Go commandline tools with pipes

## Introduction
One of the most useful ways of manipulating text is using
trusty commandline tools separated by pipes. This comes up
all the time, for example searching through large log files
or reformatting data in csv files. This blog shows how we can
incorporate our own programs written in Go into a chain of
commandline pipes.

## Pipes
On the commandline, pipes (`|` on the commandline) allow the
output from one program to be fed into the input of another.
For example

```shell
grep 'user1' large_log_file.txt | less
```

which looks for lines with the string "user1" in the log file using
the tool `grep` and passes matching lines to the program `less` which
paginates large output to make it easier for us to read.

We can also put `grep` into the pipeline using the program `cat` which
reads a file and prints the output to the terminal:

```shell
cat large_log_file.txt | grep 'user1' | less
```

has the same efect as given the file to grep itself.

The process works by the program on the left of a `|` printing output
to `stdout` which the pipe then redirects to the `stdin` of the program
on the right.

## Gogrep
To illustrate how a program can sit between these pipes, we are
going to implement `gogrep` which is a very simplified `grep` to
search files for a given regular expression. It needs the following
functionality:

1. Take arguments of a pattern to search for and a file to search through
2. If the file argument is omitted, read from stdin
3. Provide an option to count the number of matches rather than printing the matching lines
4. Provide an option to perform case insensitive matching

We'll build up the functionality in stages. Our implementation will not focus
on optimising I/O performance.

###