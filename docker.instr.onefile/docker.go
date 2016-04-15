package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/client"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/utils"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/mailgun/godebug/lib"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var client_go_scope = godebug.EnteringNewFile(main_pkg_scope, client_go_contents)

var clientFlags = &cli.ClientFlags{FlagSet: new(flag.FlagSet), Common: commonFlags}

func init() {
	client := clientFlags.FlagSet
	client.StringVar(&clientFlags.ConfigDir, []string{"-config"}, cliconfig.ConfigDir(), "Location of client config files")

	clientFlags.PostParse = func() {
		clientFlags.Common.PostParse()

		if clientFlags.ConfigDir != "" {
			cliconfig.SetConfigDir(clientFlags.ConfigDir)
		}

		if clientFlags.Common.TrustKey == "" {
			clientFlags.Common.TrustKey = filepath.Join(cliconfig.ConfigDir(), defaultTrustKeyFile)
		}

		if clientFlags.Common.Debug {
			utils.EnableDebug()
		}
	}
}

var client_go_contents = `package main

import (
	"path/filepath"

	"github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
)

var clientFlags = &cli.ClientFlags{FlagSet: new(flag.FlagSet), Common: commonFlags}

func init() {
	client := clientFlags.FlagSet
	client.StringVar(&clientFlags.ConfigDir, []string{"-config"}, cliconfig.ConfigDir(), "Location of client config files")

	clientFlags.PostParse = func() {
		clientFlags.Common.PostParse()

		if clientFlags.ConfigDir != "" {
			cliconfig.SetConfigDir(clientFlags.ConfigDir)
		}

		if clientFlags.Common.TrustKey == "" {
			clientFlags.Common.TrustKey = filepath.Join(cliconfig.ConfigDir(), defaultTrustKeyFile)
		}

		if clientFlags.Common.Debug {
			utils.EnableDebug()
		}
	}
}
`

var main_pkg_scope = &godebug.Scope{}

func init() {
	main_pkg_scope.Vars = map[string]interface{}{
		"clientFlags":     &clientFlags,
		"commonFlags":     &commonFlags,
		"dockerCertPath":  &dockerCertPath,
		"dockerTLSVerify": &dockerTLSVerify,
		"daemonCli":       &daemonCli,
		"flHelp":          &flHelp,
		"flVersion":       &flVersion,
		"dockerCommands":  &dockerCommands,
	}
	main_pkg_scope.Consts = map[string]interface{}{
		"defaultTrustKeyFile": defaultTrustKeyFile,
		"defaultCaFile":       defaultCaFile,
		"defaultKeyFile":      defaultKeyFile,
		"defaultCertFile":     defaultCertFile,
		"tlsVerifyKey":        tlsVerifyKey,
		"daemonUsage":         daemonUsage,
	}
	main_pkg_scope.Funcs = map[string]interface{}{
		"postParseCommon":                postParseCommon,
		"setDaemonLogLevel":              setDaemonLogLevel,
		"main":                           main,
		"showVersion":                    showVersion,
		"TestClientDebugEnabled":         TestClientDebugEnabled,
		"TestDockerSubcommandsAreSorted": TestDockerSubcommandsAreSorted,
	}
}

var common_go_scope = godebug.EnteringNewFile(main_pkg_scope, common_go_contents)

const (
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
	tlsVerifyKey        = "tlsverify"
)

var (
	commonFlags = &cli.CommonFlags{FlagSet: new(flag.FlagSet)}

	dockerCertPath  = os.Getenv("DOCKER_CERT_PATH")
	dockerTLSVerify = os.Getenv("DOCKER_TLS_VERIFY") != ""
)

func init() {
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.ConfigDir()
	}

	commonFlags.PostParse = postParseCommon

	cmd := commonFlags.FlagSet

	cmd.BoolVar(&commonFlags.Debug, []string{"D", "-debug"}, false, "Enable debug mode")
	cmd.StringVar(&commonFlags.LogLevel, []string{"l", "-log-level"}, "info", "Set the logging level")
	cmd.BoolVar(&commonFlags.TLS, []string{"-tls"}, false, "Use TLS; implied by --tlsverify")
	cmd.BoolVar(&commonFlags.TLSVerify, []string{"-tlsverify"}, dockerTLSVerify, "Use TLS and verify the remote")

	var tlsOptions tlsconfig.Options
	commonFlags.TLSOptions = &tlsOptions
	cmd.StringVar(&tlsOptions.CAFile, []string{"-tlscacert"}, filepath.Join(dockerCertPath, defaultCaFile), "Trust certs signed only by this CA")
	cmd.StringVar(&tlsOptions.CertFile, []string{"-tlscert"}, filepath.Join(dockerCertPath, defaultCertFile), "Path to TLS certificate file")
	cmd.StringVar(&tlsOptions.KeyFile, []string{"-tlskey"}, filepath.Join(dockerCertPath, defaultKeyFile), "Path to TLS key file")

	cmd.Var(opts.NewNamedListOptsRef("hosts", &commonFlags.Hosts, opts.ValidateHost), []string{"H", "-host"}, "Daemon socket(s) to connect to")
}

