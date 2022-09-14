# Go commandline tools with pipes

## Introduction
One of the most useful ways of manipulating text is using
trusty commandline tools separated by pipes. This comes up
all the time, for example searching through large log files
or reformatting data in csv files. This blog shows how we can
incorporate our own programs written in Go into a chain of
commandline pipes.

## Pipes
On the commandline, pipes (`|`) allow the
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
3. Provide an option to perform case insensitive matching

We'll build up the functionality in stages. Our implementation will not focus
on optimising I/O performance.

### Step 1 - search through a file
We have written in a [previous blog](https://www.evergreeninnovations.co/blog-writing-function-main-in-go/)
about how we like to structure our `main` function. The same approach is used
here in which there is a deferred function at the start of `main` which is
used to print any error and exit with the appropriate code. The advantage of
this approach is that it allows us to `defer` other operations and be certain
they will be executed when the program finishes.

```go
package main

import (
    ...
)

func main() {
	// Deferred functions called in FILO order meaning
	// this will be the last function to run when we return.
	var mainErr error
	defer func() {
		if mainErr != nil {
			fmt.Printf("%s: %s\n", os.Args[0], mainErr)
			os.Exit(1)
		}
	}()
    ...
```

Our program is going to take two arguments: a pattern to search for, and the
file to search in. The pattern is compiled into a regular expression and the
file is opened, ready to be read. Note that `os.Args[0]` is the name of the
running program; this is not required so stripped before processing the
commandline arguments.

```go
	args := os.Args[1:]

	if len(args) == 0 {
		mainErr = errors.New("pattern must be supplied")
		return
	}

	pat := args[0]
	rex, err := regexp.Compile(pat)
	if err != nil {
		mainErr = err
		return
	}

	if len(args) != 2 {
		mainErr = errors.New("file must be supplied")
        return
	}

	f, err := os.Open(args[1])
	if err != nil {
		mainErr = err
		return
	}
	defer f.Close()
```

The file is then read line by line using a [scanner](https://pkg.go.dev/bufio#Scanner)
with lines matching the scanner printed to `Stdout`

```go
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()

		if rex.MatchString(l) {
			fmt.Println(l)
		}
	}

	if err := scanner.Err(); err != nil {
		mainErr = err
		return
	}
}
```

We can see this program is action with the following file named
`test.txt`

```
$ cat test.txt
first line
second line with ending
third entry in the file
fouth entry in the file
```

For example
```shell
$ ./step1 'line' test.txt
first line
second line with ending
```

```shell
$ ./step1 'line$' test.txt
first line
```

### Step 2 - Add ability to read from stdin
To optionally read from stdin if the file postitional argument is ommited.
The only part of the code that changes is where the file was opened. Now, we
intially assign `Stdin` to the file handle `f` and then overwrite `f` if a file
has been provided.

```go
	f := os.Stdin
	if len(args) == 2 {
		f, err = os.Open(args[1])
		if err != nil {
			mainErr = err
			return
		}
		defer f.Close()
	}
```

Using providing a file as positional argument works exactly as before

```shell
$ ./step2 'line' test.txt
first line
second line with ending
```

In addition, We can now pipe input from another program into ours, for example

```shell
$ cat test.txt | ./step2 'line'
first line
second line with ending
```

As we are printing the matching lines, we can further pipe the output to
another program, to, for example, replace the word 'first' with 'changed'

```shell
$ cat test.txt | ./step2 'line' | sed 's/first/changed/'
changed line
second line with ending
```

### Step 3 Provide case insensitive matching
Following `grep` we are going to provide a flag `-i` to perform case insenstive
matching. To parse the commandline options the standard library's [flag](https://pkg.go.dev/flag) has been used. We can specify the flag name and default value and also use the `flag.Args()` function to get the remaining positional arguments.


```go
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] pattern [file] \n", os.Args[0])
		flag.PrintDefaults()
	}

	var insensitiveFlag bool
	flag.BoolVar(&insensitiveFlag, "i", false, "insensitive match")
	flag.Parse()

	args := flag.Args()
```

We are also overwriting the `flag.Usage` function to print the usage of the program in the manner we want.

```shell
$ ./step3 -h
Usage: ./step3  [options] pattern [file]
  -i    insensitive match

```

To perform case insensitive matching, we need to prepend `(?i)` to the given
pattern as described in the [Go documentation](https://pkg.go.dev/regexp/syntax).

```go
	if insensitiveFlag {
		pat = "(?i)" + pat
	}
```

The rest of the code is unchanged. An example of case insensitive matching,

```go
$ ./step3 -i 'LINE' test.txt
first line
second line with ending
```

whereas the code in `step1` would not have returned any matches.

## Conclusion
Hopefully this blog has shown that including Go programs that we write into
pipe separated commands is easily accomplished. We also saw how we can include
commandline options in our Go programs alongside positional arguments.

You can find the full code for the above steps, along with a version that can
output the number of matches found using a `-c` commandline option, on our [Github](https://github.com/evergreen-innovations/blogs) page.
