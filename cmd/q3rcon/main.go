package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/onyx-and-iris/q3rcon"

	log "github.com/sirupsen/logrus"
)

var interactive bool

func main() {
	var (
		host     string
		port     int
		password string
		loglevel int
	)

	flag.StringVar(&host, "host", "localhost", "hostname of the server")
	flag.StringVar(&host, "h", "localhost", "hostname of the server (shorthand)")
	flag.IntVar(&port, "port", 28960, "port of the server")
	flag.IntVar(&port, "p", 28960, "port of the server (shorthand)")
	flag.StringVar(&password, "password", "", "hostname of the server")
	flag.StringVar(&password, "P", "", "hostname of the server (shorthand)")

	flag.BoolVar(&interactive, "interactive", false, "run in interactive mode")
	flag.BoolVar(&interactive, "i", false, "run in interactive mode")

	flag.IntVar(&loglevel, "loglevel", int(log.WarnLevel), "log level")
	flag.IntVar(&loglevel, "l", int(log.WarnLevel), "log level (shorthand)")
	flag.Parse()

	if slices.Contains(log.AllLevels, log.Level(loglevel)) {
		log.SetLevel(log.Level(loglevel))
	}

	rcon, err := connectRcon(host, port, password)
	if err != nil {
		log.Fatal(err)
	}
	defer rcon.Close()

	if interactive {
		fmt.Printf("Enter 'Q' to exit.\n>> ")
		err := interactiveMode(rcon, os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	for _, arg := range flag.Args() {
		resp, err := rcon.Send(arg)
		if err != nil {
			log.Error(err)
			continue
		}
		fmt.Print(resp)
	}
}

func connectRcon(host string, port int, password string) (*q3rcon.Rcon, error) {
	rcon, err := q3rcon.New(
		host, port, password)
	if err != nil {
		return nil, err
	}
	return rcon, nil
}

// interactiveMode continuously reads from input until a quit signal is given.
func interactiveMode(rcon *q3rcon.Rcon, input io.Reader) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		cmd := scanner.Text()
		if strings.ToUpper(cmd) == "Q" {
			return nil
		}

		resp, err := rcon.Send(cmd)
		if err != nil {
			log.Error(err)
			continue
		}
		fmt.Printf("%s>> ", resp)
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	return nil
}
