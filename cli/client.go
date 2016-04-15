package cli

import (
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mailgun/godebug/lib"
)

var client_go_scope = godebug.EnteringNewFile(cli_pkg_scope, client_go_contents)

type ClientFlags struct {
	FlagSet   *flag.FlagSet
	Common    *CommonFlags
	PostParse func()

	ConfigDir string
}

var client_go_contents = `package cli

import flag "github.com/docker/docker/pkg/mflag"

// ClientFlags represents flags for the docker client.
type ClientFlags struct {
	FlagSet   *flag.FlagSet
	Common    *CommonFlags
	PostParse func()

	ConfigDir string
}
`
