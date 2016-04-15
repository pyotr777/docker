package cli

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

const debug_level int = 1

// Cli represents a command line interface.
type Cli struct {
	Stderr   io.Writer
	handlers []Handler
	Usage    func()
}

// Handler holds the different commands Cli will call
// It should have methods with names starting with `Cmd` like:
// 	func (h myHandler) CmdFoo(args ...string) error
type Handler interface{}

// Initializer can be optionally implemented by a Handler to
// initialize before each call to one of its commands.
type Initializer interface {
	Initialize() error
}

// New instantiates a ready-to-use Cli.
func New(handlers ...Handler) *Cli {
	if debug_level > 0 {
		logrus.Debugf("Executing cli/cli.go : New(%s)", handlers)
	}
	// make the generic Cli object the first cli handler
	// in order to handle `docker help` appropriately
	cli := new(Cli)
	cli.handlers = append([]Handler{cli}, handlers...)
	return cli
}

// initErr is an error returned upon initialization of a handler implementing Initializer.
type initErr struct{ error }

func (err initErr) Error() string {
	return err.Error()
}

func (cli *Cli) command(args ...string) (func(...string) error, error) {
	if debug_level > 0 {
		logrus.Debugf("Executing cli/cli.go : command(%s)", args)
	}
	for _, c := range cli.handlers {
		// cli handlers has 3 elements
		// []cli.Handler{(*cli.Cli)(0xc82037ca50), (*client.DockerCli)(0xc82023fe00), (*main.DaemonCli)(0xc82038c090)}
		if c == nil {
			continue
		}
		// when called with two words from docker command
		// camelArgs will be these two words Capitalised.
		camelArgs := make([]string, len(args))
		for i, s := range args {
			if len(s) == 0 {
				return nil, errors.New("empty command")
			}
			camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
		}
		// sample methodName for merge command
		// "CmdMergeHello-world"
		methodName := "Cmd" + strings.Join(camelArgs, "")
		method := reflect.ValueOf(c).MethodByName(methodName)
		if debug_level > 0 {
			logrus.Debugf("Returning method %s from cli/cli.go:command(%v)", methodName, args)
			if debug_level > 1 {
				logrus.Debug("Stack trace:")
				debug.PrintStack()
			}
		}
		if method.IsValid() {
			if c, ok := c.(Initializer); ok {
				if err := c.Initialize(); err != nil {
					return nil, initErr{err}
				}
			}
			return method.Interface().(func(...string) error), nil
		}
	}
	return nil, errors.New("command not found")
}

// Run executes the specified command.
func (cli *Cli) Run(args ...string) error {
	_ = "breakpoint"
	if debug_level > 0 {
		logrus.Debugf("Executing cli/cli.go : Run(%s)", args)
	}
	if len(args) > 1 {
		command, err := cli.command(args[:2]...)
		switch err := err.(type) {
		case nil:
			return command(args[2:]...)
		case initErr:
			return err.error
		}
	}
	if len(args) > 0 {
		command, err := cli.command(args[0])
		switch err := err.(type) {
		case nil:
			// Function CmdMerge is called here
			return command(args[1:]...)
		case initErr:
			return err.error
		}
		cli.noSuchCommand(args[0])
	}
	return cli.CmdHelp()
}

func (cli *Cli) noSuchCommand(command string) {
	if cli.Stderr == nil {
		cli.Stderr = os.Stderr
	}
	fmt.Fprintf(cli.Stderr, "docker: '%s' is not a docker command.\nSee 'docker --help'.\n", command)
	os.Exit(1)
}

// CmdHelp displays information on a Docker command.
//
// If more than one command is specified, information is only shown for the first command.
//
// Usage: docker help COMMAND or docker COMMAND --help
func (cli *Cli) CmdHelp(args ...string) error {
	if len(args) > 1 {
		command, err := cli.command(args[:2]...)
		switch err := err.(type) {
		case nil:
			command("--help")
			return nil
		case initErr:
			return err.error
		}
	}
	if len(args) > 0 {
		command, err := cli.command(args[0])
		switch err := err.(type) {
		case nil:
			command("--help")
			return nil
		case initErr:
			return err.error
		}
		cli.noSuchCommand(args[0])
	}

	if cli.Usage == nil {
		flag.Usage()
	} else {
		cli.Usage()
	}

	return nil
}

// Subcmd is a subcommand of the main "docker" command.
// A subcommand represents an action that can be performed
// from the Docker command line client.
//
// To see all available subcommands, run "docker --help".
func Subcmd(name string, synopses []string, description string, exitOnError bool) *flag.FlagSet {
	if debug_level > 0 {
		logrus.Debugf("Called Subcmd with name %s", name)
	}
	var errorHandling flag.ErrorHandling
	if exitOnError {
		errorHandling = flag.ExitOnError
	} else {
		errorHandling = flag.ContinueOnError
	}
	flags := flag.NewFlagSet(name, errorHandling)
	flags.Usage = func() {
		flags.ShortUsage()
		flags.PrintDefaults()
	}

	flags.ShortUsage = func() {
		options := ""
		if flags.FlagCountUndeprecated() > 0 {
			options = " [OPTIONS]"
		}

		if len(synopses) == 0 {
			synopses = []string{""}
		}

		// Allow for multiple command usage synopses.
		for i, synopsis := range synopses {
			lead := "\t"
			if i == 0 {
				// First line needs the word 'Usage'.
				lead = "Usage:\t"
			}

			if synopsis != "" {
				synopsis = " " + synopsis
			}

			fmt.Fprintf(flags.Out(), "\n%sdocker %s%s%s", lead, name, options, synopsis)
		}

		fmt.Fprintf(flags.Out(), "\n\n%s\n", description)
	}

	return flags
}

// StatusError reports an unsuccessful exit by a command.
type StatusError struct {
	Status     string
	StatusCode int
}

func (e StatusError) Error() string {
	return fmt.Sprintf("Status: %s, Code: %d", e.Status, e.StatusCode)
}
