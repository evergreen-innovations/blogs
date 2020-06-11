# Go code to simulate IoT devices

## Overview
Most of the time developers need to start developing and testing the code
before they have access to the actual hardware. We at Evergreen Innovations do
this a lot and have found that a combination of Go and Docker provides a great
means creating simulating particular bits of hardware. This approach allows for
development of interfaces to the hardware as well as the creation of specific
scenarios in the simlulator to both mimic real-world behaviour and error
states. In the case of the errors, the any controlling software can be developed to properly handle these scenarios.

All the code from this blog is [available](https://github.com/evergreen-innovations/blogs). We always want to post full
examples so we have included some additional description of the code structure
towards the end of this blog.

In this blog we are going to take the example of a power meter which reports
instaneous data such as frequency alongside the 3-phases of voltage and current.
An example of such an application could be a battery connected to a domestic solar
panel. The pupose of the blog is to demostrate the communications between such a
device and a supervisor over the Modbus protocol. In our case, the Modbus
communication will be over TCP.

Modbus is a messaging protocol by establishing a client-server communication and
has been widely adopted by IoT devices. You can find more information at
[modbus.org](http://modbus.org/). The specifics of of reading and writing values
from the modbus registers are not covered in this blog post. Let us know if you'd
like us to write a blog on this!

## Architecture
In order to provide a simplied interface for this demonstration a `modbus` Go package has been created to wrap
the excellent [modbus server](https://github.com/goburrow/modbus) and [modbus client](https://github.com/tbrandon/mbserver)
to encompass both the client and server functionality in a single package.

Data is transferred over the Modbus protocol by writing to and reading from registers. The documentation for the device will specify the value stored at a particular address,
for example our power meter stores the frequency at address 16384.

The power meter will act as the server, writing values to registers and the supervisor
will act as the client, reading the values from the power meter. In case we will be writing
the code for both sides of the interface though often in practice values could be read from
a device (or similarly written to supervisor) over which you have no control.

## The power meter
The code for the simulated power meter can be found in the "powermeter" folder of the repository,
and is set up as a standard Go module-enabled project.

The first step in the project is to define the Modbus address for the various values exposed
by this powermeter. These values are the frequency, three-phase voltage, and three-phase current,
given by:

```go
const (
	FrequencyAddr uint16 = 16384
	PhaseV1Addr   uint16 = 16386
	PhaseV2Addr   uint16 = 16388
	PhaseV3Addr   uint16 = 16390
	CurrentI1Addr uint16 = 16402
	CurrentI2Addr uint16 = 16404
	CurrentI3Addr uint16 = 16406
)
```

To allow easy interation over these addresses, they are packed into a slice alongside a human
readable name for convenience,

```go
var registers = []Register{
	{"Frequency", FrequencyAddr},
	{"PhaseV1", PhaseV1Addr},
	{"PhaseV2", PhaseV2Addr},
	{"PhaseV3", PhaseV3Addr},
	{"CurrentI1", CurrentI1Addr},
	{"CurrentI2", CurrentI2Addr},
	{"CurrentI3", CurrentI3Addr},
}
```
Given that our Modbus communication is going to be over TCP, commandline arguments for the host
and the port are provided using the [flag](https://golang.org/pkg/flag/) package, with default
values provided

```go
host := flag.String("host", defaultHost, "host for the modbus server")
port := flag.String("port", defaultPort, "port for the modbus server")
flag.Parse()
```

which then allows us to create our Modbus server:

```go
addr := fmt.Sprintf("%s%s", *host, *port)
s, err := modbus.NewServer(addr)
if err != nil {
	mainErr = fmt.Errorf("creating server: %v", err)
	return
}
defer s.Close()
```

There is more detail on the error handling in the main function later in this blog.

Finally, each of the registers is written to within a timed loop with a random value which we
can them observe from the supervisor, described below.

```go
rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
ticker := time.NewTicker(500 * time.Millisecond)
for range ticker.C {
	// Loop over the register address values from map and write the values
	for _, r := range registers {
        value := uint16(rnd.Int())
        fmt.Printf("writing to %v[%v] value: %v\n", r.Name, r.Address, value)
		s.WriteRegister(r.Address, value)
	}
}
```

The output of the program (using `go run main.go`) is then:

```
Modbus server for power meter running at address 0.0.0.0:1503
writing to Frequency[16384] value: 47927
writing to PhaseV1[16386] value: 3661
writing to PhaseV2[16388] value: 8259
writing to PhaseV3[16390] value: 47553
writing to CurrentI1[16402] value: 31286
writing to CurrentI2[16404] value: 6672
writing to CurrentI3[16406] value: 63385
...
```

## The supervisor
The code structure for the supervisor is similar to that of the power meter, and
must have identical Modbus register definitions. In the supervisor, however, we
create a client rather than a server and use the ip address of the power meter to
establish a connection.

```go
// Set up the commandline options
host := flag.String("host", defaultHost, "host for the modbus listener")
port := flag.String("port", defaultPort, "port for the modbus listener")
flag.Parse()

// Start a listener modbus client
addr := fmt.Sprintf("%s%s", *host, *port)
c, err := modbus.NewClient(addr)
if err != nil {
	mainErr = fmt.Errorf("error creating client: %v", err)
	return
}
defer c.Close()
```

The client can then be used to read the registers that the power meter is writing to:

```go
for range ticker.C {
	// Loop over the register address values from map and read the values
	for _, r := range registers {
		v, err := c.ReadRegister(r.Address)
		if err != nil {
			fmt.Printf("error reading %v[%v]: %v\n", r.Name, r.Address, err)
			continue
		}
		fmt.Printf("read %v[%v]: %v\n", r.Name, r.Address, v)
	}
}
```

To see the process in action open up two terminal windows. In the first open up the directory
for the power meter and the second that of the supervisor. Starting with the power meter, issue
the command `go run main.go` in both terminal windows and observe the output. Your output will
be slightly different (due to using random numbers as the value) but you will see blocks such as

```
writing to Frequency[16384] value: 61325
writing to PhaseV1[16386] value: 14234
writing to PhaseV2[16388] value: 48279
writing to PhaseV3[16390] value: 12937
writing to CurrentI1[16402] value: 43749
writing to CurrentI2[16404] value: 9852
writing to CurrentI3[16406] value: 35399
```

which matches output in the supervisor:

```
read Frequency[16384]: 61325
read PhaseV1[16386]: 14234
read PhaseV2[16388]: 48279
read PhaseV3[16390]: 12937
read CurrentI1[16402]: 43749
read CurrentI2[16404]: 9852
read CurrentI3[16406]: 35399
```

We have therefore communicated over Modbus!

## Docker integration
As outlined in the first [blog](https://www.evergreeninnovations.co/blog-elk-stack-in-docker/) the aim of this series is to create a complete IoT system
for local development, and is most easily acheived using docker containers. In the directories for
both the power meter and the supervisor there is `Dockerfile` to build the container. Both these
files have a similar structure and make use of a two-stage build to minimise the final container size
(~3MB rather than ~800MB).

To run the services together we can make sure of [docker-compose](https://docs.docker.com/compose/).
In the `docker-compose.yml` file we specify the servies that we want to run, in this case the
powermeter and the supervisor. For the supervisior, we specify the commandline arguments in the
`command` tag to specify the host for the Modbus connection - notice we can use 'powermeter' as the host
which docker will resolve into the ip address of the container associated with the 'powermeter' service.

The images can then be conveniently built by issuing the following command in the same directory as the
`docker-compose.yml`:

```bash
docker compose build
```

Then to run the power meter and the simulator together:

```bash
docker-compose up -d
```

and to view the logs:

```bash
docker-compose logs -f
```

You should see output similar to

```
powermeter_1  | writing to Frequency[16384] value: 60200
powermeter_1  | writing to PhaseV1[16386] value: 45665
powermeter_1  | writing to PhaseV2[16388] value: 16311
powermeter_1  | writing to PhaseV3[16390] value: 36347
powermeter_1  | writing to CurrentI1[16402] value: 44515
powermeter_1  | writing to CurrentI2[16404] value: 14367
powermeter_1  | writing to CurrentI3[16406] value: 54751
supervisor_1  | read Frequency[16384]: 60200
supervisor_1  | read PhaseV1[16386]: 45665
supervisor_1  | read PhaseV2[16388]: 16311
supervisor_1  | read PhaseV3[16390]: 36347
supervisor_1  | read CurrentI1[16402]: 44515
supervisor_1  | read CurrentI2[16404]: 14367
supervisor_1  | read CurrentI3[16406]: 54751
```

With this in place, we are ready to integrate these iot devices into the larger project outlined in
the previous [blog](https://www.evergreeninnovations.co/blog-elk-stack-in-docker/).


## Conclusion
We hope this blog was useful for you. Please stay tuned for the next part in the series where we continue to build our IoT framework and
do let us know if there are any other subject that would you like to know about.