func postParseCommon() {
	ctx, ok := godebug.EnterFunc(postParseCommon)
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	godebug.Line(ctx, common_go_scope, 57)
	cmd := commonFlags.FlagSet
	scope := common_go_scope.EnteringNewChildScope()
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 59)

	setDaemonLogLevel(commonFlags.LogLevel)
	godebug.Line(ctx, scope, 64)

	if cmd.IsSet("-"+tlsVerifyKey) || commonFlags.TLSVerify {
		godebug.Line(ctx, scope, 65)
		commonFlags.TLS = true
	}
	godebug.Line(ctx, scope, 68)

	if !commonFlags.TLS {
		godebug.Line(ctx, scope, 69)
		commonFlags.TLSOptions = nil
	} else {
		godebug.Line(ctx, scope, 70)
		godebug.Line(ctx, scope, 71)
		tlsOptions := commonFlags.TLSOptions
		scope := scope.EnteringNewChildScope()
		scope.Declare("tlsOptions", &tlsOptions)
		godebug.Line(ctx, scope, 72)
		tlsOptions.InsecureSkipVerify = !commonFlags.TLSVerify
		godebug.Line(ctx, scope, 76)

		if !cmd.IsSet("-tlscert") {
			godebug.Line(ctx, scope, 77)
			if _, err := os.Stat(tlsOptions.CertFile); os.IsNotExist(err) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 78)
				tlsOptions.CertFile = ""
			}
		}
		godebug.Line(ctx, scope, 81)
		if !cmd.IsSet("-tlskey") {
			godebug.Line(ctx, scope, 82)
			if _, err := os.Stat(tlsOptions.KeyFile); os.IsNotExist(err) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 83)
				tlsOptions.KeyFile = ""
			}
		}
	}
}

