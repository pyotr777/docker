package cli

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/mailgun/godebug/lib"
	flag "github.com/docker/docker/pkg/mflag"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

var cli_go_scope = godebug.EnteringNewFile(cli_pkg_scope, cli_go_contents)

const debug_level int = 1

type Cli struct {
	Stderr   io.Writer
	handlers []Handler
	Usage    func()
}

type Handler interface{}

type Initializer interface {
	Initialize() error
}

func New(handlers ...Handler) *Cli {
	var result1 *Cli
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = New(handlers...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("handlers", &handlers)
	godebug.Line(ctx, scope, 37)
	if debug_level > 0 {
		godebug.Line(ctx, scope, 38)
		logrus.Debugf("Executing cli/cli.go : New(%s)", handlers)
	}
	godebug.Line(ctx, scope, 42)

	cli := new(Cli)
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 43)
	cli.handlers = append([]Handler{cli}, handlers...)
	godebug.Line(ctx, scope, 44)
	return cli
}

type initErr struct{ error }

func (err initErr) Error() string {
	var result1 string
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = err.Error()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 51)
	return err.Error()
}

func (cli *Cli) command(args ...string) (func(...string) error, error) {
	var result1 func(...string) error
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = cli.command(args...)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 55)
	if debug_level > 0 {
		godebug.Line(ctx, scope, 56)
		logrus.Debugf("Executing cli/cli.go : command(%s)", args)
	}
	{
		scope := scope.EnteringNewChildScope()
		for _, c := range cli.handlers {
			godebug.Line(ctx, scope, 58)
			scope.Declare("c", &c)
			godebug.Line(ctx, scope, 59)
			if c == nil {
				godebug.Line(ctx, scope, 60)
				continue
			}
			godebug.Line(ctx, scope, 62)
			camelArgs := make([]string, len(args))
			scope := scope.EnteringNewChildScope()
			scope.Declare("camelArgs", &camelArgs)
			{
				scope := scope.EnteringNewChildScope()
				for i, s := range args {
					godebug.Line(ctx, scope, 63)
					scope.Declare("i", &i, "s", &s)
					godebug.Line(ctx, scope, 64)
					if len(s) == 0 {
						godebug.Line(ctx, scope, 65)
						return nil, errors.New("empty command")
					}
					godebug.Line(ctx, scope, 67)
					camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
				}
				godebug.Line(ctx, scope, 63)
			}
			godebug.Line(ctx, scope, 69)
			methodName := "Cmd" + strings.Join(camelArgs, "")
			scope.Declare("methodName", &methodName)
			godebug.Line(ctx, scope, 70)
			method := reflect.ValueOf(c).MethodByName(methodName)
			scope.Declare("method", &method)
			godebug.Line(ctx, scope, 71)
			if debug_level > 0 {
				godebug.Line(ctx, scope, 72)
				logrus.Debugf("Returning method %s from cli/cli.go:command(%v)", methodName, args)
				godebug.Line(ctx, scope, 73)
				if debug_level > 1 {
					godebug.Line(ctx, scope, 74)
					logrus.Debug("Stack trace:")
					godebug.Line(ctx, scope, 75)
					debug.PrintStack()
				}
			}
			godebug.Line(ctx, scope, 78)
			if method.IsValid() {
				godebug.Line(ctx, scope, 79)
				if c, ok := c.(Initializer); ok {
					scope := scope.EnteringNewChildScope()
					scope.Declare("c", &c, "ok", &ok)
					godebug.Line(ctx, scope, 80)
					if err := c.Initialize(); err != nil {
						scope := scope.EnteringNewChildScope()
						scope.Declare("err", &err)
						godebug.Line(ctx, scope, 81)
						return nil, initErr{err}
					}
				}
				godebug.Line(ctx, scope, 84)
				return method.Interface().(func(...string) error), nil
			}
		}
		godebug.Line(ctx, scope, 58)
	}
	godebug.Line(ctx, scope, 87)
	return nil, errors.New("command not found")
}

