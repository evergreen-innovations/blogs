Working with package time in Go
=======================

Overview
---------------------------
At Evergreen Innovations we have recently been writing back-end systems for 
our clients in Go. Go has proven itself to be a productive and robust language for these applications, with the time related functionality available in the standard library being particularly impressive.

In common with most of the Go standard library, the time package is well 
documented with a number of examples. The official documentation can be found 
in [godoc](https://golang.org/pkg/time/).

The aim of this blog post is to demonstrate some of the functionality of the time package and highlight a couple of issues we came across when first getting to grips with the package.

Creating Time
---------------------------
The [time.Time](https://golang.org/pkg/time/) type is the workhorse of `package time` and, importantly, is time-zone aware. Getting the current time uses the `time.Now()` function with returns a `time.Time` `struct`. `time.Time` has many convenience methods, for example

```go
    package main

    import (
    	"fmt"
    	"time"
    )

    func main() {
	    now := time.Now()
        day := now.Day()
        month := new.Month()

	    fmt.Println("This is day", day, "of month", month)
    }

    This is day 27 of month May
```

To get the month as an integer, we can use the `Date` [method](https://golang.org/pkg/time/#Time.Date)

```go
    ...

    year, month, day := now.Date()

    fmt.Println("Today is in the year", year, " and month", month)

    ...
```

As previously mentioned, the `time.Type` type is location aware. If we want to create a specific date-time, the `Date` [function](https://golang.org/pkg/time/#Date) is used, which has the signature

    func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time

The final location parameter can either be the constant `time.UTC` or be generated from the `LoadLocation` [function](https://golang.org/pkg/time/#LoadLocation), which takes the timezone name as the only argument. A helpful list of these names can be found [here](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) and take the form `Europe/London` etc. `LoadLocation` also accepts `"Local"` and `"UTC"` as special cases. For example, to generate a date in Auckland, New Zealand,

```go
    func main() {
    	loc, err := time.LoadLocation("Pacific/Auckland")
    	if err != nil {
    		fmt.Println("Could not load location:", err)
    		return
    	}

    	date := time.Date(2019, 5, 5, 11, 0, 0, 0, loc)

    	fmt.Println("The time in Auckland is", date)
    	fmt.Println("Or in UTC              ", date.UTC())
    }

    The time in Auckland is 2019-05-05 11:00:00 +1200 NZST
    Or in UTC               2019-05-04 23:00:00 +0000 UTC
```

The package helpfully defines the `UTC` method which returns a new `time.Time` with the location set the UTC, with the corresponding adjustments to the date and time made.

In order to use `time.LoadLocation`, the tz database with the appropriate entry must be on the host system. As with all Go code, therefore, the error should always be checked just in case the database is missing.

Parsing and Formatting
-------------------------------------------
This brings us nicely onto an aspect of the time package which seems a bit confusing at first, formatting and parsing time strings. As can be seen in the previous example, the `time.Time` type has a `String` method but often we want to use a particular format.

Most languages implement time formatting using verbs, similar to those used in with `fmt.Printf`. For example using the `datetime` [module](https://docs.python.org/3/library/datetime.html#strftime-and-strptime-behavior) in Python3 the time could be printed with `"%H:%M:%S%"`. Go takes a very different approach by using predefined formats: the same time string would be created with `"15:04:05"`. To create custom strings, we use the `Format` [method](https://golang.org/pkg/time/#Time.Format),

```go
    func main() {
    	date := time.Now()

    	fmtStr := "15:04:05"
    	customStr := date.Format(fmtStr)

    	fmt.Println("The time is", customStr)
    }

    The time is 19:00:35
```

The standard library [defines](https://golang.org/pkg/time/#pkg-constants) a number of common formats such as `RFC3339` which also allow us to work out the individual predefined format elements. These common formats can be used directly with the `Format` method

    ...
    fmt.Println("Or in RFC3339:", date.Format(time.RFC3339))
    ...

    Or in RFC3339: 2019-05-27T19:00:35Z

Parsing times from strings uses the same format strings:

```go
    func main() {
    	timeStr := "2019 05 27 11:23:45"

    	layout := "2006 01 02 15:04:05"
    	date, err := time.Parse(layout, timeStr)
    	if err != nil {
    		fmt.Println("Not able to parse time:", err)
    		return
    	}

    	fmt.Println("date: ", date)
    }

    date:  2019-05-27 11:23:45 +0000 UTC
```

If the date is in a particular time zone without the timezone information in the string itself, there is the `time.ParseInLocation` function which takes an additional `Location` argument which can be created in the same manner as described earlier.


We will explore Unmarshaling and Marshaling custom time formats from/to JSON in a later section.


Time.Duration
---------------------------
The `time.Duration` [type](https://golang.org/pkg/time/#Duration) has a straightforward implementation but makes working with time feel natural. `time.Duration` is simply an 
`int64` number of nanoseconds. The beauty arises from Go's [untyped constants](https://blog.golang.org/constants) which then defines useful intervals as 

```go 
    const (
        Nanosecond  Duration = 1
        Microsecond          = 1000 * Nanosecond
        Millisecond          = 1000 * Microsecond
        Second               = 1000 * Millisecond
        Minute               = 60 * Second
        Hour                 = 60 * Minute
    )
```

which permits code such as the following 

```go
    func sleepFor(duration time.Duration) {
    	for i := 0; i < 10; i++ {
    		fmt.Println("Iteration", i)
    		time.Sleep(duration)
    	}
    }

    func main() {
    	interval := 500 * time.Millisecond

    	sleepFor(interval)
	}
```

Using `time.Duration` as the argument type is much more robust than, say, passing an `int` and documenting that this input should be in milliseconds. The standard library functions work extensively with `time.Duration`. Some examples of using duration:

```go
    func main() {
    	now := time.Now().UTC()
    	future := now.Add(5 * time.Hour)
    	past := now.Add(-11 * time.Minute)

    	fmt.Println("The time now is", now)
    	fmt.Println("In the future  ", future)
    	fmt.Println("In the past    ", past)

    	time.Sleep(3000 * time.Millisecond)

    	elapseLimit := 10 * time.Second
    	if time.Since(now) > elapseLimit {
    		fmt.Println("More than", elapseLimit)
    	} else {
    		fmt.Println("Less than", elapseLimit)
    	}
    }

    The time now is 2019-05-27 19:32:24.663249 +0000 UTC
    In the future   2019-05-28 00:32:24.663249 +0000 UTC
    In the past     2019-05-27 19:21:24.663249 +0000 UTC
    Less than 10s
```

Note that `time.Duration` defines a `String` [method](https://golang.org/pkg/time/#Duration.String) which also prints duration including units.

The slight wart with time.Duration is trying to create duration from variables which are integers. Go does not allow mixed type arithmetic; therefore the following will not compile

```go
    func main() {
	    interval := 5

	    time.Sleep(interval * time.Second)
    }
```

with the error

    ./main.go:8:22: invalid operation: interval * time.Second (mismatched types int and time.Duration)

Instead we need to explicitly convert the variable to a `time.Duration` and then apply the correct scale

```go
    func main() {
    	interval := 5

    	time.Sleep(time.Duration(interval) * time.Second)
    }
```

This explicit conversion feels awkward but, in our experience, is not required all that often and is a small price for the overall utility of the duration type.


(Un)Marshalling Time
----------------------------
We often want to Unmarshal and Marshal times from JSON data. Take the following example

```go
    type tsData struct {
    	Timestamp time.Time `json:"ts"`
    	Value     int       `json:"value"`
    }

    func main() {
    	input := []byte(`{
    		"ts": "2019 05 27 12:52:18",
    		"value": 10
    		}`)

    	var data tsData
    	if err := json.Unmarshal(input, &data); err != nil {
    		fmt.Println("Could not unmarshal data:", err)
    	}

    	fmt.Printf("%+v\n", data)
    }
```

We have defined a type `tsData` to represent times-series data which has a `Timestamp` of type `time.Time` and a `Value` of type `int`. Our simulated json input is defined in the `input` variable which has a `ts` field with a reasonable time format.

Looking back at the documentation we can see that `time.Time` has a [method](https://golang.org/pkg/time/#Time.UnmarshalJSON) `UnmarshalJSON` meaning it implements the `Unmarshaler` [interface](https://golang.org/pkg/encoding/json/#Unmarshaler) and can be used with json.Unmarshal.

 When we come to run the program, however, we exit with the following error message 

    Could not unmarshal data: parsing time ""2019 05 27 12:52:18"" as ""2006-01-02T15:04:05Z07:00"": cannot parse " 05 27 12:52:18"" as "-"

This is indicating to us that `Unmarshal` is expecting a particular time format which is different to that we have given as input. The [documentation](https://golang.org/pkg/time/#Time.UnmarshalJSON) for `time.UnmarshalJSON` specifies "the time is expected to be a quoted string in the RFC 3339 format". We therefore need to create our own type that can handle this different format.

Our custom type is called `timestamp` which simply embeds a `time.Time`. This allows us to use `timestamp` in an almost identical manner to a `time.Time` throughout the rest of our code.

In order to satisfy the `Unmarshaler` interface we define a single method on `timestamp` called `UnmarashalJSON`. Note how we can access the anonymous field using its type in the line `ts.Time = t`.

```go
    const layout = "2006 01 02 15:04:05"

    type timestamp struct {
    	time.Time
    }

    func (ts *timestamp) UnmarshalJSON(b []byte) error {
    	// Convert to string and remove quotes
    	s := strings.Trim(string(b), "\"")

    	// Parse the time using the layout
    	t, err := time.Parse(layout, s)
    	if err != nil {
    		return err
    	}

    	// Assign the parsed time to our variable
    	ts.Time = t
    	return nil
    }

    type tsData struct {
    	Timestamp timestamp `json:"ts"`
    	Value     int       `json:"value"`
    }

    func main() {
    	input := []byte(`{
    		"ts": "2019 05 27 12:52:18",
    		"value": 10
    		}`)

    	var data tsData
    	if err := json.Unmarshal(input, &data); err != nil {
    		fmt.Println("Could not unmarshal data:", err)
    		return
    	}

    	fmt.Printf("%+v\n", data)
        fmt.Println("Month:", data.Timestamp.Month())
    }
```

Now when we run the code we get the desired output:
    {Timestamp:2019-05-27 12:52:18 +0000 UTC Value:10}
    Month: May

Marshalling is achieved by satisfying the `json.Marshaler` [interface](https://golang.org/pkg/encoding/json/#Marshaler). Note the value receiver (ie not `*timestamp`),

```go
    ...
    func (ts timestamp) MarshalJSON() ([]byte, error) {
    	// The +2 is take account of the quotation marks
    	b := make([]byte, 0, len(layout)+2)

    	// Write the json output
    	b = append(b, '"')
    	b = ts.AppendFormat(b, layout)
    	b = append(b, '"')

    	return b, nil
    }
    ...
```

Now in `func main()`:

```go
    ...
    	dataJSON, err := json.Marshal(data)
    	if err != nil {
    		fmt.Println("Could not marshal data:", err)
    		return
    	}

    	fmt.Println(bytes.Equal(input, dataJSON))
    ...

    true
```

once we have removed all the white space in the `input` variable.