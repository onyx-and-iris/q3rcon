package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/chelnak/ysmrr"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/onyx-and-iris/q3rcon"
)

var version string // Version will be set at build time

type Flags struct {
	Host        string
	Port        int
	Rconpass    string
	Interactive bool
	LogLevel    string
	Version     bool
}

func (f Flags) Validate() error {
	if f.Port < 1024 || f.Port > 65535 {
		return fmt.Errorf(
			"invalid port value, got: (%d) expected: in range 1024-65535",
			f.Port,
		)
	}

	if len(f.Rconpass) < 8 {
		return fmt.Errorf(
			"invalid rcon password, got: (%s) expected: at least 8 characters",
			f.Rconpass,
		)
	}

	return nil
}

var reColourCodes = regexp.MustCompile(`\^[0-9]`)

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

type context struct {
	client   *q3rcon.Rcon
	timeouts map[string]time.Duration
	in       io.Reader
	sm       ysmrr.SpinnerManager
}

// run executes the main logic of the application and returns a cleanup function and an error if any.
func run() (func(), error) {
	var flags Flags

	fs := ff.NewFlagSet("q3rcon - A command-line RCON client for Q3 Rcon compatible game servers")
	fs.StringVar(&flags.Host, 'H', "host", "localhost", "hostname of the gameserver")
	fs.IntVar(
		&flags.Port,
		'p',
		"port",
		28960,
		"port on which the gameserver resides, default is 28960",
	)
	fs.StringVar(
		&flags.Rconpass,
		'r',
		"rconpass",
		"",
		"rcon password of the gameserver",
	)

	fs.BoolVar(&flags.Interactive, 'i', "interactive", "run in interactive mode")
	fs.StringVar(
		&flags.LogLevel,
		'l',
		"loglevel",
		"info",
		"Log level (debug, info, warn, error, fatal, panic)",
	)
	fs.BoolVar(&flags.Version, 'v', "version", "print version information and exit")

	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("Q3RCON"),
	)
	switch {
	case errors.Is(err, ff.ErrHelp):
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs, "q3rcon [flags] <rcon commands>"))
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	if flags.Version {
		fmt.Printf("q3rcon version: %s\n", versionFromBuild())
		return nil, nil
	}

	if err := flags.Validate(); err != nil {
		return nil, err
	}

	level, err := log.ParseLevel(flags.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", flags.LogLevel)
	}
	log.SetLevel(level)

	timeouts := map[string]time.Duration{
		"map":         time.Second,
		"map_rotate":  time.Second,
		"map_restart": time.Second,
	}
	log.Debugf("using timeouts: %v", timeouts)

	client, closer, err := connectRcon(flags.Host, flags.Port, flags.Rconpass, timeouts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rcon: %w", err)
	}

	sm := ysmrr.NewSpinnerManager()
	sm.AddSpinner("")

	ctx := &context{
		client:   client,
		timeouts: timeouts,
		in:       os.Stdin,
		sm:       sm,
	}

	if flags.Interactive {
		fmt.Printf("Enter 'Q' to exit.\n>> ")
		err := interactiveMode(ctx)
		if err != nil {
			return closer, fmt.Errorf("interactive mode error: %w", err)
		}
		return closer, nil
	}

	commands := fs.GetArgs()
	if len(commands) == 0 {
		log.Debug("no commands provided, defaulting to 'status'")
		commands = append(commands, "status")
	}
	runCommands(ctx, commands)

	return closer, nil
}

// versionFromBuild retrieves the version information from the build metadata.
func versionFromBuild() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unable to read version)"
	}
	return strings.Split(info.Main.Version, "-")[0]
}

func connectRcon(
	host string,
	port int,
	password string,
	timeouts map[string]time.Duration,
) (*q3rcon.Rcon, func(), error) {
	client, err := q3rcon.New(host, port, password, q3rcon.WithTimeouts(timeouts))
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

// interactiveMode continuously reads from input until a quit signal is given.
func interactiveMode(ctx *context) error {
	scanner := bufio.NewScanner(ctx.in)
	for scanner.Scan() {
		cmd := scanner.Text()
		if strings.EqualFold(cmd, "Q") {
			return nil
		}

		if err := runCommand(ctx, cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Printf(">> ")
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}
	return nil
}

// runCommands executes a list of commands sequentially and prints any errors encountered.
func runCommands(ctx *context, commands []string) {
	for _, cmd := range commands {
		if err := runCommand(ctx, cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// runCommand sends a command to the RCON client and prints the response.
// If the command is in the timeouts map, it starts a spinner until the response is received.
func runCommand(ctx *context, cmd string) error {
	before, _, _ := strings.Cut(cmd, " ")
	_, ok := ctx.timeouts[before]
	if ok {
		ctx.sm.Start()
	}

	resp, err := ctx.client.Send(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %w", cmd, err)
	}
	if ctx.sm.Running() {
		ctx.sm.Stop()
	}
	fmt.Print(reColourCodes.ReplaceAllString(resp, ""))
	return nil
}