func (cli *Cli) Run(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.Run(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.SetTraceGen(ctx)
	godebug.Line(ctx, scope, 92)
	godebug.Line(ctx, scope, 93)

	if debug_level > 0 {
		godebug.Line(ctx, scope, 94)
		logrus.Debugf("Executing cli/cli.go : Run(%s)", args)
	}
	godebug.Line(ctx, scope, 96)
	if len(args) > 1 {
		godebug.Line(ctx, scope, 97)
		command, err := cli.command(args[:2]...)
		scope := scope.EnteringNewChildScope()
		scope.Declare("command", &command, "err", &err)
		godebug.Line(ctx, scope, 98)
		switch err := err.(type) {
		case nil:
			godebug.Line(ctx, scope, 99)
			godebug.Line(ctx, scope, 100)
			return command(args[2:]...)
		case initErr:
			godebug.Line(ctx, scope, 101)
			godebug.Line(ctx, scope, 102)
			return err.error
		}
	}
	godebug.Line(ctx, scope, 105)
	if len(args) > 0 {
		godebug.Line(ctx, scope, 106)
		command, err := cli.command(args[0])
		scope := scope.EnteringNewChildScope()
		scope.Declare("command", &command, "err", &err)
		godebug.Line(ctx, scope, 107)
		switch err := err.(type) {
		case nil:
			godebug.Line(ctx, scope, 108)
			godebug.Line(ctx, scope, 109)
			return command(args[1:]...)
		case initErr:
			godebug.Line(ctx, scope, 110)
			godebug.Line(ctx, scope, 111)
			return err.error
		}
		godebug.Line(ctx, scope, 113)
		cli.noSuchCommand(args[0])
	}
	godebug.Line(ctx, scope, 115)
	return cli.CmdHelp()
}

func (cli *Cli) noSuchCommand(command string) {
	ctx, _ok := godebug.EnterFunc(func() {
		cli.noSuchCommand(command)
	})
	if !_ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "command", &command)
	godebug.Line(ctx, scope, 119)
	if cli.Stderr == nil {
		godebug.Line(ctx, scope, 120)
		cli.Stderr = os.Stderr
	}
	godebug.Line(ctx, scope, 122)
	fmt.Fprintf(cli.Stderr, "docker: '%s' is not a docker command.\nSee 'docker --help'.\n", command)
	godebug.Line(ctx, scope, 123)
	os.Exit(1)
}

func (cli *Cli) CmdHelp(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdHelp(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 132)
	if len(args) > 1 {
		godebug.Line(ctx, scope, 133)
		command, err := cli.command(args[:2]...)
		scope := scope.EnteringNewChildScope()
		scope.Declare("command", &command, "err", &err)
		godebug.Line(ctx, scope, 134)
		switch err := err.(type) {
		case nil:
			godebug.Line(ctx, scope, 135)
			godebug.Line(ctx, scope, 136)
			command("--help")
			godebug.Line(ctx, scope, 137)
			return nil
		case initErr:
			godebug.Line(ctx, scope, 138)
			godebug.Line(ctx, scope, 139)
			return err.error
		}
	}
	godebug.Line(ctx, scope, 142)
	if len(args) > 0 {
		godebug.Line(ctx, scope, 143)
		command, err := cli.command(args[0])
		scope := scope.EnteringNewChildScope()
		scope.Declare("command", &command, "err", &err)
		godebug.Line(ctx, scope, 144)
		switch err := err.(type) {
		case nil:
			godebug.Line(ctx, scope, 145)
			godebug.Line(ctx, scope, 146)
			command("--help")
			godebug.Line(ctx, scope, 147)
			return nil
		case initErr:
			godebug.Line(ctx, scope, 148)
			godebug.Line(ctx, scope, 149)
			return err.error
		}
		godebug.Line(ctx, scope, 151)
		cli.noSuchCommand(args[0])
	}
	godebug.Line(ctx, scope, 154)

	if cli.Usage == nil {
		godebug.Line(ctx, scope, 155)
		flag.Usage()
	} else {
		godebug.Line(ctx, scope, 156)
		godebug.Line(ctx, scope, 157)
		cli.Usage()
	}
	godebug.Line(ctx, scope, 160)

	return nil
}

