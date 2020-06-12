# Writing function main in Go

This blog is written as part of our Evergreen Innovations blog series <a href="https://www.evergreeninnovations.co/tech-blog/" target="_blank">here</a>.

## Overview

An important part of writing software is dealing with errors and, in particular, their interaction with resources. For example we should close any database connections or open files before our program exits. Typically the `defer` keyword is used to perform such actions as defered functions run just before a function returns. We can use a similar approach in our `main` function (where the program starts and finishes). A slight difference, however, is that function `main` does not return an `error` like most go functions but instead indicates failure with a non-zero exit code.

## Errors generated in main
In Go, a non-zero exit code can be achieved with a call to `os.Exit(1)` for an exit code of 1. This should be avoided as deferred functions are not called and therefore we do not clean up any resources. To show this in action, consider the following example:

```go
func main() {
	withError := flag.Bool("error", false, "exit program with error")
	flag.Parse()

	defer func() {
		fmt.Println("in the deferred function")
	}()

	fmt.Println("in main")

	if *withError {
		fmt.Println("exiting with error")
		os.Exit(1)
	}
}
```

If you run this program with `go run main.go` in the `example1` directory you will see something something similar to

```
in main
in the deferred function
```

To trigger the exit code, you can run the program with `go run main.go --error` which gives an output similar to

```
in main
exiting with error
```

Notice that the deferred function is not run.

A common alternative is to panic in the case of an error such as in `example2`, which has the same struture as above save for the error condition:

```go
	if *withError {
		panic("Exiting with error")
	}
```

In the output of running `example2` with the error state, we can see the lines

```
in main
in the deferred function
panic: Exiting with error
```

Although deferred functions are called after a panic and the program exits with a code of 2, there are several downsides of panicing.
The first is that panic prints a stack trace and other debugging information which can obscure the actual error and, secondly, panic
prints to stderr which may bypass any logging infrastructure we have up. The approach we take, therefore, is to use the fact that
deferred functions are run in first-in-last-out order. Taking `example3`, we start main with

```go
func main() {
	var err error

	// Deferred functions run in reverse order so this will be the last
	// one called, after any tidy up.
	defer func() {
		if err != nil {
			fmt.Println("error encountered:", err)
			os.Exit(1)
		} else {
			fmt.Println("exiting")
		}
	}()
}
```

in which we use `err` to capture any errors we encounter. As this is the first deferred function, it will be run after all
the other deferred functions. If we have had an error (`err != nil`) then we print the error and exit with the appropriate code
and otherwise the program exits as normal.

We can then use the following pattern for any functions that might return an error in our main routine:

```go
func main() {
...
	f, err := os.Open(fileName)
	if err != nil {
		err = fmt.Errorf("opening file: %v", err)
		return
	}
	defer f.Close()
...
}
```

where in the case of an error we set the `err` and return, triggering all the defered function until we end with the one outlined
above. If the call is successful (no error) we defer the close operation until the program exits.

You can play around with this program which takes a file name as the only required argument, some outputs are (there is some additional logging in the file closing defer)

```shell
$ go run main.go
error encountered: invalid number of arguments
```

```shell
$ go run main.go file.txt
File contents: EGI structure of function main blog
closed the file
exiting
```

```shell
$ go run main.go nonexistent.txt
error encountered: opening file: open nonexistent.txt: no such file or directory
```

## Trapping signals

A second way our program could terminate early is if an interupt is received, for example by pressing `CTRL+C` at the commandline.
In this case deferred functions are not called meaning we leak resources. To alleviate this problem we can use the standard library
`signal` package to catch these signals using the following code in `example4`

```go
func main() {
	var err error

	// Deferred functions run in reverse order so this will be the last
	// one called, after any tidy up.
	defer func() {
		if err != nil {
			fmt.Println("error encountered:", err)
			os.Exit(1)
		} else {
			fmt.Println("exiting")
		}
	}()

...

	errs := make(chan error)

	go func() {
		// Simple long-running process
		for i := 0; i < 10; i++ {
			fmt.Println("in the for loop, iteration", i)
			time.Sleep(time.Second)
		}

		// Indicate normal end to the program
		errs <- nil
	}()


	// Trap any signals to exit gracefully
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("signal trapped: %v", <-c)
	}()

	// Block execution until any errors are encountered.
	// Deferred functions will be run afterwards.
	err = <-errs
}
```

We now run our code, in this case just a for loop to simulate a long-running process, within a go-routine alongside the signal capturing code and block the execution of
`main` with `err = <-errs`. When an error is triggered main will return, running all deferred functions first. If we run the
program and press `CTRL+C` during execution the output is

```
in the for loop, iteration 0
in the for loop, iteration 1
in the for loop, iteration 2
in the for loop, iteration 3
in the for loop, iteration 4
in the for loop, iteration 5
in the for loop, iteration 6
^Cexiting with error: signal trapped: interrupt
```

The last line indicates that our deferred function has run sucessfully.

## Conclusion
In this blog we have demonstrated how we tend to write our `main` function for the systems we develop. If you want us to describe anything else go-related, please get in touch.