func setDaemonLogLevel(logLevel string) {
	ctx, ok := godebug.EnterFunc(func() {
		setDaemonLogLevel(logLevel)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := common_go_scope.EnteringNewChildScope()
	scope.Declare("logLevel", &logLevel)
	godebug.Line(ctx, scope, 90)
	if logLevel != "" {
		godebug.Line(ctx, scope, 91)
		lvl, err := logrus.ParseLevel(logLevel)
		scope := scope.EnteringNewChildScope()
		scope.Declare("lvl", &lvl, "err", &err)
		godebug.Line(ctx, scope, 92)
		if err != nil {
			godebug.Line(ctx, scope, 93)
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", logLevel)
			godebug.Line(ctx, scope, 94)
			os.Exit(1)
		}
		godebug.Line(ctx, scope, 96)
		logrus.SetLevel(lvl)
	} else {
		godebug.Line(ctx, scope, 97)
		godebug.Line(ctx, scope, 98)
		logrus.SetLevel(logrus.InfoLevel)
	}
}

var common_go_contents = `package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
	tlsVerifyKey        = "tlsverify"
)

var (
	commonFlags = &cli.CommonFlags{FlagSet: new(flag.FlagSet)}

	dockerCertPath  = os.Getenv("DOCKER_CERT_PATH")
	dockerTLSVerify = os.Getenv("DOCKER_TLS_VERIFY") != ""
)

func init() {
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.ConfigDir()
	}

	commonFlags.PostParse = postParseCommon

	cmd := commonFlags.FlagSet

	cmd.BoolVar(&commonFlags.Debug, []string{"D", "-debug"}, false, "Enable debug mode")
	cmd.StringVar(&commonFlags.LogLevel, []string{"l", "-log-level"}, "info", "Set the logging level")
	cmd.BoolVar(&commonFlags.TLS, []string{"-tls"}, false, "Use TLS; implied by --tlsverify")
	cmd.BoolVar(&commonFlags.TLSVerify, []string{"-tlsverify"}, dockerTLSVerify, "Use TLS and verify the remote")

	// TODO use flag flag.String([]string{"i", "-identity"}, "", "Path to libtrust key file")

	var tlsOptions tlsconfig.Options
	commonFlags.TLSOptions = &tlsOptions
	cmd.StringVar(&tlsOptions.CAFile, []string{"-tlscacert"}, filepath.Join(dockerCertPath, defaultCaFile), "Trust certs signed only by this CA")
	cmd.StringVar(&tlsOptions.CertFile, []string{"-tlscert"}, filepath.Join(dockerCertPath, defaultCertFile), "Path to TLS certificate file")
	cmd.StringVar(&tlsOptions.KeyFile, []string{"-tlskey"}, filepath.Join(dockerCertPath, defaultKeyFile), "Path to TLS key file")

	cmd.Var(opts.NewNamedListOptsRef("hosts", &commonFlags.Hosts, opts.ValidateHost), []string{"H", "-host"}, "Daemon socket(s) to connect to")
}

func postParseCommon() {
	cmd := commonFlags.FlagSet

	setDaemonLogLevel(commonFlags.LogLevel)

	// Regardless of whether the user sets it to true or false, if they
	// specify --tlsverify at all then we need to turn on tls
	// TLSVerify can be true even if not set due to DOCKER_TLS_VERIFY env var, so we need to check that here as well
	if cmd.IsSet("-"+tlsVerifyKey) || commonFlags.TLSVerify {
		commonFlags.TLS = true
	}

	if !commonFlags.TLS {
		commonFlags.TLSOptions = nil
	} else {
		tlsOptions := commonFlags.TLSOptions
		tlsOptions.InsecureSkipVerify = !commonFlags.TLSVerify

		// Reset CertFile and KeyFile to empty string if the user did not specify
		// the respective flags and the respective default files were not found.
		if !cmd.IsSet("-tlscert") {
			if _, err := os.Stat(tlsOptions.CertFile); os.IsNotExist(err) {
				tlsOptions.CertFile = ""
			}
		}
		if !cmd.IsSet("-tlskey") {
			if _, err := os.Stat(tlsOptions.KeyFile); os.IsNotExist(err) {
				tlsOptions.KeyFile = ""
			}
		}
	}
}

func setDaemonLogLevel(logLevel string) {
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", logLevel)
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
`

var daemon_none_go_scope = godebug.EnteringNewFile(main_pkg_scope, daemon_none_go_contents)

const daemonUsage = ""

var daemonCli cli.Handler

var daemon_none_go_contents = `// +build !daemon

package main

import "github.com/docker/docker/cli"

const daemonUsage = ""

var daemonCli cli.Handler
`

var docker_go_scope = godebug.EnteringNewFile(main_pkg_scope, docker_go_contents)

func main() {
	ctx, _ok := godebug.EnterFunc(main)
	if !_ok {
		return
	}
	godebug.Line(ctx, docker_go_scope, 18)
	if reexec.Init() {
		godebug.Line(ctx, docker_go_scope, 19)
		return
	}
	godebug.Line(ctx, docker_go_scope, 23)

	stdin, stdout, stderr := term.StdStreams()
	scope := docker_go_scope.EnteringNewChildScope()
	scope.Declare("stdin", &stdin, "stdout", &stdout, "stderr", &stderr)
	godebug.Line(ctx, scope, 25)

	logrus.SetOutput(stderr)
	godebug.Line(ctx, scope, 27)

	flag.Merge(flag.CommandLine, clientFlags.FlagSet, commonFlags.FlagSet)
	godebug.Line(ctx, scope, 29)

	flag.Usage = func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 30)
			fmt.Fprint(stdout, "Usage: docker [OPTIONS] COMMAND [arg...]\n"+daemonUsage+"       docker [ --help | -v | --version ]\n\n")
			godebug.Line(ctx, scope, 31)
			fmt.Fprint(stdout, "A self-sufficient runtime for containers.\n\nOptions:\n")
			godebug.Line(ctx, scope, 33)
			flag.CommandLine.SetOutput(stdout)
			godebug.Line(ctx, scope, 34)
			flag.PrintDefaults()
			godebug.Line(ctx, scope, 36)
			help := "\nCommands:\n"
			scope := scope.EnteringNewChildScope()
			scope.Declare("help", &help)
			{
				scope := scope.EnteringNewChildScope()
				for _, cmd := range dockerCommands {
					godebug.Line(ctx, scope, 38)
					scope.Declare("cmd", &cmd)
					godebug.Line(ctx, scope, 39)
					help += fmt.Sprintf("    %-10.10s%s\n", cmd.Name, cmd.Description)
				}
				godebug.Line(ctx, scope, 38)
			}
			godebug.Line(ctx, scope, 42)
			help += "\nRun 'docker COMMAND --help' for more information on a command."
			godebug.Line(ctx, scope, 43)
			fmt.Fprintf(stdout, "%s\n", help)
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	godebug.Line(ctx, scope, 46)

	flag.Parse()
	godebug.Line(ctx, scope, 48)

	if *flVersion {
		godebug.Line(ctx, scope, 49)
		showVersion()
		godebug.Line(ctx, scope, 50)
		return
	}
	godebug.Line(ctx, scope, 53)

	if *flHelp {
		godebug.Line(ctx, scope, 56)

		flag.Usage()
		godebug.Line(ctx, scope, 57)
		return
	}
	godebug.Line(ctx, scope, 60)

	clientCli := client.NewDockerCli(stdin, stdout, stderr, clientFlags)
	scope.Declare("clientCli", &clientCli)
	godebug.Line(ctx, scope, 62)

	c := cli.New(clientCli, daemonCli)
	scope.Declare("c", &c)
	godebug.SetTraceGen(ctx)
	godebug.Line(ctx, scope, 63)
	godebug.Line(ctx, scope, 64)

	if err := c.Run(flag.Args()...); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 65)
		if sterr, ok := err.(cli.StatusError); ok {
			scope := scope.EnteringNewChildScope()
			scope.Declare("sterr", &sterr, "ok", &ok)
			godebug.Line(ctx, scope, 66)
			if sterr.Status != "" {
				godebug.Line(ctx, scope, 67)
				fmt.Fprintln(stderr, sterr.Status)
				godebug.Line(ctx, scope, 68)
				os.Exit(1)
			}
			godebug.Line(ctx, scope, 70)
			os.Exit(sterr.StatusCode)
		}
		godebug.Line(ctx, scope, 72)
		fmt.Fprintln(stderr, err)
		godebug.Line(ctx, scope, 73)
		os.Exit(1)
	}
}