func Subcmd(name string, synopses []string, description string, exitOnError bool) *flag.FlagSet {
	var result1 *flag.FlagSet
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = Subcmd(name, synopses, description, exitOnError)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("name", &name, "synopses", &synopses, "description", &description, "exitOnError", &exitOnError)
	godebug.Line(ctx, scope, 169)
	if debug_level > 0 {
		godebug.Line(ctx, scope, 170)
		logrus.Debugf("Called Subcmd with name %s", name)
	}
	godebug.Line(ctx, scope, 172)
	var errorHandling flag.ErrorHandling
	scope.Declare("errorHandling", &errorHandling)
	godebug.Line(ctx, scope, 173)
	if exitOnError {
		godebug.Line(ctx, scope, 174)
		errorHandling = flag.ExitOnError
	} else {
		godebug.Line(ctx, scope, 175)
		godebug.Line(ctx, scope, 176)
		errorHandling = flag.ContinueOnError
	}
	godebug.Line(ctx, scope, 178)
	flags := flag.NewFlagSet(name, errorHandling)
	scope.Declare("flags", &flags)
	godebug.Line(ctx, scope, 179)
	flags.Usage = func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 180)
			flags.ShortUsage()
			godebug.Line(ctx, scope, 181)
			flags.PrintDefaults()
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	godebug.Line(ctx, scope, 184)

	flags.ShortUsage = func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 185)
			options := ""
			scope := scope.EnteringNewChildScope()
			scope.Declare("options", &options)
			godebug.Line(ctx, scope, 186)
			if flags.FlagCountUndeprecated() > 0 {
				godebug.Line(ctx, scope, 187)
				options = " [OPTIONS]"
			}
			godebug.Line(ctx, scope, 190)
			if len(synopses) == 0 {
				godebug.Line(ctx, scope, 191)
				synopses = []string{""}
			}
			{
				scope := scope.EnteringNewChildScope()
				for i, synopsis := range synopses {
					godebug.Line(ctx, scope, 195)
					scope.Declare("i", &i, "synopsis", &synopsis)
					godebug.Line(ctx, scope, 196)
					lead := "\t"
					scope := scope.EnteringNewChildScope()
					scope.Declare("lead", &lead)
					godebug.Line(ctx, scope, 197)
					if i == 0 {
						godebug.Line(ctx, scope, 199)
						lead = "Usage:\t"
					}
					godebug.Line(ctx, scope, 202)
					if synopsis != "" {
						godebug.Line(ctx, scope, 203)
						synopsis = " " + synopsis
					}
					godebug.Line(ctx, scope, 206)
					fmt.Fprintf(flags.Out(), "\n%sdocker %s%s%s", lead, name, options, synopsis)
				}
				godebug.Line(ctx, scope, 195)
			}
			godebug.Line(ctx, scope, 209)
			fmt.Fprintf(flags.Out(), "\n\n%s\n", description)
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	godebug.Line(ctx, scope, 212)

	return flags
}

type StatusError struct {
	Status     string
	StatusCode int
}

func (e StatusError) Error() string {
	var result1 string
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = e.Error()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("e", &e)
	godebug.Line(ctx, scope, 222)
	return fmt.Sprintf("Status: %s, Code: %d", e.Status, e.StatusCode)
}

var cli_go_contents = `package cli

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
// It should have methods with names starting with ` + "`" + `Cmd` + "`" + ` like:
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
	// in order to handle ` + "`" + `docker help` + "`" + ` appropriately
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
		if c == nil {
			continue
		}
		camelArgs := make([]string, len(args))
		for i, s := range args {
			if len(s) == 0 {
				return nil, errors.New("empty command")
			}
			camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
		}
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
`


var cli_pkg_scope = &godebug.Scope{}

func init() {
	cli_pkg_scope.Vars = map[string]interface{}{
		"dockerCommands": &dockerCommands,
		"DockerCommands": &DockerCommands,
	}
	cli_pkg_scope.Consts = map[string]interface{}{
		"debug_level": debug_level,
	}
	cli_pkg_scope.Funcs = map[string]interface{}{
		"New": New,
		"Subcmd": Subcmd,
	}
}