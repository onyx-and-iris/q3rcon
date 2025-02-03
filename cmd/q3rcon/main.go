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

func exitOnError(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}

func main() {
	var (
		host        string
		port        int
		rconpass    string
		interactive bool
		loglevel    int
	)

	flag.StringVar(&host, "host", "localhost", "hostname of the gameserver")
	flag.StringVar(&host, "h", "localhost", "hostname of the gameserver (shorthand)")
	flag.IntVar(&port, "port", 28960, "port on which the gameserver resides, default is 28960")
	flag.IntVar(&port, "p", 28960, "port on which the gameserver resides, default is 28960 (shorthand)")
	flag.StringVar(&rconpass, "rconpass", os.Getenv("RCON_PASS"), "rcon password of the gameserver")
	flag.StringVar(&rconpass, "r", os.Getenv("RCON_PASS"), "rcon password of the gameserver (shorthand)")

	flag.BoolVar(&interactive, "interactive", false, "run in interactive mode")
	flag.BoolVar(&interactive, "i", false, "run in interactive mode")

	flag.IntVar(&loglevel, "loglevel", int(log.WarnLevel), "log level")
	flag.IntVar(&loglevel, "l", int(log.WarnLevel), "log level (shorthand)")
	flag.Parse()

	if slices.Contains(log.AllLevels, log.Level(loglevel)) {
		log.SetLevel(log.Level(loglevel))
	}

	if port < 1024 || port > 65535 {
		exitOnError(fmt.Errorf("invalid port value, got: (%d) expected: in range 1024-65535", port))
	}

	if len(rconpass) < 8 {
		exitOnError(fmt.Errorf("invalid rcon password, got: (%s) expected: at least 8 characters", rconpass))
	}

	rcon, err := connectRcon(host, port, rconpass)
	if err != nil {
		exitOnError(err)
	}
	defer rcon.Close()

	if !interactive {
		runCommands(flag.Args(), rcon)
		return
	}

	fmt.Printf("Enter 'Q' to exit.\n>> ")
	err = interactiveMode(rcon, os.Stdin)
	if err != nil {
		exitOnError(err)
	}
}

func connectRcon(host string, port int, password string) (*q3rcon.Rcon, error) {
	rcon, err := q3rcon.New(host, port, password)
	if err != nil {
		return nil, err
	}
	return rcon, nil
}

// runCommands runs the commands given in the flag.Args slice.
// If no commands are given, it defaults to running the "status" command.
func runCommands(commands []string, rcon *q3rcon.Rcon) {
	if len(commands) == 0 {
		commands = append(commands, "status")
	}

	for _, cmd := range commands {
		resp, err := rcon.Send(cmd)
		if err != nil {
			log.Error(err)
			continue
		}
		fmt.Print(resp)
	}
}

// interactiveMode continuously reads from input until a quit signal is given.
func interactiveMode(rcon *q3rcon.Rcon, input io.Reader) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		cmd := scanner.Text()
		if strings.EqualFold(cmd, "Q") {
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