func showVersion() {
	ctx, _ok := godebug.EnterFunc(showVersion)
	if !_ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	godebug.Line(ctx, docker_go_scope, 78)
	if utils.ExperimentalBuild() {
		godebug.Line(ctx, docker_go_scope, 79)
		fmt.Printf("Docker version %s, build %s, experimental\n", dockerversion.Version, dockerversion.GitCommit)
	} else {
		godebug.Line(ctx, docker_go_scope, 80)
		godebug.Line(ctx, docker_go_scope, 81)
		fmt.Printf("Docker version %s, build %s\n", dockerversion.Version, dockerversion.GitCommit)
	}
}

var docker_go_contents = `package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/client"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/dockerversion"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/utils"
)

func main() {
	if reexec.Init() {
		return
	}

	// Set terminal emulation based on platform as required.
	stdin, stdout, stderr := term.StdStreams()

	logrus.SetOutput(stderr)

	flag.Merge(flag.CommandLine, clientFlags.FlagSet, commonFlags.FlagSet)

	flag.Usage = func() {
		fmt.Fprint(stdout, "Usage: docker [OPTIONS] COMMAND [arg...]\n"+daemonUsage+"       docker [ --help | -v | --version ]\n\n")
		fmt.Fprint(stdout, "A self-sufficient runtime for containers.\n\nOptions:\n")

		flag.CommandLine.SetOutput(stdout)
		flag.PrintDefaults()

		help := "\nCommands:\n"

		for _, cmd := range dockerCommands {
			help += fmt.Sprintf("    %-10.10s%s\n", cmd.Name, cmd.Description)
		}

		help += "\nRun 'docker COMMAND --help' for more information on a command."
		fmt.Fprintf(stdout, "%s\n", help)
	}

	flag.Parse()

	if *flVersion {
		showVersion()
		return
	}

	if *flHelp {
		// if global flag --help is present, regardless of what other options and commands there are,
		// just print the usage.
		flag.Usage()
		return
	}

	clientCli := client.NewDockerCli(stdin, stdout, stderr, clientFlags)

	c := cli.New(clientCli, daemonCli)
	_ = "breakpoint"
	if err := c.Run(flag.Args()...); err != nil {
		if sterr, ok := err.(cli.StatusError); ok {
			if sterr.Status != "" {
				fmt.Fprintln(stderr, sterr.Status)
				os.Exit(1)
			}
			os.Exit(sterr.StatusCode)
		}
		fmt.Fprintln(stderr, err)
		os.Exit(1)
	}
}

func showVersion() {
	if utils.ExperimentalBuild() {
		fmt.Printf("Docker version %s, build %s, experimental\n", dockerversion.Version, dockerversion.GitCommit)
	} else {
		fmt.Printf("Docker version %s, build %s\n", dockerversion.Version, dockerversion.GitCommit)
	}
}
`

