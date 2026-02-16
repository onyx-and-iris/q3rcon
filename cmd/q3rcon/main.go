package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/onyx-and-iris/q3rcon"
)

var version string // Version will be set at build time

func main() {
	var exitCode int

	// Defer exit with the final exit code
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	closer, err := run()
	if closer != nil {
		defer closer()
	}
	if err != nil {
		log.Error(err)
		exitCode = 1
	}
}

// run executes the main logic of the application and returns a cleanup function and an error if any.
func run() (func(), error) {
	var (
		host        string
		port        int
		rconpass    string
		interactive bool
		loglevel    string
		versionFlag bool
	)

	flag.StringVar(&host, "host", "localhost", "hostname of the gameserver")
	flag.StringVar(&host, "h", "localhost", "hostname of the gameserver (shorthand)")
	flag.IntVar(&port, "port", 28960, "port on which the gameserver resides, default is 28960")
	flag.IntVar(
		&port,
		"p",
		28960,
		"port on which the gameserver resides, default is 28960 (shorthand)",
	)
	flag.StringVar(&rconpass, "rconpass", os.Getenv("RCON_PASS"), "rcon password of the gameserver")
	flag.StringVar(
		&rconpass,
		"r",
		os.Getenv("RCON_PASS"),
		"rcon password of the gameserver (shorthand)",
	)

	flag.BoolVar(&interactive, "interactive", false, "run in interactive mode")
	flag.BoolVar(&interactive, "i", false, "run in interactive mode")

	flag.StringVar(&loglevel, "loglevel", "warn", "log level")
	flag.StringVar(&loglevel, "l", "warn", "log level (shorthand)")

	flag.BoolVar(&versionFlag, "version", false, "print version information and exit")
	flag.BoolVar(&versionFlag, "v", false, "print version information and exit (shorthand)")

	flag.Parse()

	if versionFlag {
		fmt.Printf("q3rcon version: %s\n", versionFromBuild())
		return nil, nil
	}

	level, err := log.ParseLevel(loglevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", loglevel)
	}
	log.SetLevel(level)

	if port < 1024 || port > 65535 {
		return nil, fmt.Errorf("invalid port value, got: (%d) expected: in range 1024-65535", port)
	}

	if len(rconpass) < 8 {
		return nil, fmt.Errorf(
			"invalid rcon password, got: (%s) expected: at least 8 characters",
			rconpass,
		)
	}

	client, closer, err := connectRcon(host, port, rconpass)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rcon: %w", err)
	}

	if interactive {
		fmt.Printf("Enter 'Q' to exit.\n>> ")
		err := interactiveMode(client, os.Stdin)
		if err != nil {
			return closer, fmt.Errorf("interactive mode error: %w", err)
		}
		return closer, nil
	}

	commands := flag.Args()
	if len(commands) == 0 {
		log.Debug("no commands provided, defaulting to 'status'")
		commands = append(commands, "status")
	}
	runCommands(client, commands)

	return closer, nil
}

// versionFromBuild retrieves the version information from the build metadata.
func versionFromBuild() string {
	if version == "" {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "(unable to read build info)"
		}
		version = strings.Split(info.Main.Version, "-")[0]
	}
	return version
}

func connectRcon(host string, port int, password string) (*q3rcon.Rcon, func(), error) {
	client, err := q3rcon.New(host, port, password, q3rcon.WithTimeouts(map[string]time.Duration{
		"map":         time.Second,
		"map_rotate":  time.Second,
		"map_restart": time.Second,
	}))
	if err != nil {
		return nil, nil, err
	}

	closer := func() {
		if err := client.Close(); err != nil {
			log.Error(err)
		}
	}

	return client, closer, nil
}

// runCommands runs the commands given in the flag.Args slice.
// If no commands are given, it defaults to running the "status" command.
func runCommands(client *q3rcon.Rcon, commands []string) {
	for _, cmd := range commands {
		resp, err := client.Send(cmd)
		if err != nil {
			log.Error(err)
			continue
		}
		fmt.Print(resp)
	}
}

// interactiveMode continuously reads from input until a quit signal is given.
func interactiveMode(client *q3rcon.Rcon, input io.Reader) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		cmd := scanner.Text()
		if strings.EqualFold(cmd, "Q") {
			return nil
		}

		resp, err := client.Send(cmd)
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
