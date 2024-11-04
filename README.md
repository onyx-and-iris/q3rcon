![Windows](https://img.shields.io/badge/Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)

# Q3Rcon

Send rcon commands to Q3 compatible servers.

For an outline of past/future changes refer to: [CHANGELOG](CHANGELOG.md)

## Requirements

-   The game must implement RCON using the Q3 protocol.

## Background

Quake3 Rcon works by firing UDP packets to the game server port, responses may be returned at once or in fragments (depending on the size of the response). For this reason I've made this package quite flexible, timeouts for responses can be set by request kind using a timeouts map. The default timeout for a response is 20ms, although this can be changed as well.

Rcon itself is insecure and each packet includes the password so I don't suggest using it remotely. If you have direct access to the server then SSH in first, then use this tool locally.

## Use

`go get github.com/onyx-and-iris/q3rcon`

```go
package main

import (
	"fmt"
	"log"

	"github.com/onyx-and-iris/q3rcon"
)

func main() {
	var (
		host     string = "localhost"
		port     int    = 30000
		password string = "rconpassword"
	)

	rcon, err := q3rcon.New(host, port, password)
	if err != nil {
		log.Fatal(err)
	}
	defer rcon.Close()

	resp, err := rcon.Send("mapname")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
```

#### `WithDefaultTimeout(timeout time.Duration)`

You may want to change the default timeout if some of your responses are getting mixed together (stability can depend on connection to the server). For example, on LAN I can leave the default at 20ms, when connecting remotely I normally increase this to 50ms.

For example:

```go
	rcon, err := q3rcon.New(
		host, port, password,
		q3rcon.WithDefaultTimeout(50*time.Millisecond))
```

#### `WithTimeouts(timeouts map[string]time.Duration)`

Perhaps there are some requests that take a long time to receive a response but you want the whole response in one chunk. Pass a timeouts map, for example:

```go
	timeouts := map[string]time.Duration{
		"map_rotate":  1200 * time.Millisecond,
		"map_restart": 1200 * time.Millisecond,
	}

	rcon, err := q3rcon.New(
		host, port, password,
		q3rcon.WithTimeouts(timeouts))
```

## Command line

Pass `host`, `port` and `password` as flags, for example:

```
q3rcon -h=localhost -p=30000 -P="rconpassword" "mapname"
```

-   `host` defaults to "localhost"
-   `port` defaults to 28960
-   `password` defaults to ""

Arguments following the flags will be sent as rcon commands. You may send multiple arguments.

#### Interactive mode

Pass `interactive (-i shorthand)` flag to enable interactive mode, for example:

```
q3rcon -h=localhost -p=30000 -P="rconpassword" -i
```

If interactive mode is enabled, any arguments sent on the command line will be ignored.

## Your own implementation

I've separated the q3rcon package from the CLI precisely so that you can write your own implementation, since I don't know the target game or it's commands. For example, in my case I've added terminal colouring and reformatted some of the command responses.

## `Logging`

Log level may be set by passing the `-l` flag with a number from 0 up to 6 where

0 = Panic, 1 = Fatal, 2 = Error, 3 = Warning, 4 = Info, 5 = Debug, 6 = Trace