var flags_go_scope = godebug.EnteringNewFile(main_pkg_scope, flags_go_contents)

var (
	flHelp    = flag.Bool([]string{"h", "-help"}, false, "Print usage")
	flVersion = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
)

type byName []cli.Command

func (a byName) Len() int {
	var result1 int
	ctx, ok := godebug.EnterFunc(func() {
		result1 = a.Len()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := flags_go_scope.EnteringNewChildScope()
	scope.Declare("a", &a)
	godebug.Line(ctx, scope, 17)
	return len(a)
}
func (a byName) Swap(i, j int) {
	ctx, ok := godebug.EnterFunc(func() {
		a.Swap(i, j)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := flags_go_scope.EnteringNewChildScope()
	scope.Declare("a", &a, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 18)
	a[i], a[j] = a[j], a[i]
}
func (a byName) Less(i, j int) bool {
	var result1 bool
	ctx, ok := godebug.EnterFunc(func() {
		result1 = a.Less(i, j)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := flags_go_scope.EnteringNewChildScope()
	scope.Declare("a", &a, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 19)
	return a[i].Name < a[j].Name
}

var dockerCommands []cli.Command

func init() {
	for _, cmd := range cli.DockerCommands {
		dockerCommands = append(dockerCommands, cmd)
	}
	sort.Sort(byName(dockerCommands))
}

var flags_go_contents = `package main

import (
	"sort"

	"github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

var (
	flHelp    = flag.Bool([]string{"h", "-help"}, false, "Print usage")
	flVersion = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
)

type byName []cli.Command

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

var dockerCommands []cli.Command

// TODO(tiborvass): do not show 'daemon' on client-only binaries

func init() {
	for _, cmd := range cli.DockerCommands {
		dockerCommands = append(dockerCommands, cmd)
	}
	sort.Sort(byName(dockerCommands))
}
`

var client_test_go_scope = godebug.EnteringNewFile(main_pkg_scope, client_test_go_contents)

func TestClientDebugEnabled(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestClientDebugEnabled(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := client_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 12)
	defer utils.DisableDebug()
	defer godebug.Defer(ctx, scope, 12)
	godebug.Line(ctx, scope, 14)

	clientFlags.Common.FlagSet.Parse([]string{"-D"})
	godebug.Line(ctx, scope, 15)
	clientFlags.PostParse()
	godebug.Line(ctx, scope, 17)

	if os.Getenv("DEBUG") != "1" {
		godebug.Line(ctx, scope, 18)
		t.Fatal("expected debug enabled, got false")
	}
	godebug.Line(ctx, scope, 20)
	if logrus.GetLevel() != logrus.DebugLevel {
		godebug.Line(ctx, scope, 21)
		t.Fatalf("expected logrus debug level, got %v", logrus.GetLevel())
	}
}

var client_test_go_contents = `package main

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/utils"
)

func TestClientDebugEnabled(t *testing.T) {
	defer utils.DisableDebug()

	clientFlags.Common.FlagSet.Parse([]string{"-D"})
	clientFlags.PostParse()

	if os.Getenv("DEBUG") != "1" {
		t.Fatal("expected debug enabled, got false")
	}
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Fatalf("expected logrus debug level, got %v", logrus.GetLevel())
	}
}
`

var flags_test_go_scope = godebug.EnteringNewFile(main_pkg_scope, flags_test_go_contents)

func TestDockerSubcommandsAreSorted(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestDockerSubcommandsAreSorted(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := flags_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 10)
	if !sort.IsSorted(byName(dockerCommands)) {
		godebug.Line(ctx, scope, 11)
		t.Fatal("Docker subcommands are not in sorted order")
	}
}

var flags_test_go_contents = `package main

import (
	"sort"
	"testing"
)

// Tests if the subcommands of docker are sorted
func TestDockerSubcommandsAreSorted(t *testing.T) {
	if !sort.IsSorted(byName(dockerCommands)) {
		t.Fatal("Docker subcommands are not in sorted order")
	}
}
`
