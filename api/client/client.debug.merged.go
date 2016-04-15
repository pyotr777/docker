package client

import (
	"fmt"
	"io"
	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	"github.com/mailgun/godebug/lib"
	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/engine-api/types"
	"errors"
	"sync"
	"github.com/docker/docker/pkg/stdcopy"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/stringutils"
	"github.com/docker/go-units"
	"github.com/docker/docker/api/client/formatter"
	"github.com/docker/docker/opts"
	"github.com/docker/engine-api/types/filters"
	"os"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/urlutil"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/utils"
	"github.com/docker/docker/api/client/inspect"
	"github.com/docker/docker/utils/templates"
	"bufio"
	"runtime"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/cliconfig/credentials"
	"github.com/docker/docker/pkg/term"
	"net/http/httputil"
	"github.com/docker/docker/pkg/signal"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/libnetwork/resolvconf/dns"
	"runtime/debug"
	"net"
	"sort"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/registry"
	"net/url"
	registrytypes "github.com/docker/engine-api/types/registry"
	"github.com/docker/engine-api/types/events"
	"encoding/json"
	"encoding/hex"
	"net/http"
	"path"
	"path/filepath"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/distribution"
	apiclient "github.com/docker/engine-api/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/notary/client"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/store"
	"github.com/docker/engine-api/types/container"
	"encoding/base64"
	"io/ioutil"
	gosignal "os/signal"
	"text/template"
	"github.com/docker/docker/dockerversion"
	"testing"
	"bytes"
	"github.com/docker/docker/pkg/jsonlog"
	eventtypes "github.com/docker/engine-api/types/events"
	"github.com/docker/docker/pkg/archive"
	networktypes "github.com/docker/engine-api/types/network"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/api"
	"github.com/docker/docker/cli"
	"github.com/docker/go-connections/sockets"
	"archive/tar"
	"regexp"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
)

var attach_go_scope = godebug.EnteringNewFile(client_pkg_scope, attach_go_contents)

func (cli *DockerCli) CmdAttach(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdAttach(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := attach_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 21)
	cmd := Cli.Subcmd("attach", []string{"CONTAINER"}, Cli.DockerCommands["attach"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 22)
	noStdin := cmd.Bool([]string{"-no-stdin"}, false, "Do not attach STDIN")
	scope.Declare("noStdin", &noStdin)
	godebug.Line(ctx, scope, 23)
	proxy := cmd.Bool([]string{"-sig-proxy"}, true, "Proxy all received signals to the process")
	scope.Declare("proxy", &proxy)
	godebug.Line(ctx, scope, 24)
	detachKeys := cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
	scope.Declare("detachKeys", &detachKeys)
	godebug.Line(ctx, scope, 26)

	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 28)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 30)

	c, err := cli.client.ContainerInspect(context.Background(), cmd.Arg(0))
	scope.Declare("c", &c, "err", &err)
	godebug.Line(ctx, scope, 31)
	if err != nil {
		godebug.Line(ctx, scope, 32)
		return err
	}
	godebug.Line(ctx, scope, 35)

	if !c.State.Running {
		godebug.Line(ctx, scope, 36)
		return fmt.Errorf("You cannot attach to a stopped container, start it first")
	}
	godebug.Line(ctx, scope, 39)

	if c.State.Paused {
		godebug.Line(ctx, scope, 40)
		return fmt.Errorf("You cannot attach to a paused container, unpause it first")
	}
	godebug.Line(ctx, scope, 43)

	if err := cli.CheckTtyInput(!*noStdin, c.Config.Tty); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 44)
		return err
	}
	godebug.Line(ctx, scope, 47)

	if *detachKeys != "" {
		godebug.Line(ctx, scope, 48)
		cli.configFile.DetachKeys = *detachKeys
	}
	godebug.Line(ctx, scope, 51)

	options := types.ContainerAttachOptions{
		ContainerID: cmd.Arg(0),
		Stream:      true,
		Stdin:       !*noStdin && c.Config.OpenStdin,
		Stdout:      true,
		Stderr:      true,
		DetachKeys:  cli.configFile.DetachKeys,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 60)

	var in io.ReadCloser
	scope.Declare("in", &in)
	godebug.Line(ctx, scope, 61)
	if options.Stdin {
		godebug.Line(ctx, scope, 62)
		in = cli.in
	}
	godebug.Line(ctx, scope, 65)

	if *proxy && !c.Config.Tty {
		godebug.Line(ctx, scope, 66)
		sigc := cli.forwardAllSignals(options.ContainerID)
		scope := scope.EnteringNewChildScope()
		scope.Declare("sigc", &sigc)
		godebug.Line(ctx, scope, 67)
		defer signal.StopCatch(sigc)
		defer godebug.Defer(ctx, scope, 67)
	}
	godebug.Line(ctx, scope, 70)

	resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
	scope.Declare("resp", &resp, "errAttach", &errAttach)
	godebug.Line(ctx, scope, 71)
	if errAttach != nil && errAttach != httputil.ErrPersistEOF {
		godebug.Line(ctx, scope, 75)

		return errAttach
	}
	godebug.Line(ctx, scope, 77)
	defer resp.Close()
	defer godebug.Defer(ctx, scope, 77)
	godebug.Line(ctx, scope, 79)

	if c.Config.Tty && cli.isTerminalOut {
		godebug.Line(ctx, scope, 80)
		height, width := cli.getTtySize()
		scope := scope.EnteringNewChildScope()
		scope.Declare("height", &height, "width", &width)
		godebug.Line(ctx, scope, 85)

		cli.resizeTtyTo(cmd.Arg(0), height+1, width+1, false)
		godebug.Line(ctx, scope, 89)

		if err := cli.monitorTtySize(cmd.Arg(0), false); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 90)
			logrus.Debugf("Error monitoring TTY size: %s", err)
		}
	}
	godebug.Line(ctx, scope, 93)
	if err := cli.holdHijackedConnection(context.Background(), c.Config.Tty, in, cli.out, cli.err, resp); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 94)
		return err
	}
	godebug.Line(ctx, scope, 97)

	if errAttach != nil {
		godebug.Line(ctx, scope, 98)
		return errAttach
	}
	godebug.Line(ctx, scope, 101)

	_, status, err := getExitCode(cli, options.ContainerID)
	scope.Declare("status", &status)
	godebug.Line(ctx, scope, 102)
	if err != nil {
		godebug.Line(ctx, scope, 103)
		return err
	}
	godebug.Line(ctx, scope, 105)
	if status != 0 {
		godebug.Line(ctx, scope, 106)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(ctx, scope, 109)

	return nil
}

var attach_go_contents = `package client

import (
	"fmt"
	"io"
	"net/http/httputil"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/engine-api/types"
)

// CmdAttach attaches to a running container.
//
// Usage: docker attach [OPTIONS] CONTAINER
func (cli *DockerCli) CmdAttach(args ...string) error {
	cmd := Cli.Subcmd("attach", []string{"CONTAINER"}, Cli.DockerCommands["attach"].Description, true)
	noStdin := cmd.Bool([]string{"-no-stdin"}, false, "Do not attach STDIN")
	proxy := cmd.Bool([]string{"-sig-proxy"}, true, "Proxy all received signals to the process")
	detachKeys := cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")

	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	c, err := cli.client.ContainerInspect(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}

	if !c.State.Running {
		return fmt.Errorf("You cannot attach to a stopped container, start it first")
	}

	if c.State.Paused {
		return fmt.Errorf("You cannot attach to a paused container, unpause it first")
	}

	if err := cli.CheckTtyInput(!*noStdin, c.Config.Tty); err != nil {
		return err
	}

	if *detachKeys != "" {
		cli.configFile.DetachKeys = *detachKeys
	}

	options := types.ContainerAttachOptions{
		ContainerID: cmd.Arg(0),
		Stream:      true,
		Stdin:       !*noStdin && c.Config.OpenStdin,
		Stdout:      true,
		Stderr:      true,
		DetachKeys:  cli.configFile.DetachKeys,
	}

	var in io.ReadCloser
	if options.Stdin {
		in = cli.in
	}

	if *proxy && !c.Config.Tty {
		sigc := cli.forwardAllSignals(options.ContainerID)
		defer signal.StopCatch(sigc)
	}

	resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
	if errAttach != nil && errAttach != httputil.ErrPersistEOF {
		// ContainerAttach returns an ErrPersistEOF (connection closed)
		// means server met an error and put it in Hijacked connection
		// keep the error and read detailed error message from hijacked connection later
		return errAttach
	}
	defer resp.Close()

	if c.Config.Tty && cli.isTerminalOut {
		height, width := cli.getTtySize()
		// To handle the case where a user repeatedly attaches/detaches without resizing their
		// terminal, the only way to get the shell prompt to display for attaches 2+ is to artificially
		// resize it, then go back to normal. Without this, every attach after the first will
		// require the user to manually resize or hit enter.
		cli.resizeTtyTo(cmd.Arg(0), height+1, width+1, false)

		// After the above resizing occurs, the call to monitorTtySize below will handle resetting back
		// to the actual size.
		if err := cli.monitorTtySize(cmd.Arg(0), false); err != nil {
			logrus.Debugf("Error monitoring TTY size: %s", err)
		}
	}
	if err := cli.holdHijackedConnection(context.Background(), c.Config.Tty, in, cli.out, cli.err, resp); err != nil {
		return err
	}

	if errAttach != nil {
		return errAttach
	}

	_, status, err := getExitCode(cli, options.ContainerID)
	if err != nil {
		return err
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}

	return nil
}
`


var client_pkg_scope = &godebug.Scope{}

func init() {
	client_pkg_scope.Vars = map[string]interface{}{
		"dockerfileFromLinePattern": &dockerfileFromLinePattern,
		"validDrivers": &validDrivers,
		"releasesRole": &releasesRole,
		"untrusted": &untrusted,
		"versionTemplate": &versionTemplate,
	}
	client_pkg_scope.Consts = map[string]interface{}{
		"fromContainer": fromContainer,
		"toContainer": toContainer,
		"acrossContainers": acrossContainers,
		"debug_level": debug_level,
		"errCmdNotFound": errCmdNotFound,
		"errCmdCouldNotBeInvoked": errCmdCouldNotBeInvoked,
	}
	client_pkg_scope.Funcs = map[string]interface{}{
		"validateTag": validateTag,
		"rewriteDockerfileFrom": rewriteDockerfileFrom,
		"replaceDockerfileTarWrapper": replaceDockerfileTarWrapper,
		"NewDockerCli": NewDockerCli,
		"getServerHost": getServerHost,
		"newHTTPClient": newHTTPClient,
		"clientUserAgent": clientUserAgent,
		"splitCpArg": splitCpArg,
		"resolveLocalPath": resolveLocalPath,
		"newCIDFile": newCIDFile,
		"streamEvents": streamEvents,
		"decodeEvents": decodeEvents,
		"printOutput": printOutput,
		"ParseExec": ParseExec,
		"readInput": readInput,
		"getCredentials": getCredentials,
		"getAllCredentials": getAllCredentials,
		"storeCredentials": storeCredentials,
		"eraseCredentials": eraseCredentials,
		"loadCredentialsStore": loadCredentialsStore,
		"consolidateIpam": consolidateIpam,
		"subnetMatches": subnetMatches,
		"networkUsage": networkUsage,
		"runStartContainerErr": runStartContainerErr,
		"calculateCPUPercent": calculateCPUPercent,
		"calculateBlockIO": calculateBlockIO,
		"calculateNetwork": calculateNetwork,
		"addTrustedFlags": addTrustedFlags,
		"isTrusted": isTrusted,
		"trustServer": trustServer,
		"convertTarget": convertTarget,
		"notaryError": notaryError,
		"encodeAuthToBase64": encodeAuthToBase64,
		"getExitCode": getExitCode,
		"getExecExitCode": getExecExitCode,
		"copyToFile": copyToFile,
		"TestParseExec": TestParseExec,
		"compareExecConfig": compareExecConfig,
		"TestDisplay": TestDisplay,
		"TestCalculBlockIO": TestCalculBlockIO,
		"unsetENV": unsetENV,
		"TestENVTrustServer": TestENVTrustServer,
		"TestHTTPENVTrustServer": TestHTTPENVTrustServer,
		"TestOfficialTrustServer": TestOfficialTrustServer,
		"TestNonOfficialTrustServer": TestNonOfficialTrustServer,
	}
}

var build_go_scope = godebug.EnteringNewFile(client_pkg_scope, build_go_contents)

type translatorFunc func(reference.NamedTagged) (reference.Canonical, error)

func (cli *DockerCli) CmdBuild(args ...string) error {
	var result1 error
	_ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdBuild(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := build_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(_ctx, scope, 43)
	cmd := Cli.Subcmd("build", []string{"PATH | URL | -"}, Cli.DockerCommands["build"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(_ctx, scope, 44)
	flTags := opts.NewListOpts(validateTag)
	scope.Declare("flTags", &flTags)
	godebug.Line(_ctx, scope, 45)
	cmd.Var(&flTags, []string{"t", "-tag"}, "Name and optionally a tag in the 'name:tag' format")
	godebug.Line(_ctx, scope, 46)
	suppressOutput := cmd.Bool([]string{"q", "-quiet"}, false, "Suppress the build output and print image ID on success")
	scope.Declare("suppressOutput", &suppressOutput)
	godebug.Line(_ctx, scope, 47)
	noCache := cmd.Bool([]string{"-no-cache"}, false, "Do not use cache when building the image")
	scope.Declare("noCache", &noCache)
	godebug.Line(_ctx, scope, 48)
	rm := cmd.Bool([]string{"-rm"}, true, "Remove intermediate containers after a successful build")
	scope.Declare("rm", &rm)
	godebug.Line(_ctx, scope, 49)
	forceRm := cmd.Bool([]string{"-force-rm"}, false, "Always remove intermediate containers")
	scope.Declare("forceRm", &forceRm)
	godebug.Line(_ctx, scope, 50)
	pull := cmd.Bool([]string{"-pull"}, false, "Always attempt to pull a newer version of the image")
	scope.Declare("pull", &pull)
	godebug.Line(_ctx, scope, 51)
	dockerfileName := cmd.String([]string{"f", "-file"}, "", "Name of the Dockerfile (Default is 'PATH/Dockerfile')")
	scope.Declare("dockerfileName", &dockerfileName)
	godebug.Line(_ctx, scope, 52)
	flMemoryString := cmd.String([]string{"m", "-memory"}, "", "Memory limit")
	scope.Declare("flMemoryString", &flMemoryString)
	godebug.Line(_ctx, scope, 53)
	flMemorySwap := cmd.String([]string{"-memory-swap"}, "", "Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	scope.Declare("flMemorySwap", &flMemorySwap)
	godebug.Line(_ctx, scope, 54)
	flShmSize := cmd.String([]string{"-shm-size"}, "", "Size of /dev/shm, default value is 64MB")
	scope.Declare("flShmSize", &flShmSize)
	godebug.Line(_ctx, scope, 55)
	flCPUShares := cmd.Int64([]string{"#c", "-cpu-shares"}, 0, "CPU shares (relative weight)")
	scope.Declare("flCPUShares", &flCPUShares)
	godebug.Line(_ctx, scope, 56)
	flCPUPeriod := cmd.Int64([]string{"-cpu-period"}, 0, "Limit the CPU CFS (Completely Fair Scheduler) period")
	scope.Declare("flCPUPeriod", &flCPUPeriod)
	godebug.Line(_ctx, scope, 57)
	flCPUQuota := cmd.Int64([]string{"-cpu-quota"}, 0, "Limit the CPU CFS (Completely Fair Scheduler) quota")
	scope.Declare("flCPUQuota", &flCPUQuota)
	godebug.Line(_ctx, scope, 58)
	flCPUSetCpus := cmd.String([]string{"-cpuset-cpus"}, "", "CPUs in which to allow execution (0-3, 0,1)")
	scope.Declare("flCPUSetCpus", &flCPUSetCpus)
	godebug.Line(_ctx, scope, 59)
	flCPUSetMems := cmd.String([]string{"-cpuset-mems"}, "", "MEMs in which to allow execution (0-3, 0,1)")
	scope.Declare("flCPUSetMems", &flCPUSetMems)
	godebug.Line(_ctx, scope, 60)
	flCgroupParent := cmd.String([]string{"-cgroup-parent"}, "", "Optional parent cgroup for the container")
	scope.Declare("flCgroupParent", &flCgroupParent)
	godebug.Line(_ctx, scope, 61)
	flBuildArg := opts.NewListOpts(runconfigopts.ValidateEnv)
	scope.Declare("flBuildArg", &flBuildArg)
	godebug.Line(_ctx, scope, 62)
	cmd.Var(&flBuildArg, []string{"-build-arg"}, "Set build-time variables")
	godebug.Line(_ctx, scope, 63)
	isolation := cmd.String([]string{"-isolation"}, "", "Container isolation technology")
	scope.Declare("isolation", &isolation)
	godebug.Line(_ctx, scope, 65)

	flLabels := opts.NewListOpts(nil)
	scope.Declare("flLabels", &flLabels)
	godebug.Line(_ctx, scope, 66)
	cmd.Var(&flLabels, []string{"-label"}, "Set metadata for an image")
	godebug.Line(_ctx, scope, 68)

	ulimits := make(map[string]*units.Ulimit)
	scope.Declare("ulimits", &ulimits)
	godebug.Line(_ctx, scope, 69)
	flUlimits := runconfigopts.NewUlimitOpt(&ulimits)
	scope.Declare("flUlimits", &flUlimits)
	godebug.Line(_ctx, scope, 70)
	cmd.Var(flUlimits, []string{"-ulimit"}, "Ulimit options")
	godebug.Line(_ctx, scope, 72)

	cmd.Require(flag.Exact, 1)
	godebug.Line(_ctx, scope, 75)

	addTrustedFlags(cmd, true)
	godebug.Line(_ctx, scope, 77)

	cmd.ParseFlags(args, true)
	godebug.Line(_ctx, scope, 79)

	var (
		ctx io.ReadCloser
		err error
	)
	scope.Declare("ctx", &ctx, "err", &err)
	godebug.Line(_ctx, scope, 84)

	specifiedContext := cmd.Arg(0)
	scope.Declare("specifiedContext", &specifiedContext)
	godebug.Line(_ctx, scope, 86)

	var (
		contextDir    string
		tempDir       string
		relDockerfile string
		progBuff      io.Writer
		buildBuff     io.Writer
	)
	scope.Declare("contextDir", &contextDir, "tempDir", &tempDir, "relDockerfile", &relDockerfile, "progBuff", &progBuff, "buildBuff", &buildBuff)
	godebug.Line(_ctx, scope, 94)

	progBuff = cli.out
	godebug.Line(_ctx, scope, 95)
	buildBuff = cli.out
	godebug.Line(_ctx, scope, 96)
	if *suppressOutput {
		godebug.Line(_ctx, scope, 97)
		progBuff = bytes.NewBuffer(nil)
		godebug.Line(_ctx, scope, 98)
		buildBuff = bytes.NewBuffer(nil)
	}
	godebug.Line(_ctx, scope, 101)

	switch {
	case godebug.Case(_ctx, scope, 102):
		fallthrough
	case specifiedContext == "-":
		godebug.Line(_ctx, scope, 103)
		ctx, relDockerfile, err = builder.GetContextFromReader(cli.in, *dockerfileName)
	case godebug.Case(_ctx, scope, 104):
		fallthrough
	case urlutil.IsGitURL(specifiedContext):
		godebug.Line(_ctx, scope, 105)
		tempDir, relDockerfile, err = builder.GetContextFromGitURL(specifiedContext, *dockerfileName)
	case godebug.Case(_ctx, scope, 106):
		fallthrough
	case urlutil.IsURL(specifiedContext):
		godebug.Line(_ctx, scope, 107)
		ctx, relDockerfile, err = builder.GetContextFromURL(progBuff, specifiedContext, *dockerfileName)
	default:
		godebug.Line(_ctx, scope, 108)
		godebug.Line(_ctx, scope, 109)
		contextDir, relDockerfile, err = builder.GetContextFromLocalDir(specifiedContext, *dockerfileName)
	}
	godebug.Line(_ctx, scope, 112)

	if err != nil {
		godebug.Line(_ctx, scope, 113)
		if *suppressOutput && urlutil.IsURL(specifiedContext) {
			godebug.Line(_ctx, scope, 114)
			fmt.Fprintln(cli.err, progBuff)
		}
		godebug.Line(_ctx, scope, 116)
		return fmt.Errorf("unable to prepare context: %s", err)
	}
	godebug.Line(_ctx, scope, 119)

	if tempDir != "" {
		godebug.Line(_ctx, scope, 120)
		defer os.RemoveAll(tempDir)
		defer godebug.Defer(_ctx, scope, 120)
		godebug.Line(_ctx, scope, 121)
		contextDir = tempDir
	}
	godebug.Line(_ctx, scope, 124)

	if ctx == nil {
		godebug.Line(_ctx, scope, 126)

		relDockerfile, err = archive.CanonicalTarNameForPath(relDockerfile)
		godebug.Line(_ctx, scope, 127)
		if err != nil {
			godebug.Line(_ctx, scope, 128)
			return fmt.Errorf("cannot canonicalize dockerfile path %s: %v", relDockerfile, err)
		}
		godebug.Line(_ctx, scope, 131)

		f, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
		scope := scope.EnteringNewChildScope()
		scope.Declare("f", &f, "err", &err)
		godebug.Line(_ctx, scope, 132)
		if err != nil && !os.IsNotExist(err) {
			godebug.Line(_ctx, scope, 133)
			return err
		}
		godebug.Line(_ctx, scope, 136)

		var excludes []string
		scope.Declare("excludes", &excludes)
		godebug.Line(_ctx, scope, 137)
		if err == nil {
			godebug.Line(_ctx, scope, 138)
			excludes, err = dockerignore.ReadAll(f)
			godebug.Line(_ctx, scope, 139)
			if err != nil {
				godebug.Line(_ctx, scope, 140)
				return err
			}
		}
		godebug.Line(_ctx, scope, 144)

		if err := builder.ValidateContextDirectory(contextDir, excludes); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 145)
			return fmt.Errorf("Error checking context: '%s'.", err)
		}
		godebug.Line(_ctx, scope, 155)

		var includes = []string{"."}
		scope.Declare("includes", &includes)
		godebug.Line(_ctx, scope, 156)
		keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
		scope.Declare("keepThem1", &keepThem1)
		godebug.Line(_ctx, scope, 157)
		keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
		scope.Declare("keepThem2", &keepThem2)
		godebug.Line(_ctx, scope, 158)
		if keepThem1 || keepThem2 {
			godebug.Line(_ctx, scope, 159)
			includes = append(includes, ".dockerignore", relDockerfile)
		}
		godebug.Line(_ctx, scope, 162)

		ctx, err = archive.TarWithOptions(contextDir, &archive.TarOptions{
			Compression:     archive.Uncompressed,
			ExcludePatterns: excludes,
			IncludeFiles:    includes,
		})
		godebug.Line(_ctx, scope, 167)
		if err != nil {
			godebug.Line(_ctx, scope, 168)
			return err
		}
	}
	godebug.Line(_ctx, scope, 172)

	var resolvedTags []*resolvedTag
	scope.Declare("resolvedTags", &resolvedTags)
	godebug.Line(_ctx, scope, 173)
	if isTrusted() {
		godebug.Line(_ctx, scope, 176)

		ctx = replaceDockerfileTarWrapper(ctx, relDockerfile, cli.trustedReference, &resolvedTags)
	}
	godebug.Line(_ctx, scope, 180)

	progressOutput := streamformatter.NewStreamFormatter().NewProgressOutput(progBuff, true)
	scope.Declare("progressOutput", &progressOutput)
	godebug.Line(_ctx, scope, 182)

	var body io.Reader = progress.NewProgressReader(ctx, progressOutput, 0, "", "Sending build context to Docker daemon")
	scope.Declare("body", &body)
	godebug.Line(_ctx, scope, 184)

	var memory int64
	scope.Declare("memory", &memory)
	godebug.Line(_ctx, scope, 185)
	if *flMemoryString != "" {
		godebug.Line(_ctx, scope, 186)
		parsedMemory, err := units.RAMInBytes(*flMemoryString)
		scope := scope.EnteringNewChildScope()
		scope.Declare("parsedMemory", &parsedMemory, "err", &err)
		godebug.Line(_ctx, scope, 187)
		if err != nil {
			godebug.Line(_ctx, scope, 188)
			return err
		}
		godebug.Line(_ctx, scope, 190)
		memory = parsedMemory
	}
	godebug.Line(_ctx, scope, 193)

	var memorySwap int64
	scope.Declare("memorySwap", &memorySwap)
	godebug.Line(_ctx, scope, 194)
	if *flMemorySwap != "" {
		godebug.Line(_ctx, scope, 195)
		if *flMemorySwap == "-1" {
			godebug.Line(_ctx, scope, 196)
			memorySwap = -1
		} else {
			godebug.Line(_ctx, scope, 197)
			godebug.Line(_ctx, scope, 198)
			parsedMemorySwap, err := units.RAMInBytes(*flMemorySwap)
			scope := scope.EnteringNewChildScope()
			scope.Declare("parsedMemorySwap", &parsedMemorySwap, "err", &err)
			godebug.Line(_ctx, scope, 199)
			if err != nil {
				godebug.Line(_ctx, scope, 200)
				return err
			}
			godebug.Line(_ctx, scope, 202)
			memorySwap = parsedMemorySwap
		}
	}
	godebug.Line(_ctx, scope, 206)

	var shmSize int64
	scope.Declare("shmSize", &shmSize)
	godebug.Line(_ctx, scope, 207)
	if *flShmSize != "" {
		godebug.Line(_ctx, scope, 208)
		shmSize, err = units.RAMInBytes(*flShmSize)
		godebug.Line(_ctx, scope, 209)
		if err != nil {
			godebug.Line(_ctx, scope, 210)
			return err
		}
	}
	godebug.Line(_ctx, scope, 214)

	options := types.ImageBuildOptions{
		Context:        body,
		Memory:         memory,
		MemorySwap:     memorySwap,
		Tags:           flTags.GetAll(),
		SuppressOutput: *suppressOutput,
		NoCache:        *noCache,
		Remove:         *rm,
		ForceRemove:    *forceRm,
		PullParent:     *pull,
		Isolation:      container.Isolation(*isolation),
		CPUSetCPUs:     *flCPUSetCpus,
		CPUSetMems:     *flCPUSetMems,
		CPUShares:      *flCPUShares,
		CPUQuota:       *flCPUQuota,
		CPUPeriod:      *flCPUPeriod,
		CgroupParent:   *flCgroupParent,
		Dockerfile:     relDockerfile,
		ShmSize:        shmSize,
		Ulimits:        flUlimits.GetList(),
		BuildArgs:      runconfigopts.ConvertKVStringsToMap(flBuildArg.GetAll()),
		AuthConfigs:    cli.retrieveAuthConfigs(),
		Labels:         runconfigopts.ConvertKVStringsToMap(flLabels.GetAll()),
	}
	scope.Declare("options", &options)
	godebug.Line(_ctx, scope, 239)

	response, err := cli.client.ImageBuild(context.Background(), options)
	scope.Declare("response", &response)
	godebug.Line(_ctx, scope, 240)
	if err != nil {
		godebug.Line(_ctx, scope, 241)
		return err
	}
	godebug.Line(_ctx, scope, 243)
	defer response.Body.Close()
	defer godebug.Defer(_ctx, scope, 243)
	godebug.Line(_ctx, scope, 245)

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, buildBuff, cli.outFd, cli.isTerminalOut, nil)
	godebug.Line(_ctx, scope, 246)
	if err != nil {
		godebug.Line(_ctx, scope, 247)
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			scope := scope.EnteringNewChildScope()
			scope.Declare("jerr", &jerr, "ok", &ok)
			godebug.Line(_ctx, scope, 249)

			if jerr.Code == 0 {
				godebug.Line(_ctx, scope, 250)
				jerr.Code = 1
			}
			godebug.Line(_ctx, scope, 252)
			if *suppressOutput {
				godebug.Line(_ctx, scope, 253)
				fmt.Fprintf(cli.err, "%s%s", progBuff, buildBuff)
			}
			godebug.Line(_ctx, scope, 255)
			return Cli.StatusError{Status: jerr.Message, StatusCode: jerr.Code}
		}
	}
	godebug.Line(_ctx, scope, 261)

	if response.OSType != "windows" && runtime.GOOS == "windows" {
		godebug.Line(_ctx, scope, 262)
		fmt.Fprintln(cli.err, `SECURITY WARNING: You are building a Docker image from Windows against a non-Windows Docker host. All files and directories added to build context will have '-rwxr-xr-x' permissions. It is recommended to double check and reset permissions for sensitive files and directories.`)
	}
	godebug.Line(_ctx, scope, 267)

	if *suppressOutput {
		godebug.Line(_ctx, scope, 268)
		fmt.Fprintf(cli.out, "%s", buildBuff)
	}
	godebug.Line(_ctx, scope, 271)

	if isTrusted() {
		{
			scope := scope.EnteringNewChildScope()

			for _, resolved := range resolvedTags {
				godebug.Line(_ctx, scope, 274)
				scope.Declare("resolved", &resolved)
				godebug.Line(_ctx, scope, 275)
				if err := cli.tagTrusted(resolved.digestRef, resolved.tagRef); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(_ctx, scope, 276)
					return err
				}
			}
			godebug.Line(_ctx, scope, 274)
		}
	}
	godebug.Line(_ctx, scope, 281)

	return nil
}

func validateTag(rawRepo string) (string, error) {
	var result1 string
	var result2 error
	_ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = validateTag(rawRepo)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(_ctx)
	scope := build_go_scope.EnteringNewChildScope()
	scope.Declare("rawRepo", &rawRepo)
	godebug.Line(_ctx, scope, 286)
	_, err := reference.ParseNamed(rawRepo)
	scope.Declare("err", &err)
	godebug.Line(_ctx, scope, 287)
	if err != nil {
		godebug.Line(_ctx, scope, 288)
		return "", err
	}
	godebug.Line(_ctx, scope, 291)

	return rawRepo, nil
}

var dockerfileFromLinePattern = regexp.MustCompile(`(?i)^[\s]*FROM[ \f\r\t\v]+(?P<image>[^ \f\r\t\v\n#]+)`)

type resolvedTag struct {
	digestRef reference.Canonical
	tagRef    reference.NamedTagged
}

func rewriteDockerfileFrom(dockerfile io.Reader, translator translatorFunc) (newDockerfile []byte, resolvedTags []*resolvedTag, err error) {
	_ctx, _ok := godebug.EnterFunc(func() {
		newDockerfile, resolvedTags, err = rewriteDockerfileFrom(dockerfile, translator)
	})
	if !_ok {
		return newDockerfile, resolvedTags, err
	}
	defer godebug.ExitFunc(_ctx)
	scope := build_go_scope.EnteringNewChildScope()
	scope.Declare("dockerfile", &dockerfile, "translator", &translator, "newDockerfile", &newDockerfile, "resolvedTags", &resolvedTags, "err", &err)
	godebug.Line(_ctx, scope, 308)
	scanner := bufio.NewScanner(dockerfile)
	scope.Declare("scanner", &scanner)
	godebug.Line(_ctx, scope, 309)
	buf := bytes.NewBuffer(nil)
	scope.Declare("buf", &buf)
	godebug.Line(_ctx, scope, 312)

	for scanner.Scan() {
		godebug.Line(_ctx, scope, 313)
		line := scanner.Text()
		scope := scope.EnteringNewChildScope()
		scope.Declare("line", &line)
		godebug.Line(_ctx, scope, 315)

		matches := dockerfileFromLinePattern.FindStringSubmatch(line)
		scope.Declare("matches", &matches)
		godebug.Line(_ctx, scope, 316)
		if matches != nil && matches[1] != api.NoBaseImageSpecifier {
			godebug.Line(_ctx, scope, 318)

			ref, err := reference.ParseNamed(matches[1])
			scope := scope.EnteringNewChildScope()
			scope.Declare("ref", &ref, "err", &err)
			godebug.Line(_ctx, scope, 319)
			if err != nil {
				godebug.Line(_ctx, scope, 320)
				return nil, nil, err
			}
			godebug.Line(_ctx, scope, 322)
			ref = reference.WithDefaultTag(ref)
			godebug.Line(_ctx, scope, 323)
			if ref, ok := ref.(reference.NamedTagged); ok && isTrusted() {
				scope := scope.EnteringNewChildScope()
				scope.Declare("ref", &ref, "ok", &ok)
				godebug.Line(_ctx, scope, 324)
				trustedRef, err := translator(ref)
				scope.Declare("trustedRef", &trustedRef, "err", &err)
				godebug.Line(_ctx, scope, 325)
				if err != nil {
					godebug.Line(_ctx, scope, 326)
					return nil, nil, err
				}
				godebug.Line(_ctx, scope, 329)

				line = dockerfileFromLinePattern.ReplaceAllLiteralString(line, fmt.Sprintf("FROM %s", trustedRef.String()))
				godebug.Line(_ctx, scope, 330)
				resolvedTags = append(resolvedTags, &resolvedTag{
					digestRef: trustedRef,
					tagRef:    ref,
				})
			}
		}
		godebug.Line(_ctx, scope, 337)

		_, err := fmt.Fprintln(buf, line)
		scope.Declare("err", &err)
		godebug.Line(_ctx, scope, 338)
		if err != nil {
			godebug.Line(_ctx, scope, 339)
			return nil, nil, err
		}
		godebug.Line(_ctx, scope, 312)
	}
	godebug.Line(_ctx, scope, 343)

	return buf.Bytes(), resolvedTags, scanner.Err()
}

func replaceDockerfileTarWrapper(inputTarStream io.ReadCloser, dockerfileName string, translator translatorFunc, resolvedTags *[]*resolvedTag) io.ReadCloser {
	var result1 io.ReadCloser
	_ctx, _ok := godebug.EnterFunc(func() {
		result1 = replaceDockerfileTarWrapper(inputTarStream, dockerfileName, translator, resolvedTags)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := build_go_scope.EnteringNewChildScope()
	scope.Declare("inputTarStream", &inputTarStream, "dockerfileName", &dockerfileName, "translator", &translator, "resolvedTags", &resolvedTags)
	godebug.Line(_ctx, scope, 351)
	pipeReader, pipeWriter := io.Pipe()
	scope.Declare("pipeReader", &pipeReader, "pipeWriter", &pipeWriter)
	godebug.Line(_ctx, scope, 352)
	go func() {
		fn := func(_ctx *godebug.Context) {
			godebug.Line(_ctx, scope, 353)
			tarReader := tar.NewReader(inputTarStream)
			scope := scope.EnteringNewChildScope()
			scope.Declare("tarReader", &tarReader)
			godebug.Line(_ctx, scope, 354)
			tarWriter := tar.NewWriter(pipeWriter)
			scope.Declare("tarWriter", &tarWriter)
			godebug.Line(_ctx, scope, 356)
			defer inputTarStream.Close()
			defer godebug.Defer(_ctx, scope, 356)
			godebug.Line(_ctx, scope, 358)
			for {
				godebug.Line(_ctx, scope, 359)
				hdr, err := tarReader.Next()
				scope := scope.EnteringNewChildScope()
				scope.Declare("hdr", &hdr, "err", &err)
				godebug.Line(_ctx, scope, 360)
				if err == io.EOF {
					godebug.Line(_ctx, scope, 362)
					tarWriter.Close()
					godebug.Line(_ctx, scope, 363)
					pipeWriter.Close()
					godebug.Line(_ctx, scope, 364)
					return
				}
				godebug.Line(_ctx, scope, 366)
				if err != nil {
					godebug.Line(_ctx, scope, 367)
					pipeWriter.CloseWithError(err)
					godebug.Line(_ctx, scope, 368)
					return
				}
				godebug.Line(_ctx, scope, 371)
				var content io.Reader = tarReader
				scope.Declare("content", &content)
				godebug.Line(_ctx, scope, 372)
				if hdr.Name == dockerfileName {
					godebug.Line(_ctx, scope, 376)
					var newDockerfile []byte
					scope := scope.EnteringNewChildScope()
					scope.Declare("newDockerfile", &newDockerfile)
					godebug.Line(_ctx, scope, 377)
					newDockerfile, *resolvedTags, err = rewriteDockerfileFrom(content, translator)
					godebug.Line(_ctx, scope, 378)
					if err != nil {
						godebug.Line(_ctx, scope, 379)
						pipeWriter.CloseWithError(err)
						godebug.Line(_ctx, scope, 380)
						return
					}
					godebug.Line(_ctx, scope, 382)
					hdr.Size = int64(len(newDockerfile))
					godebug.Line(_ctx, scope, 383)
					content = bytes.NewBuffer(newDockerfile)
				}
				godebug.Line(_ctx, scope, 386)
				if err := tarWriter.WriteHeader(hdr); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(_ctx, scope, 387)
					pipeWriter.CloseWithError(err)
					godebug.Line(_ctx, scope, 388)
					return
				}
				godebug.Line(_ctx, scope, 391)
				if _, err := io.Copy(tarWriter, content); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(_ctx, scope, 392)
					pipeWriter.CloseWithError(err)
					godebug.Line(_ctx, scope, 393)
					return
				}
				godebug.Line(_ctx, scope, 358)
			}
		}
		if _ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(_ctx)
			fn(_ctx)
		}
	}()
	godebug.Line(_ctx, scope, 398)

	return pipeReader
}

var build_go_contents = `package client

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"golang.org/x/net/context"

	"github.com/docker/docker/api"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerignore"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/pkg/urlutil"
	"github.com/docker/docker/reference"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/go-units"
)

type translatorFunc func(reference.NamedTagged) (reference.Canonical, error)

// CmdBuild builds a new image from the source code at a given path.
//
// If '-' is provided instead of a path or URL, Docker will build an image from either a Dockerfile or tar archive read from STDIN.
//
// Usage: docker build [OPTIONS] PATH | URL | -
func (cli *DockerCli) CmdBuild(args ...string) error {
	cmd := Cli.Subcmd("build", []string{"PATH | URL | -"}, Cli.DockerCommands["build"].Description, true)
	flTags := opts.NewListOpts(validateTag)
	cmd.Var(&flTags, []string{"t", "-tag"}, "Name and optionally a tag in the 'name:tag' format")
	suppressOutput := cmd.Bool([]string{"q", "-quiet"}, false, "Suppress the build output and print image ID on success")
	noCache := cmd.Bool([]string{"-no-cache"}, false, "Do not use cache when building the image")
	rm := cmd.Bool([]string{"-rm"}, true, "Remove intermediate containers after a successful build")
	forceRm := cmd.Bool([]string{"-force-rm"}, false, "Always remove intermediate containers")
	pull := cmd.Bool([]string{"-pull"}, false, "Always attempt to pull a newer version of the image")
	dockerfileName := cmd.String([]string{"f", "-file"}, "", "Name of the Dockerfile (Default is 'PATH/Dockerfile')")
	flMemoryString := cmd.String([]string{"m", "-memory"}, "", "Memory limit")
	flMemorySwap := cmd.String([]string{"-memory-swap"}, "", "Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	flShmSize := cmd.String([]string{"-shm-size"}, "", "Size of /dev/shm, default value is 64MB")
	flCPUShares := cmd.Int64([]string{"#c", "-cpu-shares"}, 0, "CPU shares (relative weight)")
	flCPUPeriod := cmd.Int64([]string{"-cpu-period"}, 0, "Limit the CPU CFS (Completely Fair Scheduler) period")
	flCPUQuota := cmd.Int64([]string{"-cpu-quota"}, 0, "Limit the CPU CFS (Completely Fair Scheduler) quota")
	flCPUSetCpus := cmd.String([]string{"-cpuset-cpus"}, "", "CPUs in which to allow execution (0-3, 0,1)")
	flCPUSetMems := cmd.String([]string{"-cpuset-mems"}, "", "MEMs in which to allow execution (0-3, 0,1)")
	flCgroupParent := cmd.String([]string{"-cgroup-parent"}, "", "Optional parent cgroup for the container")
	flBuildArg := opts.NewListOpts(runconfigopts.ValidateEnv)
	cmd.Var(&flBuildArg, []string{"-build-arg"}, "Set build-time variables")
	isolation := cmd.String([]string{"-isolation"}, "", "Container isolation technology")

	flLabels := opts.NewListOpts(nil)
	cmd.Var(&flLabels, []string{"-label"}, "Set metadata for an image")

	ulimits := make(map[string]*units.Ulimit)
	flUlimits := runconfigopts.NewUlimitOpt(&ulimits)
	cmd.Var(flUlimits, []string{"-ulimit"}, "Ulimit options")

	cmd.Require(flag.Exact, 1)

	// For trusted pull on "FROM <image>" instruction.
	addTrustedFlags(cmd, true)

	cmd.ParseFlags(args, true)

	var (
		ctx io.ReadCloser
		err error
	)

	specifiedContext := cmd.Arg(0)

	var (
		contextDir    string
		tempDir       string
		relDockerfile string
		progBuff      io.Writer
		buildBuff     io.Writer
	)

	progBuff = cli.out
	buildBuff = cli.out
	if *suppressOutput {
		progBuff = bytes.NewBuffer(nil)
		buildBuff = bytes.NewBuffer(nil)
	}

	switch {
	case specifiedContext == "-":
		ctx, relDockerfile, err = builder.GetContextFromReader(cli.in, *dockerfileName)
	case urlutil.IsGitURL(specifiedContext):
		tempDir, relDockerfile, err = builder.GetContextFromGitURL(specifiedContext, *dockerfileName)
	case urlutil.IsURL(specifiedContext):
		ctx, relDockerfile, err = builder.GetContextFromURL(progBuff, specifiedContext, *dockerfileName)
	default:
		contextDir, relDockerfile, err = builder.GetContextFromLocalDir(specifiedContext, *dockerfileName)
	}

	if err != nil {
		if *suppressOutput && urlutil.IsURL(specifiedContext) {
			fmt.Fprintln(cli.err, progBuff)
		}
		return fmt.Errorf("unable to prepare context: %s", err)
	}

	if tempDir != "" {
		defer os.RemoveAll(tempDir)
		contextDir = tempDir
	}

	if ctx == nil {
		// And canonicalize dockerfile name to a platform-independent one
		relDockerfile, err = archive.CanonicalTarNameForPath(relDockerfile)
		if err != nil {
			return fmt.Errorf("cannot canonicalize dockerfile path %s: %v", relDockerfile, err)
		}

		f, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		var excludes []string
		if err == nil {
			excludes, err = dockerignore.ReadAll(f)
			if err != nil {
				return err
			}
		}

		if err := builder.ValidateContextDirectory(contextDir, excludes); err != nil {
			return fmt.Errorf("Error checking context: '%s'.", err)
		}

		// If .dockerignore mentions .dockerignore or the Dockerfile
		// then make sure we send both files over to the daemon
		// because Dockerfile is, obviously, needed no matter what, and
		// .dockerignore is needed to know if either one needs to be
		// removed. The daemon will remove them for us, if needed, after it
		// parses the Dockerfile. Ignore errors here, as they will have been
		// caught by validateContextDirectory above.
		var includes = []string{"."}
		keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
		keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
		if keepThem1 || keepThem2 {
			includes = append(includes, ".dockerignore", relDockerfile)
		}

		ctx, err = archive.TarWithOptions(contextDir, &archive.TarOptions{
			Compression:     archive.Uncompressed,
			ExcludePatterns: excludes,
			IncludeFiles:    includes,
		})
		if err != nil {
			return err
		}
	}

	var resolvedTags []*resolvedTag
	if isTrusted() {
		// Wrap the tar archive to replace the Dockerfile entry with the rewritten
		// Dockerfile which uses trusted pulls.
		ctx = replaceDockerfileTarWrapper(ctx, relDockerfile, cli.trustedReference, &resolvedTags)
	}

	// Setup an upload progress bar
	progressOutput := streamformatter.NewStreamFormatter().NewProgressOutput(progBuff, true)

	var body io.Reader = progress.NewProgressReader(ctx, progressOutput, 0, "", "Sending build context to Docker daemon")

	var memory int64
	if *flMemoryString != "" {
		parsedMemory, err := units.RAMInBytes(*flMemoryString)
		if err != nil {
			return err
		}
		memory = parsedMemory
	}

	var memorySwap int64
	if *flMemorySwap != "" {
		if *flMemorySwap == "-1" {
			memorySwap = -1
		} else {
			parsedMemorySwap, err := units.RAMInBytes(*flMemorySwap)
			if err != nil {
				return err
			}
			memorySwap = parsedMemorySwap
		}
	}

	var shmSize int64
	if *flShmSize != "" {
		shmSize, err = units.RAMInBytes(*flShmSize)
		if err != nil {
			return err
		}
	}

	options := types.ImageBuildOptions{
		Context:        body,
		Memory:         memory,
		MemorySwap:     memorySwap,
		Tags:           flTags.GetAll(),
		SuppressOutput: *suppressOutput,
		NoCache:        *noCache,
		Remove:         *rm,
		ForceRemove:    *forceRm,
		PullParent:     *pull,
		Isolation:      container.Isolation(*isolation),
		CPUSetCPUs:     *flCPUSetCpus,
		CPUSetMems:     *flCPUSetMems,
		CPUShares:      *flCPUShares,
		CPUQuota:       *flCPUQuota,
		CPUPeriod:      *flCPUPeriod,
		CgroupParent:   *flCgroupParent,
		Dockerfile:     relDockerfile,
		ShmSize:        shmSize,
		Ulimits:        flUlimits.GetList(),
		BuildArgs:      runconfigopts.ConvertKVStringsToMap(flBuildArg.GetAll()),
		AuthConfigs:    cli.retrieveAuthConfigs(),
		Labels:         runconfigopts.ConvertKVStringsToMap(flLabels.GetAll()),
	}

	response, err := cli.client.ImageBuild(context.Background(), options)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, buildBuff, cli.outFd, cli.isTerminalOut, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// If no error code is set, default to 1
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			if *suppressOutput {
				fmt.Fprintf(cli.err, "%s%s", progBuff, buildBuff)
			}
			return Cli.StatusError{Status: jerr.Message, StatusCode: jerr.Code}
		}
	}

	// Windows: show error message about modified file permissions if the
	// daemon isn't running Windows.
	if response.OSType != "windows" && runtime.GOOS == "windows" {
		fmt.Fprintln(cli.err, ` + "`" + `SECURITY WARNING: You are building a Docker image from Windows against a non-Windows Docker host. All files and directories added to build context will have '-rwxr-xr-x' permissions. It is recommended to double check and reset permissions for sensitive files and directories.` + "`" + `)
	}

	// Everything worked so if -q was provided the output from the daemon
	// should be just the image ID and we'll print that to stdout.
	if *suppressOutput {
		fmt.Fprintf(cli.out, "%s", buildBuff)
	}

	if isTrusted() {
		// Since the build was successful, now we must tag any of the resolved
		// images from the above Dockerfile rewrite.
		for _, resolved := range resolvedTags {
			if err := cli.tagTrusted(resolved.digestRef, resolved.tagRef); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateTag checks if the given image name can be resolved.
func validateTag(rawRepo string) (string, error) {
	_, err := reference.ParseNamed(rawRepo)
	if err != nil {
		return "", err
	}

	return rawRepo, nil
}

var dockerfileFromLinePattern = regexp.MustCompile(` + "`" + `(?i)^[\s]*FROM[ \f\r\t\v]+(?P<image>[^ \f\r\t\v\n#]+)` + "`" + `)

// resolvedTag records the repository, tag, and resolved digest reference
// from a Dockerfile rewrite.
type resolvedTag struct {
	digestRef reference.Canonical
	tagRef    reference.NamedTagged
}

// rewriteDockerfileFrom rewrites the given Dockerfile by resolving images in
// "FROM <image>" instructions to a digest reference. ` + "`" + `translator` + "`" + ` is a
// function that takes a repository name and tag reference and returns a
// trusted digest reference.
func rewriteDockerfileFrom(dockerfile io.Reader, translator translatorFunc) (newDockerfile []byte, resolvedTags []*resolvedTag, err error) {
	scanner := bufio.NewScanner(dockerfile)
	buf := bytes.NewBuffer(nil)

	// Scan the lines of the Dockerfile, looking for a "FROM" line.
	for scanner.Scan() {
		line := scanner.Text()

		matches := dockerfileFromLinePattern.FindStringSubmatch(line)
		if matches != nil && matches[1] != api.NoBaseImageSpecifier {
			// Replace the line with a resolved "FROM repo@digest"
			ref, err := reference.ParseNamed(matches[1])
			if err != nil {
				return nil, nil, err
			}
			ref = reference.WithDefaultTag(ref)
			if ref, ok := ref.(reference.NamedTagged); ok && isTrusted() {
				trustedRef, err := translator(ref)
				if err != nil {
					return nil, nil, err
				}

				line = dockerfileFromLinePattern.ReplaceAllLiteralString(line, fmt.Sprintf("FROM %s", trustedRef.String()))
				resolvedTags = append(resolvedTags, &resolvedTag{
					digestRef: trustedRef,
					tagRef:    ref,
				})
			}
		}

		_, err := fmt.Fprintln(buf, line)
		if err != nil {
			return nil, nil, err
		}
	}

	return buf.Bytes(), resolvedTags, scanner.Err()
}

// replaceDockerfileTarWrapper wraps the given input tar archive stream and
// replaces the entry with the given Dockerfile name with the contents of the
// new Dockerfile. Returns a new tar archive stream with the replaced
// Dockerfile.
func replaceDockerfileTarWrapper(inputTarStream io.ReadCloser, dockerfileName string, translator translatorFunc, resolvedTags *[]*resolvedTag) io.ReadCloser {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		tarReader := tar.NewReader(inputTarStream)
		tarWriter := tar.NewWriter(pipeWriter)

		defer inputTarStream.Close()

		for {
			hdr, err := tarReader.Next()
			if err == io.EOF {
				// Signals end of archive.
				tarWriter.Close()
				pipeWriter.Close()
				return
			}
			if err != nil {
				pipeWriter.CloseWithError(err)
				return
			}

			var content io.Reader = tarReader
			if hdr.Name == dockerfileName {
				// This entry is the Dockerfile. Since the tar archive was
				// generated from a directory on the local filesystem, the
				// Dockerfile will only appear once in the archive.
				var newDockerfile []byte
				newDockerfile, *resolvedTags, err = rewriteDockerfileFrom(content, translator)
				if err != nil {
					pipeWriter.CloseWithError(err)
					return
				}
				hdr.Size = int64(len(newDockerfile))
				content = bytes.NewBuffer(newDockerfile)
			}

			if err := tarWriter.WriteHeader(hdr); err != nil {
				pipeWriter.CloseWithError(err)
				return
			}

			if _, err := io.Copy(tarWriter, content); err != nil {
				pipeWriter.CloseWithError(err)
				return
			}
		}
	}()

	return pipeReader
}
`


var cli_go_scope = godebug.EnteringNewFile(client_pkg_scope, cli_go_contents)

type DockerCli struct {
	init func() error

	configFile *cliconfig.ConfigFile

	in io.ReadCloser

	out io.Writer

	err io.Writer

	keyFile string

	inFd uintptr

	outFd uintptr

	isTerminalIn bool

	isTerminalOut bool

	client apiclient.APIClient

	state *term.State
}

func (cli *DockerCli) Initialize() error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.Initialize()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 56)
	if cli.init == nil {
		godebug.Line(ctx, scope, 57)
		return nil
	}
	godebug.Line(ctx, scope, 59)
	return cli.init()
}

func (cli *DockerCli) CheckTtyInput(attachStdin, ttyMode bool) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CheckTtyInput(attachStdin, ttyMode)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "attachStdin", &attachStdin, "ttyMode", &ttyMode)
	godebug.Line(ctx, scope, 68)

	if ttyMode && attachStdin && !cli.isTerminalIn {
		godebug.Line(ctx, scope, 69)
		return errors.New("cannot enable tty mode on non tty input")
	}
	godebug.Line(ctx, scope, 71)
	return nil
}

func (cli *DockerCli) PsFormat() string {
	var result1 string
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.PsFormat()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 77)
	return cli.configFile.PsFormat
}

func (cli *DockerCli) ImagesFormat() string {
	var result1 string
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.ImagesFormat()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 83)
	return cli.configFile.ImagesFormat
}

func (cli *DockerCli) setRawTerminal() error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.setRawTerminal()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 87)
	if cli.isTerminalIn && os.Getenv("NORAW") == "" {
		godebug.Line(ctx, scope, 88)
		state, err := term.SetRawTerminal(cli.inFd)
		scope := scope.EnteringNewChildScope()
		scope.Declare("state", &state, "err", &err)
		godebug.Line(ctx, scope, 89)
		if err != nil {
			godebug.Line(ctx, scope, 90)
			return err
		}
		godebug.Line(ctx, scope, 92)
		cli.state = state
	}
	godebug.Line(ctx, scope, 94)
	return nil
}

func (cli *DockerCli) restoreTerminal(in io.Closer) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.restoreTerminal(in)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "in", &in)
	godebug.Line(ctx, scope, 98)
	if cli.state != nil {
		godebug.Line(ctx, scope, 99)
		term.RestoreTerminal(cli.inFd, cli.state)
	}
	godebug.Line(ctx, scope, 105)

	if in != nil && runtime.GOOS != "darwin" {
		godebug.Line(ctx, scope, 106)
		return in.Close()
	}
	godebug.Line(ctx, scope, 108)
	return nil
}

func NewDockerCli(in io.ReadCloser, out, err io.Writer, clientFlags *cli.ClientFlags) *DockerCli {
	var result1 *DockerCli
	ctx, ok := godebug.EnterFunc(func() {
		result1 = NewDockerCli(in, out, err, clientFlags)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("in", &in, "out", &out, "err", &err, "clientFlags", &clientFlags)
	godebug.Line(ctx, scope, 116)
	cli := &DockerCli{
		in:      in,
		out:     out,
		err:     err,
		keyFile: clientFlags.Common.TrustKey,
	}
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 123)

	cli.init = func() error {
		var result1 error
		fn := func(ctx *godebug.Context) {
			result1 = func() error {
				godebug.Line(ctx, scope, 124)
				clientFlags.PostParse()
				godebug.Line(ctx, scope, 125)
				configFile, e := cliconfig.Load(cliconfig.ConfigDir())
				scope := scope.EnteringNewChildScope()
				scope.Declare("configFile", &configFile, "e", &e)
				godebug.Line(ctx, scope, 126)
				if e != nil {
					godebug.Line(ctx, scope, 127)
					fmt.Fprintf(cli.err, "WARNING: Error loading config file:%v\n", e)
				}
				godebug.Line(ctx, scope, 129)
				if !configFile.ContainsAuth() {
					godebug.Line(ctx, scope, 130)
					credentials.DetectDefaultStore(configFile)
				}
				godebug.Line(ctx, scope, 132)
				cli.configFile = configFile
				godebug.Line(ctx, scope, 134)
				host, err := getServerHost(clientFlags.Common.Hosts, clientFlags.Common.TLSOptions)
				scope.Declare("host", &host, "err", &err)
				godebug.Line(ctx, scope, 135)
				if err != nil {
					godebug.Line(ctx, scope, 136)
					return err
				}
				godebug.Line(ctx, scope, 139)
				customHeaders := cli.configFile.HTTPHeaders
				scope.Declare("customHeaders", &customHeaders)
				godebug.Line(ctx, scope, 140)
				if customHeaders == nil {
					godebug.Line(ctx, scope, 141)
					customHeaders = map[string]string{}
				}
				godebug.Line(ctx, scope, 143)
				customHeaders["User-Agent"] = clientUserAgent()
				godebug.Line(ctx, scope, 145)
				verStr := api.DefaultVersion.String()
				scope.Declare("verStr", &verStr)
				godebug.Line(ctx, scope, 146)
				if tmpStr := os.Getenv("DOCKER_API_VERSION"); tmpStr != "" {
					scope := scope.EnteringNewChildScope()
					scope.Declare("tmpStr", &tmpStr)
					godebug.Line(ctx, scope, 147)
					verStr = tmpStr
				}
				godebug.Line(ctx, scope, 150)
				httpClient, err := newHTTPClient(host, clientFlags.Common.TLSOptions)
				scope.Declare("httpClient", &httpClient)
				godebug.Line(ctx, scope, 151)
				if err != nil {
					godebug.Line(ctx, scope, 152)
					return err
				}
				godebug.Line(ctx, scope, 155)
				client, err := apiclient.NewClient(host, verStr, httpClient, customHeaders)
				scope.Declare("client", &client)
				godebug.Line(ctx, scope, 156)
				if err != nil {
					godebug.Line(ctx, scope, 157)
					return err
				}
				godebug.Line(ctx, scope, 159)
				cli.client = client
				godebug.Line(ctx, scope, 161)
				if cli.in != nil {
					godebug.Line(ctx, scope, 162)
					cli.inFd, cli.isTerminalIn = term.GetFdInfo(cli.in)
				}
				godebug.Line(ctx, scope, 164)
				if cli.out != nil {
					godebug.Line(ctx, scope, 165)
					cli.outFd, cli.isTerminalOut = term.GetFdInfo(cli.out)
				}
				godebug.Line(ctx, scope, 168)
				return nil
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1
	}
	godebug.Line(ctx, scope, 171)

	return cli
}

func getServerHost(hosts []string, tlsOptions *tlsconfig.Options) (host string, err error) {
	ctx, ok := godebug.EnterFunc(func() {
		host, err = getServerHost(hosts, tlsOptions)
	})
	if !ok {
		return host, err
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("hosts", &hosts, "tlsOptions", &tlsOptions, "host", &host, "err", &err)
	godebug.Line(ctx, scope, 175)
	switch len(hosts) {
	case godebug.Case(ctx, scope, 176):
		fallthrough
	case 0:
		godebug.Line(ctx, scope, 177)
		host = os.Getenv("DOCKER_HOST")
	case godebug.Case(ctx, scope, 178):
		fallthrough
	case 1:
		godebug.Line(ctx, scope, 179)
		host = hosts[0]
	default:
		godebug.Line(ctx, scope, 180)
		godebug.Line(ctx, scope, 181)
		return "", errors.New("Please specify only one -H")
	}
	godebug.Line(ctx, scope, 184)

	host, err = opts.ParseHost(tlsOptions != nil, host)
	godebug.Line(ctx, scope, 185)
	return
}

func newHTTPClient(host string, tlsOptions *tlsconfig.Options) (*http.Client, error) {
	var result1 *http.Client
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = newHTTPClient(host, tlsOptions)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := cli_go_scope.EnteringNewChildScope()
	scope.Declare("host", &host, "tlsOptions", &tlsOptions)
	godebug.Line(ctx, scope, 189)
	if tlsOptions == nil {
		godebug.Line(ctx, scope, 191)

		return nil, nil
	}
	godebug.Line(ctx, scope, 194)

	config, err := tlsconfig.Client(*tlsOptions)
	scope.Declare("config", &config, "err", &err)
	godebug.Line(ctx, scope, 195)
	if err != nil {
		godebug.Line(ctx, scope, 196)
		return nil, err
	}
	godebug.Line(ctx, scope, 198)
	tr := &http.Transport{
		TLSClientConfig: config,
	}
	scope.Declare("tr", &tr)
	godebug.Line(ctx, scope, 201)

	proto, addr, _, err := apiclient.ParseHost(host)
	scope.Declare("proto", &proto, "addr", &addr)
	godebug.Line(ctx, scope, 202)
	if err != nil {
		godebug.Line(ctx, scope, 203)
		return nil, err
	}
	godebug.Line(ctx, scope, 206)

	sockets.ConfigureTransport(tr, proto, addr)
	godebug.Line(ctx, scope, 208)

	return &http.Client{
		Transport: tr,
	}, nil
}

func clientUserAgent() string {
	var result1 string
	ctx, ok := godebug.EnterFunc(func() {
		result1 = clientUserAgent()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	godebug.Line(ctx, cli_go_scope, 214)
	return "Docker-Client/" + dockerversion.Version + " (" + runtime.GOOS + ")"
}

var cli_go_contents = `package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/docker/docker/api"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/cliconfig/credentials"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
)

// DockerCli represents the docker command line client.
// Instances of the client can be returned from NewDockerCli.
type DockerCli struct {
	// initializing closure
	init func() error

	// configFile has the client configuration file
	configFile *cliconfig.ConfigFile
	// in holds the input stream and closer (io.ReadCloser) for the client.
	in io.ReadCloser
	// out holds the output stream (io.Writer) for the client.
	out io.Writer
	// err holds the error stream (io.Writer) for the client.
	err io.Writer
	// keyFile holds the key file as a string.
	keyFile string
	// inFd holds the file descriptor of the client's STDIN (if valid).
	inFd uintptr
	// outFd holds file descriptor of the client's STDOUT (if valid).
	outFd uintptr
	// isTerminalIn indicates whether the client's STDIN is a TTY
	isTerminalIn bool
	// isTerminalOut indicates whether the client's STDOUT is a TTY
	isTerminalOut bool
	// client is the http client that performs all API operations
	client client.APIClient
	// state holds the terminal state
	state *term.State
}

// Initialize calls the init function that will setup the configuration for the client
// such as the TLS, tcp and other parameters used to run the client.
func (cli *DockerCli) Initialize() error {
	if cli.init == nil {
		return nil
	}
	return cli.init()
}

// CheckTtyInput checks if we are trying to attach to a container tty
// from a non-tty client input stream, and if so, returns an error.
func (cli *DockerCli) CheckTtyInput(attachStdin, ttyMode bool) error {
	// In order to attach to a container tty, input stream for the client must
	// be a tty itself: redirecting or piping the client standard input is
	// incompatible with ` + "`" + `docker run -t` + "`" + `, ` + "`" + `docker exec -t` + "`" + ` or ` + "`" + `docker attach` + "`" + `.
	if ttyMode && attachStdin && !cli.isTerminalIn {
		return errors.New("cannot enable tty mode on non tty input")
	}
	return nil
}

// PsFormat returns the format string specified in the configuration.
// String contains columns and format specification, for example {{ID}}\t{{Name}}.
func (cli *DockerCli) PsFormat() string {
	return cli.configFile.PsFormat
}

// ImagesFormat returns the format string specified in the configuration.
// String contains columns and format specification, for example {{ID}}\t{{Name}}.
func (cli *DockerCli) ImagesFormat() string {
	return cli.configFile.ImagesFormat
}

func (cli *DockerCli) setRawTerminal() error {
	if cli.isTerminalIn && os.Getenv("NORAW") == "" {
		state, err := term.SetRawTerminal(cli.inFd)
		if err != nil {
			return err
		}
		cli.state = state
	}
	return nil
}

func (cli *DockerCli) restoreTerminal(in io.Closer) error {
	if cli.state != nil {
		term.RestoreTerminal(cli.inFd, cli.state)
	}
	// WARNING: DO NOT REMOVE THE OS CHECK !!!
	// For some reason this Close call blocks on darwin..
	// As the client exists right after, simply discard the close
	// until we find a better solution.
	if in != nil && runtime.GOOS != "darwin" {
		return in.Close()
	}
	return nil
}

// NewDockerCli returns a DockerCli instance with IO output and error streams set by in, out and err.
// The key file, protocol (i.e. unix) and address are passed in as strings, along with the tls.Config. If the tls.Config
// is set the client scheme will be set to https.
// The client will be given a 32-second timeout (see https://github.com/docker/docker/pull/8035).
func NewDockerCli(in io.ReadCloser, out, err io.Writer, clientFlags *cli.ClientFlags) *DockerCli {
	cli := &DockerCli{
		in:      in,
		out:     out,
		err:     err,
		keyFile: clientFlags.Common.TrustKey,
	}

	cli.init = func() error {
		clientFlags.PostParse()
		configFile, e := cliconfig.Load(cliconfig.ConfigDir())
		if e != nil {
			fmt.Fprintf(cli.err, "WARNING: Error loading config file:%v\n", e)
		}
		if !configFile.ContainsAuth() {
			credentials.DetectDefaultStore(configFile)
		}
		cli.configFile = configFile

		host, err := getServerHost(clientFlags.Common.Hosts, clientFlags.Common.TLSOptions)
		if err != nil {
			return err
		}

		customHeaders := cli.configFile.HTTPHeaders
		if customHeaders == nil {
			customHeaders = map[string]string{}
		}
		customHeaders["User-Agent"] = clientUserAgent()

		verStr := api.DefaultVersion.String()
		if tmpStr := os.Getenv("DOCKER_API_VERSION"); tmpStr != "" {
			verStr = tmpStr
		}

		httpClient, err := newHTTPClient(host, clientFlags.Common.TLSOptions)
		if err != nil {
			return err
		}

		client, err := client.NewClient(host, verStr, httpClient, customHeaders)
		if err != nil {
			return err
		}
		cli.client = client

		if cli.in != nil {
			cli.inFd, cli.isTerminalIn = term.GetFdInfo(cli.in)
		}
		if cli.out != nil {
			cli.outFd, cli.isTerminalOut = term.GetFdInfo(cli.out)
		}

		return nil
	}

	return cli
}

func getServerHost(hosts []string, tlsOptions *tlsconfig.Options) (host string, err error) {
	switch len(hosts) {
	case 0:
		host = os.Getenv("DOCKER_HOST")
	case 1:
		host = hosts[0]
	default:
		return "", errors.New("Please specify only one -H")
	}

	host, err = opts.ParseHost(tlsOptions != nil, host)
	return
}

func newHTTPClient(host string, tlsOptions *tlsconfig.Options) (*http.Client, error) {
	if tlsOptions == nil {
		// let the api client configure the default transport.
		return nil, nil
	}

	config, err := tlsconfig.Client(*tlsOptions)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: config,
	}
	proto, addr, _, err := client.ParseHost(host)
	if err != nil {
		return nil, err
	}

	sockets.ConfigureTransport(tr, proto, addr)

	return &http.Client{
		Transport: tr,
	}, nil
}

func clientUserAgent() string {
	return "Docker-Client/" + dockerversion.Version + " (" + runtime.GOOS + ")"
}
`


var client_go_scope = godebug.EnteringNewFile(client_pkg_scope, client_go_contents)

var client_go_contents = `// Package client provides a command-line interface for Docker.
//
// Run "docker help SUBCOMMAND" or "docker SUBCOMMAND --help" to see more information on any Docker subcommand, including the full list of options supported for the subcommand.
// See https://docs.docker.com/installation/ for instructions on installing Docker.
package client
`


var commit_go_scope = godebug.EnteringNewFile(client_pkg_scope, commit_go_contents)

func (cli *DockerCli) CmdCommit(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdCommit(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := commit_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 22)
	cmd := Cli.Subcmd("commit", []string{"CONTAINER [REPOSITORY[:TAG]]"}, Cli.DockerCommands["commit"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 23)
	flPause := cmd.Bool([]string{"p", "-pause"}, true, "Pause container during commit")
	scope.Declare("flPause", &flPause)
	godebug.Line(ctx, scope, 24)
	flComment := cmd.String([]string{"m", "-message"}, "", "Commit message")
	scope.Declare("flComment", &flComment)
	godebug.Line(ctx, scope, 25)
	flAuthor := cmd.String([]string{"a", "-author"}, "", "Author (e.g., \"John Hannibal Smith <hannibal@a-team.com>\")")
	scope.Declare("flAuthor", &flAuthor)
	godebug.Line(ctx, scope, 26)
	flChanges := opts.NewListOpts(nil)
	scope.Declare("flChanges", &flChanges)
	godebug.Line(ctx, scope, 27)
	cmd.Var(&flChanges, []string{"c", "-change"}, "Apply Dockerfile instruction to the created image")
	godebug.Line(ctx, scope, 29)

	flConfig := cmd.String([]string{"#-run"}, "", "This option is deprecated and will be removed in a future version in favor of inline Dockerfile-compatible commands")
	scope.Declare("flConfig", &flConfig)
	godebug.Line(ctx, scope, 30)
	cmd.Require(flag.Max, 2)
	godebug.Line(ctx, scope, 31)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 33)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 35)

	var (
		name             = cmd.Arg(0)
		repositoryAndTag = cmd.Arg(1)
		repositoryName   string
		tag              string
	)
	scope.Declare("name", &name, "repositoryAndTag", &repositoryAndTag, "repositoryName", &repositoryName, "tag", &tag)
	godebug.Line(ctx, scope, 43)

	if repositoryAndTag != "" {
		godebug.Line(ctx, scope, 44)
		ref, err := reference.ParseNamed(repositoryAndTag)
		scope := scope.EnteringNewChildScope()
		scope.Declare("ref", &ref, "err", &err)
		godebug.Line(ctx, scope, 45)
		if err != nil {
			godebug.Line(ctx, scope, 46)
			return err
		}
		godebug.Line(ctx, scope, 49)

		repositoryName = ref.Name()
		godebug.Line(ctx, scope, 51)

		switch x := ref.(type) {
		case reference.Canonical:
			godebug.Line(ctx, scope, 52)
			godebug.Line(ctx, scope, 53)
			return errors.New("cannot commit to digest reference")
		case reference.NamedTagged:
			godebug.Line(ctx, scope, 54)
			godebug.Line(ctx, scope, 55)
			tag = x.Tag()
		}
	}
	godebug.Line(ctx, scope, 59)

	var config *container.Config
	scope.Declare("config", &config)
	godebug.Line(ctx, scope, 60)
	if *flConfig != "" {
		godebug.Line(ctx, scope, 61)
		config = &container.Config{}
		godebug.Line(ctx, scope, 62)
		if err := json.Unmarshal([]byte(*flConfig), config); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 63)
			return err
		}
	}
	godebug.Line(ctx, scope, 67)

	options := types.ContainerCommitOptions{
		ContainerID:    name,
		RepositoryName: repositoryName,
		Tag:            tag,
		Comment:        *flComment,
		Author:         *flAuthor,
		Changes:        flChanges.GetAll(),
		Pause:          *flPause,
		Config:         config,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 78)

	response, err := cli.client.ContainerCommit(context.Background(), options)
	scope.Declare("response", &response, "err", &err)
	godebug.Line(ctx, scope, 79)
	if err != nil {
		godebug.Line(ctx, scope, 80)
		return err
	}
	godebug.Line(ctx, scope, 83)

	fmt.Fprintln(cli.out, response.ID)
	godebug.Line(ctx, scope, 84)
	return nil
}

var commit_go_contents = `package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/reference"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
)

// CmdCommit creates a new image from a container's changes.
//
// Usage: docker commit [OPTIONS] CONTAINER [REPOSITORY[:TAG]]
func (cli *DockerCli) CmdCommit(args ...string) error {
	cmd := Cli.Subcmd("commit", []string{"CONTAINER [REPOSITORY[:TAG]]"}, Cli.DockerCommands["commit"].Description, true)
	flPause := cmd.Bool([]string{"p", "-pause"}, true, "Pause container during commit")
	flComment := cmd.String([]string{"m", "-message"}, "", "Commit message")
	flAuthor := cmd.String([]string{"a", "-author"}, "", "Author (e.g., \"John Hannibal Smith <hannibal@a-team.com>\")")
	flChanges := opts.NewListOpts(nil)
	cmd.Var(&flChanges, []string{"c", "-change"}, "Apply Dockerfile instruction to the created image")
	// FIXME: --run is deprecated, it will be replaced with inline Dockerfile commands.
	flConfig := cmd.String([]string{"#-run"}, "", "This option is deprecated and will be removed in a future version in favor of inline Dockerfile-compatible commands")
	cmd.Require(flag.Max, 2)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var (
		name             = cmd.Arg(0)
		repositoryAndTag = cmd.Arg(1)
		repositoryName   string
		tag              string
	)

	//Check if the given image name can be resolved
	if repositoryAndTag != "" {
		ref, err := reference.ParseNamed(repositoryAndTag)
		if err != nil {
			return err
		}

		repositoryName = ref.Name()

		switch x := ref.(type) {
		case reference.Canonical:
			return errors.New("cannot commit to digest reference")
		case reference.NamedTagged:
			tag = x.Tag()
		}
	}

	var config *container.Config
	if *flConfig != "" {
		config = &container.Config{}
		if err := json.Unmarshal([]byte(*flConfig), config); err != nil {
			return err
		}
	}

	options := types.ContainerCommitOptions{
		ContainerID:    name,
		RepositoryName: repositoryName,
		Tag:            tag,
		Comment:        *flComment,
		Author:         *flAuthor,
		Changes:        flChanges.GetAll(),
		Pause:          *flPause,
		Config:         config,
	}

	response, err := cli.client.ContainerCommit(context.Background(), options)
	if err != nil {
		return err
	}

	fmt.Fprintln(cli.out, response.ID)
	return nil
}
`


var cp_go_scope = godebug.EnteringNewFile(client_pkg_scope, cp_go_contents)

type copyDirection int

const (
	fromContainer copyDirection = (1 << iota)
	toContainer
	acrossContainers = fromContainer | toContainer
)

type cpConfig struct {
	followLink bool
}

func (cli *DockerCli) CmdCp(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdCp(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := cp_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 44)
	cmd := Cli.Subcmd(
		"cp",
		[]string{"CONTAINER:SRC_PATH DEST_PATH|-", "SRC_PATH|- CONTAINER:DEST_PATH"},
		strings.Join([]string{
			Cli.DockerCommands["cp"].Description,
			"\nUse '-' as the source to read a tar archive from stdin\n",
			"and extract it to a directory destination in a container.\n",
			"Use '-' as the destination to stream a tar archive of a\n",
			"container source to stdout.",
		}, ""),
		true,
	)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 57)

	followLink := cmd.Bool([]string{"L", "-follow-link"}, false, "Always follow symbol link in SRC_PATH")
	scope.Declare("followLink", &followLink)
	godebug.Line(ctx, scope, 59)

	cmd.Require(flag.Exact, 2)
	godebug.Line(ctx, scope, 60)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 62)

	if cmd.Arg(0) == "" {
		godebug.Line(ctx, scope, 63)
		return fmt.Errorf("source can not be empty")
	}
	godebug.Line(ctx, scope, 65)
	if cmd.Arg(1) == "" {
		godebug.Line(ctx, scope, 66)
		return fmt.Errorf("destination can not be empty")
	}
	godebug.Line(ctx, scope, 69)

	srcContainer, srcPath := splitCpArg(cmd.Arg(0))
	scope.Declare("srcContainer", &srcContainer, "srcPath", &srcPath)
	godebug.Line(ctx, scope, 70)
	dstContainer, dstPath := splitCpArg(cmd.Arg(1))
	scope.Declare("dstContainer", &dstContainer, "dstPath", &dstPath)
	godebug.Line(ctx, scope, 72)

	var direction copyDirection
	scope.Declare("direction", &direction)
	godebug.Line(ctx, scope, 73)
	if srcContainer != "" {
		godebug.Line(ctx, scope, 74)
		direction |= fromContainer
	}
	godebug.Line(ctx, scope, 76)
	if dstContainer != "" {
		godebug.Line(ctx, scope, 77)
		direction |= toContainer
	}
	godebug.Line(ctx, scope, 80)

	cpParam := &cpConfig{
		followLink: *followLink,
	}
	scope.Declare("cpParam", &cpParam)
	godebug.Line(ctx, scope, 84)

	switch direction {
	case godebug.Case(ctx, scope, 85):
		fallthrough
	case fromContainer:
		godebug.Line(ctx, scope, 86)
		return cli.copyFromContainer(srcContainer, srcPath, dstPath, cpParam)
	case godebug.Case(ctx, scope, 87):
		fallthrough
	case toContainer:
		godebug.Line(ctx, scope, 88)
		return cli.copyToContainer(srcPath, dstContainer, dstPath, cpParam)
	case godebug.Case(ctx, scope, 89):
		fallthrough
	case acrossContainers:
		godebug.Line(ctx, scope, 91)

		return fmt.Errorf("copying between containers is not supported")
	default:
		godebug.Line(ctx, scope, 92)
		godebug.Line(ctx, scope, 94)

		return fmt.Errorf("must specify at least one container source")
	}
}

func splitCpArg(arg string) (container, path string) {
	ctx, ok := godebug.EnterFunc(func() {
		container, path = splitCpArg(arg)
	})
	if !ok {
		return container, path
	}
	defer godebug.ExitFunc(ctx)
	scope := cp_go_scope.EnteringNewChildScope()
	scope.Declare("arg", &arg, "container", &container, "path", &path)
	godebug.Line(ctx, scope, 113)
	if system.IsAbs(arg) {
		godebug.Line(ctx, scope, 115)

		return "", arg
	}
	godebug.Line(ctx, scope, 118)

	parts := strings.SplitN(arg, ":", 2)
	scope.Declare("parts", &parts)
	godebug.Line(ctx, scope, 120)

	if len(parts) == 1 || strings.HasPrefix(parts[0], ".") {
		godebug.Line(ctx, scope, 123)

		return "", arg
	}
	godebug.Line(ctx, scope, 126)

	return parts[0], parts[1]
}

func (cli *DockerCli) statContainerPath(containerName, path string) (types.ContainerPathStat, error) {
	var result1 types.ContainerPathStat
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = cli.statContainerPath(containerName, path)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := cp_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "containerName", &containerName, "path", &path)
	godebug.Line(ctx, scope, 130)
	return cli.client.ContainerStatPath(context.Background(), containerName, path)
}

func resolveLocalPath(localPath string) (absPath string, err error) {
	ctx, ok := godebug.EnterFunc(func() {
		absPath, err = resolveLocalPath(localPath)
	})
	if !ok {
		return absPath, err
	}
	defer godebug.ExitFunc(ctx)
	scope := cp_go_scope.EnteringNewChildScope()
	scope.Declare("localPath", &localPath, "absPath", &absPath, "err", &err)
	godebug.Line(ctx, scope, 134)
	if absPath, err = filepath.Abs(localPath); err != nil {
		godebug.Line(ctx, scope, 135)
		return
	}
	godebug.Line(ctx, scope, 138)

	return archive.PreserveTrailingDotOrSeparator(absPath, localPath), nil
}

func (cli *DockerCli) copyFromContainer(srcContainer, srcPath, dstPath string, cpParam *cpConfig) (err error) {
	ctx, ok := godebug.EnterFunc(func() {
		err = cli.copyFromContainer(srcContainer, srcPath, dstPath, cpParam)
	})
	if !ok {
		return err
	}
	defer godebug.ExitFunc(ctx)
	scope := cp_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "srcContainer", &srcContainer, "srcPath", &srcPath, "dstPath", &dstPath, "cpParam", &cpParam, "err", &err)
	godebug.Line(ctx, scope, 142)
	if dstPath != "-" {
		godebug.Line(ctx, scope, 144)

		dstPath, err = resolveLocalPath(dstPath)
		godebug.Line(ctx, scope, 145)
		if err != nil {
			godebug.Line(ctx, scope, 146)
			return err
		}
	}
	godebug.Line(ctx, scope, 151)

	var rebaseName string
	scope.Declare("rebaseName", &rebaseName)
	godebug.Line(ctx, scope, 152)
	if cpParam.followLink {
		godebug.Line(ctx, scope, 153)
		srcStat, err := cli.statContainerPath(srcContainer, srcPath)
		scope := scope.EnteringNewChildScope()
		scope.Declare("srcStat", &srcStat, "err", &err)
		godebug.Line(ctx, scope, 156)

		if err == nil && srcStat.Mode&os.ModeSymlink != 0 {
			godebug.Line(ctx, scope, 157)
			linkTarget := srcStat.LinkTarget
			scope := scope.EnteringNewChildScope()
			scope.Declare("linkTarget", &linkTarget)
			godebug.Line(ctx, scope, 158)
			if !system.IsAbs(linkTarget) {
				godebug.Line(ctx, scope, 160)

				srcParent, _ := archive.SplitPathDirEntry(srcPath)
				scope := scope.EnteringNewChildScope()
				scope.Declare("srcParent", &srcParent)
				godebug.Line(ctx, scope, 161)
				linkTarget = filepath.Join(srcParent, linkTarget)
			}
			godebug.Line(ctx, scope, 164)

			linkTarget, rebaseName = archive.GetRebaseName(srcPath, linkTarget)
			godebug.Line(ctx, scope, 165)
			srcPath = linkTarget
		}

	}
	godebug.Line(ctx, scope, 170)

	content, stat, err := cli.client.CopyFromContainer(context.Background(), srcContainer, srcPath)
	scope.Declare("content", &content, "stat", &stat)
	godebug.Line(ctx, scope, 171)
	if err != nil {
		godebug.Line(ctx, scope, 172)
		return err
	}
	godebug.Line(ctx, scope, 174)
	defer content.Close()
	defer godebug.Defer(ctx, scope, 174)
	godebug.Line(ctx, scope, 176)

	if dstPath == "-" {
		godebug.Line(ctx, scope, 178)

		_, err = io.Copy(os.Stdout, content)
		godebug.Line(ctx, scope, 180)

		return err
	}
	godebug.Line(ctx, scope, 184)

	srcInfo := archive.CopyInfo{
		Path:       srcPath,
		Exists:     true,
		IsDir:      stat.Mode.IsDir(),
		RebaseName: rebaseName,
	}
	scope.Declare("srcInfo", &srcInfo)
	godebug.Line(ctx, scope, 191)

	preArchive := content
	scope.Declare("preArchive", &preArchive)
	godebug.Line(ctx, scope, 192)
	if len(srcInfo.RebaseName) != 0 {
		godebug.Line(ctx, scope, 193)
		_, srcBase := archive.SplitPathDirEntry(srcInfo.Path)
		scope := scope.EnteringNewChildScope()
		scope.Declare("srcBase", &srcBase)
		godebug.Line(ctx, scope, 194)
		preArchive = archive.RebaseArchiveEntries(content, srcBase, srcInfo.RebaseName)
	}
	godebug.Line(ctx, scope, 199)

	return archive.CopyTo(preArchive, srcInfo, dstPath)
}

func (cli *DockerCli) copyToContainer(srcPath, dstContainer, dstPath string, cpParam *cpConfig) (err error) {
	ctx, ok := godebug.EnterFunc(func() {
		err = cli.copyToContainer(srcPath, dstContainer, dstPath, cpParam)
	})
	if !ok {
		return err
	}
	defer godebug.ExitFunc(ctx)
	scope := cp_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "srcPath", &srcPath, "dstContainer", &dstContainer, "dstPath", &dstPath, "cpParam", &cpParam, "err", &err)
	godebug.Line(ctx, scope, 203)
	if srcPath != "-" {
		godebug.Line(ctx, scope, 205)

		srcPath, err = resolveLocalPath(srcPath)
		godebug.Line(ctx, scope, 206)
		if err != nil {
			godebug.Line(ctx, scope, 207)
			return err
		}
	}
	godebug.Line(ctx, scope, 217)

	dstInfo := archive.CopyInfo{Path: dstPath}
	scope.Declare("dstInfo", &dstInfo)
	godebug.Line(ctx, scope, 218)
	dstStat, err := cli.statContainerPath(dstContainer, dstPath)
	scope.Declare("dstStat", &dstStat)
	godebug.Line(ctx, scope, 221)

	if err == nil && dstStat.Mode&os.ModeSymlink != 0 {
		godebug.Line(ctx, scope, 222)
		linkTarget := dstStat.LinkTarget
		scope := scope.EnteringNewChildScope()
		scope.Declare("linkTarget", &linkTarget)
		godebug.Line(ctx, scope, 223)
		if !system.IsAbs(linkTarget) {
			godebug.Line(ctx, scope, 225)

			dstParent, _ := archive.SplitPathDirEntry(dstPath)
			scope := scope.EnteringNewChildScope()
			scope.Declare("dstParent", &dstParent)
			godebug.Line(ctx, scope, 226)
			linkTarget = filepath.Join(dstParent, linkTarget)
		}
		godebug.Line(ctx, scope, 229)

		dstInfo.Path = linkTarget
		godebug.Line(ctx, scope, 230)
		dstStat, err = cli.statContainerPath(dstContainer, linkTarget)
	}
	godebug.Line(ctx, scope, 239)

	if err == nil {
		godebug.Line(ctx, scope, 240)
		dstInfo.Exists, dstInfo.IsDir = true, dstStat.Mode.IsDir()
	}
	godebug.Line(ctx, scope, 243)

	var (
		content         io.Reader
		resolvedDstPath string
	)
	scope.Declare("content", &content, "resolvedDstPath", &resolvedDstPath)
	godebug.Line(ctx, scope, 248)

	if srcPath == "-" {
		godebug.Line(ctx, scope, 250)

		content = os.Stdin
		godebug.Line(ctx, scope, 251)
		resolvedDstPath = dstInfo.Path
		godebug.Line(ctx, scope, 252)
		if !dstInfo.IsDir {
			godebug.Line(ctx, scope, 253)
			return fmt.Errorf("destination %q must be a directory", fmt.Sprintf("%s:%s", dstContainer, dstPath))
		}
	} else {
		godebug.Line(ctx, scope, 255)
		godebug.Line(ctx, scope, 257)

		srcInfo, err := archive.CopyInfoSourcePath(srcPath, cpParam.followLink)
		scope := scope.EnteringNewChildScope()
		scope.Declare("srcInfo", &srcInfo, "err", &err)
		godebug.Line(ctx, scope, 258)
		if err != nil {
			godebug.Line(ctx, scope, 259)
			return err
		}
		godebug.Line(ctx, scope, 262)

		srcArchive, err := archive.TarResource(srcInfo)
		scope.Declare("srcArchive", &srcArchive)
		godebug.Line(ctx, scope, 263)
		if err != nil {
			godebug.Line(ctx, scope, 264)
			return err
		}
		godebug.Line(ctx, scope, 266)
		defer srcArchive.Close()
		defer godebug.Defer(ctx, scope, 266)
		godebug.Line(ctx, scope, 280)

		dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
		scope.Declare("dstDir", &dstDir, "preparedArchive", &preparedArchive)
		godebug.Line(ctx, scope, 281)
		if err != nil {
			godebug.Line(ctx, scope, 282)
			return err
		}
		godebug.Line(ctx, scope, 284)
		defer preparedArchive.Close()
		defer godebug.Defer(ctx, scope, 284)
		godebug.Line(ctx, scope, 286)

		resolvedDstPath = dstDir
		godebug.Line(ctx, scope, 287)
		content = preparedArchive
	}
	godebug.Line(ctx, scope, 290)

	options := types.CopyToContainerOptions{
		ContainerID:               dstContainer,
		Path:                      resolvedDstPath,
		Content:                   content,
		AllowOverwriteDirWithFile: false,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 297)

	return cli.client.CopyToContainer(context.Background(), options)
}

var cp_go_contents = `package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/archive"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/engine-api/types"
)

type copyDirection int

const (
	fromContainer copyDirection = (1 << iota)
	toContainer
	acrossContainers = fromContainer | toContainer
)

type cpConfig struct {
	followLink bool
}

// CmdCp copies files/folders to or from a path in a container.
//
// When copying from a container, if DEST_PATH is '-' the data is written as a
// tar archive file to STDOUT.
//
// When copying to a container, if SRC_PATH is '-' the data is read as a tar
// archive file from STDIN, and the destination CONTAINER:DEST_PATH, must specify
// a directory.
//
// Usage:
// 	docker cp CONTAINER:SRC_PATH DEST_PATH|-
// 	docker cp SRC_PATH|- CONTAINER:DEST_PATH
func (cli *DockerCli) CmdCp(args ...string) error {
	cmd := Cli.Subcmd(
		"cp",
		[]string{"CONTAINER:SRC_PATH DEST_PATH|-", "SRC_PATH|- CONTAINER:DEST_PATH"},
		strings.Join([]string{
			Cli.DockerCommands["cp"].Description,
			"\nUse '-' as the source to read a tar archive from stdin\n",
			"and extract it to a directory destination in a container.\n",
			"Use '-' as the destination to stream a tar archive of a\n",
			"container source to stdout.",
		}, ""),
		true,
	)

	followLink := cmd.Bool([]string{"L", "-follow-link"}, false, "Always follow symbol link in SRC_PATH")

	cmd.Require(flag.Exact, 2)
	cmd.ParseFlags(args, true)

	if cmd.Arg(0) == "" {
		return fmt.Errorf("source can not be empty")
	}
	if cmd.Arg(1) == "" {
		return fmt.Errorf("destination can not be empty")
	}

	srcContainer, srcPath := splitCpArg(cmd.Arg(0))
	dstContainer, dstPath := splitCpArg(cmd.Arg(1))

	var direction copyDirection
	if srcContainer != "" {
		direction |= fromContainer
	}
	if dstContainer != "" {
		direction |= toContainer
	}

	cpParam := &cpConfig{
		followLink: *followLink,
	}

	switch direction {
	case fromContainer:
		return cli.copyFromContainer(srcContainer, srcPath, dstPath, cpParam)
	case toContainer:
		return cli.copyToContainer(srcPath, dstContainer, dstPath, cpParam)
	case acrossContainers:
		// Copying between containers isn't supported.
		return fmt.Errorf("copying between containers is not supported")
	default:
		// User didn't specify any container.
		return fmt.Errorf("must specify at least one container source")
	}
}

// We use ` + "`" + `:` + "`" + ` as a delimiter between CONTAINER and PATH, but ` + "`" + `:` + "`" + ` could also be
// in a valid LOCALPATH, like ` + "`" + `file:name.txt` + "`" + `. We can resolve this ambiguity by
// requiring a LOCALPATH with a ` + "`" + `:` + "`" + ` to be made explicit with a relative or
// absolute path:
// 	` + "`" + `/path/to/file:name.txt` + "`" + ` or ` + "`" + `./file:name.txt` + "`" + `
//
// This is apparently how ` + "`" + `scp` + "`" + ` handles this as well:
// 	http://www.cyberciti.biz/faq/rsync-scp-file-name-with-colon-punctuation-in-it/
//
// We can't simply check for a filepath separator because container names may
// have a separator, e.g., "host0/cname1" if container is in a Docker cluster,
// so we have to check for a ` + "`" + `/` + "`" + ` or ` + "`" + `.` + "`" + ` prefix. Also, in the case of a Windows
// client, a ` + "`" + `:` + "`" + ` could be part of an absolute Windows path, in which case it
// is immediately proceeded by a backslash.
func splitCpArg(arg string) (container, path string) {
	if system.IsAbs(arg) {
		// Explicit local absolute path, e.g., ` + "`" + `C:\foo` + "`" + ` or ` + "`" + `/foo` + "`" + `.
		return "", arg
	}

	parts := strings.SplitN(arg, ":", 2)

	if len(parts) == 1 || strings.HasPrefix(parts[0], ".") {
		// Either there's no ` + "`" + `:` + "`" + ` in the arg
		// OR it's an explicit local relative path like ` + "`" + `./file:name.txt` + "`" + `.
		return "", arg
	}

	return parts[0], parts[1]
}

func (cli *DockerCli) statContainerPath(containerName, path string) (types.ContainerPathStat, error) {
	return cli.client.ContainerStatPath(context.Background(), containerName, path)
}

func resolveLocalPath(localPath string) (absPath string, err error) {
	if absPath, err = filepath.Abs(localPath); err != nil {
		return
	}

	return archive.PreserveTrailingDotOrSeparator(absPath, localPath), nil
}

func (cli *DockerCli) copyFromContainer(srcContainer, srcPath, dstPath string, cpParam *cpConfig) (err error) {
	if dstPath != "-" {
		// Get an absolute destination path.
		dstPath, err = resolveLocalPath(dstPath)
		if err != nil {
			return err
		}
	}

	// if client requests to follow symbol link, then must decide target file to be copied
	var rebaseName string
	if cpParam.followLink {
		srcStat, err := cli.statContainerPath(srcContainer, srcPath)

		// If the destination is a symbolic link, we should follow it.
		if err == nil && srcStat.Mode&os.ModeSymlink != 0 {
			linkTarget := srcStat.LinkTarget
			if !system.IsAbs(linkTarget) {
				// Join with the parent directory.
				srcParent, _ := archive.SplitPathDirEntry(srcPath)
				linkTarget = filepath.Join(srcParent, linkTarget)
			}

			linkTarget, rebaseName = archive.GetRebaseName(srcPath, linkTarget)
			srcPath = linkTarget
		}

	}

	content, stat, err := cli.client.CopyFromContainer(context.Background(), srcContainer, srcPath)
	if err != nil {
		return err
	}
	defer content.Close()

	if dstPath == "-" {
		// Send the response to STDOUT.
		_, err = io.Copy(os.Stdout, content)

		return err
	}

	// Prepare source copy info.
	srcInfo := archive.CopyInfo{
		Path:       srcPath,
		Exists:     true,
		IsDir:      stat.Mode.IsDir(),
		RebaseName: rebaseName,
	}

	preArchive := content
	if len(srcInfo.RebaseName) != 0 {
		_, srcBase := archive.SplitPathDirEntry(srcInfo.Path)
		preArchive = archive.RebaseArchiveEntries(content, srcBase, srcInfo.RebaseName)
	}
	// See comments in the implementation of ` + "`" + `archive.CopyTo` + "`" + ` for exactly what
	// goes into deciding how and whether the source archive needs to be
	// altered for the correct copy behavior.
	return archive.CopyTo(preArchive, srcInfo, dstPath)
}

func (cli *DockerCli) copyToContainer(srcPath, dstContainer, dstPath string, cpParam *cpConfig) (err error) {
	if srcPath != "-" {
		// Get an absolute source path.
		srcPath, err = resolveLocalPath(srcPath)
		if err != nil {
			return err
		}
	}

	// In order to get the copy behavior right, we need to know information
	// about both the source and destination. The API is a simple tar
	// archive/extract API but we can use the stat info header about the
	// destination to be more informed about exactly what the destination is.

	// Prepare destination copy info by stat-ing the container path.
	dstInfo := archive.CopyInfo{Path: dstPath}
	dstStat, err := cli.statContainerPath(dstContainer, dstPath)

	// If the destination is a symbolic link, we should evaluate it.
	if err == nil && dstStat.Mode&os.ModeSymlink != 0 {
		linkTarget := dstStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// Join with the parent directory.
			dstParent, _ := archive.SplitPathDirEntry(dstPath)
			linkTarget = filepath.Join(dstParent, linkTarget)
		}

		dstInfo.Path = linkTarget
		dstStat, err = cli.statContainerPath(dstContainer, linkTarget)
	}

	// Ignore any error and assume that the parent directory of the destination
	// path exists, in which case the copy may still succeed. If there is any
	// type of conflict (e.g., non-directory overwriting an existing directory
	// or vice versa) the extraction will fail. If the destination simply did
	// not exist, but the parent directory does, the extraction will still
	// succeed.
	if err == nil {
		dstInfo.Exists, dstInfo.IsDir = true, dstStat.Mode.IsDir()
	}

	var (
		content         io.Reader
		resolvedDstPath string
	)

	if srcPath == "-" {
		// Use STDIN.
		content = os.Stdin
		resolvedDstPath = dstInfo.Path
		if !dstInfo.IsDir {
			return fmt.Errorf("destination %q must be a directory", fmt.Sprintf("%s:%s", dstContainer, dstPath))
		}
	} else {
		// Prepare source copy info.
		srcInfo, err := archive.CopyInfoSourcePath(srcPath, cpParam.followLink)
		if err != nil {
			return err
		}

		srcArchive, err := archive.TarResource(srcInfo)
		if err != nil {
			return err
		}
		defer srcArchive.Close()

		// With the stat info about the local source as well as the
		// destination, we have enough information to know whether we need to
		// alter the archive that we upload so that when the server extracts
		// it to the specified directory in the container we get the desired
		// copy behavior.

		// See comments in the implementation of ` + "`" + `archive.PrepareArchiveCopy` + "`" + `
		// for exactly what goes into deciding how and whether the source
		// archive needs to be altered for the correct copy behavior when it is
		// extracted. This function also infers from the source and destination
		// info which directory to extract to, which may be the parent of the
		// destination that the user specified.
		dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
		if err != nil {
			return err
		}
		defer preparedArchive.Close()

		resolvedDstPath = dstDir
		content = preparedArchive
	}

	options := types.CopyToContainerOptions{
		ContainerID:               dstContainer,
		Path:                      resolvedDstPath,
		Content:                   content,
		AllowOverwriteDirWithFile: false,
	}

	return cli.client.CopyToContainer(context.Background(), options)
}
`


var create_go_scope = godebug.EnteringNewFile(client_pkg_scope, create_go_contents)

func (cli *DockerCli) pullImage(image string, out io.Writer) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.pullImage(image, out)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := create_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "image", &image, "out", &out)
	godebug.Line(ctx, scope, 22)
	ref, err := reference.ParseNamed(image)
	scope.Declare("ref", &ref, "err", &err)
	godebug.Line(ctx, scope, 23)
	if err != nil {
		godebug.Line(ctx, scope, 24)
		return err
	}
	godebug.Line(ctx, scope, 27)

	var tag string
	scope.Declare("tag", &tag)
	godebug.Line(ctx, scope, 28)
	switch x := reference.WithDefaultTag(ref).(type) {
	case reference.Canonical:
		godebug.Line(ctx, scope, 29)
		godebug.Line(ctx, scope, 30)
		tag = x.Digest().String()
	case reference.NamedTagged:
		godebug.Line(ctx, scope, 31)
		godebug.Line(ctx, scope, 32)
		tag = x.Tag()
	}
	godebug.Line(ctx, scope, 36)

	repoInfo, err := registry.ParseRepositoryInfo(ref)
	scope.Declare("repoInfo", &repoInfo)
	godebug.Line(ctx, scope, 37)
	if err != nil {
		godebug.Line(ctx, scope, 38)
		return err
	}
	godebug.Line(ctx, scope, 41)

	authConfig := cli.resolveAuthConfig(repoInfo.Index)
	scope.Declare("authConfig", &authConfig)
	godebug.Line(ctx, scope, 42)
	encodedAuth, err := encodeAuthToBase64(authConfig)
	scope.Declare("encodedAuth", &encodedAuth)
	godebug.Line(ctx, scope, 43)
	if err != nil {
		godebug.Line(ctx, scope, 44)
		return err
	}
	godebug.Line(ctx, scope, 47)

	options := types.ImageCreateOptions{
		Parent:       ref.Name(),
		Tag:          tag,
		RegistryAuth: encodedAuth,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 53)

	responseBody, err := cli.client.ImageCreate(context.Background(), options)
	scope.Declare("responseBody", &responseBody)
	godebug.Line(ctx, scope, 54)
	if err != nil {
		godebug.Line(ctx, scope, 55)
		return err
	}
	godebug.Line(ctx, scope, 57)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 57)
	godebug.Line(ctx, scope, 59)

	return jsonmessage.DisplayJSONMessagesStream(responseBody, out, cli.outFd, cli.isTerminalOut, nil)
}

type cidFile struct {
	path    string
	file    *os.File
	written bool
}

func newCIDFile(path string) (*cidFile, error) {
	var result1 *cidFile
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = newCIDFile(path)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := create_go_scope.EnteringNewChildScope()
	scope.Declare("path", &path)
	godebug.Line(ctx, scope, 69)
	if _, err := os.Stat(path); err == nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 70)
		return nil, fmt.Errorf("Container ID file found, make sure the other container isn't running or delete %s", path)
	}
	godebug.Line(ctx, scope, 73)

	f, err := os.Create(path)
	scope.Declare("f", &f, "err", &err)
	godebug.Line(ctx, scope, 74)
	if err != nil {
		godebug.Line(ctx, scope, 75)
		return nil, fmt.Errorf("Failed to create the container ID file: %s", err)
	}
	godebug.Line(ctx, scope, 78)

	return &cidFile{path: path, file: f}, nil
}

func (cli *DockerCli) createContainer(config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, cidfile, name string) (*types.ContainerCreateResponse, error) {
	var result1 *types.ContainerCreateResponse
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = cli.createContainer(config, hostConfig, networkingConfig, cidfile, name)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := create_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "config", &config, "hostConfig", &hostConfig, "networkingConfig", &networkingConfig, "cidfile", &cidfile, "name", &name)
	godebug.Line(ctx, scope, 82)
	var containerIDFile *cidFile
	scope.Declare("containerIDFile", &containerIDFile)
	godebug.Line(ctx, scope, 83)
	if cidfile != "" {
		godebug.Line(ctx, scope, 84)
		var err error
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 85)
		if containerIDFile, err = newCIDFile(cidfile); err != nil {
			godebug.Line(ctx, scope, 86)
			return nil, err
		}
		godebug.Line(ctx, scope, 88)
		defer containerIDFile.Close()
		defer godebug.Defer(ctx, scope, 88)
	}
	godebug.Line(ctx, scope, 91)

	var trustedRef reference.Canonical
	scope.Declare("trustedRef", &trustedRef)
	godebug.Line(ctx, scope, 92)
	_, ref, err := reference.ParseIDOrReference(config.Image)
	scope.Declare("ref", &ref, "err", &err)
	godebug.Line(ctx, scope, 93)
	if err != nil {
		godebug.Line(ctx, scope, 94)
		return nil, err
	}
	godebug.Line(ctx, scope, 96)
	if ref != nil {
		godebug.Line(ctx, scope, 97)
		ref = reference.WithDefaultTag(ref)
		godebug.Line(ctx, scope, 99)

		if ref, ok := ref.(reference.NamedTagged); ok && isTrusted() {
			scope := scope.EnteringNewChildScope()
			scope.Declare("ref", &ref, "ok", &ok)
			godebug.Line(ctx, scope, 100)
			var err error
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 101)
			trustedRef, err = cli.trustedReference(ref)
			godebug.Line(ctx, scope, 102)
			if err != nil {
				godebug.Line(ctx, scope, 103)
				return nil, err
			}
			godebug.Line(ctx, scope, 105)
			config.Image = trustedRef.String()
		}
	}
	godebug.Line(ctx, scope, 110)

	response, err := cli.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
	scope.Declare("response", &response)
	godebug.Line(ctx, scope, 113)

	if err != nil {
		godebug.Line(ctx, scope, 114)
		if apiclient.IsErrImageNotFound(err) && ref != nil {
			godebug.Line(ctx, scope, 115)
			fmt.Fprintf(cli.err, "Unable to find image '%s' locally\n", ref.String())
			godebug.Line(ctx, scope, 118)

			if err = cli.pullImage(config.Image, cli.err); err != nil {
				godebug.Line(ctx, scope, 119)
				return nil, err
			}
			godebug.Line(ctx, scope, 121)
			if ref, ok := ref.(reference.NamedTagged); ok && trustedRef != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("ref", &ref, "ok", &ok)
				godebug.Line(ctx, scope, 122)
				if err := cli.tagTrusted(trustedRef, ref); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(ctx, scope, 123)
					return nil, err
				}
			}
			godebug.Line(ctx, scope, 127)

			var retryErr error
			scope := scope.EnteringNewChildScope()
			scope.Declare("retryErr", &retryErr)
			godebug.Line(ctx, scope, 128)
			response, retryErr = cli.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
			godebug.Line(ctx, scope, 129)
			if retryErr != nil {
				godebug.Line(ctx, scope, 130)
				return nil, retryErr
			}
		} else {
			godebug.Line(ctx, scope, 132)
			godebug.Line(ctx, scope, 133)
			return nil, err
		}
	}
	{
		scope := scope.EnteringNewChildScope()

		for _, warning := range response.Warnings {
			godebug.Line(ctx, scope, 137)
			scope.Declare("warning", &warning)
			godebug.Line(ctx, scope, 138)
			fmt.Fprintf(cli.err, "WARNING: %s\n", warning)
		}
		godebug.Line(ctx, scope, 137)
	}
	godebug.Line(ctx, scope, 140)
	if containerIDFile != nil {
		godebug.Line(ctx, scope, 141)
		if err = containerIDFile.Write(response.ID); err != nil {
			godebug.Line(ctx, scope, 142)
			return nil, err
		}
	}
	godebug.Line(ctx, scope, 145)
	return &response, nil
}

func (cli *DockerCli) CmdCreate(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdCreate(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := create_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 152)
	cmd := Cli.Subcmd("create", []string{"IMAGE [COMMAND] [ARG...]"}, Cli.DockerCommands["create"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 153)
	addTrustedFlags(cmd, true)
	godebug.Line(ctx, scope, 156)

	var (
		flName = cmd.String([]string{"-name"}, "", "Assign a name to the container")
	)
	scope.Declare("flName", &flName)
	godebug.Line(ctx, scope, 160)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)
	scope.Declare("config", &config, "hostConfig", &hostConfig, "networkingConfig", &networkingConfig, "err", &err)
	godebug.Line(ctx, scope, 162)

	if err != nil {
		godebug.Line(ctx, scope, 163)
		cmd.ReportError(err.Error(), true)
		godebug.Line(ctx, scope, 164)
		os.Exit(1)
	}
	godebug.Line(ctx, scope, 166)
	if config.Image == "" {
		godebug.Line(ctx, scope, 167)
		cmd.Usage()
		godebug.Line(ctx, scope, 168)
		return nil
	}
	godebug.Line(ctx, scope, 170)
	response, err := cli.createContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	scope.Declare("response", &response)
	godebug.Line(ctx, scope, 171)
	if err != nil {
		godebug.Line(ctx, scope, 172)
		return err
	}
	godebug.Line(ctx, scope, 174)
	fmt.Fprintf(cli.out, "%s\n", response.ID)
	godebug.Line(ctx, scope, 175)
	return nil
}

var create_go_contents = `package client

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	networktypes "github.com/docker/engine-api/types/network"
)

func (cli *DockerCli) pullImage(image string, out io.Writer) error {
	ref, err := reference.ParseNamed(image)
	if err != nil {
		return err
	}

	var tag string
	switch x := reference.WithDefaultTag(ref).(type) {
	case reference.Canonical:
		tag = x.Digest().String()
	case reference.NamedTagged:
		tag = x.Tag()
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}

	authConfig := cli.resolveAuthConfig(repoInfo.Index)
	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	options := types.ImageCreateOptions{
		Parent:       ref.Name(),
		Tag:          tag,
		RegistryAuth: encodedAuth,
	}

	responseBody, err := cli.client.ImageCreate(context.Background(), options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, out, cli.outFd, cli.isTerminalOut, nil)
}

type cidFile struct {
	path    string
	file    *os.File
	written bool
}

func newCIDFile(path string) (*cidFile, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("Container ID file found, make sure the other container isn't running or delete %s", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to create the container ID file: %s", err)
	}

	return &cidFile{path: path, file: f}, nil
}

func (cli *DockerCli) createContainer(config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, cidfile, name string) (*types.ContainerCreateResponse, error) {
	var containerIDFile *cidFile
	if cidfile != "" {
		var err error
		if containerIDFile, err = newCIDFile(cidfile); err != nil {
			return nil, err
		}
		defer containerIDFile.Close()
	}

	var trustedRef reference.Canonical
	_, ref, err := reference.ParseIDOrReference(config.Image)
	if err != nil {
		return nil, err
	}
	if ref != nil {
		ref = reference.WithDefaultTag(ref)

		if ref, ok := ref.(reference.NamedTagged); ok && isTrusted() {
			var err error
			trustedRef, err = cli.trustedReference(ref)
			if err != nil {
				return nil, err
			}
			config.Image = trustedRef.String()
		}
	}

	//create the container
	response, err := cli.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)

	//if image not found try to pull it
	if err != nil {
		if client.IsErrImageNotFound(err) && ref != nil {
			fmt.Fprintf(cli.err, "Unable to find image '%s' locally\n", ref.String())

			// we don't want to write to stdout anything apart from container.ID
			if err = cli.pullImage(config.Image, cli.err); err != nil {
				return nil, err
			}
			if ref, ok := ref.(reference.NamedTagged); ok && trustedRef != nil {
				if err := cli.tagTrusted(trustedRef, ref); err != nil {
					return nil, err
				}
			}
			// Retry
			var retryErr error
			response, retryErr = cli.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, name)
			if retryErr != nil {
				return nil, retryErr
			}
		} else {
			return nil, err
		}
	}

	for _, warning := range response.Warnings {
		fmt.Fprintf(cli.err, "WARNING: %s\n", warning)
	}
	if containerIDFile != nil {
		if err = containerIDFile.Write(response.ID); err != nil {
			return nil, err
		}
	}
	return &response, nil
}

// CmdCreate creates a new container from a given image.
//
// Usage: docker create [OPTIONS] IMAGE [COMMAND] [ARG...]
func (cli *DockerCli) CmdCreate(args ...string) error {
	cmd := Cli.Subcmd("create", []string{"IMAGE [COMMAND] [ARG...]"}, Cli.DockerCommands["create"].Description, true)
	addTrustedFlags(cmd, true)

	// These are flags not stored in Config/HostConfig
	var (
		flName = cmd.String([]string{"-name"}, "", "Assign a name to the container")
	)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)

	if err != nil {
		cmd.ReportError(err.Error(), true)
		os.Exit(1)
	}
	if config.Image == "" {
		cmd.Usage()
		return nil
	}
	response, err := cli.createContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s\n", response.ID)
	return nil
}
`


var diff_go_scope = godebug.EnteringNewFile(client_pkg_scope, diff_go_contents)

func (cli *DockerCli) CmdDiff(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdDiff(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := diff_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 21)
	cmd := Cli.Subcmd("diff", []string{"CONTAINER"}, Cli.DockerCommands["diff"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 22)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 24)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 26)

	if cmd.Arg(0) == "" {
		godebug.Line(ctx, scope, 27)
		return fmt.Errorf("Container name cannot be empty")
	}
	godebug.Line(ctx, scope, 30)

	changes, err := cli.client.ContainerDiff(context.Background(), cmd.Arg(0))
	scope.Declare("changes", &changes, "err", &err)
	godebug.Line(ctx, scope, 31)
	if err != nil {
		godebug.Line(ctx, scope, 32)
		return err
	}
	{
		scope := scope.EnteringNewChildScope()

		for _, change := range changes {
			godebug.Line(ctx, scope, 35)
			scope.Declare("change", &change)
			godebug.Line(ctx, scope, 36)
			var kind string
			scope := scope.EnteringNewChildScope()
			scope.Declare("kind", &kind)
			godebug.Line(ctx, scope, 37)
			switch change.Kind {
			case godebug.Case(ctx, scope, 38):
				fallthrough
			case archive.ChangeModify:
				godebug.Line(ctx, scope, 39)
				kind = "C"
			case godebug.Case(ctx, scope, 40):
				fallthrough
			case archive.ChangeAdd:
				godebug.Line(ctx, scope, 41)
				kind = "A"
			case godebug.Case(ctx, scope, 42):
				fallthrough
			case archive.ChangeDelete:
				godebug.Line(ctx, scope, 43)
				kind = "D"
			}
			godebug.Line(ctx, scope, 45)
			fmt.Fprintf(cli.out, "%s %s\n", kind, change.Path)
		}
		godebug.Line(ctx, scope, 35)
	}
	godebug.Line(ctx, scope, 48)

	return nil
}

var diff_go_contents = `package client

import (
	"fmt"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/archive"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdDiff shows changes on a container's filesystem.
//
// Each changed file is printed on a separate line, prefixed with a single
// character that indicates the status of the file: C (modified), A (added),
// or D (deleted).
//
// Usage: docker diff CONTAINER
func (cli *DockerCli) CmdDiff(args ...string) error {
	cmd := Cli.Subcmd("diff", []string{"CONTAINER"}, Cli.DockerCommands["diff"].Description, true)
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	if cmd.Arg(0) == "" {
		return fmt.Errorf("Container name cannot be empty")
	}

	changes, err := cli.client.ContainerDiff(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}

	for _, change := range changes {
		var kind string
		switch change.Kind {
		case archive.ChangeModify:
			kind = "C"
		case archive.ChangeAdd:
			kind = "A"
		case archive.ChangeDelete:
			kind = "D"
		}
		fmt.Fprintf(cli.out, "%s %s\n", kind, change.Path)
	}

	return nil
}
`


var events_go_scope = godebug.EnteringNewFile(client_pkg_scope, events_go_contents)

func (cli *DockerCli) CmdEvents(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdEvents(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := events_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 28)
	cmd := Cli.Subcmd("events", nil, Cli.DockerCommands["events"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 29)
	since := cmd.String([]string{"-since"}, "", "Show all events created since timestamp")
	scope.Declare("since", &since)
	godebug.Line(ctx, scope, 30)
	until := cmd.String([]string{"-until"}, "", "Stream events until this timestamp")
	scope.Declare("until", &until)
	godebug.Line(ctx, scope, 31)
	flFilter := opts.NewListOpts(nil)
	scope.Declare("flFilter", &flFilter)
	godebug.Line(ctx, scope, 32)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	godebug.Line(ctx, scope, 33)
	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 35)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 37)

	eventFilterArgs := filters.NewArgs()
	scope.Declare("eventFilterArgs", &eventFilterArgs)
	{
		scope := scope.EnteringNewChildScope()

		for _, f := range flFilter.GetAll() {
			godebug.Line(ctx, scope, 41)
			scope.Declare("f", &f)
			godebug.Line(ctx, scope, 42)
			var err error
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 43)
			eventFilterArgs, err = filters.ParseFlag(f, eventFilterArgs)
			godebug.Line(ctx, scope, 44)
			if err != nil {
				godebug.Line(ctx, scope, 45)
				return err
			}
		}
		godebug.Line(ctx, scope, 41)
	}
	godebug.Line(ctx, scope, 49)

	options := types.EventsOptions{
		Since:   *since,
		Until:   *until,
		Filters: eventFilterArgs,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 55)

	responseBody, err := cli.client.Events(context.Background(), options)
	scope.Declare("responseBody", &responseBody, "err", &err)
	godebug.Line(ctx, scope, 56)
	if err != nil {
		godebug.Line(ctx, scope, 57)
		return err
	}
	godebug.Line(ctx, scope, 59)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 59)
	godebug.Line(ctx, scope, 61)

	return streamEvents(responseBody, cli.out)
}

func streamEvents(input io.Reader, output io.Writer) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = streamEvents(input, output)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := events_go_scope.EnteringNewChildScope()
	scope.Declare("input", &input, "output", &output)
	godebug.Line(ctx, scope, 66)
	return decodeEvents(input, func(event eventtypes.Message, err error) error {
		var result1 error
		fn := func(ctx *godebug.Context) {
			result1 = func() error {
				scope := scope.EnteringNewChildScope()
				scope.Declare("event", &event, "err", &err)
				godebug.Line(ctx, scope, 67)
				if err != nil {
					godebug.Line(ctx, scope, 68)
					return err
				}
				godebug.Line(ctx, scope, 70)
				printOutput(event, output)
				godebug.Line(ctx, scope, 71)
				return nil
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1
	},
	)
}

type eventProcessor func(event eventtypes.Message, err error) error

func decodeEvents(input io.Reader, ep eventProcessor) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = decodeEvents(input, ep)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := events_go_scope.EnteringNewChildScope()
	scope.Declare("input", &input, "ep", &ep)
	godebug.Line(ctx, scope, 78)
	dec := json.NewDecoder(input)
	scope.Declare("dec", &dec)
	godebug.Line(ctx, scope, 79)
	for {
		godebug.Line(ctx, scope, 80)
		var event eventtypes.Message
		scope := scope.EnteringNewChildScope()
		scope.Declare("event", &event)
		godebug.Line(ctx, scope, 81)
		err := dec.Decode(&event)
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 82)
		if err != nil && err == io.EOF {
			godebug.Line(ctx, scope, 83)
			break
		}
		godebug.Line(ctx, scope, 86)

		if procErr := ep(event, err); procErr != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("procErr", &procErr)
			godebug.Line(ctx, scope, 87)
			return procErr
		}
		godebug.Line(ctx, scope, 79)
	}
	godebug.Line(ctx, scope, 90)
	return nil
}

func printOutput(event eventtypes.Message, output io.Writer) {
	ctx, ok := godebug.EnterFunc(func() {
		printOutput(event, output)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := events_go_scope.EnteringNewChildScope()
	scope.Declare("event", &event, "output", &output)
	godebug.Line(ctx, scope, 97)
	if event.TimeNano != 0 {
		godebug.Line(ctx, scope, 98)
		fmt.Fprintf(output, "%s ", time.Unix(0, event.TimeNano).Format(jsonlog.RFC3339NanoFixed))
	} else {
		godebug.ElseIfExpr(ctx, scope, 99)
		if event.Time != 0 {
			godebug.Line(ctx, scope, 100)
			fmt.Fprintf(output, "%s ", time.Unix(event.Time, 0).Format(jsonlog.RFC3339NanoFixed))
		}
	}
	godebug.Line(ctx, scope, 103)

	fmt.Fprintf(output, "%s %s %s", event.Type, event.Action, event.Actor.ID)
	godebug.Line(ctx, scope, 105)

	if len(event.Actor.Attributes) > 0 {
		godebug.Line(ctx, scope, 106)
		var attrs []string
		scope := scope.EnteringNewChildScope()
		scope.Declare("attrs", &attrs)
		godebug.Line(ctx, scope, 107)
		var keys []string
		scope.Declare("keys", &keys)
		{
			scope := scope.EnteringNewChildScope()
			for k := range event.Actor.Attributes {
				godebug.Line(ctx, scope, 108)
				scope.Declare("k", &k)
				godebug.Line(ctx, scope, 109)
				keys = append(keys, k)
			}
			godebug.Line(ctx, scope, 108)
		}
		godebug.Line(ctx, scope, 111)
		sort.Strings(keys)
		{
			scope := scope.EnteringNewChildScope()
			for _, k := range keys {
				godebug.Line(ctx, scope, 112)
				scope.Declare("k", &k)
				godebug.Line(ctx, scope, 113)
				v := event.Actor.Attributes[k]
				scope := scope.EnteringNewChildScope()
				scope.Declare("v", &v)
				godebug.Line(ctx, scope, 114)
				attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
			}
			godebug.Line(ctx, scope, 112)
		}
		godebug.Line(ctx, scope, 116)
		fmt.Fprintf(output, " (%s)", strings.Join(attrs, ", "))
	}
	godebug.Line(ctx, scope, 118)
	fmt.Fprint(output, "\n")
}

type eventHandler struct {
	handlers map[string]func(eventtypes.Message)
	mu       sync.Mutex
}

func (w *eventHandler) Handle(action string, h func(eventtypes.Message)) {
	ctx, ok := godebug.EnterFunc(func() {
		w.Handle(action, h)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := events_go_scope.EnteringNewChildScope()
	scope.Declare("w", &w, "action", &action, "h", &h)
	godebug.Line(ctx, scope, 127)
	w.mu.Lock()
	godebug.Line(ctx, scope, 128)
	w.handlers[action] = h
	godebug.Line(ctx, scope, 129)
	w.mu.Unlock()
}

func (w *eventHandler) Watch(c <-chan eventtypes.Message) {
	ctx, ok := godebug.EnterFunc(func() {
		w.Watch(c)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := events_go_scope.EnteringNewChildScope()
	scope.Declare("w", &w, "c", &c)
	{
		scope := scope.EnteringNewChildScope()
		for e := range c {
			godebug.Line(ctx, scope, 136)
			scope.Declare("e", &e)
			godebug.Line(ctx, scope, 137)
			w.mu.Lock()
			godebug.Line(ctx, scope, 138)
			h, exists := w.handlers[e.Action]
			scope := scope.EnteringNewChildScope()
			scope.Declare("h", &h, "exists", &exists)
			godebug.Line(ctx, scope, 139)
			w.mu.Unlock()
			godebug.Line(ctx, scope, 140)
			if !exists {
				godebug.Line(ctx, scope, 141)
				continue
			}
			godebug.Line(ctx, scope, 143)
			logrus.Debugf("event handler: received event: %v", e)
			godebug.Line(ctx, scope, 144)
			go h(e)
		}
		godebug.Line(ctx, scope, 136)
	}
}

var events_go_contents = `package client

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/jsonlog"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
)

// CmdEvents prints a live stream of real time events from the server.
//
// Usage: docker events [OPTIONS]
func (cli *DockerCli) CmdEvents(args ...string) error {
	cmd := Cli.Subcmd("events", nil, Cli.DockerCommands["events"].Description, true)
	since := cmd.String([]string{"-since"}, "", "Show all events created since timestamp")
	until := cmd.String([]string{"-until"}, "", "Stream events until this timestamp")
	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	cmd.Require(flag.Exact, 0)

	cmd.ParseFlags(args, true)

	eventFilterArgs := filters.NewArgs()

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process in the daemon/server.
	for _, f := range flFilter.GetAll() {
		var err error
		eventFilterArgs, err = filters.ParseFlag(f, eventFilterArgs)
		if err != nil {
			return err
		}
	}

	options := types.EventsOptions{
		Since:   *since,
		Until:   *until,
		Filters: eventFilterArgs,
	}

	responseBody, err := cli.client.Events(context.Background(), options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return streamEvents(responseBody, cli.out)
}

// streamEvents decodes prints the incoming events in the provided output.
func streamEvents(input io.Reader, output io.Writer) error {
	return decodeEvents(input, func(event eventtypes.Message, err error) error {
		if err != nil {
			return err
		}
		printOutput(event, output)
		return nil
	})
}

type eventProcessor func(event eventtypes.Message, err error) error

func decodeEvents(input io.Reader, ep eventProcessor) error {
	dec := json.NewDecoder(input)
	for {
		var event eventtypes.Message
		err := dec.Decode(&event)
		if err != nil && err == io.EOF {
			break
		}

		if procErr := ep(event, err); procErr != nil {
			return procErr
		}
	}
	return nil
}

// printOutput prints all types of event information.
// Each output includes the event type, actor id, name and action.
// Actor attributes are printed at the end if the actor has any.
func printOutput(event eventtypes.Message, output io.Writer) {
	if event.TimeNano != 0 {
		fmt.Fprintf(output, "%s ", time.Unix(0, event.TimeNano).Format(jsonlog.RFC3339NanoFixed))
	} else if event.Time != 0 {
		fmt.Fprintf(output, "%s ", time.Unix(event.Time, 0).Format(jsonlog.RFC3339NanoFixed))
	}

	fmt.Fprintf(output, "%s %s %s", event.Type, event.Action, event.Actor.ID)

	if len(event.Actor.Attributes) > 0 {
		var attrs []string
		var keys []string
		for k := range event.Actor.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := event.Actor.Attributes[k]
			attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(output, " (%s)", strings.Join(attrs, ", "))
	}
	fmt.Fprint(output, "\n")
}

type eventHandler struct {
	handlers map[string]func(eventtypes.Message)
	mu       sync.Mutex
}

func (w *eventHandler) Handle(action string, h func(eventtypes.Message)) {
	w.mu.Lock()
	w.handlers[action] = h
	w.mu.Unlock()
}

// Watch ranges over the passed in event chan and processes the events based on the
// handlers created for a given action.
// To stop watching, close the event chan.
func (w *eventHandler) Watch(c <-chan eventtypes.Message) {
	for e := range c {
		w.mu.Lock()
		h, exists := w.handlers[e.Action]
		w.mu.Unlock()
		if !exists {
			continue
		}
		logrus.Debugf("event handler: received event: %v", e)
		go h(e)
	}
}
`


var exec_go_scope = godebug.EnteringNewFile(client_pkg_scope, exec_go_contents)

func (cli *DockerCli) CmdExec(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdExec(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := exec_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 20)
	cmd := Cli.Subcmd("exec", []string{"CONTAINER COMMAND [ARG...]"}, Cli.DockerCommands["exec"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 21)
	detachKeys := cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
	scope.Declare("detachKeys", &detachKeys)
	godebug.Line(ctx, scope, 23)

	execConfig, err := ParseExec(cmd, args)
	scope.Declare("execConfig", &execConfig, "err", &err)
	godebug.Line(ctx, scope, 25)

	if execConfig.Container == "" || err != nil {
		godebug.Line(ctx, scope, 26)
		return Cli.StatusError{StatusCode: 1}
	}
	godebug.Line(ctx, scope, 29)

	if *detachKeys != "" {
		godebug.Line(ctx, scope, 30)
		cli.configFile.DetachKeys = *detachKeys
	}
	godebug.Line(ctx, scope, 34)

	execConfig.DetachKeys = cli.configFile.DetachKeys
	godebug.Line(ctx, scope, 36)

	response, err := cli.client.ContainerExecCreate(context.Background(), *execConfig)
	scope.Declare("response", &response)
	godebug.Line(ctx, scope, 37)
	if err != nil {
		godebug.Line(ctx, scope, 38)
		return err
	}
	godebug.Line(ctx, scope, 41)

	execID := response.ID
	scope.Declare("execID", &execID)
	godebug.Line(ctx, scope, 42)
	if execID == "" {
		godebug.Line(ctx, scope, 43)
		fmt.Fprintf(cli.out, "exec ID empty")
		godebug.Line(ctx, scope, 44)
		return nil
	}
	godebug.Line(ctx, scope, 48)

	if !execConfig.Detach {
		godebug.Line(ctx, scope, 49)
		if err := cli.CheckTtyInput(execConfig.AttachStdin, execConfig.Tty); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 50)
			return err
		}
	} else {
		godebug.Line(ctx, scope, 52)
		godebug.Line(ctx, scope, 53)
		execStartCheck := types.ExecStartCheck{
			Detach: execConfig.Detach,
			Tty:    execConfig.Tty,
		}
		scope := scope.EnteringNewChildScope()
		scope.Declare("execStartCheck", &execStartCheck)
		godebug.Line(ctx, scope, 58)

		if err := cli.client.ContainerExecStart(context.Background(), execID, execStartCheck); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 59)
			return err
		}
		godebug.Line(ctx, scope, 63)

		return nil
	}
	godebug.Line(ctx, scope, 67)

	var (
		out, stderr io.Writer
		in          io.ReadCloser
		errCh       chan error
	)
	scope.Declare("out", &out, "stderr", &stderr, "in", &in, "errCh", &errCh)
	godebug.Line(ctx, scope, 73)

	if execConfig.AttachStdin {
		godebug.Line(ctx, scope, 74)
		in = cli.in
	}
	godebug.Line(ctx, scope, 76)
	if execConfig.AttachStdout {
		godebug.Line(ctx, scope, 77)
		out = cli.out
	}
	godebug.Line(ctx, scope, 79)
	if execConfig.AttachStderr {
		godebug.Line(ctx, scope, 80)
		if execConfig.Tty {
			godebug.Line(ctx, scope, 81)
			stderr = cli.out
		} else {
			godebug.Line(ctx, scope, 82)
			godebug.Line(ctx, scope, 83)
			stderr = cli.err
		}
	}
	godebug.Line(ctx, scope, 87)

	resp, err := cli.client.ContainerExecAttach(context.Background(), execID, *execConfig)
	scope.Declare("resp", &resp)
	godebug.Line(ctx, scope, 88)
	if err != nil {
		godebug.Line(ctx, scope, 89)
		return err
	}
	godebug.Line(ctx, scope, 91)
	defer resp.Close()
	defer godebug.Defer(ctx, scope, 91)
	godebug.Line(ctx, scope, 92)
	errCh = promise.Go(func() error {
		var result1 error
		fn := func(ctx *godebug.Context) {
			result1 = func() error {
				godebug.Line(ctx, scope, 93)
				return cli.holdHijackedConnection(context.Background(), execConfig.Tty, in, out, stderr, resp)
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1
	},
	)
	godebug.Line(ctx, scope, 96)

	if execConfig.Tty && cli.isTerminalIn {
		godebug.Line(ctx, scope, 97)
		if err := cli.monitorTtySize(execID, true); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 98)
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}
	godebug.Line(ctx, scope, 102)

	if err := <-errCh; err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 103)
		logrus.Debugf("Error hijack: %s", err)
		godebug.Line(ctx, scope, 104)
		return err
	}
	godebug.Line(ctx, scope, 107)

	var status int
	scope.Declare("status", &status)
	godebug.Line(ctx, scope, 108)
	if _, status, err = getExecExitCode(cli, execID); err != nil {
		godebug.Line(ctx, scope, 109)
		return err
	}
	godebug.Line(ctx, scope, 112)

	if status != 0 {
		godebug.Line(ctx, scope, 113)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(ctx, scope, 116)

	return nil
}

func ParseExec(cmd *flag.FlagSet, args []string) (*types.ExecConfig, error) {
	var result1 *types.ExecConfig
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = ParseExec(cmd, args)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := exec_go_scope.EnteringNewChildScope()
	scope.Declare("cmd", &cmd, "args", &args)
	godebug.Line(ctx, scope, 124)
	var (
		flStdin      = cmd.Bool([]string{"i", "-interactive"}, false, "Keep STDIN open even if not attached")
		flTty        = cmd.Bool([]string{"t", "-tty"}, false, "Allocate a pseudo-TTY")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Detached mode: run command in the background")
		flUser       = cmd.String([]string{"u", "-user"}, "", "Username or UID (format: <name|uid>[:<group|gid>])")
		flPrivileged = cmd.Bool([]string{"-privileged"}, false, "Give extended privileges to the command")
		execCmd      []string
		container    string
	)
	scope.Declare("flStdin", &flStdin, "flTty", &flTty, "flDetach", &flDetach, "flUser", &flUser, "flPrivileged", &flPrivileged, "execCmd", &execCmd, "container", &container)
	godebug.Line(ctx, scope, 133)

	cmd.Require(flag.Min, 2)
	godebug.Line(ctx, scope, 134)
	if err := cmd.ParseFlags(args, true); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 135)
		return nil, err
	}
	godebug.Line(ctx, scope, 137)
	container = cmd.Arg(0)
	godebug.Line(ctx, scope, 138)
	parsedArgs := cmd.Args()
	scope.Declare("parsedArgs", &parsedArgs)
	godebug.Line(ctx, scope, 139)
	execCmd = parsedArgs[1:]
	godebug.Line(ctx, scope, 141)

	execConfig := &types.ExecConfig{
		User:       *flUser,
		Privileged: *flPrivileged,
		Tty:        *flTty,
		Cmd:        execCmd,
		Container:  container,
		Detach:     *flDetach,
	}
	scope.Declare("execConfig", &execConfig)
	godebug.Line(ctx, scope, 151)

	if !*flDetach {
		godebug.Line(ctx, scope, 152)
		execConfig.AttachStdout = true
		godebug.Line(ctx, scope, 153)
		execConfig.AttachStderr = true
		godebug.Line(ctx, scope, 154)
		if *flStdin {
			godebug.Line(ctx, scope, 155)
			execConfig.AttachStdin = true
		}
	}
	godebug.Line(ctx, scope, 159)

	return execConfig, nil
}

var exec_go_contents = `package client

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/engine-api/types"
)

// CmdExec runs a command in a running container.
//
// Usage: docker exec [OPTIONS] CONTAINER COMMAND [ARG...]
func (cli *DockerCli) CmdExec(args ...string) error {
	cmd := Cli.Subcmd("exec", []string{"CONTAINER COMMAND [ARG...]"}, Cli.DockerCommands["exec"].Description, true)
	detachKeys := cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")

	execConfig, err := ParseExec(cmd, args)
	// just in case the ParseExec does not exit
	if execConfig.Container == "" || err != nil {
		return Cli.StatusError{StatusCode: 1}
	}

	if *detachKeys != "" {
		cli.configFile.DetachKeys = *detachKeys
	}

	// Send client escape keys
	execConfig.DetachKeys = cli.configFile.DetachKeys

	response, err := cli.client.ContainerExecCreate(context.Background(), *execConfig)
	if err != nil {
		return err
	}

	execID := response.ID
	if execID == "" {
		fmt.Fprintf(cli.out, "exec ID empty")
		return nil
	}

	//Temp struct for execStart so that we don't need to transfer all the execConfig
	if !execConfig.Detach {
		if err := cli.CheckTtyInput(execConfig.AttachStdin, execConfig.Tty); err != nil {
			return err
		}
	} else {
		execStartCheck := types.ExecStartCheck{
			Detach: execConfig.Detach,
			Tty:    execConfig.Tty,
		}

		if err := cli.client.ContainerExecStart(context.Background(), execID, execStartCheck); err != nil {
			return err
		}
		// For now don't print this - wait for when we support exec wait()
		// fmt.Fprintf(cli.out, "%s\n", execID)
		return nil
	}

	// Interactive exec requested.
	var (
		out, stderr io.Writer
		in          io.ReadCloser
		errCh       chan error
	)

	if execConfig.AttachStdin {
		in = cli.in
	}
	if execConfig.AttachStdout {
		out = cli.out
	}
	if execConfig.AttachStderr {
		if execConfig.Tty {
			stderr = cli.out
		} else {
			stderr = cli.err
		}
	}

	resp, err := cli.client.ContainerExecAttach(context.Background(), execID, *execConfig)
	if err != nil {
		return err
	}
	defer resp.Close()
	errCh = promise.Go(func() error {
		return cli.holdHijackedConnection(context.Background(), execConfig.Tty, in, out, stderr, resp)
	})

	if execConfig.Tty && cli.isTerminalIn {
		if err := cli.monitorTtySize(execID, true); err != nil {
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}

	if err := <-errCh; err != nil {
		logrus.Debugf("Error hijack: %s", err)
		return err
	}

	var status int
	if _, status, err = getExecExitCode(cli, execID); err != nil {
		return err
	}

	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}

	return nil
}

// ParseExec parses the specified args for the specified command and generates
// an ExecConfig from it.
// If the minimal number of specified args is not right or if specified args are
// not valid, it will return an error.
func ParseExec(cmd *flag.FlagSet, args []string) (*types.ExecConfig, error) {
	var (
		flStdin      = cmd.Bool([]string{"i", "-interactive"}, false, "Keep STDIN open even if not attached")
		flTty        = cmd.Bool([]string{"t", "-tty"}, false, "Allocate a pseudo-TTY")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Detached mode: run command in the background")
		flUser       = cmd.String([]string{"u", "-user"}, "", "Username or UID (format: <name|uid>[:<group|gid>])")
		flPrivileged = cmd.Bool([]string{"-privileged"}, false, "Give extended privileges to the command")
		execCmd      []string
		container    string
	)
	cmd.Require(flag.Min, 2)
	if err := cmd.ParseFlags(args, true); err != nil {
		return nil, err
	}
	container = cmd.Arg(0)
	parsedArgs := cmd.Args()
	execCmd = parsedArgs[1:]

	execConfig := &types.ExecConfig{
		User:       *flUser,
		Privileged: *flPrivileged,
		Tty:        *flTty,
		Cmd:        execCmd,
		Container:  container,
		Detach:     *flDetach,
	}

	// If -d is not set, attach to everything by default
	if !*flDetach {
		execConfig.AttachStdout = true
		execConfig.AttachStderr = true
		if *flStdin {
			execConfig.AttachStdin = true
		}
	}

	return execConfig, nil
}
`


var export_go_scope = godebug.EnteringNewFile(client_pkg_scope, export_go_contents)

func (cli *DockerCli) CmdExport(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdExport(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := export_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("export", []string{"CONTAINER"}, Cli.DockerCommands["export"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	outfile := cmd.String([]string{"o", "-output"}, "", "Write to a file, instead of STDOUT")
	scope.Declare("outfile", &outfile)
	godebug.Line(ctx, scope, 21)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 23)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 25)

	if *outfile == "" && cli.isTerminalOut {
		godebug.Line(ctx, scope, 26)
		return errors.New("Cowardly refusing to save to a terminal. Use the -o flag or redirect.")
	}
	godebug.Line(ctx, scope, 29)

	responseBody, err := cli.client.ContainerExport(context.Background(), cmd.Arg(0))
	scope.Declare("responseBody", &responseBody, "err", &err)
	godebug.Line(ctx, scope, 30)
	if err != nil {
		godebug.Line(ctx, scope, 31)
		return err
	}
	godebug.Line(ctx, scope, 33)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 33)
	godebug.Line(ctx, scope, 35)

	if *outfile == "" {
		godebug.Line(ctx, scope, 36)
		_, err := io.Copy(cli.out, responseBody)
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 37)
		return err
	}
	godebug.Line(ctx, scope, 40)

	return copyToFile(*outfile, responseBody)

}

var export_go_contents = `package client

import (
	"errors"
	"io"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdExport exports a filesystem as a tar archive.
//
// The tar archive is streamed to STDOUT by default or written to a file.
//
// Usage: docker export [OPTIONS] CONTAINER
func (cli *DockerCli) CmdExport(args ...string) error {
	cmd := Cli.Subcmd("export", []string{"CONTAINER"}, Cli.DockerCommands["export"].Description, true)
	outfile := cmd.String([]string{"o", "-output"}, "", "Write to a file, instead of STDOUT")
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	if *outfile == "" && cli.isTerminalOut {
		return errors.New("Cowardly refusing to save to a terminal. Use the -o flag or redirect.")
	}

	responseBody, err := cli.client.ContainerExport(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if *outfile == "" {
		_, err := io.Copy(cli.out, responseBody)
		return err
	}

	return copyToFile(*outfile, responseBody)

}
`


var hijack_go_scope = godebug.EnteringNewFile(client_pkg_scope, hijack_go_contents)

func (cli *DockerCli) holdHijackedConnection(ctx context.Context, tty bool, inputStream io.ReadCloser, outputStream, errorStream io.Writer, resp types.HijackedResponse) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.holdHijackedConnection(ctx, tty, inputStream, outputStream, errorStream, resp)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := hijack_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "ctx", &ctx, "tty", &tty, "inputStream", &inputStream, "outputStream", &outputStream, "errorStream", &errorStream, "resp", &resp)
	godebug.Line(_ctx, scope, 15)
	var (
		err         error
		restoreOnce sync.Once
	)
	scope.Declare("err", &err, "restoreOnce", &restoreOnce)
	godebug.Line(_ctx, scope, 19)

	if inputStream != nil && tty {
		godebug.Line(_ctx, scope, 20)
		if err := cli.setRawTerminal(); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 21)
			return err
		}
		godebug.Line(_ctx, scope, 23)
		defer func() {
			fn := func(_ctx *godebug.Context) {
				godebug.Line(_ctx, scope, 24)
				restoreOnce.Do(func() {
					fn := func(_ctx *godebug.Context) {
						godebug.Line(_ctx, scope, 25)
						cli.restoreTerminal(inputStream)
					}
					if _ctx, ok := godebug.EnterFuncLit(fn); ok {
						defer godebug.ExitFunc(_ctx)
						fn(_ctx)
					}
				})
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
		}()
		defer godebug.Defer(_ctx, scope, 23)
	}
	godebug.Line(_ctx, scope, 30)

	receiveStdout := make(chan error, 1)
	scope.Declare("receiveStdout", &receiveStdout)
	godebug.Line(_ctx, scope, 31)
	if outputStream != nil || errorStream != nil {
		godebug.Line(_ctx, scope, 32)
		go func() {
			fn := func(_ctx *godebug.Context) {
				godebug.Line(_ctx, scope, 34)
				if tty && outputStream != nil {
					godebug.Line(_ctx, scope, 35)
					_, err = io.Copy(outputStream, resp.Reader)
					godebug.Line(_ctx, scope, 38)
					if inputStream != nil {
						godebug.Line(_ctx, scope, 39)
						restoreOnce.Do(func() {
							fn := func(_ctx *godebug.Context) {
								godebug.Line(_ctx, scope, 40)
								cli.restoreTerminal(inputStream)
							}
							if _ctx, ok := godebug.EnterFuncLit(fn); ok {
								defer godebug.ExitFunc(_ctx)
								fn(_ctx)
							}
						})
					}
				} else {
					godebug.Line(_ctx, scope, 43)
					godebug.Line(_ctx, scope, 44)
					_, err = stdcopy.StdCopy(outputStream, errorStream, resp.Reader)
				}
				godebug.Line(_ctx, scope, 47)
				logrus.Debugf("[hijack] End of stdout")
				godebug.Line(_ctx, scope, 48)
				receiveStdout <- err
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
		}()
	}
	godebug.Line(_ctx, scope, 52)

	stdinDone := make(chan struct{})
	scope.Declare("stdinDone", &stdinDone)
	godebug.Line(_ctx, scope, 53)
	go func() {
		fn := func(_ctx *godebug.Context) {
			godebug.Line(_ctx, scope, 54)
			if inputStream != nil {
				godebug.Line(_ctx, scope, 55)
				io.Copy(resp.Conn, inputStream)
				godebug.Line(_ctx, scope, 58)
				if tty {
					godebug.Line(_ctx, scope, 59)
					restoreOnce.Do(func() {
						fn := func(_ctx *godebug.Context) {
							godebug.Line(_ctx, scope, 60)
							cli.restoreTerminal(inputStream)
						}
						if _ctx, ok := godebug.EnterFuncLit(fn); ok {
							defer godebug.ExitFunc(_ctx)
							fn(_ctx)
						}
					})
				}
				godebug.Line(_ctx, scope, 63)
				logrus.Debugf("[hijack] End of stdin")
			}
			godebug.Line(_ctx, scope, 66)
			if err := resp.CloseWrite(); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(_ctx, scope, 67)
				logrus.Debugf("Couldn't send EOF: %s", err)
			}
			godebug.Line(_ctx, scope, 69)
			close(stdinDone)
		}
		if _ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(_ctx)
			fn(_ctx)
		}
	}()
	godebug.Select(_ctx, scope, 72)

	select {
	case <-godebug.Comm(_ctx, scope, 73):
		panic("impossible")
	case err := <-receiveStdout:
		godebug.Line(_ctx, scope, 73)
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(_ctx, scope, 74)
		if err != nil {
			logrus.Debugf("Error receiveStdout: %s", err)
			return err
		}
	case <-godebug.Comm(_ctx, scope, 78):
		panic("impossible")
	case <-stdinDone:
		godebug.Line(_ctx, scope, 78)
		godebug.Line(_ctx, scope, 79)
		if outputStream != nil || errorStream != nil {
			select {
			case err := <-receiveStdout:
				if err != nil {
					logrus.Debugf("Error receiveStdout: %s", err)
					return err
				}
			case <-ctx.Done():
			}
		}
	case <-godebug.Comm(_ctx, scope, 89):
		panic("impossible")
	case <-ctx.Done():
		godebug.Line(_ctx, scope, 89)
	case <-godebug.EndSelect(_ctx, scope):
		panic("impossible")
	}
	godebug.Line(_ctx, scope, 92)

	return nil
}

var hijack_go_contents = `package client

import (
	"io"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/engine-api/types"
)

func (cli *DockerCli) holdHijackedConnection(ctx context.Context, tty bool, inputStream io.ReadCloser, outputStream, errorStream io.Writer, resp types.HijackedResponse) error {
	var (
		err         error
		restoreOnce sync.Once
	)
	if inputStream != nil && tty {
		if err := cli.setRawTerminal(); err != nil {
			return err
		}
		defer func() {
			restoreOnce.Do(func() {
				cli.restoreTerminal(inputStream)
			})
		}()
	}

	receiveStdout := make(chan error, 1)
	if outputStream != nil || errorStream != nil {
		go func() {
			// When TTY is ON, use regular copy
			if tty && outputStream != nil {
				_, err = io.Copy(outputStream, resp.Reader)
				// we should restore the terminal as soon as possible once connection end
				// so any following print messages will be in normal type.
				if inputStream != nil {
					restoreOnce.Do(func() {
						cli.restoreTerminal(inputStream)
					})
				}
			} else {
				_, err = stdcopy.StdCopy(outputStream, errorStream, resp.Reader)
			}

			logrus.Debugf("[hijack] End of stdout")
			receiveStdout <- err
		}()
	}

	stdinDone := make(chan struct{})
	go func() {
		if inputStream != nil {
			io.Copy(resp.Conn, inputStream)
			// we should restore the terminal as soon as possible once connection end
			// so any following print messages will be in normal type.
			if tty {
				restoreOnce.Do(func() {
					cli.restoreTerminal(inputStream)
				})
			}
			logrus.Debugf("[hijack] End of stdin")
		}

		if err := resp.CloseWrite(); err != nil {
			logrus.Debugf("Couldn't send EOF: %s", err)
		}
		close(stdinDone)
	}()

	select {
	case err := <-receiveStdout:
		if err != nil {
			logrus.Debugf("Error receiveStdout: %s", err)
			return err
		}
	case <-stdinDone:
		if outputStream != nil || errorStream != nil {
			select {
			case err := <-receiveStdout:
				if err != nil {
					logrus.Debugf("Error receiveStdout: %s", err)
					return err
				}
			case <-ctx.Done():
			}
		}
	case <-ctx.Done():
	}

	return nil
}
`


var history_go_scope = godebug.EnteringNewFile(client_pkg_scope, history_go_contents)

func (cli *DockerCli) CmdHistory(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdHistory(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := history_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 23)
	cmd := Cli.Subcmd("history", []string{"IMAGE"}, Cli.DockerCommands["history"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 24)
	human := cmd.Bool([]string{"H", "-human"}, true, "Print sizes and dates in human readable format")
	scope.Declare("human", &human)
	godebug.Line(ctx, scope, 25)
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only show numeric IDs")
	scope.Declare("quiet", &quiet)
	godebug.Line(ctx, scope, 26)
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
	scope.Declare("noTrunc", &noTrunc)
	godebug.Line(ctx, scope, 27)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 29)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 31)

	history, err := cli.client.ImageHistory(context.Background(), cmd.Arg(0))
	scope.Declare("history", &history, "err", &err)
	godebug.Line(ctx, scope, 32)
	if err != nil {
		godebug.Line(ctx, scope, 33)
		return err
	}
	godebug.Line(ctx, scope, 36)

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	scope.Declare("w", &w)
	godebug.Line(ctx, scope, 38)

	if *quiet {
		{
			scope := scope.EnteringNewChildScope()
			for _, entry := range history {
				godebug.Line(ctx, scope, 39)
				scope.Declare("entry", &entry)
				godebug.Line(ctx, scope, 40)
				if *noTrunc {
					godebug.Line(ctx, scope, 41)
					fmt.Fprintf(w, "%s\n", entry.ID)
				} else {
					godebug.Line(ctx, scope, 42)
					godebug.Line(ctx, scope, 43)
					fmt.Fprintf(w, "%s\n", stringid.TruncateID(entry.ID))
				}
			}
			godebug.Line(ctx, scope, 39)
		}
		godebug.Line(ctx, scope, 46)
		w.Flush()
		godebug.Line(ctx, scope, 47)
		return nil
	}
	godebug.Line(ctx, scope, 50)

	var imageID string
	scope.Declare("imageID", &imageID)
	godebug.Line(ctx, scope, 51)
	var createdBy string
	scope.Declare("createdBy", &createdBy)
	godebug.Line(ctx, scope, 52)
	var created string
	scope.Declare("created", &created)
	godebug.Line(ctx, scope, 53)
	var size string
	scope.Declare("size", &size)
	godebug.Line(ctx, scope, 55)

	fmt.Fprintln(w, "IMAGE\tCREATED\tCREATED BY\tSIZE\tCOMMENT")
	{
		scope := scope.EnteringNewChildScope()
		for _, entry := range history {
			godebug.Line(ctx, scope, 56)
			scope.Declare("entry", &entry)
			godebug.Line(ctx, scope, 57)
			imageID = entry.ID
			godebug.Line(ctx, scope, 58)
			createdBy = strings.Replace(entry.CreatedBy, "\t", " ", -1)
			godebug.Line(ctx, scope, 59)
			if *noTrunc == false {
				godebug.Line(ctx, scope, 60)
				createdBy = stringutils.Truncate(createdBy, 45)
				godebug.Line(ctx, scope, 61)
				imageID = stringid.TruncateID(entry.ID)
			}
			godebug.Line(ctx, scope, 64)

			if *human {
				godebug.Line(ctx, scope, 65)
				created = units.HumanDuration(time.Now().UTC().Sub(time.Unix(entry.Created, 0))) + " ago"
				godebug.Line(ctx, scope, 66)
				size = units.HumanSize(float64(entry.Size))
			} else {
				godebug.Line(ctx, scope, 67)
				godebug.Line(ctx, scope, 68)
				created = time.Unix(entry.Created, 0).Format(time.RFC3339)
				godebug.Line(ctx, scope, 69)
				size = strconv.FormatInt(entry.Size, 10)
			}
			godebug.Line(ctx, scope, 72)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", imageID, created, createdBy, size, entry.Comment)
		}
		godebug.Line(ctx, scope, 56)
	}
	godebug.Line(ctx, scope, 74)
	w.Flush()
	godebug.Line(ctx, scope, 75)
	return nil
}

var history_go_contents = `package client

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/stringutils"
	"github.com/docker/go-units"
)

// CmdHistory shows the history of an image.
//
// Usage: docker history [OPTIONS] IMAGE
func (cli *DockerCli) CmdHistory(args ...string) error {
	cmd := Cli.Subcmd("history", []string{"IMAGE"}, Cli.DockerCommands["history"].Description, true)
	human := cmd.Bool([]string{"H", "-human"}, true, "Print sizes and dates in human readable format")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only show numeric IDs")
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	history, err := cli.client.ImageHistory(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)

	if *quiet {
		for _, entry := range history {
			if *noTrunc {
				fmt.Fprintf(w, "%s\n", entry.ID)
			} else {
				fmt.Fprintf(w, "%s\n", stringid.TruncateID(entry.ID))
			}
		}
		w.Flush()
		return nil
	}

	var imageID string
	var createdBy string
	var created string
	var size string

	fmt.Fprintln(w, "IMAGE\tCREATED\tCREATED BY\tSIZE\tCOMMENT")
	for _, entry := range history {
		imageID = entry.ID
		createdBy = strings.Replace(entry.CreatedBy, "\t", " ", -1)
		if *noTrunc == false {
			createdBy = stringutils.Truncate(createdBy, 45)
			imageID = stringid.TruncateID(entry.ID)
		}

		if *human {
			created = units.HumanDuration(time.Now().UTC().Sub(time.Unix(entry.Created, 0))) + " ago"
			size = units.HumanSize(float64(entry.Size))
		} else {
			created = time.Unix(entry.Created, 0).Format(time.RFC3339)
			size = strconv.FormatInt(entry.Size, 10)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", imageID, created, createdBy, size, entry.Comment)
	}
	w.Flush()
	return nil
}
`


var images_go_scope = godebug.EnteringNewFile(client_pkg_scope, images_go_contents)

func (cli *DockerCli) CmdImages(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdImages(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := images_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 18)
	cmd := Cli.Subcmd("images", []string{"[REPOSITORY[:TAG]]"}, Cli.DockerCommands["images"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 19)
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only show numeric IDs")
	scope.Declare("quiet", &quiet)
	godebug.Line(ctx, scope, 20)
	all := cmd.Bool([]string{"a", "-all"}, false, "Show all images (default hides intermediate images)")
	scope.Declare("all", &all)
	godebug.Line(ctx, scope, 21)
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
	scope.Declare("noTrunc", &noTrunc)
	godebug.Line(ctx, scope, 22)
	showDigests := cmd.Bool([]string{"-digests"}, false, "Show digests")
	scope.Declare("showDigests", &showDigests)
	godebug.Line(ctx, scope, 23)
	format := cmd.String([]string{"-format"}, "", "Pretty-print images using a Go template")
	scope.Declare("format", &format)
	godebug.Line(ctx, scope, 25)

	flFilter := opts.NewListOpts(nil)
	scope.Declare("flFilter", &flFilter)
	godebug.Line(ctx, scope, 26)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	godebug.Line(ctx, scope, 27)
	cmd.Require(flag.Max, 1)
	godebug.Line(ctx, scope, 29)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 33)

	imageFilterArgs := filters.NewArgs()
	scope.Declare("imageFilterArgs", &imageFilterArgs)
	{
		scope := scope.EnteringNewChildScope()
		for _, f := range flFilter.GetAll() {
			godebug.Line(ctx, scope, 34)
			scope.Declare("f", &f)
			godebug.Line(ctx, scope, 35)
			var err error
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 36)
			imageFilterArgs, err = filters.ParseFlag(f, imageFilterArgs)
			godebug.Line(ctx, scope, 37)
			if err != nil {
				godebug.Line(ctx, scope, 38)
				return err
			}
		}
		godebug.Line(ctx, scope, 34)
	}
	godebug.Line(ctx, scope, 42)

	var matchName string
	scope.Declare("matchName", &matchName)
	godebug.Line(ctx, scope, 43)
	if cmd.NArg() == 1 {
		godebug.Line(ctx, scope, 44)
		matchName = cmd.Arg(0)
	}
	godebug.Line(ctx, scope, 47)

	options := types.ImageListOptions{
		MatchName: matchName,
		All:       *all,
		Filters:   imageFilterArgs,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 53)

	images, err := cli.client.ImageList(context.Background(), options)
	scope.Declare("images", &images, "err", &err)
	godebug.Line(ctx, scope, 54)
	if err != nil {
		godebug.Line(ctx, scope, 55)
		return err
	}
	godebug.Line(ctx, scope, 58)

	f := *format
	scope.Declare("f", &f)
	godebug.Line(ctx, scope, 59)
	if len(f) == 0 {
		godebug.Line(ctx, scope, 60)
		if len(cli.ImagesFormat()) > 0 && !*quiet {
			godebug.Line(ctx, scope, 61)
			f = cli.ImagesFormat()
		} else {
			godebug.Line(ctx, scope, 62)
			godebug.Line(ctx, scope, 63)
			f = "table"
		}
	}
	godebug.Line(ctx, scope, 67)

	imagesCtx := formatter.ImageContext{
		Context: formatter.Context{
			Output: cli.out,
			Format: f,
			Quiet:  *quiet,
			Trunc:  !*noTrunc,
		},
		Digest: *showDigests,
		Images: images,
	}
	scope.Declare("imagesCtx", &imagesCtx)
	godebug.Line(ctx, scope, 78)

	imagesCtx.Write()
	godebug.Line(ctx, scope, 80)

	return nil
}

var images_go_contents = `package client

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/client/formatter"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// CmdImages lists the images in a specified repository, or all top-level images if no repository is specified.
//
// Usage: docker images [OPTIONS] [REPOSITORY]
func (cli *DockerCli) CmdImages(args ...string) error {
	cmd := Cli.Subcmd("images", []string{"[REPOSITORY[:TAG]]"}, Cli.DockerCommands["images"].Description, true)
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only show numeric IDs")
	all := cmd.Bool([]string{"a", "-all"}, false, "Show all images (default hides intermediate images)")
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
	showDigests := cmd.Bool([]string{"-digests"}, false, "Show digests")
	format := cmd.String([]string{"-format"}, "", "Pretty-print images using a Go template")

	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	cmd.Require(flag.Max, 1)

	cmd.ParseFlags(args, true)

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process in the daemon/server.
	imageFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		var err error
		imageFilterArgs, err = filters.ParseFlag(f, imageFilterArgs)
		if err != nil {
			return err
		}
	}

	var matchName string
	if cmd.NArg() == 1 {
		matchName = cmd.Arg(0)
	}

	options := types.ImageListOptions{
		MatchName: matchName,
		All:       *all,
		Filters:   imageFilterArgs,
	}

	images, err := cli.client.ImageList(context.Background(), options)
	if err != nil {
		return err
	}

	f := *format
	if len(f) == 0 {
		if len(cli.ImagesFormat()) > 0 && !*quiet {
			f = cli.ImagesFormat()
		} else {
			f = "table"
		}
	}

	imagesCtx := formatter.ImageContext{
		Context: formatter.Context{
			Output: cli.out,
			Format: f,
			Quiet:  *quiet,
			Trunc:  !*noTrunc,
		},
		Digest: *showDigests,
		Images: images,
	}

	imagesCtx.Write()

	return nil
}
`


var import_go_scope = godebug.EnteringNewFile(client_pkg_scope, import_go_contents)

func (cli *DockerCli) CmdImport(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdImport(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := import_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 25)
	cmd := Cli.Subcmd("import", []string{"file|URL|- [REPOSITORY[:TAG]]"}, Cli.DockerCommands["import"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 26)
	flChanges := opts.NewListOpts(nil)
	scope.Declare("flChanges", &flChanges)
	godebug.Line(ctx, scope, 27)
	cmd.Var(&flChanges, []string{"c", "-change"}, "Apply Dockerfile instruction to the created image")
	godebug.Line(ctx, scope, 28)
	message := cmd.String([]string{"m", "-message"}, "", "Set commit message for imported image")
	scope.Declare("message", &message)
	godebug.Line(ctx, scope, 29)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 31)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 33)

	var (
		in         io.Reader
		tag        string
		src        = cmd.Arg(0)
		srcName    = src
		repository = cmd.Arg(1)
		changes    = flChanges.GetAll()
	)
	scope.Declare("in", &in, "tag", &tag, "src", &src, "srcName", &srcName, "repository", &repository, "changes", &changes)
	godebug.Line(ctx, scope, 42)

	if cmd.NArg() == 3 {
		godebug.Line(ctx, scope, 43)
		fmt.Fprintf(cli.err, "[DEPRECATED] The format 'file|URL|- [REPOSITORY [TAG]]' has been deprecated. Please use file|URL|- [REPOSITORY[:TAG]]\n")
		godebug.Line(ctx, scope, 44)
		tag = cmd.Arg(2)
	}
	godebug.Line(ctx, scope, 47)

	if repository != "" {
		godebug.Line(ctx, scope, 49)

		if _, err := reference.ParseNamed(repository); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 50)
			return err
		}
	}
	godebug.Line(ctx, scope, 54)

	if src == "-" {
		godebug.Line(ctx, scope, 55)
		in = cli.in
	} else {
		godebug.ElseIfExpr(ctx, scope, 56)
		if !urlutil.IsURL(src) {
			godebug.Line(ctx, scope, 57)
			srcName = "-"
			godebug.Line(ctx, scope, 58)
			file, err := os.Open(src)
			scope := scope.EnteringNewChildScope()
			scope.Declare("file", &file, "err", &err)
			godebug.Line(ctx, scope, 59)
			if err != nil {
				godebug.Line(ctx, scope, 60)
				return err
			}
			godebug.Line(ctx, scope, 62)
			defer file.Close()
			defer godebug.Defer(ctx, scope, 62)
			godebug.Line(ctx, scope, 63)
			in = file
		}
	}
	godebug.Line(ctx, scope, 66)

	options := types.ImageImportOptions{
		Source:         in,
		SourceName:     srcName,
		RepositoryName: repository,
		Message:        *message,
		Tag:            tag,
		Changes:        changes,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 75)

	responseBody, err := cli.client.ImageImport(context.Background(), options)
	scope.Declare("responseBody", &responseBody, "err", &err)
	godebug.Line(ctx, scope, 76)
	if err != nil {
		godebug.Line(ctx, scope, 77)
		return err
	}
	godebug.Line(ctx, scope, 79)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 79)
	godebug.Line(ctx, scope, 81)

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}

var import_go_contents = `package client

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/urlutil"
	"github.com/docker/docker/reference"
	"github.com/docker/engine-api/types"
)

// CmdImport creates an empty filesystem image, imports the contents of the tarball into the image, and optionally tags the image.
//
// The URL argument is the address of a tarball (.tar, .tar.gz, .tgz, .bzip, .tar.xz, .txz) file or a path to local file relative to docker client. If the URL is '-', then the tar file is read from STDIN.
//
// Usage: docker import [OPTIONS] file|URL|- [REPOSITORY[:TAG]]
func (cli *DockerCli) CmdImport(args ...string) error {
	cmd := Cli.Subcmd("import", []string{"file|URL|- [REPOSITORY[:TAG]]"}, Cli.DockerCommands["import"].Description, true)
	flChanges := opts.NewListOpts(nil)
	cmd.Var(&flChanges, []string{"c", "-change"}, "Apply Dockerfile instruction to the created image")
	message := cmd.String([]string{"m", "-message"}, "", "Set commit message for imported image")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var (
		in         io.Reader
		tag        string
		src        = cmd.Arg(0)
		srcName    = src
		repository = cmd.Arg(1)
		changes    = flChanges.GetAll()
	)

	if cmd.NArg() == 3 {
		fmt.Fprintf(cli.err, "[DEPRECATED] The format 'file|URL|- [REPOSITORY [TAG]]' has been deprecated. Please use file|URL|- [REPOSITORY[:TAG]]\n")
		tag = cmd.Arg(2)
	}

	if repository != "" {
		//Check if the given image name can be resolved
		if _, err := reference.ParseNamed(repository); err != nil {
			return err
		}
	}

	if src == "-" {
		in = cli.in
	} else if !urlutil.IsURL(src) {
		srcName = "-"
		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()
		in = file
	}

	options := types.ImageImportOptions{
		Source:         in,
		SourceName:     srcName,
		RepositoryName: repository,
		Message:        *message,
		Tag:            tag,
		Changes:        changes,
	}

	responseBody, err := cli.client.ImageImport(context.Background(), options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}
`


var info_go_scope = godebug.EnteringNewFile(client_pkg_scope, info_go_contents)

func (cli *DockerCli) CmdInfo(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdInfo(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := info_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 20)
	cmd := Cli.Subcmd("info", nil, Cli.DockerCommands["info"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 21)
	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 23)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 25)

	info, err := cli.client.Info(context.Background())
	scope.Declare("info", &info, "err", &err)
	godebug.Line(ctx, scope, 26)
	if err != nil {
		godebug.Line(ctx, scope, 27)
		return err
	}
	godebug.Line(ctx, scope, 30)

	fmt.Fprintf(cli.out, "Containers: %d\n", info.Containers)
	godebug.Line(ctx, scope, 31)
	fmt.Fprintf(cli.out, " Running: %d\n", info.ContainersRunning)
	godebug.Line(ctx, scope, 32)
	fmt.Fprintf(cli.out, " Paused: %d\n", info.ContainersPaused)
	godebug.Line(ctx, scope, 33)
	fmt.Fprintf(cli.out, " Stopped: %d\n", info.ContainersStopped)
	godebug.Line(ctx, scope, 34)
	fmt.Fprintf(cli.out, "Images: %d\n", info.Images)
	godebug.Line(ctx, scope, 35)
	ioutils.FprintfIfNotEmpty(cli.out, "Server Version: %s\n", info.ServerVersion)
	godebug.Line(ctx, scope, 36)
	ioutils.FprintfIfNotEmpty(cli.out, "Storage Driver: %s\n", info.Driver)
	godebug.Line(ctx, scope, 37)
	if info.DriverStatus != nil {
		{
			scope := scope.EnteringNewChildScope()
			for _, pair := range info.DriverStatus {
				godebug.Line(ctx, scope, 38)
				scope.Declare("pair", &pair)
				godebug.Line(ctx, scope, 39)
				fmt.Fprintf(cli.out, " %s: %s\n", pair[0], pair[1])
				godebug.Line(ctx, scope, 42)

				if pair[0] == "Data loop file" {
					godebug.Line(ctx, scope, 43)
					fmt.Fprintln(cli.err, " WARNING: Usage of loopback devices is strongly discouraged for production use. Either use `--storage-opt dm.thinpooldev` or use `--storage-opt dm.no_warn_on_loop_devices=true` to suppress this warning.")
				}
			}
			godebug.Line(ctx, scope, 38)
		}

	}
	godebug.Line(ctx, scope, 48)
	if info.SystemStatus != nil {
		{
			scope := scope.EnteringNewChildScope()
			for _, pair := range info.SystemStatus {
				godebug.Line(ctx, scope, 49)
				scope.Declare("pair", &pair)
				godebug.Line(ctx, scope, 50)
				fmt.Fprintf(cli.out, "%s: %s\n", pair[0], pair[1])
			}
			godebug.Line(ctx, scope, 49)
		}
	}
	godebug.Line(ctx, scope, 53)
	ioutils.FprintfIfNotEmpty(cli.out, "Execution Driver: %s\n", info.ExecutionDriver)
	godebug.Line(ctx, scope, 54)
	ioutils.FprintfIfNotEmpty(cli.out, "Logging Driver: %s\n", info.LoggingDriver)
	godebug.Line(ctx, scope, 55)
	ioutils.FprintfIfNotEmpty(cli.out, "Cgroup Driver: %s\n", info.CgroupDriver)
	godebug.Line(ctx, scope, 57)

	fmt.Fprintf(cli.out, "Plugins: \n")
	godebug.Line(ctx, scope, 58)
	fmt.Fprintf(cli.out, " Volume:")
	godebug.Line(ctx, scope, 59)
	fmt.Fprintf(cli.out, " %s", strings.Join(info.Plugins.Volume, " "))
	godebug.Line(ctx, scope, 60)
	fmt.Fprintf(cli.out, "\n")
	godebug.Line(ctx, scope, 61)
	fmt.Fprintf(cli.out, " Network:")
	godebug.Line(ctx, scope, 62)
	fmt.Fprintf(cli.out, " %s", strings.Join(info.Plugins.Network, " "))
	godebug.Line(ctx, scope, 63)
	fmt.Fprintf(cli.out, "\n")
	godebug.Line(ctx, scope, 65)

	if len(info.Plugins.Authorization) != 0 {
		godebug.Line(ctx, scope, 66)
		fmt.Fprintf(cli.out, " Authorization:")
		godebug.Line(ctx, scope, 67)
		fmt.Fprintf(cli.out, " %s", strings.Join(info.Plugins.Authorization, " "))
		godebug.Line(ctx, scope, 68)
		fmt.Fprintf(cli.out, "\n")
	}
	godebug.Line(ctx, scope, 71)

	ioutils.FprintfIfNotEmpty(cli.out, "Kernel Version: %s\n", info.KernelVersion)
	godebug.Line(ctx, scope, 72)
	ioutils.FprintfIfNotEmpty(cli.out, "Operating System: %s\n", info.OperatingSystem)
	godebug.Line(ctx, scope, 73)
	ioutils.FprintfIfNotEmpty(cli.out, "OSType: %s\n", info.OSType)
	godebug.Line(ctx, scope, 74)
	ioutils.FprintfIfNotEmpty(cli.out, "Architecture: %s\n", info.Architecture)
	godebug.Line(ctx, scope, 75)
	fmt.Fprintf(cli.out, "CPUs: %d\n", info.NCPU)
	godebug.Line(ctx, scope, 76)
	fmt.Fprintf(cli.out, "Total Memory: %s\n", units.BytesSize(float64(info.MemTotal)))
	godebug.Line(ctx, scope, 77)
	ioutils.FprintfIfNotEmpty(cli.out, "Name: %s\n", info.Name)
	godebug.Line(ctx, scope, 78)
	ioutils.FprintfIfNotEmpty(cli.out, "ID: %s\n", info.ID)
	godebug.Line(ctx, scope, 79)
	fmt.Fprintf(cli.out, "Docker Root Dir: %s\n", info.DockerRootDir)
	godebug.Line(ctx, scope, 80)
	fmt.Fprintf(cli.out, "Debug Mode (client): %v\n", utils.IsDebugEnabled())
	godebug.Line(ctx, scope, 81)
	fmt.Fprintf(cli.out, "Debug Mode (server): %v\n", info.Debug)
	godebug.Line(ctx, scope, 83)

	if info.Debug {
		godebug.Line(ctx, scope, 84)
		fmt.Fprintf(cli.out, " File Descriptors: %d\n", info.NFd)
		godebug.Line(ctx, scope, 85)
		fmt.Fprintf(cli.out, " Goroutines: %d\n", info.NGoroutines)
		godebug.Line(ctx, scope, 86)
		fmt.Fprintf(cli.out, " System Time: %s\n", info.SystemTime)
		godebug.Line(ctx, scope, 87)
		fmt.Fprintf(cli.out, " EventsListeners: %d\n", info.NEventsListener)
	}
	godebug.Line(ctx, scope, 90)

	ioutils.FprintfIfNotEmpty(cli.out, "Http Proxy: %s\n", info.HTTPProxy)
	godebug.Line(ctx, scope, 91)
	ioutils.FprintfIfNotEmpty(cli.out, "Https Proxy: %s\n", info.HTTPSProxy)
	godebug.Line(ctx, scope, 92)
	ioutils.FprintfIfNotEmpty(cli.out, "No Proxy: %s\n", info.NoProxy)
	godebug.Line(ctx, scope, 94)

	if info.IndexServerAddress != "" {
		godebug.Line(ctx, scope, 95)
		u := cli.configFile.AuthConfigs[info.IndexServerAddress].Username
		scope := scope.EnteringNewChildScope()
		scope.Declare("u", &u)
		godebug.Line(ctx, scope, 96)
		if len(u) > 0 {
			godebug.Line(ctx, scope, 97)
			fmt.Fprintf(cli.out, "Username: %v\n", u)
		}
		godebug.Line(ctx, scope, 99)
		fmt.Fprintf(cli.out, "Registry: %v\n", info.IndexServerAddress)
	}
	godebug.Line(ctx, scope, 103)

	if info.OSType != "windows" {
		godebug.Line(ctx, scope, 104)
		if !info.MemoryLimit {
			godebug.Line(ctx, scope, 105)
			fmt.Fprintln(cli.err, "WARNING: No memory limit support")
		}
		godebug.Line(ctx, scope, 107)
		if !info.SwapLimit {
			godebug.Line(ctx, scope, 108)
			fmt.Fprintln(cli.err, "WARNING: No swap limit support")
		}
		godebug.Line(ctx, scope, 110)
		if !info.KernelMemory {
			godebug.Line(ctx, scope, 111)
			fmt.Fprintln(cli.err, "WARNING: No kernel memory limit support")
		}
		godebug.Line(ctx, scope, 113)
		if !info.OomKillDisable {
			godebug.Line(ctx, scope, 114)
			fmt.Fprintln(cli.err, "WARNING: No oom kill disable support")
		}
		godebug.Line(ctx, scope, 116)
		if !info.CPUCfsQuota {
			godebug.Line(ctx, scope, 117)
			fmt.Fprintln(cli.err, "WARNING: No cpu cfs quota support")
		}
		godebug.Line(ctx, scope, 119)
		if !info.CPUCfsPeriod {
			godebug.Line(ctx, scope, 120)
			fmt.Fprintln(cli.err, "WARNING: No cpu cfs period support")
		}
		godebug.Line(ctx, scope, 122)
		if !info.CPUShares {
			godebug.Line(ctx, scope, 123)
			fmt.Fprintln(cli.err, "WARNING: No cpu shares support")
		}
		godebug.Line(ctx, scope, 125)
		if !info.CPUSet {
			godebug.Line(ctx, scope, 126)
			fmt.Fprintln(cli.err, "WARNING: No cpuset support")
		}
		godebug.Line(ctx, scope, 128)
		if !info.IPv4Forwarding {
			godebug.Line(ctx, scope, 129)
			fmt.Fprintln(cli.err, "WARNING: IPv4 forwarding is disabled")
		}
		godebug.Line(ctx, scope, 131)
		if !info.BridgeNfIptables {
			godebug.Line(ctx, scope, 132)
			fmt.Fprintln(cli.err, "WARNING: bridge-nf-call-iptables is disabled")
		}
		godebug.Line(ctx, scope, 134)
		if !info.BridgeNfIP6tables {
			godebug.Line(ctx, scope, 135)
			fmt.Fprintln(cli.err, "WARNING: bridge-nf-call-ip6tables is disabled")
		}
	}
	godebug.Line(ctx, scope, 139)

	if info.Labels != nil {
		godebug.Line(ctx, scope, 140)
		fmt.Fprintln(cli.out, "Labels:")
		{
			scope := scope.EnteringNewChildScope()
			for _, attribute := range info.Labels {
				godebug.Line(ctx, scope, 141)
				scope.Declare("attribute", &attribute)
				godebug.Line(ctx, scope, 142)
				fmt.Fprintf(cli.out, " %s\n", attribute)
			}
			godebug.Line(ctx, scope, 141)
		}
	}
	godebug.Line(ctx, scope, 146)

	ioutils.FprintfIfTrue(cli.out, "Experimental: %v\n", info.ExperimentalBuild)
	godebug.Line(ctx, scope, 147)
	if info.ClusterStore != "" {
		godebug.Line(ctx, scope, 148)
		fmt.Fprintf(cli.out, "Cluster Store: %s\n", info.ClusterStore)
	}
	godebug.Line(ctx, scope, 151)

	if info.ClusterAdvertise != "" {
		godebug.Line(ctx, scope, 152)
		fmt.Fprintf(cli.out, "Cluster Advertise: %s\n", info.ClusterAdvertise)
	}
	godebug.Line(ctx, scope, 155)

	if info.RegistryConfig != nil && (len(info.RegistryConfig.InsecureRegistryCIDRs) > 0 || len(info.RegistryConfig.IndexConfigs) > 0) {
		godebug.Line(ctx, scope, 156)
		fmt.Fprintln(cli.out, "Insecure registries:")
		{
			scope := scope.EnteringNewChildScope()
			for _, registry := range info.RegistryConfig.IndexConfigs {
				godebug.Line(ctx, scope, 157)
				scope.Declare("registry", &registry)
				godebug.Line(ctx, scope, 158)
				if registry.Secure == false {
					godebug.Line(ctx, scope, 159)
					fmt.Fprintf(cli.out, " %s\n", registry.Name)
				}
			}
			godebug.Line(ctx, scope, 157)
		}
		{
			scope := scope.EnteringNewChildScope()

			for _, registry := range info.RegistryConfig.InsecureRegistryCIDRs {
				godebug.Line(ctx, scope, 163)
				scope.Declare("registry", &registry)
				godebug.Line(ctx, scope, 164)
				mask, _ := registry.Mask.Size()
				scope := scope.EnteringNewChildScope()
				scope.Declare("mask", &mask)
				godebug.Line(ctx, scope, 165)
				fmt.Fprintf(cli.out, " %s/%d\n", registry.IP.String(), mask)
			}
			godebug.Line(ctx, scope, 163)
		}
	}
	godebug.Line(ctx, scope, 168)
	return nil
}

var info_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/ioutils"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/go-units"
)

// CmdInfo displays system-wide information.
//
// Usage: docker info
func (cli *DockerCli) CmdInfo(args ...string) error {
	cmd := Cli.Subcmd("info", nil, Cli.DockerCommands["info"].Description, true)
	cmd.Require(flag.Exact, 0)

	cmd.ParseFlags(args, true)

	info, err := cli.client.Info(context.Background())
	if err != nil {
		return err
	}

	fmt.Fprintf(cli.out, "Containers: %d\n", info.Containers)
	fmt.Fprintf(cli.out, " Running: %d\n", info.ContainersRunning)
	fmt.Fprintf(cli.out, " Paused: %d\n", info.ContainersPaused)
	fmt.Fprintf(cli.out, " Stopped: %d\n", info.ContainersStopped)
	fmt.Fprintf(cli.out, "Images: %d\n", info.Images)
	ioutils.FprintfIfNotEmpty(cli.out, "Server Version: %s\n", info.ServerVersion)
	ioutils.FprintfIfNotEmpty(cli.out, "Storage Driver: %s\n", info.Driver)
	if info.DriverStatus != nil {
		for _, pair := range info.DriverStatus {
			fmt.Fprintf(cli.out, " %s: %s\n", pair[0], pair[1])

			// print a warning if devicemapper is using a loopback file
			if pair[0] == "Data loop file" {
				fmt.Fprintln(cli.err, " WARNING: Usage of loopback devices is strongly discouraged for production use. Either use ` + "`" + `--storage-opt dm.thinpooldev` + "`" + ` or use ` + "`" + `--storage-opt dm.no_warn_on_loop_devices=true` + "`" + ` to suppress this warning.")
			}
		}

	}
	if info.SystemStatus != nil {
		for _, pair := range info.SystemStatus {
			fmt.Fprintf(cli.out, "%s: %s\n", pair[0], pair[1])
		}
	}
	ioutils.FprintfIfNotEmpty(cli.out, "Execution Driver: %s\n", info.ExecutionDriver)
	ioutils.FprintfIfNotEmpty(cli.out, "Logging Driver: %s\n", info.LoggingDriver)
	ioutils.FprintfIfNotEmpty(cli.out, "Cgroup Driver: %s\n", info.CgroupDriver)

	fmt.Fprintf(cli.out, "Plugins: \n")
	fmt.Fprintf(cli.out, " Volume:")
	fmt.Fprintf(cli.out, " %s", strings.Join(info.Plugins.Volume, " "))
	fmt.Fprintf(cli.out, "\n")
	fmt.Fprintf(cli.out, " Network:")
	fmt.Fprintf(cli.out, " %s", strings.Join(info.Plugins.Network, " "))
	fmt.Fprintf(cli.out, "\n")

	if len(info.Plugins.Authorization) != 0 {
		fmt.Fprintf(cli.out, " Authorization:")
		fmt.Fprintf(cli.out, " %s", strings.Join(info.Plugins.Authorization, " "))
		fmt.Fprintf(cli.out, "\n")
	}

	ioutils.FprintfIfNotEmpty(cli.out, "Kernel Version: %s\n", info.KernelVersion)
	ioutils.FprintfIfNotEmpty(cli.out, "Operating System: %s\n", info.OperatingSystem)
	ioutils.FprintfIfNotEmpty(cli.out, "OSType: %s\n", info.OSType)
	ioutils.FprintfIfNotEmpty(cli.out, "Architecture: %s\n", info.Architecture)
	fmt.Fprintf(cli.out, "CPUs: %d\n", info.NCPU)
	fmt.Fprintf(cli.out, "Total Memory: %s\n", units.BytesSize(float64(info.MemTotal)))
	ioutils.FprintfIfNotEmpty(cli.out, "Name: %s\n", info.Name)
	ioutils.FprintfIfNotEmpty(cli.out, "ID: %s\n", info.ID)
	fmt.Fprintf(cli.out, "Docker Root Dir: %s\n", info.DockerRootDir)
	fmt.Fprintf(cli.out, "Debug Mode (client): %v\n", utils.IsDebugEnabled())
	fmt.Fprintf(cli.out, "Debug Mode (server): %v\n", info.Debug)

	if info.Debug {
		fmt.Fprintf(cli.out, " File Descriptors: %d\n", info.NFd)
		fmt.Fprintf(cli.out, " Goroutines: %d\n", info.NGoroutines)
		fmt.Fprintf(cli.out, " System Time: %s\n", info.SystemTime)
		fmt.Fprintf(cli.out, " EventsListeners: %d\n", info.NEventsListener)
	}

	ioutils.FprintfIfNotEmpty(cli.out, "Http Proxy: %s\n", info.HTTPProxy)
	ioutils.FprintfIfNotEmpty(cli.out, "Https Proxy: %s\n", info.HTTPSProxy)
	ioutils.FprintfIfNotEmpty(cli.out, "No Proxy: %s\n", info.NoProxy)

	if info.IndexServerAddress != "" {
		u := cli.configFile.AuthConfigs[info.IndexServerAddress].Username
		if len(u) > 0 {
			fmt.Fprintf(cli.out, "Username: %v\n", u)
		}
		fmt.Fprintf(cli.out, "Registry: %v\n", info.IndexServerAddress)
	}

	// Only output these warnings if the server does not support these features
	if info.OSType != "windows" {
		if !info.MemoryLimit {
			fmt.Fprintln(cli.err, "WARNING: No memory limit support")
		}
		if !info.SwapLimit {
			fmt.Fprintln(cli.err, "WARNING: No swap limit support")
		}
		if !info.KernelMemory {
			fmt.Fprintln(cli.err, "WARNING: No kernel memory limit support")
		}
		if !info.OomKillDisable {
			fmt.Fprintln(cli.err, "WARNING: No oom kill disable support")
		}
		if !info.CPUCfsQuota {
			fmt.Fprintln(cli.err, "WARNING: No cpu cfs quota support")
		}
		if !info.CPUCfsPeriod {
			fmt.Fprintln(cli.err, "WARNING: No cpu cfs period support")
		}
		if !info.CPUShares {
			fmt.Fprintln(cli.err, "WARNING: No cpu shares support")
		}
		if !info.CPUSet {
			fmt.Fprintln(cli.err, "WARNING: No cpuset support")
		}
		if !info.IPv4Forwarding {
			fmt.Fprintln(cli.err, "WARNING: IPv4 forwarding is disabled")
		}
		if !info.BridgeNfIptables {
			fmt.Fprintln(cli.err, "WARNING: bridge-nf-call-iptables is disabled")
		}
		if !info.BridgeNfIP6tables {
			fmt.Fprintln(cli.err, "WARNING: bridge-nf-call-ip6tables is disabled")
		}
	}

	if info.Labels != nil {
		fmt.Fprintln(cli.out, "Labels:")
		for _, attribute := range info.Labels {
			fmt.Fprintf(cli.out, " %s\n", attribute)
		}
	}

	ioutils.FprintfIfTrue(cli.out, "Experimental: %v\n", info.ExperimentalBuild)
	if info.ClusterStore != "" {
		fmt.Fprintf(cli.out, "Cluster Store: %s\n", info.ClusterStore)
	}

	if info.ClusterAdvertise != "" {
		fmt.Fprintf(cli.out, "Cluster Advertise: %s\n", info.ClusterAdvertise)
	}

	if info.RegistryConfig != nil && (len(info.RegistryConfig.InsecureRegistryCIDRs) > 0 || len(info.RegistryConfig.IndexConfigs) > 0) {
		fmt.Fprintln(cli.out, "Insecure registries:")
		for _, registry := range info.RegistryConfig.IndexConfigs {
			if registry.Secure == false {
				fmt.Fprintf(cli.out, " %s\n", registry.Name)
			}
		}

		for _, registry := range info.RegistryConfig.InsecureRegistryCIDRs {
			mask, _ := registry.Mask.Size()
			fmt.Fprintf(cli.out, " %s/%d\n", registry.IP.String(), mask)
		}
	}
	return nil
}
`


var inspect_go_scope = godebug.EnteringNewFile(client_pkg_scope, inspect_go_contents)

func (cli *DockerCli) CmdInspect(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdInspect(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("inspect", []string{"CONTAINER|IMAGE [CONTAINER|IMAGE...]"}, Cli.DockerCommands["inspect"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")
	scope.Declare("tmplStr", &tmplStr)
	godebug.Line(ctx, scope, 21)
	inspectType := cmd.String([]string{"-type"}, "", "Return JSON for specified type, (e.g image or container)")
	scope.Declare("inspectType", &inspectType)
	godebug.Line(ctx, scope, 22)
	size := cmd.Bool([]string{"s", "-size"}, false, "Display total file sizes if the type is container")
	scope.Declare("size", &size)
	godebug.Line(ctx, scope, 23)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 25)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 27)

	if *inspectType != "" && *inspectType != "container" && *inspectType != "image" {
		godebug.Line(ctx, scope, 28)
		return fmt.Errorf("%q is not a valid value for --type", *inspectType)
	}
	godebug.Line(ctx, scope, 31)

	var elementSearcher inspectSearcher
	scope.Declare("elementSearcher", &elementSearcher)
	godebug.Line(ctx, scope, 32)
	switch *inspectType {
	case godebug.Case(ctx, scope, 33):
		fallthrough
	case "container":
		godebug.Line(ctx, scope, 34)
		elementSearcher = cli.inspectContainers(*size)
	case godebug.Case(ctx, scope, 35):
		fallthrough
	case "image":
		godebug.Line(ctx, scope, 36)
		elementSearcher = cli.inspectImages(*size)
	default:
		godebug.Line(ctx, scope, 37)
		godebug.Line(ctx, scope, 38)
		elementSearcher = cli.inspectAll(*size)
	}
	godebug.Line(ctx, scope, 41)

	return cli.inspectElements(*tmplStr, cmd.Args(), elementSearcher)
}

func (cli *DockerCli) inspectContainers(getSize bool) inspectSearcher {
	var result1 inspectSearcher
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.inspectContainers(getSize)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "getSize", &getSize)
	godebug.Line(ctx, scope, 45)
	return func(ref string) (interface{}, []byte, error) {
		var result1 interface {
		}
		var result2 []byte
		var result3 error
		fn := func(ctx *godebug.Context) {
			result1, result2, result3 = func() (interface {
			}, []byte, error) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("ref", &ref)
				godebug.Line(ctx, scope, 46)
				return cli.client.ContainerInspectWithRaw(context.Background(), ref, getSize)
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2, result3
	}

}

func (cli *DockerCli) inspectImages(getSize bool) inspectSearcher {
	var result1 inspectSearcher
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.inspectImages(getSize)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "getSize", &getSize)
	godebug.Line(ctx, scope, 51)
	return func(ref string) (interface{}, []byte, error) {
		var result1 interface {
		}
		var result2 []byte
		var result3 error
		fn := func(ctx *godebug.Context) {
			result1, result2, result3 = func() (interface {
			}, []byte, error) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("ref", &ref)
				godebug.Line(ctx, scope, 52)
				return cli.client.ImageInspectWithRaw(context.Background(), ref, getSize)
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2, result3
	}

}

func (cli *DockerCli) inspectAll(getSize bool) inspectSearcher {
	var result1 inspectSearcher
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.inspectAll(getSize)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "getSize", &getSize)
	godebug.Line(ctx, scope, 57)
	return func(ref string) (interface{}, []byte, error) {
		var result1 interface {
		}
		var result2 []byte
		var result3 error
		fn := func(ctx *godebug.Context) {
			result1, result2, result3 = func() (interface {
			}, []byte, error) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("ref", &ref)
				godebug.Line(ctx, scope, 58)
				c, rawContainer, err := cli.client.ContainerInspectWithRaw(context.Background(), ref, getSize)
				scope.Declare("c", &c, "rawContainer", &rawContainer, "err", &err)
				godebug.Line(ctx, scope, 59)
				if err != nil {
					godebug.Line(ctx, scope, 61)
					if apiclient.IsErrContainerNotFound(err) {
						godebug.Line(ctx, scope, 62)
						i, rawImage, err := cli.client.ImageInspectWithRaw(context.Background(), ref, getSize)
						scope := scope.EnteringNewChildScope()
						scope.Declare("i", &i, "rawImage", &rawImage, "err", &err)
						godebug.Line(ctx, scope, 63)
						if err != nil {
							godebug.Line(ctx, scope, 64)
							if apiclient.IsErrImageNotFound(err) {
								godebug.Line(ctx, scope, 65)
								return nil, nil, fmt.Errorf("Error: No such image or container: %s", ref)
							}
							godebug.Line(ctx, scope, 67)
							return nil, nil, err
						}
						godebug.Line(ctx, scope, 69)
						return i, rawImage, err
					}
					godebug.Line(ctx, scope, 71)
					return nil, nil, err
				}
				godebug.Line(ctx, scope, 73)
				return c, rawContainer, err
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2, result3
	}

}

type inspectSearcher func(ref string) (interface{}, []byte, error)

func (cli *DockerCli) inspectElements(tmplStr string, references []string, searchByReference inspectSearcher) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.inspectElements(tmplStr, references, searchByReference)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "tmplStr", &tmplStr, "references", &references, "searchByReference", &searchByReference)
	godebug.Line(ctx, scope, 80)
	elementInspector, err := cli.newInspectorWithTemplate(tmplStr)
	scope.Declare("elementInspector", &elementInspector, "err", &err)
	godebug.Line(ctx, scope, 81)
	if err != nil {
		godebug.Line(ctx, scope, 82)
		return Cli.StatusError{StatusCode: 64, Status: err.Error()}
	}
	godebug.Line(ctx, scope, 85)

	var inspectErr error
	scope.Declare("inspectErr", &inspectErr)
	{
		scope := scope.EnteringNewChildScope()
		for _, ref := range references {
			godebug.Line(ctx, scope, 86)
			scope.Declare("ref", &ref)
			godebug.Line(ctx, scope, 87)
			element, raw, err := searchByReference(ref)
			scope := scope.EnteringNewChildScope()
			scope.Declare("element", &element, "raw", &raw, "err", &err)
			godebug.Line(ctx, scope, 88)
			if err != nil {
				godebug.Line(ctx, scope, 89)
				inspectErr = err
				godebug.Line(ctx, scope, 90)
				break
			}
			godebug.Line(ctx, scope, 93)

			if err := elementInspector.Inspect(element, raw); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 94)
				inspectErr = err
				godebug.Line(ctx, scope, 95)
				break
			}
		}
		godebug.Line(ctx, scope, 86)
	}
	godebug.Line(ctx, scope, 99)

	if err := elementInspector.Flush(); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 100)
		cli.inspectErrorStatus(err)
	}
	godebug.Line(ctx, scope, 103)

	if status := cli.inspectErrorStatus(inspectErr); status != 0 {
		scope := scope.EnteringNewChildScope()
		scope.Declare("status", &status)
		godebug.Line(ctx, scope, 104)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(ctx, scope, 106)
	return nil
}

func (cli *DockerCli) inspectErrorStatus(err error) (status int) {
	ctx, ok := godebug.EnterFunc(func() {
		status = cli.inspectErrorStatus(err)
	})
	if !ok {
		return status
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "err", &err, "status", &status)
	godebug.Line(ctx, scope, 110)
	if err != nil {
		godebug.Line(ctx, scope, 111)
		fmt.Fprintf(cli.err, "%s\n", err)
		godebug.Line(ctx, scope, 112)
		status = 1
	}
	godebug.Line(ctx, scope, 114)
	return
}

func (cli *DockerCli) newInspectorWithTemplate(tmplStr string) (inspect.Inspector, error) {
	var result1 inspect.Inspector
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = cli.newInspectorWithTemplate(tmplStr)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := inspect_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "tmplStr", &tmplStr)
	godebug.Line(ctx, scope, 118)
	elementInspector := inspect.NewIndentedInspector(cli.out)
	scope.Declare("elementInspector", &elementInspector)
	godebug.Line(ctx, scope, 119)
	if tmplStr != "" {
		godebug.Line(ctx, scope, 120)
		tmpl, err := templates.Parse(tmplStr)
		scope := scope.EnteringNewChildScope()
		scope.Declare("tmpl", &tmpl, "err", &err)
		godebug.Line(ctx, scope, 121)
		if err != nil {
			godebug.Line(ctx, scope, 122)
			return nil, fmt.Errorf("Template parsing error: %s", err)
		}
		godebug.Line(ctx, scope, 124)
		elementInspector = inspect.NewTemplateInspector(cli.out, tmpl)
	}
	godebug.Line(ctx, scope, 126)
	return elementInspector, nil
}

var inspect_go_contents = `package client

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client/inspect"
	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils/templates"
	"github.com/docker/engine-api/client"
)

// CmdInspect displays low-level information on one or more containers or images.
//
// Usage: docker inspect [OPTIONS] CONTAINER|IMAGE [CONTAINER|IMAGE...]
func (cli *DockerCli) CmdInspect(args ...string) error {
	cmd := Cli.Subcmd("inspect", []string{"CONTAINER|IMAGE [CONTAINER|IMAGE...]"}, Cli.DockerCommands["inspect"].Description, true)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")
	inspectType := cmd.String([]string{"-type"}, "", "Return JSON for specified type, (e.g image or container)")
	size := cmd.Bool([]string{"s", "-size"}, false, "Display total file sizes if the type is container")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	if *inspectType != "" && *inspectType != "container" && *inspectType != "image" {
		return fmt.Errorf("%q is not a valid value for --type", *inspectType)
	}

	var elementSearcher inspectSearcher
	switch *inspectType {
	case "container":
		elementSearcher = cli.inspectContainers(*size)
	case "image":
		elementSearcher = cli.inspectImages(*size)
	default:
		elementSearcher = cli.inspectAll(*size)
	}

	return cli.inspectElements(*tmplStr, cmd.Args(), elementSearcher)
}

func (cli *DockerCli) inspectContainers(getSize bool) inspectSearcher {
	return func(ref string) (interface{}, []byte, error) {
		return cli.client.ContainerInspectWithRaw(context.Background(), ref, getSize)
	}
}

func (cli *DockerCli) inspectImages(getSize bool) inspectSearcher {
	return func(ref string) (interface{}, []byte, error) {
		return cli.client.ImageInspectWithRaw(context.Background(), ref, getSize)
	}
}

func (cli *DockerCli) inspectAll(getSize bool) inspectSearcher {
	return func(ref string) (interface{}, []byte, error) {
		c, rawContainer, err := cli.client.ContainerInspectWithRaw(context.Background(), ref, getSize)
		if err != nil {
			// Search for image with that id if a container doesn't exist.
			if client.IsErrContainerNotFound(err) {
				i, rawImage, err := cli.client.ImageInspectWithRaw(context.Background(), ref, getSize)
				if err != nil {
					if client.IsErrImageNotFound(err) {
						return nil, nil, fmt.Errorf("Error: No such image or container: %s", ref)
					}
					return nil, nil, err
				}
				return i, rawImage, err
			}
			return nil, nil, err
		}
		return c, rawContainer, err
	}
}

type inspectSearcher func(ref string) (interface{}, []byte, error)

func (cli *DockerCli) inspectElements(tmplStr string, references []string, searchByReference inspectSearcher) error {
	elementInspector, err := cli.newInspectorWithTemplate(tmplStr)
	if err != nil {
		return Cli.StatusError{StatusCode: 64, Status: err.Error()}
	}

	var inspectErr error
	for _, ref := range references {
		element, raw, err := searchByReference(ref)
		if err != nil {
			inspectErr = err
			break
		}

		if err := elementInspector.Inspect(element, raw); err != nil {
			inspectErr = err
			break
		}
	}

	if err := elementInspector.Flush(); err != nil {
		cli.inspectErrorStatus(err)
	}

	if status := cli.inspectErrorStatus(inspectErr); status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

func (cli *DockerCli) inspectErrorStatus(err error) (status int) {
	if err != nil {
		fmt.Fprintf(cli.err, "%s\n", err)
		status = 1
	}
	return
}

func (cli *DockerCli) newInspectorWithTemplate(tmplStr string) (inspect.Inspector, error) {
	elementInspector := inspect.NewIndentedInspector(cli.out)
	if tmplStr != "" {
		tmpl, err := templates.Parse(tmplStr)
		if err != nil {
			return nil, fmt.Errorf("Template parsing error: %s", err)
		}
		elementInspector = inspect.NewTemplateInspector(cli.out, tmpl)
	}
	return elementInspector, nil
}
`


var kill_go_scope = godebug.EnteringNewFile(client_pkg_scope, kill_go_contents)

func (cli *DockerCli) CmdKill(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdKill(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := kill_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 17)
	cmd := Cli.Subcmd("kill", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["kill"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 18)
	signal := cmd.String([]string{"s", "-signal"}, "KILL", "Signal to send to the container")
	scope.Declare("signal", &signal)
	godebug.Line(ctx, scope, 19)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 21)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 23)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 24)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 25)
			if err := cli.client.ContainerKill(context.Background(), name, *signal); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 26)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 27)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 28)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 24)
	}
	godebug.Line(ctx, scope, 31)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 32)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 34)
	return nil
}

var kill_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdKill kills one or more running container using SIGKILL or a specified signal.
//
// Usage: docker kill [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdKill(args ...string) error {
	cmd := Cli.Subcmd("kill", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["kill"].Description, true)
	signal := cmd.String([]string{"s", "-signal"}, "KILL", "Signal to send to the container")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		if err := cli.client.ContainerKill(context.Background(), name, *signal); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var load_go_scope = godebug.EnteringNewFile(client_pkg_scope, load_go_contents)

func (cli *DockerCli) CmdLoad(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdLoad(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := load_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 20)
	cmd := Cli.Subcmd("load", nil, Cli.DockerCommands["load"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 21)
	infile := cmd.String([]string{"i", "-input"}, "", "Read from a tar archive file, instead of STDIN")
	scope.Declare("infile", &infile)
	godebug.Line(ctx, scope, 22)
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Suppress the load output")
	scope.Declare("quiet", &quiet)
	godebug.Line(ctx, scope, 23)
	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 24)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 26)

	var input io.Reader = cli.in
	scope.Declare("input", &input)
	godebug.Line(ctx, scope, 27)
	if *infile != "" {
		godebug.Line(ctx, scope, 28)
		file, err := os.Open(*infile)
		scope := scope.EnteringNewChildScope()
		scope.Declare("file", &file, "err", &err)
		godebug.Line(ctx, scope, 29)
		if err != nil {
			godebug.Line(ctx, scope, 30)
			return err
		}
		godebug.Line(ctx, scope, 32)
		defer file.Close()
		defer godebug.Defer(ctx, scope, 32)
		godebug.Line(ctx, scope, 33)
		input = file
	}
	godebug.Line(ctx, scope, 35)
	if !cli.isTerminalOut {
		godebug.Line(ctx, scope, 36)
		*quiet = true
	}
	godebug.Line(ctx, scope, 38)
	response, err := cli.client.ImageLoad(context.Background(), input, *quiet)
	scope.Declare("response", &response, "err", &err)
	godebug.Line(ctx, scope, 39)
	if err != nil {
		godebug.Line(ctx, scope, 40)
		return err
	}
	godebug.Line(ctx, scope, 42)
	defer response.Body.Close()
	defer godebug.Defer(ctx, scope, 42)
	godebug.Line(ctx, scope, 44)

	if response.JSON {
		godebug.Line(ctx, scope, 45)
		return jsonmessage.DisplayJSONMessagesStream(response.Body, cli.out, cli.outFd, cli.isTerminalOut, nil)
	}
	godebug.Line(ctx, scope, 48)

	_, err = io.Copy(cli.out, response.Body)
	godebug.Line(ctx, scope, 49)
	return err
}

var load_go_contents = `package client

import (
	"io"
	"os"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdLoad loads an image from a tar archive.
//
// The tar archive is read from STDIN by default, or from a tar archive file.
//
// Usage: docker load [OPTIONS]
func (cli *DockerCli) CmdLoad(args ...string) error {
	cmd := Cli.Subcmd("load", nil, Cli.DockerCommands["load"].Description, true)
	infile := cmd.String([]string{"i", "-input"}, "", "Read from a tar archive file, instead of STDIN")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Suppress the load output")
	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	var input io.Reader = cli.in
	if *infile != "" {
		file, err := os.Open(*infile)
		if err != nil {
			return err
		}
		defer file.Close()
		input = file
	}
	if !cli.isTerminalOut {
		*quiet = true
	}
	response, err := cli.client.ImageLoad(context.Background(), input, *quiet)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.JSON {
		return jsonmessage.DisplayJSONMessagesStream(response.Body, cli.out, cli.outFd, cli.isTerminalOut, nil)
	}

	_, err = io.Copy(cli.out, response.Body)
	return err
}
`


var login_go_scope = godebug.EnteringNewFile(client_pkg_scope, login_go_contents)

func (cli *DockerCli) CmdLogin(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdLogin(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 27)
	cmd := Cli.Subcmd("login", []string{"[SERVER]"}, Cli.DockerCommands["login"].Description+".\nIf no server is specified, the default is defined by the daemon.", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 28)
	cmd.Require(flag.Max, 1)
	godebug.Line(ctx, scope, 30)

	flUser := cmd.String([]string{"u", "-username"}, "", "Username")
	scope.Declare("flUser", &flUser)
	godebug.Line(ctx, scope, 31)
	flPassword := cmd.String([]string{"p", "-password"}, "", "Password")
	scope.Declare("flPassword", &flPassword)
	godebug.Line(ctx, scope, 34)

	cmd.String([]string{"#e", "#-email"}, "", "Email")
	godebug.Line(ctx, scope, 36)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 39)

	if runtime.GOOS == "windows" {
		godebug.Line(ctx, scope, 40)
		cli.in = os.Stdin
	}
	godebug.Line(ctx, scope, 43)

	var serverAddress string
	scope.Declare("serverAddress", &serverAddress)
	godebug.Line(ctx, scope, 44)
	var isDefaultRegistry bool
	scope.Declare("isDefaultRegistry", &isDefaultRegistry)
	godebug.Line(ctx, scope, 45)
	if len(cmd.Args()) > 0 {
		godebug.Line(ctx, scope, 46)
		serverAddress = cmd.Arg(0)
	} else {
		godebug.Line(ctx, scope, 47)
		godebug.Line(ctx, scope, 48)
		serverAddress = cli.electAuthServer()
		godebug.Line(ctx, scope, 49)
		isDefaultRegistry = true
	}
	godebug.Line(ctx, scope, 52)

	authConfig, err := cli.configureAuth(*flUser, *flPassword, serverAddress, isDefaultRegistry)
	scope.Declare("authConfig", &authConfig, "err", &err)
	godebug.Line(ctx, scope, 53)
	if err != nil {
		godebug.Line(ctx, scope, 54)
		return err
	}
	godebug.Line(ctx, scope, 57)

	response, err := cli.client.RegistryLogin(context.Background(), authConfig)
	scope.Declare("response", &response)
	godebug.Line(ctx, scope, 58)
	if err != nil {
		godebug.Line(ctx, scope, 59)
		return err
	}
	godebug.Line(ctx, scope, 62)

	if response.IdentityToken != "" {
		godebug.Line(ctx, scope, 63)
		authConfig.Password = ""
		godebug.Line(ctx, scope, 64)
		authConfig.IdentityToken = response.IdentityToken
	}
	godebug.Line(ctx, scope, 66)
	if err := storeCredentials(cli.configFile, authConfig); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 67)
		return fmt.Errorf("Error saving credentials: %v", err)
	}
	godebug.Line(ctx, scope, 70)

	if response.Status != "" {
		godebug.Line(ctx, scope, 71)
		fmt.Fprintln(cli.out, response.Status)
	}
	godebug.Line(ctx, scope, 73)
	return nil
}

func (cli *DockerCli) promptWithDefault(prompt string, configDefault string) {
	ctx, ok := godebug.EnterFunc(func() {
		cli.promptWithDefault(prompt, configDefault)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "prompt", &prompt, "configDefault", &configDefault)
	godebug.Line(ctx, scope, 77)
	if configDefault == "" {
		godebug.Line(ctx, scope, 78)
		fmt.Fprintf(cli.out, "%s: ", prompt)
	} else {
		godebug.Line(ctx, scope, 79)
		godebug.Line(ctx, scope, 80)
		fmt.Fprintf(cli.out, "%s (%s): ", prompt, configDefault)
	}
}

func (cli *DockerCli) configureAuth(flUser, flPassword, serverAddress string, isDefaultRegistry bool) (types.AuthConfig, error) {
	var result1 types.AuthConfig
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = cli.configureAuth(flUser, flPassword, serverAddress, isDefaultRegistry)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "flUser", &flUser, "flPassword", &flPassword, "serverAddress", &serverAddress, "isDefaultRegistry", &isDefaultRegistry)
	godebug.Line(ctx, scope, 85)
	authconfig, err := getCredentials(cli.configFile, serverAddress)
	scope.Declare("authconfig", &authconfig, "err", &err)
	godebug.Line(ctx, scope, 86)
	if err != nil {
		godebug.Line(ctx, scope, 87)
		return authconfig, err
	}
	godebug.Line(ctx, scope, 90)

	authconfig.Username = strings.TrimSpace(authconfig.Username)
	godebug.Line(ctx, scope, 92)

	if flUser = strings.TrimSpace(flUser); flUser == "" {
		godebug.Line(ctx, scope, 93)
		if isDefaultRegistry {
			godebug.Line(ctx, scope, 95)

			fmt.Fprintln(cli.out, "Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.")
		}
		godebug.Line(ctx, scope, 97)
		cli.promptWithDefault("Username", authconfig.Username)
		godebug.Line(ctx, scope, 98)
		flUser = readInput(cli.in, cli.out)
		godebug.Line(ctx, scope, 99)
		flUser = strings.TrimSpace(flUser)
		godebug.Line(ctx, scope, 100)
		if flUser == "" {
			godebug.Line(ctx, scope, 101)
			flUser = authconfig.Username
		}
	}
	godebug.Line(ctx, scope, 105)

	if flUser == "" {
		godebug.Line(ctx, scope, 106)
		return authconfig, fmt.Errorf("Error: Non-null Username Required")
	}
	godebug.Line(ctx, scope, 109)

	if flPassword == "" {
		godebug.Line(ctx, scope, 110)
		oldState, err := term.SaveState(cli.inFd)
		scope := scope.EnteringNewChildScope()
		scope.Declare("oldState", &oldState, "err", &err)
		godebug.Line(ctx, scope, 111)
		if err != nil {
			godebug.Line(ctx, scope, 112)
			return authconfig, err
		}
		godebug.Line(ctx, scope, 114)
		fmt.Fprintf(cli.out, "Password: ")
		godebug.Line(ctx, scope, 115)
		term.DisableEcho(cli.inFd, oldState)
		godebug.Line(ctx, scope, 117)

		flPassword = readInput(cli.in, cli.out)
		godebug.Line(ctx, scope, 118)
		fmt.Fprint(cli.out, "\n")
		godebug.Line(ctx, scope, 120)

		term.RestoreTerminal(cli.inFd, oldState)
		godebug.Line(ctx, scope, 121)
		if flPassword == "" {
			godebug.Line(ctx, scope, 122)
			return authconfig, fmt.Errorf("Error: Password Required")
		}
	}
	godebug.Line(ctx, scope, 126)

	authconfig.Username = flUser
	godebug.Line(ctx, scope, 127)
	authconfig.Password = flPassword
	godebug.Line(ctx, scope, 128)
	authconfig.ServerAddress = serverAddress
	godebug.Line(ctx, scope, 129)
	authconfig.IdentityToken = ""
	godebug.Line(ctx, scope, 131)

	return authconfig, nil
}

func readInput(in io.Reader, out io.Writer) string {
	var result1 string
	ctx, ok := godebug.EnterFunc(func() {
		result1 = readInput(in, out)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("in", &in, "out", &out)
	godebug.Line(ctx, scope, 135)
	reader := bufio.NewReader(in)
	scope.Declare("reader", &reader)
	godebug.Line(ctx, scope, 136)
	line, _, err := reader.ReadLine()
	scope.Declare("line", &line, "err", &err)
	godebug.Line(ctx, scope, 137)
	if err != nil {
		godebug.Line(ctx, scope, 138)
		fmt.Fprintln(out, err.Error())
		godebug.Line(ctx, scope, 139)
		os.Exit(1)
	}
	godebug.Line(ctx, scope, 141)
	return string(line)
}

func getCredentials(c *cliconfig.ConfigFile, serverAddress string) (types.AuthConfig, error) {
	var result1 types.AuthConfig
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = getCredentials(c, serverAddress)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("c", &c, "serverAddress", &serverAddress)
	godebug.Line(ctx, scope, 147)
	s := loadCredentialsStore(c)
	scope.Declare("s", &s)
	godebug.Line(ctx, scope, 148)
	return s.Get(serverAddress)
}

func getAllCredentials(c *cliconfig.ConfigFile) (map[string]types.AuthConfig, error) {
	var result1 map[string]types.AuthConfig
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = getAllCredentials(c)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("c", &c)
	godebug.Line(ctx, scope, 152)
	s := loadCredentialsStore(c)
	scope.Declare("s", &s)
	godebug.Line(ctx, scope, 153)
	return s.GetAll()
}

func storeCredentials(c *cliconfig.ConfigFile, auth types.AuthConfig) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = storeCredentials(c, auth)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("c", &c, "auth", &auth)
	godebug.Line(ctx, scope, 159)
	s := loadCredentialsStore(c)
	scope.Declare("s", &s)
	godebug.Line(ctx, scope, 160)
	return s.Store(auth)
}

func eraseCredentials(c *cliconfig.ConfigFile, serverAddress string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = eraseCredentials(c, serverAddress)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("c", &c, "serverAddress", &serverAddress)
	godebug.Line(ctx, scope, 166)
	s := loadCredentialsStore(c)
	scope.Declare("s", &s)
	godebug.Line(ctx, scope, 167)
	return s.Erase(serverAddress)
}

func loadCredentialsStore(c *cliconfig.ConfigFile) credentials.Store {
	var result1 credentials.Store
	ctx, ok := godebug.EnterFunc(func() {
		result1 = loadCredentialsStore(c)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := login_go_scope.EnteringNewChildScope()
	scope.Declare("c", &c)
	godebug.Line(ctx, scope, 173)
	if c.CredentialsStore != "" {
		godebug.Line(ctx, scope, 174)
		return credentials.NewNativeStore(c)
	}
	godebug.Line(ctx, scope, 176)
	return credentials.NewFileStore(c)
}

var login_go_contents = `package client

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/cliconfig/credentials"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/types"
)

// CmdLogin logs in a user to a Docker registry service.
//
// If no server is specified, the user will be logged into or registered to the registry's index server.
//
// Usage: docker login SERVER
func (cli *DockerCli) CmdLogin(args ...string) error {
	cmd := Cli.Subcmd("login", []string{"[SERVER]"}, Cli.DockerCommands["login"].Description+".\nIf no server is specified, the default is defined by the daemon.", true)
	cmd.Require(flag.Max, 1)

	flUser := cmd.String([]string{"u", "-username"}, "", "Username")
	flPassword := cmd.String([]string{"p", "-password"}, "", "Password")

	// Deprecated in 1.11: Should be removed in docker 1.13
	cmd.String([]string{"#e", "#-email"}, "", "Email")

	cmd.ParseFlags(args, true)

	// On Windows, force the use of the regular OS stdin stream. Fixes #14336/#14210
	if runtime.GOOS == "windows" {
		cli.in = os.Stdin
	}

	var serverAddress string
	var isDefaultRegistry bool
	if len(cmd.Args()) > 0 {
		serverAddress = cmd.Arg(0)
	} else {
		serverAddress = cli.electAuthServer()
		isDefaultRegistry = true
	}

	authConfig, err := cli.configureAuth(*flUser, *flPassword, serverAddress, isDefaultRegistry)
	if err != nil {
		return err
	}

	response, err := cli.client.RegistryLogin(context.Background(), authConfig)
	if err != nil {
		return err
	}

	if response.IdentityToken != "" {
		authConfig.Password = ""
		authConfig.IdentityToken = response.IdentityToken
	}
	if err := storeCredentials(cli.configFile, authConfig); err != nil {
		return fmt.Errorf("Error saving credentials: %v", err)
	}

	if response.Status != "" {
		fmt.Fprintln(cli.out, response.Status)
	}
	return nil
}

func (cli *DockerCli) promptWithDefault(prompt string, configDefault string) {
	if configDefault == "" {
		fmt.Fprintf(cli.out, "%s: ", prompt)
	} else {
		fmt.Fprintf(cli.out, "%s (%s): ", prompt, configDefault)
	}
}

func (cli *DockerCli) configureAuth(flUser, flPassword, serverAddress string, isDefaultRegistry bool) (types.AuthConfig, error) {
	authconfig, err := getCredentials(cli.configFile, serverAddress)
	if err != nil {
		return authconfig, err
	}

	authconfig.Username = strings.TrimSpace(authconfig.Username)

	if flUser = strings.TrimSpace(flUser); flUser == "" {
		if isDefaultRegistry {
			// if this is a defauly registry (docker hub), then display the following message.
			fmt.Fprintln(cli.out, "Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.")
		}
		cli.promptWithDefault("Username", authconfig.Username)
		flUser = readInput(cli.in, cli.out)
		flUser = strings.TrimSpace(flUser)
		if flUser == "" {
			flUser = authconfig.Username
		}
	}

	if flUser == "" {
		return authconfig, fmt.Errorf("Error: Non-null Username Required")
	}

	if flPassword == "" {
		oldState, err := term.SaveState(cli.inFd)
		if err != nil {
			return authconfig, err
		}
		fmt.Fprintf(cli.out, "Password: ")
		term.DisableEcho(cli.inFd, oldState)

		flPassword = readInput(cli.in, cli.out)
		fmt.Fprint(cli.out, "\n")

		term.RestoreTerminal(cli.inFd, oldState)
		if flPassword == "" {
			return authconfig, fmt.Errorf("Error: Password Required")
		}
	}

	authconfig.Username = flUser
	authconfig.Password = flPassword
	authconfig.ServerAddress = serverAddress
	authconfig.IdentityToken = ""

	return authconfig, nil
}

func readInput(in io.Reader, out io.Writer) string {
	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		fmt.Fprintln(out, err.Error())
		os.Exit(1)
	}
	return string(line)
}

// getCredentials loads the user credentials from a credentials store.
// The store is determined by the config file settings.
func getCredentials(c *cliconfig.ConfigFile, serverAddress string) (types.AuthConfig, error) {
	s := loadCredentialsStore(c)
	return s.Get(serverAddress)
}

func getAllCredentials(c *cliconfig.ConfigFile) (map[string]types.AuthConfig, error) {
	s := loadCredentialsStore(c)
	return s.GetAll()
}

// storeCredentials saves the user credentials in a credentials store.
// The store is determined by the config file settings.
func storeCredentials(c *cliconfig.ConfigFile, auth types.AuthConfig) error {
	s := loadCredentialsStore(c)
	return s.Store(auth)
}

// eraseCredentials removes the user credentials from a credentials store.
// The store is determined by the config file settings.
func eraseCredentials(c *cliconfig.ConfigFile, serverAddress string) error {
	s := loadCredentialsStore(c)
	return s.Erase(serverAddress)
}

// loadCredentialsStore initializes a new credentials store based
// in the settings provided in the configuration file.
func loadCredentialsStore(c *cliconfig.ConfigFile) credentials.Store {
	if c.CredentialsStore != "" {
		return credentials.NewNativeStore(c)
	}
	return credentials.NewFileStore(c)
}
`

var logout_go_scope = godebug.EnteringNewFile(client_pkg_scope, logout_go_contents)

func (cli *DockerCli) CmdLogout(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdLogout(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := logout_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 16)
	cmd := Cli.Subcmd("logout", []string{"[SERVER]"}, Cli.DockerCommands["logout"].Description+".\nIf no server is specified, the default is defined by the daemon.", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 17)
	cmd.Require(flag.Max, 1)
	godebug.Line(ctx, scope, 19)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 21)

	var serverAddress string
	scope.Declare("serverAddress", &serverAddress)
	godebug.Line(ctx, scope, 22)
	if len(cmd.Args()) > 0 {
		godebug.Line(ctx, scope, 23)
		serverAddress = cmd.Arg(0)
	} else {
		godebug.Line(ctx, scope, 24)
		godebug.Line(ctx, scope, 25)
		serverAddress = cli.electAuthServer()
	}
	godebug.Line(ctx, scope, 30)

	if _, ok := cli.configFile.AuthConfigs[serverAddress]; !ok {
		scope := scope.EnteringNewChildScope()
		scope.Declare("ok", &ok)
		godebug.Line(ctx, scope, 31)
		fmt.Fprintf(cli.out, "Not logged in to %s\n", serverAddress)
		godebug.Line(ctx, scope, 32)
		return nil
	}
	godebug.Line(ctx, scope, 35)

	fmt.Fprintf(cli.out, "Remove login credentials for %s\n", serverAddress)
	godebug.Line(ctx, scope, 36)
	if err := eraseCredentials(cli.configFile, serverAddress); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 37)
		fmt.Fprintf(cli.out, "WARNING: could not erase credentials: %v\n", err)
	}
	godebug.Line(ctx, scope, 40)

	return nil
}

var logout_go_contents = `package client

import (
	"fmt"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdLogout logs a user out from a Docker registry.
//
// If no server is specified, the user will be logged out from the registry's index server.
//
// Usage: docker logout [SERVER]
func (cli *DockerCli) CmdLogout(args ...string) error {
	cmd := Cli.Subcmd("logout", []string{"[SERVER]"}, Cli.DockerCommands["logout"].Description+".\nIf no server is specified, the default is defined by the daemon.", true)
	cmd.Require(flag.Max, 1)

	cmd.ParseFlags(args, true)

	var serverAddress string
	if len(cmd.Args()) > 0 {
		serverAddress = cmd.Arg(0)
	} else {
		serverAddress = cli.electAuthServer()
	}

	// check if we're logged in based on the records in the config file
	// which means it couldn't have user/pass cause they may be in the creds store
	if _, ok := cli.configFile.AuthConfigs[serverAddress]; !ok {
		fmt.Fprintf(cli.out, "Not logged in to %s\n", serverAddress)
		return nil
	}

	fmt.Fprintf(cli.out, "Remove login credentials for %s\n", serverAddress)
	if err := eraseCredentials(cli.configFile, serverAddress); err != nil {
		fmt.Fprintf(cli.out, "WARNING: could not erase credentials: %v\n", err)
	}

	return nil
}
`


var logs_go_scope = godebug.EnteringNewFile(client_pkg_scope, logs_go_contents)

var validDrivers = map[string]bool{
	"json-file": true,
	"journald":  true,
}

func (cli *DockerCli) CmdLogs(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdLogs(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := logs_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 24)
	cmd := Cli.Subcmd("logs", []string{"CONTAINER"}, Cli.DockerCommands["logs"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 25)
	follow := cmd.Bool([]string{"f", "-follow"}, false, "Follow log output")
	scope.Declare("follow", &follow)
	godebug.Line(ctx, scope, 26)
	since := cmd.String([]string{"-since"}, "", "Show logs since timestamp")
	scope.Declare("since", &since)
	godebug.Line(ctx, scope, 27)
	times := cmd.Bool([]string{"t", "-timestamps"}, false, "Show timestamps")
	scope.Declare("times", &times)
	godebug.Line(ctx, scope, 28)
	tail := cmd.String([]string{"-tail"}, "all", "Number of lines to show from the end of the logs")
	scope.Declare("tail", &tail)
	godebug.Line(ctx, scope, 29)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 31)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 33)

	name := cmd.Arg(0)
	scope.Declare("name", &name)
	godebug.Line(ctx, scope, 35)

	c, err := cli.client.ContainerInspect(context.Background(), name)
	scope.Declare("c", &c, "err", &err)
	godebug.Line(ctx, scope, 36)
	if err != nil {
		godebug.Line(ctx, scope, 37)
		return err
	}
	godebug.Line(ctx, scope, 40)

	if !validDrivers[c.HostConfig.LogConfig.Type] {
		godebug.Line(ctx, scope, 41)
		return fmt.Errorf("\"logs\" command is supported only for \"json-file\" and \"journald\" logging drivers (got: %s)", c.HostConfig.LogConfig.Type)
	}
	godebug.Line(ctx, scope, 44)

	options := types.ContainerLogsOptions{
		ContainerID: name,
		ShowStdout:  true,
		ShowStderr:  true,
		Since:       *since,
		Timestamps:  *times,
		Follow:      *follow,
		Tail:        *tail,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 53)

	responseBody, err := cli.client.ContainerLogs(context.Background(), options)
	scope.Declare("responseBody", &responseBody)
	godebug.Line(ctx, scope, 54)
	if err != nil {
		godebug.Line(ctx, scope, 55)
		return err
	}
	godebug.Line(ctx, scope, 57)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 57)
	godebug.Line(ctx, scope, 59)

	if c.Config.Tty {
		godebug.Line(ctx, scope, 60)
		_, err = io.Copy(cli.out, responseBody)
	} else {
		godebug.Line(ctx, scope, 61)
		godebug.Line(ctx, scope, 62)
		_, err = stdcopy.StdCopy(cli.out, cli.err, responseBody)
	}
	godebug.Line(ctx, scope, 64)
	return err
}

var logs_go_contents = `package client

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/engine-api/types"
)

var validDrivers = map[string]bool{
	"json-file": true,
	"journald":  true,
}

// CmdLogs fetches the logs of a given container.
//
// docker logs [OPTIONS] CONTAINER
func (cli *DockerCli) CmdLogs(args ...string) error {
	cmd := Cli.Subcmd("logs", []string{"CONTAINER"}, Cli.DockerCommands["logs"].Description, true)
	follow := cmd.Bool([]string{"f", "-follow"}, false, "Follow log output")
	since := cmd.String([]string{"-since"}, "", "Show logs since timestamp")
	times := cmd.Bool([]string{"t", "-timestamps"}, false, "Show timestamps")
	tail := cmd.String([]string{"-tail"}, "all", "Number of lines to show from the end of the logs")
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	name := cmd.Arg(0)

	c, err := cli.client.ContainerInspect(context.Background(), name)
	if err != nil {
		return err
	}

	if !validDrivers[c.HostConfig.LogConfig.Type] {
		return fmt.Errorf("\"logs\" command is supported only for \"json-file\" and \"journald\" logging drivers (got: %s)", c.HostConfig.LogConfig.Type)
	}

	options := types.ContainerLogsOptions{
		ContainerID: name,
		ShowStdout:  true,
		ShowStderr:  true,
		Since:       *since,
		Timestamps:  *times,
		Follow:      *follow,
		Tail:        *tail,
	}
	responseBody, err := cli.client.ContainerLogs(context.Background(), options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if c.Config.Tty {
		_, err = io.Copy(cli.out, responseBody)
	} else {
		_, err = stdcopy.StdCopy(cli.out, cli.err, responseBody)
	}
	return err
}
`


var merge_go_scope = godebug.EnteringNewFile(client_pkg_scope, merge_go_contents)

const debug_level int = 1

func (cli *DockerCli) CmdMerge(args ...string) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdMerge(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := merge_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.SetTraceGen(_ctx)
	godebug.Line(_ctx, scope, 28)
	godebug.Line(_ctx, scope, 29)

	if debug_level > 0 {
		godebug.Line(_ctx, scope, 30)
		logrus.Debugf("Executing api/client/merge.go : CmdMerge(%s)", args)
		godebug.Line(_ctx, scope, 31)
		if debug_level > 1 {
			godebug.Line(_ctx, scope, 32)
			logrus.Debug("Stack trace:")
			godebug.Line(_ctx, scope, 33)
			debug.PrintStack()
		}
	}
	godebug.Line(_ctx, scope, 36)
	cmd := Cli.Subcmd("merge", []string{"IMAGE1 IMAGE2 [COMMAND] [ARG...]"}, Cli.DockerCommands["merge"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(_ctx, scope, 37)
	addTrustedFlags(cmd, true)
	godebug.Line(_ctx, scope, 40)

	var (
		flAutoRemove = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Run container in background and print container ID")
		flSigProxy   = cmd.Bool([]string{"-sig-proxy"}, true, "Proxy received signals to the process")
		flName       = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		flDetachKeys = cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
		flAttach     *opts.ListOpts

		ErrConflictAttachDetach               = fmt.Errorf("Conflicting options: -a and -d")
		ErrConflictRestartPolicyAndAutoRemove = fmt.Errorf("Conflicting options: --restart and --rm")
		ErrConflictDetachAutoRemove           = fmt.Errorf("Conflicting options: --rm and -d")
	)
	scope.Declare("flAutoRemove", &flAutoRemove, "flDetach", &flDetach, "flSigProxy", &flSigProxy, "flName", &flName, "flDetachKeys", &flDetachKeys, "flAttach", &flAttach, "ErrConflictAttachDetach", &ErrConflictAttachDetach, "ErrConflictRestartPolicyAndAutoRemove", &ErrConflictRestartPolicyAndAutoRemove, "ErrConflictDetachAutoRemove", &ErrConflictDetachAutoRemove)
	godebug.Line(_ctx, scope, 53)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)
	scope.Declare("config", &config, "hostConfig", &hostConfig, "networkingConfig", &networkingConfig, "err", &err)
	godebug.Line(_ctx, scope, 55)

	if debug_level > 0 {
		godebug.Line(_ctx, scope, 56)
		logrus.Debugf("Config in CmdMerge(): %s", config)
		godebug.Line(_ctx, scope, 57)
		logrus.Debugf("cmd   : %s", cmd)
	}
	godebug.Line(_ctx, scope, 61)

	if err != nil {
		godebug.Line(_ctx, scope, 62)
		cmd.ReportError(err.Error(), true)
		godebug.Line(_ctx, scope, 63)
		os.Exit(125)
	}
	godebug.Line(_ctx, scope, 66)

	if hostConfig.OomKillDisable != nil && *hostConfig.OomKillDisable && hostConfig.Memory == 0 {
		godebug.Line(_ctx, scope, 67)
		fmt.Fprintf(cli.err, "WARNING: Disabling the OOM killer on containers without setting a '-m/--memory' limit may be dangerous.\n")
	}
	godebug.Line(_ctx, scope, 70)

	if len(hostConfig.DNS) > 0 {
		{
			scope := scope.EnteringNewChildScope()

			for _, dnsIP := range hostConfig.DNS {
				godebug.Line(_ctx, scope, 74)
				scope.Declare("dnsIP", &dnsIP)
				godebug.Line(_ctx, scope, 75)
				if dns.IsLocalhost(dnsIP) {
					godebug.Line(_ctx, scope, 76)
					fmt.Fprintf(cli.err, "WARNING: Localhost DNS setting (--dns=%s) may fail in containers.\n", dnsIP)
					godebug.Line(_ctx, scope, 77)
					break
				}
			}
			godebug.Line(_ctx, scope, 74)
		}
	}
	godebug.Line(_ctx, scope, 81)
	if config.Image == "" {
		godebug.Line(_ctx, scope, 82)
		cmd.Usage()
		godebug.Line(_ctx, scope, 83)
		return nil
	}
	godebug.Line(_ctx, scope, 86)

	config.ArgsEscaped = false
	godebug.Line(_ctx, scope, 88)

	if !*flDetach {
		godebug.Line(_ctx, scope, 89)
		if err := cli.CheckTtyInput(config.AttachStdin, config.Tty); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 90)
			return err
		}
	} else {
		godebug.Line(_ctx, scope, 92)
		godebug.Line(_ctx, scope, 93)
		if fl := cmd.Lookup("-attach"); fl != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("fl", &fl)
			godebug.Line(_ctx, scope, 94)
			flAttach = fl.Value.(*opts.ListOpts)
			godebug.Line(_ctx, scope, 95)
			if flAttach.Len() != 0 {
				godebug.Line(_ctx, scope, 96)
				return ErrConflictAttachDetach
			}
		}
		godebug.Line(_ctx, scope, 99)
		if *flAutoRemove {
			godebug.Line(_ctx, scope, 100)
			return ErrConflictDetachAutoRemove
		}
		godebug.Line(_ctx, scope, 103)

		config.AttachStdin = false
		godebug.Line(_ctx, scope, 104)
		config.AttachStdout = false
		godebug.Line(_ctx, scope, 105)
		config.AttachStderr = false
		godebug.Line(_ctx, scope, 106)
		config.StdinOnce = false
	}
	godebug.Line(_ctx, scope, 110)

	sigProxy := *flSigProxy
	scope.Declare("sigProxy", &sigProxy)
	godebug.Line(_ctx, scope, 111)
	if config.Tty {
		godebug.Line(_ctx, scope, 112)
		sigProxy = false
	}
	godebug.Line(_ctx, scope, 118)

	if runtime.GOOS == "windows" {
		godebug.Line(_ctx, scope, 119)
		hostConfig.ConsoleSize[0], hostConfig.ConsoleSize[1] = cli.getTtySize()
	}
	godebug.Line(_ctx, scope, 121)
	if debug_level > 0 {
		godebug.Line(_ctx, scope, 122)
		logrus.Debug("Calling cli.createContainer(config,... ")
	}
	godebug.Line(_ctx, scope, 124)
	createResponse, err := cli.createContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	scope.Declare("createResponse", &createResponse)
	godebug.Line(_ctx, scope, 125)
	if err != nil {
		godebug.Line(_ctx, scope, 126)
		cmd.ReportError(err.Error(), true)
		godebug.Line(_ctx, scope, 127)
		return runStartContainerErr(err)
	}
	godebug.Line(_ctx, scope, 129)
	if sigProxy {
		godebug.Line(_ctx, scope, 130)
		sigc := cli.forwardAllSignals(createResponse.ID)
		scope := scope.EnteringNewChildScope()
		scope.Declare("sigc", &sigc)
		godebug.Line(_ctx, scope, 131)
		defer signal.StopCatch(sigc)
		defer godebug.Defer(_ctx, scope, 131)
	}
	godebug.Line(_ctx, scope, 133)
	var (
		waitDisplayID chan struct{}
		errCh         chan error
		cancelFun     context.CancelFunc
		ctx           context.Context
	)
	scope.Declare("waitDisplayID", &waitDisplayID, "errCh", &errCh, "cancelFun", &cancelFun, "ctx", &ctx)
	godebug.Line(_ctx, scope, 139)

	if !config.AttachStdout && !config.AttachStderr {
		godebug.Line(_ctx, scope, 141)

		waitDisplayID = make(chan struct{})
		godebug.Line(_ctx, scope, 142)
		go func() {
			fn := func(_ctx *godebug.Context) {
				godebug.Line(_ctx, scope, 143)
				defer close(waitDisplayID)
				defer godebug.Defer(_ctx, scope, 143)
				godebug.Line(_ctx, scope, 144)
				fmt.Fprintf(cli.out, "%s\n", createResponse.ID)
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
		}()
	}
	godebug.Line(_ctx, scope, 147)
	if *flAutoRemove && (hostConfig.RestartPolicy.IsAlways() || hostConfig.RestartPolicy.IsOnFailure()) {
		godebug.Line(_ctx, scope, 148)
		return ErrConflictRestartPolicyAndAutoRemove
	}
	godebug.Line(_ctx, scope, 150)
	attach := config.AttachStdin || config.AttachStdout || config.AttachStderr
	scope.Declare("attach", &attach)
	godebug.Line(_ctx, scope, 151)
	if attach {
		godebug.Line(_ctx, scope, 152)
		var (
			out, stderr io.Writer
			in          io.ReadCloser
		)
		scope := scope.EnteringNewChildScope()
		scope.Declare("out", &out, "stderr", &stderr, "in", &in)
		godebug.Line(_ctx, scope, 156)

		if config.AttachStdin {
			godebug.Line(_ctx, scope, 157)
			in = cli.in
		}
		godebug.Line(_ctx, scope, 159)
		if config.AttachStdout {
			godebug.Line(_ctx, scope, 160)
			out = cli.out
		}
		godebug.Line(_ctx, scope, 162)
		if config.AttachStderr {
			godebug.Line(_ctx, scope, 163)
			if config.Tty {
				godebug.Line(_ctx, scope, 164)
				stderr = cli.out
			} else {
				godebug.Line(_ctx, scope, 165)
				godebug.Line(_ctx, scope, 166)
				stderr = cli.err
			}
		}
		godebug.Line(_ctx, scope, 170)

		if *flDetachKeys != "" {
			godebug.Line(_ctx, scope, 171)
			cli.configFile.DetachKeys = *flDetachKeys
		}
		godebug.Line(_ctx, scope, 174)

		options := types.ContainerAttachOptions{
			ContainerID: createResponse.ID,
			Stream:      true,
			Stdin:       config.AttachStdin,
			Stdout:      config.AttachStdout,
			Stderr:      config.AttachStderr,
			DetachKeys:  cli.configFile.DetachKeys,
		}
		scope.Declare("options", &options)
		godebug.Line(_ctx, scope, 183)

		resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
		scope.Declare("resp", &resp, "errAttach", &errAttach)
		godebug.Line(_ctx, scope, 184)
		if errAttach != nil && errAttach != httputil.ErrPersistEOF {
			godebug.Line(_ctx, scope, 188)

			return errAttach
		}
		godebug.Line(_ctx, scope, 190)
		ctx, cancelFun = context.WithCancel(context.Background())
		godebug.Line(_ctx, scope, 191)
		errCh = promise.Go(func() error {
			var result1 error
			fn := func(_ctx *godebug.Context) {
				result1 = func() error {
					godebug.Line(_ctx, scope, 192)
					errHijack := cli.holdHijackedConnection(ctx, config.Tty, in, out, stderr, resp)
					scope := scope.EnteringNewChildScope()
					scope.Declare("errHijack", &errHijack)
					godebug.Line(_ctx, scope, 193)
					if errHijack == nil {
						godebug.Line(_ctx, scope, 194)
						return errAttach
					}
					godebug.Line(_ctx, scope, 196)
					return errHijack
				}()
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
			return result1
		},
		)
	}
	godebug.Line(_ctx, scope, 200)

	if *flAutoRemove {
		godebug.Line(_ctx, scope, 201)
		defer func() {
			fn := func(_ctx *godebug.Context) {
				godebug.Line(_ctx, scope, 202)
				if err := cli.removeContainer(createResponse.ID, true, false, true); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(_ctx, scope, 203)
					fmt.Fprintf(cli.err, "%v\n", err)
				}
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
		}()
		defer godebug.Defer(_ctx, scope, 201)
	}
	godebug.Line(_ctx, scope, 209)

	if err := cli.client.ContainerStart(context.Background(), createResponse.ID); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(_ctx, scope, 213)

		if attach {
			godebug.Line(_ctx, scope, 214)
			cancelFun()
			godebug.Line(_ctx, scope, 215)
			<-errCh
		}
		godebug.Line(_ctx, scope, 218)

		cmd.ReportError(err.Error(), false)
		godebug.Line(_ctx, scope, 219)
		return runStartContainerErr(err)
	}
	godebug.Line(_ctx, scope, 222)

	if (config.AttachStdin || config.AttachStdout || config.AttachStderr) && config.Tty && cli.isTerminalOut {
		godebug.Line(_ctx, scope, 223)
		if err := cli.monitorTtySize(createResponse.ID, false); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 224)
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}
	godebug.Line(_ctx, scope, 228)

	if errCh != nil {
		godebug.Line(_ctx, scope, 229)
		if err := <-errCh; err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 230)
			logrus.Debugf("Error hijack: %s", err)
			godebug.Line(_ctx, scope, 231)
			return err
		}
	}
	godebug.Line(_ctx, scope, 236)

	if !config.AttachStdout && !config.AttachStderr {
		godebug.Line(_ctx, scope, 238)

		<-waitDisplayID
		godebug.Line(_ctx, scope, 239)
		return nil
	}
	godebug.Line(_ctx, scope, 242)

	var status int
	scope.Declare("status", &status)
	godebug.Line(_ctx, scope, 245)

	if *flAutoRemove {
		godebug.Line(_ctx, scope, 248)

		if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
			godebug.Line(_ctx, scope, 249)
			return runStartContainerErr(err)
		}
		godebug.Line(_ctx, scope, 251)
		if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
			godebug.Line(_ctx, scope, 252)
			return err
		}
	} else {
		godebug.Line(_ctx, scope, 254)
		godebug.Line(_ctx, scope, 256)

		if !config.Tty {
			godebug.Line(_ctx, scope, 258)

			if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
				godebug.Line(_ctx, scope, 259)
				return err
			}
		} else {
			godebug.Line(_ctx, scope, 261)
			godebug.Line(_ctx, scope, 264)

			if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
				godebug.Line(_ctx, scope, 265)
				return err
			}
		}
	}
	godebug.Line(_ctx, scope, 269)
	if status != 0 {
		godebug.Line(_ctx, scope, 270)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(_ctx, scope, 272)
	return nil
}

var merge_go_contents = `package client

import (
	"fmt"
	"io"
	"net/http/httputil"
	"os"
	"runtime"

	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/signal"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types"
	"github.com/docker/libnetwork/resolvconf/dns"
	"golang.org/x/net/context"
	"runtime/debug"
)

const debug_level int = 1

// CmdMerge will runs a command in a new container from two images.
//
// Usage: docker merge [OPTIONS] IMAGE1 IMAGE2 [COMMAND] [ARG...]
func (cli *DockerCli) CmdMerge(args ...string) error {
	_ = "breakpoint"
	if debug_level > 0 {
		logrus.Debugf("Executing api/client/merge.go : CmdMerge(%s)", args)
		if debug_level > 1 {
			logrus.Debug("Stack trace:")
			debug.PrintStack()
		}
	}
	cmd := Cli.Subcmd("merge", []string{"IMAGE1 IMAGE2 [COMMAND] [ARG...]"}, Cli.DockerCommands["merge"].Description, true)
	addTrustedFlags(cmd, true)

	// These are flags not stored in Config/HostConfig
	var (
		flAutoRemove = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Run container in background and print container ID")
		flSigProxy   = cmd.Bool([]string{"-sig-proxy"}, true, "Proxy received signals to the process")
		flName       = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		flDetachKeys = cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
		flAttach     *opts.ListOpts

		ErrConflictAttachDetach               = fmt.Errorf("Conflicting options: -a and -d")
		ErrConflictRestartPolicyAndAutoRemove = fmt.Errorf("Conflicting options: --restart and --rm")
		ErrConflictDetachAutoRemove           = fmt.Errorf("Conflicting options: --rm and -d")
	)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)

	if debug_level > 0 {
		logrus.Debugf("Config in CmdMerge(): %s", config)
		logrus.Debugf("cmd   : %s", cmd)
	}

	// just in case the Parse does not exit
	if err != nil {
		cmd.ReportError(err.Error(), true)
		os.Exit(125)
	}

	if hostConfig.OomKillDisable != nil && *hostConfig.OomKillDisable && hostConfig.Memory == 0 {
		fmt.Fprintf(cli.err, "WARNING: Disabling the OOM killer on containers without setting a '-m/--memory' limit may be dangerous.\n")
	}

	if len(hostConfig.DNS) > 0 {
		// check the DNS settings passed via --dns against
		// localhost regexp to warn if they are trying to
		// set a DNS to a localhost address
		for _, dnsIP := range hostConfig.DNS {
			if dns.IsLocalhost(dnsIP) {
				fmt.Fprintf(cli.err, "WARNING: Localhost DNS setting (--dns=%s) may fail in containers.\n", dnsIP)
				break
			}
		}
	}
	if config.Image == "" {
		cmd.Usage()
		return nil
	}

	config.ArgsEscaped = false

	if !*flDetach {
		if err := cli.CheckTtyInput(config.AttachStdin, config.Tty); err != nil {
			return err
		}
	} else {
		if fl := cmd.Lookup("-attach"); fl != nil {
			flAttach = fl.Value.(*opts.ListOpts)
			if flAttach.Len() != 0 {
				return ErrConflictAttachDetach
			}
		}
		if *flAutoRemove {
			return ErrConflictDetachAutoRemove
		}

		config.AttachStdin = false
		config.AttachStdout = false
		config.AttachStderr = false
		config.StdinOnce = false
	}

	// Disable flSigProxy when in TTY mode
	sigProxy := *flSigProxy
	if config.Tty {
		sigProxy = false
	}

	// Telling the Windows daemon the initial size of the tty during start makes
	// a far better user experience rather than relying on subsequent resizes
	// to cause things to catch up.
	if runtime.GOOS == "windows" {
		hostConfig.ConsoleSize[0], hostConfig.ConsoleSize[1] = cli.getTtySize()
	}
	if debug_level > 0 {
		logrus.Debug("Calling cli.createContainer(config,... ")
	}
	createResponse, err := cli.createContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	if err != nil {
		cmd.ReportError(err.Error(), true)
		return runStartContainerErr(err)
	}
	if sigProxy {
		sigc := cli.forwardAllSignals(createResponse.ID)
		defer signal.StopCatch(sigc)
	}
	var (
		waitDisplayID chan struct{}
		errCh         chan error
		cancelFun     context.CancelFunc
		ctx           context.Context
	)
	if !config.AttachStdout && !config.AttachStderr {
		// Make this asynchronous to allow the client to write to stdin before having to read the ID
		waitDisplayID = make(chan struct{})
		go func() {
			defer close(waitDisplayID)
			fmt.Fprintf(cli.out, "%s\n", createResponse.ID)
		}()
	}
	if *flAutoRemove && (hostConfig.RestartPolicy.IsAlways() || hostConfig.RestartPolicy.IsOnFailure()) {
		return ErrConflictRestartPolicyAndAutoRemove
	}
	attach := config.AttachStdin || config.AttachStdout || config.AttachStderr
	if attach {
		var (
			out, stderr io.Writer
			in          io.ReadCloser
		)
		if config.AttachStdin {
			in = cli.in
		}
		if config.AttachStdout {
			out = cli.out
		}
		if config.AttachStderr {
			if config.Tty {
				stderr = cli.out
			} else {
				stderr = cli.err
			}
		}

		if *flDetachKeys != "" {
			cli.configFile.DetachKeys = *flDetachKeys
		}

		options := types.ContainerAttachOptions{
			ContainerID: createResponse.ID,
			Stream:      true,
			Stdin:       config.AttachStdin,
			Stdout:      config.AttachStdout,
			Stderr:      config.AttachStderr,
			DetachKeys:  cli.configFile.DetachKeys,
		}

		resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
		if errAttach != nil && errAttach != httputil.ErrPersistEOF {
			// ContainerAttach returns an ErrPersistEOF (connection closed)
			// means server met an error and put it in Hijacked connection
			// keep the error and read detailed error message from hijacked connection later
			return errAttach
		}
		ctx, cancelFun = context.WithCancel(context.Background())
		errCh = promise.Go(func() error {
			errHijack := cli.holdHijackedConnection(ctx, config.Tty, in, out, stderr, resp)
			if errHijack == nil {
				return errAttach
			}
			return errHijack
		})
	}

	if *flAutoRemove {
		defer func() {
			if err := cli.removeContainer(createResponse.ID, true, false, true); err != nil {
				fmt.Fprintf(cli.err, "%v\n", err)
			}
		}()
	}

	//start the container
	if err := cli.client.ContainerStart(context.Background(), createResponse.ID); err != nil {
		// If we have holdHijackedConnection, we should notify
		// holdHijackedConnection we are going to exit and wait
		// to avoid the terminal are not restored.
		if attach {
			cancelFun()
			<-errCh
		}

		cmd.ReportError(err.Error(), false)
		return runStartContainerErr(err)
	}

	if (config.AttachStdin || config.AttachStdout || config.AttachStderr) && config.Tty && cli.isTerminalOut {
		if err := cli.monitorTtySize(createResponse.ID, false); err != nil {
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}

	if errCh != nil {
		if err := <-errCh; err != nil {
			logrus.Debugf("Error hijack: %s", err)
			return err
		}
	}

	// Detached mode: wait for the id to be displayed and return.
	if !config.AttachStdout && !config.AttachStderr {
		// Detached mode
		<-waitDisplayID
		return nil
	}

	var status int

	// Attached mode
	if *flAutoRemove {
		// Autoremove: wait for the container to finish, retrieve
		// the exit code and remove the container
		if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
			return runStartContainerErr(err)
		}
		if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
			return err
		}
	} else {
		// No Autoremove: Simply retrieve the exit code
		if !config.Tty {
			// In non-TTY mode, we can't detach, so we must wait for container exit
			if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
				return err
			}
		} else {
			// In TTY mode, there is a race: if the process dies too slowly, the state could
			// be updated after the getExitCode call and result in the wrong exit code being reported
			if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
				return err
			}
		}
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}
`


var network_go_scope = godebug.EnteringNewFile(client_pkg_scope, network_go_contents)

func (cli *DockerCli) CmdNetwork(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetwork(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 26)
	cmd := Cli.Subcmd("network", []string{"COMMAND [OPTIONS]"}, networkUsage(), false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 27)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 28)
	err := cmd.ParseFlags(args, true)
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 29)
	cmd.Usage()
	godebug.Line(ctx, scope, 30)
	return err
}

func (cli *DockerCli) CmdNetworkCreate(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetworkCreate(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 37)
	cmd := Cli.Subcmd("network create", []string{"NETWORK-NAME"}, "Creates a new network with a name specified by the user", false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 38)
	flDriver := cmd.String([]string{"d", "-driver"}, "bridge", "Driver to manage the Network")
	scope.Declare("flDriver", &flDriver)
	godebug.Line(ctx, scope, 39)
	flOpts := opts.NewMapOpts(nil, nil)
	scope.Declare("flOpts", &flOpts)
	godebug.Line(ctx, scope, 41)

	flIpamDriver := cmd.String([]string{"-ipam-driver"}, "default", "IP Address Management Driver")
	scope.Declare("flIpamDriver", &flIpamDriver)
	godebug.Line(ctx, scope, 42)
	flIpamSubnet := opts.NewListOpts(nil)
	scope.Declare("flIpamSubnet", &flIpamSubnet)
	godebug.Line(ctx, scope, 43)
	flIpamIPRange := opts.NewListOpts(nil)
	scope.Declare("flIpamIPRange", &flIpamIPRange)
	godebug.Line(ctx, scope, 44)
	flIpamGateway := opts.NewListOpts(nil)
	scope.Declare("flIpamGateway", &flIpamGateway)
	godebug.Line(ctx, scope, 45)
	flIpamAux := opts.NewMapOpts(nil, nil)
	scope.Declare("flIpamAux", &flIpamAux)
	godebug.Line(ctx, scope, 46)
	flIpamOpt := opts.NewMapOpts(nil, nil)
	scope.Declare("flIpamOpt", &flIpamOpt)
	godebug.Line(ctx, scope, 47)
	flLabels := opts.NewListOpts(nil)
	scope.Declare("flLabels", &flLabels)
	godebug.Line(ctx, scope, 49)

	cmd.Var(&flIpamSubnet, []string{"-subnet"}, "subnet in CIDR format that represents a network segment")
	godebug.Line(ctx, scope, 50)
	cmd.Var(&flIpamIPRange, []string{"-ip-range"}, "allocate container ip from a sub-range")
	godebug.Line(ctx, scope, 51)
	cmd.Var(&flIpamGateway, []string{"-gateway"}, "ipv4 or ipv6 Gateway for the master subnet")
	godebug.Line(ctx, scope, 52)
	cmd.Var(flIpamAux, []string{"-aux-address"}, "auxiliary ipv4 or ipv6 addresses used by Network driver")
	godebug.Line(ctx, scope, 53)
	cmd.Var(flOpts, []string{"o", "-opt"}, "set driver specific options")
	godebug.Line(ctx, scope, 54)
	cmd.Var(flIpamOpt, []string{"-ipam-opt"}, "set IPAM driver specific options")
	godebug.Line(ctx, scope, 55)
	cmd.Var(&flLabels, []string{"-label"}, "set metadata on a network")
	godebug.Line(ctx, scope, 57)

	flInternal := cmd.Bool([]string{"-internal"}, false, "restricts external access to the network")
	scope.Declare("flInternal", &flInternal)
	godebug.Line(ctx, scope, 58)
	flIPv6 := cmd.Bool([]string{"-ipv6"}, false, "enable IPv6 networking")
	scope.Declare("flIPv6", &flIPv6)
	godebug.Line(ctx, scope, 60)

	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 61)
	err := cmd.ParseFlags(args, true)
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 62)
	if err != nil {
		godebug.Line(ctx, scope, 63)
		return err
	}
	godebug.Line(ctx, scope, 68)

	driver := *flDriver
	scope.Declare("driver", &driver)
	godebug.Line(ctx, scope, 69)
	if !cmd.IsSet("-driver") && !cmd.IsSet("d") {
		godebug.Line(ctx, scope, 70)
		driver = ""
	}
	godebug.Line(ctx, scope, 73)

	ipamCfg, err := consolidateIpam(flIpamSubnet.GetAll(), flIpamIPRange.GetAll(), flIpamGateway.GetAll(), flIpamAux.GetAll())
	scope.Declare("ipamCfg", &ipamCfg)
	godebug.Line(ctx, scope, 74)
	if err != nil {
		godebug.Line(ctx, scope, 75)
		return err
	}
	godebug.Line(ctx, scope, 79)

	nc := types.NetworkCreate{
		Name:           cmd.Arg(0),
		Driver:         driver,
		IPAM:           network.IPAM{Driver: *flIpamDriver, Config: ipamCfg, Options: flIpamOpt.GetAll()},
		Options:        flOpts.GetAll(),
		CheckDuplicate: true,
		Internal:       *flInternal,
		EnableIPv6:     *flIPv6,
		Labels:         runconfigopts.ConvertKVStringsToMap(flLabels.GetAll()),
	}
	scope.Declare("nc", &nc)
	godebug.Line(ctx, scope, 90)

	resp, err := cli.client.NetworkCreate(context.Background(), nc)
	scope.Declare("resp", &resp)
	godebug.Line(ctx, scope, 91)
	if err != nil {
		godebug.Line(ctx, scope, 92)
		return err
	}
	godebug.Line(ctx, scope, 94)
	fmt.Fprintf(cli.out, "%s\n", resp.ID)
	godebug.Line(ctx, scope, 95)
	return nil
}

func (cli *DockerCli) CmdNetworkRm(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetworkRm(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 102)
	cmd := Cli.Subcmd("network rm", []string{"NETWORK [NETWORK...]"}, "Deletes one or more networks", false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 103)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 104)
	if err := cmd.ParseFlags(args, true); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 105)
		return err
	}
	godebug.Line(ctx, scope, 108)

	status := 0
	scope.Declare("status", &status)
	{
		scope := scope.EnteringNewChildScope()
		for _, net := range cmd.Args() {
			godebug.Line(ctx, scope, 109)
			scope.Declare("net", &net)
			godebug.Line(ctx, scope, 110)
			if err := cli.client.NetworkRemove(context.Background(), net); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 111)
				fmt.Fprintf(cli.err, "%s\n", err)
				godebug.Line(ctx, scope, 112)
				status = 1
				godebug.Line(ctx, scope, 113)
				continue
			}
		}
		godebug.Line(ctx, scope, 109)
	}
	godebug.Line(ctx, scope, 116)
	if status != 0 {
		godebug.Line(ctx, scope, 117)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(ctx, scope, 119)
	return nil
}

func (cli *DockerCli) CmdNetworkConnect(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetworkConnect(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 126)
	cmd := Cli.Subcmd("network connect", []string{"NETWORK CONTAINER"}, "Connects a container to a network", false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 127)
	flIPAddress := cmd.String([]string{"-ip"}, "", "IP Address")
	scope.Declare("flIPAddress", &flIPAddress)
	godebug.Line(ctx, scope, 128)
	flIPv6Address := cmd.String([]string{"-ip6"}, "", "IPv6 Address")
	scope.Declare("flIPv6Address", &flIPv6Address)
	godebug.Line(ctx, scope, 129)
	flLinks := opts.NewListOpts(runconfigopts.ValidateLink)
	scope.Declare("flLinks", &flLinks)
	godebug.Line(ctx, scope, 130)
	cmd.Var(&flLinks, []string{"-link"}, "Add link to another container")
	godebug.Line(ctx, scope, 131)
	flAliases := opts.NewListOpts(nil)
	scope.Declare("flAliases", &flAliases)
	godebug.Line(ctx, scope, 132)
	cmd.Var(&flAliases, []string{"-alias"}, "Add network-scoped alias for the container")
	godebug.Line(ctx, scope, 133)
	cmd.Require(flag.Min, 2)
	godebug.Line(ctx, scope, 134)
	if err := cmd.ParseFlags(args, true); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 135)
		return err
	}
	godebug.Line(ctx, scope, 137)
	epConfig := &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: *flIPAddress,
			IPv6Address: *flIPv6Address,
		},
		Links:   flLinks.GetAll(),
		Aliases: flAliases.GetAll(),
	}
	scope.Declare("epConfig", &epConfig)
	godebug.Line(ctx, scope, 145)

	return cli.client.NetworkConnect(context.Background(), cmd.Arg(0), cmd.Arg(1), epConfig)
}

func (cli *DockerCli) CmdNetworkDisconnect(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetworkDisconnect(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 152)
	cmd := Cli.Subcmd("network disconnect", []string{"NETWORK CONTAINER"}, "Disconnects container from a network", false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 153)
	force := cmd.Bool([]string{"f", "-force"}, false, "Force the container to disconnect from a network")
	scope.Declare("force", &force)
	godebug.Line(ctx, scope, 154)
	cmd.Require(flag.Exact, 2)
	godebug.Line(ctx, scope, 155)
	if err := cmd.ParseFlags(args, true); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 156)
		return err
	}
	godebug.Line(ctx, scope, 159)

	return cli.client.NetworkDisconnect(context.Background(), cmd.Arg(0), cmd.Arg(1), *force)
}

func (cli *DockerCli) CmdNetworkLs(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetworkLs(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 166)
	cmd := Cli.Subcmd("network ls", nil, "Lists networks", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 167)
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")
	scope.Declare("quiet", &quiet)
	godebug.Line(ctx, scope, 168)
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Do not truncate the output")
	scope.Declare("noTrunc", &noTrunc)
	godebug.Line(ctx, scope, 170)

	flFilter := opts.NewListOpts(nil)
	scope.Declare("flFilter", &flFilter)
	godebug.Line(ctx, scope, 171)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	godebug.Line(ctx, scope, 173)

	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 174)
	err := cmd.ParseFlags(args, true)
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 175)
	if err != nil {
		godebug.Line(ctx, scope, 176)
		return err
	}
	godebug.Line(ctx, scope, 181)

	netFilterArgs := filters.NewArgs()
	scope.Declare("netFilterArgs", &netFilterArgs)
	{
		scope := scope.EnteringNewChildScope()
		for _, f := range flFilter.GetAll() {
			godebug.Line(ctx, scope, 182)
			scope.Declare("f", &f)
			godebug.Line(ctx, scope, 183)
			if netFilterArgs, err = filters.ParseFlag(f, netFilterArgs); err != nil {
				godebug.Line(ctx, scope, 184)
				return err
			}
		}
		godebug.Line(ctx, scope, 182)
	}
	godebug.Line(ctx, scope, 188)

	options := types.NetworkListOptions{
		Filters: netFilterArgs,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 192)

	networkResources, err := cli.client.NetworkList(context.Background(), options)
	scope.Declare("networkResources", &networkResources)
	godebug.Line(ctx, scope, 193)
	if err != nil {
		godebug.Line(ctx, scope, 194)
		return err
	}
	godebug.Line(ctx, scope, 197)

	wr := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	scope.Declare("wr", &wr)
	godebug.Line(ctx, scope, 200)

	if !*quiet {
		godebug.Line(ctx, scope, 201)
		fmt.Fprintln(wr, "NETWORK ID\tNAME\tDRIVER")
	}
	godebug.Line(ctx, scope, 203)
	sort.Sort(byNetworkName(networkResources))
	{
		scope := scope.EnteringNewChildScope()
		for _, networkResource := range networkResources {
			godebug.Line(ctx, scope, 204)
			scope.Declare("networkResource", &networkResource)
			godebug.Line(ctx, scope, 205)
			ID := networkResource.ID
			scope := scope.EnteringNewChildScope()
			scope.Declare("ID", &ID)
			godebug.Line(ctx, scope, 206)
			netName := networkResource.Name
			scope.Declare("netName", &netName)
			godebug.Line(ctx, scope, 207)
			if !*noTrunc {
				godebug.Line(ctx, scope, 208)
				ID = stringid.TruncateID(ID)
			}
			godebug.Line(ctx, scope, 210)
			if *quiet {
				godebug.Line(ctx, scope, 211)
				fmt.Fprintln(wr, ID)
				godebug.Line(ctx, scope, 212)
				continue
			}
			godebug.Line(ctx, scope, 214)
			driver := networkResource.Driver
			scope.Declare("driver", &driver)
			godebug.Line(ctx, scope, 215)
			fmt.Fprintf(wr, "%s\t%s\t%s\t",
				ID,
				netName,
				driver)
			godebug.Line(ctx, scope, 219)
			fmt.Fprint(wr, "\n")
		}
		godebug.Line(ctx, scope, 204)
	}
	godebug.Line(ctx, scope, 221)
	wr.Flush()
	godebug.Line(ctx, scope, 222)
	return nil
}

type byNetworkName []types.NetworkResource

func (r byNetworkName) Len() int {
	var result1 int
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = r.Len()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r)
	godebug.Line(ctx, scope, 227)
	return len(r)
}
func (r byNetworkName) Swap(i, j int) {
	ctx, _ok := godebug.EnterFunc(func() {
		r.Swap(i, j)
	})
	if !_ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 228)
	r[i], r[j] = r[j], r[i]
}
func (r byNetworkName) Less(i, j int) bool {
	var result1 bool
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = r.Less(i, j)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 229)
	return r[i].Name < r[j].Name
}

func (cli *DockerCli) CmdNetworkInspect(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdNetworkInspect(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 235)
	cmd := Cli.Subcmd("network inspect", []string{"NETWORK [NETWORK...]"}, "Displays detailed information on one or more networks", false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 236)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")
	scope.Declare("tmplStr", &tmplStr)
	godebug.Line(ctx, scope, 237)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 239)

	if err := cmd.ParseFlags(args, true); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 240)
		return err
	}
	godebug.Line(ctx, scope, 243)

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		var result1 interface {
		}
		var result2 []byte
		var result3 error
		fn := func(ctx *godebug.Context) {
			result1, result2, result3 = func() (interface {
			}, []byte, error) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("name", &name)
				godebug.Line(ctx, scope, 244)
				i, err := cli.client.NetworkInspect(context.Background(), name)
				scope.Declare("i", &i, "err", &err)
				godebug.Line(ctx, scope, 245)
				return i, nil, err
			}()
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2, result3
	}
	scope.Declare("inspectSearcher", &inspectSearcher)
	godebug.Line(ctx, scope, 248)

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}

func consolidateIpam(subnets, ranges, gateways []string, auxaddrs map[string]string) ([]network.IPAMConfig, error) {
	var result1 []network.IPAMConfig
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = consolidateIpam(subnets, ranges, gateways, auxaddrs)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("subnets", &subnets, "ranges", &ranges, "gateways", &gateways, "auxaddrs", &auxaddrs)
	godebug.Line(ctx, scope, 257)
	if len(subnets) < len(ranges) || len(subnets) < len(gateways) {
		godebug.Line(ctx, scope, 258)
		return nil, fmt.Errorf("every ip-range or gateway must have a corresponding subnet")
	}
	godebug.Line(ctx, scope, 260)
	iData := map[string]*network.IPAMConfig{}
	scope.Declare("iData", &iData)
	{
		scope := scope.EnteringNewChildScope()

		for _, s := range subnets {
			godebug.Line(ctx, scope, 263)
			scope.Declare("s", &s)
			{
				scope := scope.EnteringNewChildScope()
				for k := range iData {
					godebug.Line(ctx, scope, 264)
					scope.Declare("k", &k)
					godebug.Line(ctx, scope, 265)
					ok1, err := subnetMatches(s, k)
					scope := scope.EnteringNewChildScope()
					scope.Declare("ok1", &ok1, "err", &err)
					godebug.Line(ctx, scope, 266)
					if err != nil {
						godebug.Line(ctx, scope, 267)
						return nil, err
					}
					godebug.Line(ctx, scope, 269)
					ok2, err := subnetMatches(k, s)
					scope.Declare("ok2", &ok2)
					godebug.Line(ctx, scope, 270)
					if err != nil {
						godebug.Line(ctx, scope, 271)
						return nil, err
					}
					godebug.Line(ctx, scope, 273)
					if ok1 || ok2 {
						godebug.Line(ctx, scope, 274)
						return nil, fmt.Errorf("multiple overlapping subnet configuration is not supported")
					}
				}
				godebug.Line(ctx, scope, 264)
			}
			godebug.Line(ctx, scope, 277)
			iData[s] = &network.IPAMConfig{Subnet: s, AuxAddress: map[string]string{}}
		}
		godebug.Line(ctx, scope, 263)
	}
	{
		scope := scope.EnteringNewChildScope()

		for _, r := range ranges {
			godebug.Line(ctx, scope, 281)
			scope.Declare("r", &r)
			godebug.Line(ctx, scope, 282)
			match := false
			scope := scope.EnteringNewChildScope()
			scope.Declare("match", &match)
			{
				scope := scope.EnteringNewChildScope()
				for _, s := range subnets {
					godebug.Line(ctx, scope, 283)
					scope.Declare("s", &s)
					godebug.Line(ctx, scope, 284)
					ok, err := subnetMatches(s, r)
					scope := scope.EnteringNewChildScope()
					scope.Declare("ok", &ok, "err", &err)
					godebug.Line(ctx, scope, 285)
					if err != nil {
						godebug.Line(ctx, scope, 286)
						return nil, err
					}
					godebug.Line(ctx, scope, 288)
					if !ok {
						godebug.Line(ctx, scope, 289)
						continue
					}
					godebug.Line(ctx, scope, 291)
					if iData[s].IPRange != "" {
						godebug.Line(ctx, scope, 292)
						return nil, fmt.Errorf("cannot configure multiple ranges (%s, %s) on the same subnet (%s)", r, iData[s].IPRange, s)
					}
					godebug.Line(ctx, scope, 294)
					d := iData[s]
					scope.Declare("d", &d)
					godebug.Line(ctx, scope, 295)
					d.IPRange = r
					godebug.Line(ctx, scope, 296)
					match = true
				}
				godebug.Line(ctx, scope, 283)
			}
			godebug.Line(ctx, scope, 298)
			if !match {
				godebug.Line(ctx, scope, 299)
				return nil, fmt.Errorf("no matching subnet for range %s", r)
			}
		}
		godebug.Line(ctx, scope, 281)
	}
	{
		scope := scope.EnteringNewChildScope()

		for _, g := range gateways {
			godebug.Line(ctx, scope, 304)
			scope.Declare("g", &g)
			godebug.Line(ctx, scope, 305)
			match := false
			scope := scope.EnteringNewChildScope()
			scope.Declare("match", &match)
			{
				scope := scope.EnteringNewChildScope()
				for _, s := range subnets {
					godebug.Line(ctx, scope, 306)
					scope.Declare("s", &s)
					godebug.Line(ctx, scope, 307)
					ok, err := subnetMatches(s, g)
					scope := scope.EnteringNewChildScope()
					scope.Declare("ok", &ok, "err", &err)
					godebug.Line(ctx, scope, 308)
					if err != nil {
						godebug.Line(ctx, scope, 309)
						return nil, err
					}
					godebug.Line(ctx, scope, 311)
					if !ok {
						godebug.Line(ctx, scope, 312)
						continue
					}
					godebug.Line(ctx, scope, 314)
					if iData[s].Gateway != "" {
						godebug.Line(ctx, scope, 315)
						return nil, fmt.Errorf("cannot configure multiple gateways (%s, %s) for the same subnet (%s)", g, iData[s].Gateway, s)
					}
					godebug.Line(ctx, scope, 317)
					d := iData[s]
					scope.Declare("d", &d)
					godebug.Line(ctx, scope, 318)
					d.Gateway = g
					godebug.Line(ctx, scope, 319)
					match = true
				}
				godebug.Line(ctx, scope, 306)
			}
			godebug.Line(ctx, scope, 321)
			if !match {
				godebug.Line(ctx, scope, 322)
				return nil, fmt.Errorf("no matching subnet for gateway %s", g)
			}
		}
		godebug.Line(ctx, scope, 304)
	}
	{
		scope := scope.EnteringNewChildScope()

		for key, aa := range auxaddrs {
			godebug.Line(ctx, scope, 327)
			scope.Declare("key", &key, "aa", &aa)
			godebug.Line(ctx, scope, 328)
			match := false
			scope := scope.EnteringNewChildScope()
			scope.Declare("match", &match)
			{
				scope := scope.EnteringNewChildScope()
				for _, s := range subnets {
					godebug.Line(ctx, scope, 329)
					scope.Declare("s", &s)
					godebug.Line(ctx, scope, 330)
					ok, err := subnetMatches(s, aa)
					scope := scope.EnteringNewChildScope()
					scope.Declare("ok", &ok, "err", &err)
					godebug.Line(ctx, scope, 331)
					if err != nil {
						godebug.Line(ctx, scope, 332)
						return nil, err
					}
					godebug.Line(ctx, scope, 334)
					if !ok {
						godebug.Line(ctx, scope, 335)
						continue
					}
					godebug.Line(ctx, scope, 337)
					iData[s].AuxAddress[key] = aa
					godebug.Line(ctx, scope, 338)
					match = true
				}
				godebug.Line(ctx, scope, 329)
			}
			godebug.Line(ctx, scope, 340)
			if !match {
				godebug.Line(ctx, scope, 341)
				return nil, fmt.Errorf("no matching subnet for aux-address %s", aa)
			}
		}
		godebug.Line(ctx, scope, 327)
	}
	godebug.Line(ctx, scope, 345)

	idl := []network.IPAMConfig{}
	scope.Declare("idl", &idl)
	{
		scope := scope.EnteringNewChildScope()
		for _, v := range iData {
			godebug.Line(ctx, scope, 346)
			scope.Declare("v", &v)
			godebug.Line(ctx, scope, 347)
			idl = append(idl, *v)
		}
		godebug.Line(ctx, scope, 346)
	}
	godebug.Line(ctx, scope, 349)
	return idl, nil
}

func subnetMatches(subnet, data string) (bool, error) {
	var result1 bool
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = subnetMatches(subnet, data)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("subnet", &subnet, "data", &data)
	godebug.Line(ctx, scope, 353)
	var (
		ip net.IP
	)
	scope.Declare("ip", &ip)
	godebug.Line(ctx, scope, 357)

	_, s, err := net.ParseCIDR(subnet)
	scope.Declare("s", &s, "err", &err)
	godebug.Line(ctx, scope, 358)
	if err != nil {
		godebug.Line(ctx, scope, 359)
		return false, fmt.Errorf("Invalid subnet %s : %v", s, err)
	}
	godebug.Line(ctx, scope, 362)

	if strings.Contains(data, "/") {
		godebug.Line(ctx, scope, 363)
		ip, _, err = net.ParseCIDR(data)
		godebug.Line(ctx, scope, 364)
		if err != nil {
			godebug.Line(ctx, scope, 365)
			return false, fmt.Errorf("Invalid cidr %s : %v", data, err)
		}
	} else {
		godebug.Line(ctx, scope, 367)
		godebug.Line(ctx, scope, 368)
		ip = net.ParseIP(data)
	}
	godebug.Line(ctx, scope, 371)

	return s.Contains(ip), nil
}

func networkUsage() string {
	var result1 string
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = networkUsage()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	godebug.Line(ctx, network_go_scope, 375)
	networkCommands := [][]string{
		{"create", "Create a network"},
		{"connect", "Connect container to a network"},
		{"disconnect", "Disconnect container from a network"},
		{"inspect", "Display detailed network information"},
		{"ls", "List all networks"},
		{"rm", "Remove a network"},
	}
	scope := network_go_scope.EnteringNewChildScope()
	scope.Declare("networkCommands", &networkCommands)
	godebug.Line(ctx, scope, 384)

	help := "Commands:\n"
	scope.Declare("help", &help)
	{
		scope := scope.EnteringNewChildScope()

		for _, cmd := range networkCommands {
			godebug.Line(ctx, scope, 386)
			scope.Declare("cmd", &cmd)
			godebug.Line(ctx, scope, 387)
			help += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
		}
		godebug.Line(ctx, scope, 386)
	}
	godebug.Line(ctx, scope, 390)

	help += fmt.Sprintf("\nRun 'docker network COMMAND --help' for more information on a command.")
	godebug.Line(ctx, scope, 391)
	return help
}

var network_go_contents = `package client

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/stringid"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/network"
)

// CmdNetwork is the parent subcommand for all network commands
//
// Usage: docker network <COMMAND> [OPTIONS]
func (cli *DockerCli) CmdNetwork(args ...string) error {
	cmd := Cli.Subcmd("network", []string{"COMMAND [OPTIONS]"}, networkUsage(), false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdNetworkCreate creates a new network with a given name
//
// Usage: docker network create [OPTIONS] <NETWORK-NAME>
func (cli *DockerCli) CmdNetworkCreate(args ...string) error {
	cmd := Cli.Subcmd("network create", []string{"NETWORK-NAME"}, "Creates a new network with a name specified by the user", false)
	flDriver := cmd.String([]string{"d", "-driver"}, "bridge", "Driver to manage the Network")
	flOpts := opts.NewMapOpts(nil, nil)

	flIpamDriver := cmd.String([]string{"-ipam-driver"}, "default", "IP Address Management Driver")
	flIpamSubnet := opts.NewListOpts(nil)
	flIpamIPRange := opts.NewListOpts(nil)
	flIpamGateway := opts.NewListOpts(nil)
	flIpamAux := opts.NewMapOpts(nil, nil)
	flIpamOpt := opts.NewMapOpts(nil, nil)
	flLabels := opts.NewListOpts(nil)

	cmd.Var(&flIpamSubnet, []string{"-subnet"}, "subnet in CIDR format that represents a network segment")
	cmd.Var(&flIpamIPRange, []string{"-ip-range"}, "allocate container ip from a sub-range")
	cmd.Var(&flIpamGateway, []string{"-gateway"}, "ipv4 or ipv6 Gateway for the master subnet")
	cmd.Var(flIpamAux, []string{"-aux-address"}, "auxiliary ipv4 or ipv6 addresses used by Network driver")
	cmd.Var(flOpts, []string{"o", "-opt"}, "set driver specific options")
	cmd.Var(flIpamOpt, []string{"-ipam-opt"}, "set IPAM driver specific options")
	cmd.Var(&flLabels, []string{"-label"}, "set metadata on a network")

	flInternal := cmd.Bool([]string{"-internal"}, false, "restricts external access to the network")
	flIPv6 := cmd.Bool([]string{"-ipv6"}, false, "enable IPv6 networking")

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	// Set the default driver to "" if the user didn't set the value.
	// That way we can know whether it was user input or not.
	driver := *flDriver
	if !cmd.IsSet("-driver") && !cmd.IsSet("d") {
		driver = ""
	}

	ipamCfg, err := consolidateIpam(flIpamSubnet.GetAll(), flIpamIPRange.GetAll(), flIpamGateway.GetAll(), flIpamAux.GetAll())
	if err != nil {
		return err
	}

	// Construct network create request body
	nc := types.NetworkCreate{
		Name:           cmd.Arg(0),
		Driver:         driver,
		IPAM:           network.IPAM{Driver: *flIpamDriver, Config: ipamCfg, Options: flIpamOpt.GetAll()},
		Options:        flOpts.GetAll(),
		CheckDuplicate: true,
		Internal:       *flInternal,
		EnableIPv6:     *flIPv6,
		Labels:         runconfigopts.ConvertKVStringsToMap(flLabels.GetAll()),
	}

	resp, err := cli.client.NetworkCreate(context.Background(), nc)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s\n", resp.ID)
	return nil
}

// CmdNetworkRm deletes one or more networks
//
// Usage: docker network rm NETWORK-NAME|NETWORK-ID [NETWORK-NAME|NETWORK-ID...]
func (cli *DockerCli) CmdNetworkRm(args ...string) error {
	cmd := Cli.Subcmd("network rm", []string{"NETWORK [NETWORK...]"}, "Deletes one or more networks", false)
	cmd.Require(flag.Min, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	status := 0
	for _, net := range cmd.Args() {
		if err := cli.client.NetworkRemove(context.Background(), net); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

// CmdNetworkConnect connects a container to a network
//
// Usage: docker network connect [OPTIONS] <NETWORK> <CONTAINER>
func (cli *DockerCli) CmdNetworkConnect(args ...string) error {
	cmd := Cli.Subcmd("network connect", []string{"NETWORK CONTAINER"}, "Connects a container to a network", false)
	flIPAddress := cmd.String([]string{"-ip"}, "", "IP Address")
	flIPv6Address := cmd.String([]string{"-ip6"}, "", "IPv6 Address")
	flLinks := opts.NewListOpts(runconfigopts.ValidateLink)
	cmd.Var(&flLinks, []string{"-link"}, "Add link to another container")
	flAliases := opts.NewListOpts(nil)
	cmd.Var(&flAliases, []string{"-alias"}, "Add network-scoped alias for the container")
	cmd.Require(flag.Min, 2)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}
	epConfig := &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: *flIPAddress,
			IPv6Address: *flIPv6Address,
		},
		Links:   flLinks.GetAll(),
		Aliases: flAliases.GetAll(),
	}
	return cli.client.NetworkConnect(context.Background(), cmd.Arg(0), cmd.Arg(1), epConfig)
}

// CmdNetworkDisconnect disconnects a container from a network
//
// Usage: docker network disconnect <NETWORK> <CONTAINER>
func (cli *DockerCli) CmdNetworkDisconnect(args ...string) error {
	cmd := Cli.Subcmd("network disconnect", []string{"NETWORK CONTAINER"}, "Disconnects container from a network", false)
	force := cmd.Bool([]string{"f", "-force"}, false, "Force the container to disconnect from a network")
	cmd.Require(flag.Exact, 2)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	return cli.client.NetworkDisconnect(context.Background(), cmd.Arg(0), cmd.Arg(1), *force)
}

// CmdNetworkLs lists all the networks managed by docker daemon
//
// Usage: docker network ls [OPTIONS]
func (cli *DockerCli) CmdNetworkLs(args ...string) error {
	cmd := Cli.Subcmd("network ls", nil, "Lists networks", true)
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Do not truncate the output")

	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process after get response from server.
	netFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		if netFilterArgs, err = filters.ParseFlag(f, netFilterArgs); err != nil {
			return err
		}
	}

	options := types.NetworkListOptions{
		Filters: netFilterArgs,
	}

	networkResources, err := cli.client.NetworkList(context.Background(), options)
	if err != nil {
		return err
	}

	wr := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)

	// unless quiet (-q) is specified, print field titles
	if !*quiet {
		fmt.Fprintln(wr, "NETWORK ID\tNAME\tDRIVER")
	}
	sort.Sort(byNetworkName(networkResources))
	for _, networkResource := range networkResources {
		ID := networkResource.ID
		netName := networkResource.Name
		if !*noTrunc {
			ID = stringid.TruncateID(ID)
		}
		if *quiet {
			fmt.Fprintln(wr, ID)
			continue
		}
		driver := networkResource.Driver
		fmt.Fprintf(wr, "%s\t%s\t%s\t",
			ID,
			netName,
			driver)
		fmt.Fprint(wr, "\n")
	}
	wr.Flush()
	return nil
}

type byNetworkName []types.NetworkResource

func (r byNetworkName) Len() int           { return len(r) }
func (r byNetworkName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byNetworkName) Less(i, j int) bool { return r[i].Name < r[j].Name }

// CmdNetworkInspect inspects the network object for more details
//
// Usage: docker network inspect [OPTIONS] <NETWORK> [NETWORK...]
func (cli *DockerCli) CmdNetworkInspect(args ...string) error {
	cmd := Cli.Subcmd("network inspect", []string{"NETWORK [NETWORK...]"}, "Displays detailed information on one or more networks", false)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")
	cmd.Require(flag.Min, 1)

	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		i, err := cli.client.NetworkInspect(context.Background(), name)
		return i, nil, err
	}

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}

// Consolidates the ipam configuration as a group from different related configurations
// user can configure network with multiple non-overlapping subnets and hence it is
// possible to correlate the various related parameters and consolidate them.
// consoidateIpam consolidates subnets, ip-ranges, gateways and auxiliary addresses into
// structured ipam data.
func consolidateIpam(subnets, ranges, gateways []string, auxaddrs map[string]string) ([]network.IPAMConfig, error) {
	if len(subnets) < len(ranges) || len(subnets) < len(gateways) {
		return nil, fmt.Errorf("every ip-range or gateway must have a corresponding subnet")
	}
	iData := map[string]*network.IPAMConfig{}

	// Populate non-overlapping subnets into consolidation map
	for _, s := range subnets {
		for k := range iData {
			ok1, err := subnetMatches(s, k)
			if err != nil {
				return nil, err
			}
			ok2, err := subnetMatches(k, s)
			if err != nil {
				return nil, err
			}
			if ok1 || ok2 {
				return nil, fmt.Errorf("multiple overlapping subnet configuration is not supported")
			}
		}
		iData[s] = &network.IPAMConfig{Subnet: s, AuxAddress: map[string]string{}}
	}

	// Validate and add valid ip ranges
	for _, r := range ranges {
		match := false
		for _, s := range subnets {
			ok, err := subnetMatches(s, r)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if iData[s].IPRange != "" {
				return nil, fmt.Errorf("cannot configure multiple ranges (%s, %s) on the same subnet (%s)", r, iData[s].IPRange, s)
			}
			d := iData[s]
			d.IPRange = r
			match = true
		}
		if !match {
			return nil, fmt.Errorf("no matching subnet for range %s", r)
		}
	}

	// Validate and add valid gateways
	for _, g := range gateways {
		match := false
		for _, s := range subnets {
			ok, err := subnetMatches(s, g)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if iData[s].Gateway != "" {
				return nil, fmt.Errorf("cannot configure multiple gateways (%s, %s) for the same subnet (%s)", g, iData[s].Gateway, s)
			}
			d := iData[s]
			d.Gateway = g
			match = true
		}
		if !match {
			return nil, fmt.Errorf("no matching subnet for gateway %s", g)
		}
	}

	// Validate and add aux-addresses
	for key, aa := range auxaddrs {
		match := false
		for _, s := range subnets {
			ok, err := subnetMatches(s, aa)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			iData[s].AuxAddress[key] = aa
			match = true
		}
		if !match {
			return nil, fmt.Errorf("no matching subnet for aux-address %s", aa)
		}
	}

	idl := []network.IPAMConfig{}
	for _, v := range iData {
		idl = append(idl, *v)
	}
	return idl, nil
}

func subnetMatches(subnet, data string) (bool, error) {
	var (
		ip net.IP
	)

	_, s, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, fmt.Errorf("Invalid subnet %s : %v", s, err)
	}

	if strings.Contains(data, "/") {
		ip, _, err = net.ParseCIDR(data)
		if err != nil {
			return false, fmt.Errorf("Invalid cidr %s : %v", data, err)
		}
	} else {
		ip = net.ParseIP(data)
	}

	return s.Contains(ip), nil
}

func networkUsage() string {
	networkCommands := [][]string{
		{"create", "Create a network"},
		{"connect", "Connect container to a network"},
		{"disconnect", "Disconnect container from a network"},
		{"inspect", "Display detailed network information"},
		{"ls", "List all networks"},
		{"rm", "Remove a network"},
	}

	help := "Commands:\n"

	for _, cmd := range networkCommands {
		help += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	help += fmt.Sprintf("\nRun 'docker network COMMAND --help' for more information on a command.")
	return help
}
`


var pause_go_scope = godebug.EnteringNewFile(client_pkg_scope, pause_go_contents)

func (cli *DockerCli) CmdPause(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdPause(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := pause_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 17)
	cmd := Cli.Subcmd("pause", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["pause"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 18)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 20)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 22)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 23)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 24)
			if err := cli.client.ContainerPause(context.Background(), name); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 25)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 26)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 27)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 23)
	}
	godebug.Line(ctx, scope, 30)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 31)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 33)
	return nil
}

var pause_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdPause pauses all processes within one or more containers.
//
// Usage: docker pause CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdPause(args ...string) error {
	cmd := Cli.Subcmd("pause", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["pause"].Description, true)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		if err := cli.client.ContainerPause(context.Background(), name); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var port_go_scope = godebug.EnteringNewFile(client_pkg_scope, port_go_contents)

func (cli *DockerCli) CmdPort(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdPort(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := port_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("port", []string{"CONTAINER [PRIVATE_PORT[/PROTO]]"}, Cli.DockerCommands["port"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 22)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 24)

	c, err := cli.client.ContainerInspect(context.Background(), cmd.Arg(0))
	scope.Declare("c", &c, "err", &err)
	godebug.Line(ctx, scope, 25)
	if err != nil {
		godebug.Line(ctx, scope, 26)
		return err
	}
	godebug.Line(ctx, scope, 29)

	if cmd.NArg() == 2 {
		godebug.Line(ctx, scope, 30)
		var (
			port  = cmd.Arg(1)
			proto = "tcp"
			parts = strings.SplitN(port, "/", 2)
		)
		scope := scope.EnteringNewChildScope()
		scope.Declare("port", &port, "proto", &proto, "parts", &parts)
		godebug.Line(ctx, scope, 36)

		if len(parts) == 2 && len(parts[1]) != 0 {
			godebug.Line(ctx, scope, 37)
			port = parts[0]
			godebug.Line(ctx, scope, 38)
			proto = parts[1]
		}
		godebug.Line(ctx, scope, 40)
		natPort := port + "/" + proto
		scope.Declare("natPort", &natPort)
		godebug.Line(ctx, scope, 41)
		newP, err := nat.NewPort(proto, port)
		scope.Declare("newP", &newP, "err", &err)
		godebug.Line(ctx, scope, 42)
		if err != nil {
			godebug.Line(ctx, scope, 43)
			return err
		}
		godebug.Line(ctx, scope, 45)
		if frontends, exists := c.NetworkSettings.Ports[newP]; exists && frontends != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("frontends", &frontends, "exists", &exists)
			{
				scope := scope.EnteringNewChildScope()
				for _, frontend := range frontends {
					godebug.Line(ctx, scope, 46)
					scope.Declare("frontend", &frontend)
					godebug.Line(ctx, scope, 47)
					fmt.Fprintf(cli.out, "%s:%s\n", frontend.HostIP, frontend.HostPort)
				}
				godebug.Line(ctx, scope, 46)
			}
			godebug.Line(ctx, scope, 49)
			return nil
		}
		godebug.Line(ctx, scope, 51)
		return fmt.Errorf("Error: No public port '%s' published for %s", natPort, cmd.Arg(0))
	}
	{
		scope := scope.EnteringNewChildScope()

		for from, frontends := range c.NetworkSettings.Ports {
			godebug.Line(ctx, scope, 54)
			scope.Declare("from", &from, "frontends", &frontends)
			{
				scope := scope.EnteringNewChildScope()
				for _, frontend := range frontends {
					godebug.Line(ctx, scope, 55)
					scope.Declare("frontend", &frontend)
					godebug.Line(ctx, scope, 56)
					fmt.Fprintf(cli.out, "%s -> %s:%s\n", from, frontend.HostIP, frontend.HostPort)
				}
				godebug.Line(ctx, scope, 55)
			}
		}
		godebug.Line(ctx, scope, 54)
	}
	godebug.Line(ctx, scope, 60)

	return nil
}

var port_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/go-connections/nat"
)

// CmdPort lists port mappings for a container.
// If a private port is specified, it also shows the public-facing port that is NATed to the private port.
//
// Usage: docker port CONTAINER [PRIVATE_PORT[/PROTO]]
func (cli *DockerCli) CmdPort(args ...string) error {
	cmd := Cli.Subcmd("port", []string{"CONTAINER [PRIVATE_PORT[/PROTO]]"}, Cli.DockerCommands["port"].Description, true)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	c, err := cli.client.ContainerInspect(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}

	if cmd.NArg() == 2 {
		var (
			port  = cmd.Arg(1)
			proto = "tcp"
			parts = strings.SplitN(port, "/", 2)
		)

		if len(parts) == 2 && len(parts[1]) != 0 {
			port = parts[0]
			proto = parts[1]
		}
		natPort := port + "/" + proto
		newP, err := nat.NewPort(proto, port)
		if err != nil {
			return err
		}
		if frontends, exists := c.NetworkSettings.Ports[newP]; exists && frontends != nil {
			for _, frontend := range frontends {
				fmt.Fprintf(cli.out, "%s:%s\n", frontend.HostIP, frontend.HostPort)
			}
			return nil
		}
		return fmt.Errorf("Error: No public port '%s' published for %s", natPort, cmd.Arg(0))
	}

	for from, frontends := range c.NetworkSettings.Ports {
		for _, frontend := range frontends {
			fmt.Fprintf(cli.out, "%s -> %s:%s\n", from, frontend.HostIP, frontend.HostPort)
		}
	}

	return nil
}
`


var ps_go_scope = godebug.EnteringNewFile(client_pkg_scope, ps_go_contents)

func (cli *DockerCli) CmdPs(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdPs(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := ps_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 18)
	var (
		err error

		psFilterArgs = filters.NewArgs()

		cmd      = Cli.Subcmd("ps", nil, Cli.DockerCommands["ps"].Description, true)
		quiet    = cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")
		size     = cmd.Bool([]string{"s", "-size"}, false, "Display total file sizes")
		all      = cmd.Bool([]string{"a", "-all"}, false, "Show all containers (default shows just running)")
		noTrunc  = cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
		nLatest  = cmd.Bool([]string{"l", "-latest"}, false, "Show the latest created container (includes all states)")
		since    = cmd.String([]string{"#-since"}, "", "Show containers created since Id or Name (includes all states)")
		before   = cmd.String([]string{"#-before"}, "", "Only show containers created before Id or Name")
		last     = cmd.Int([]string{"n"}, -1, "Show n last created containers (includes all states)")
		format   = cmd.String([]string{"-format"}, "", "Pretty-print containers using a Go template")
		flFilter = opts.NewListOpts(nil)
	)
	scope.Declare("err", &err, "psFilterArgs", &psFilterArgs, "cmd", &cmd, "quiet", &quiet, "size", &size, "all", &all, "noTrunc", &noTrunc, "nLatest", &nLatest, "since", &since, "before", &before, "last", &last, "format", &format, "flFilter", &flFilter)
	godebug.Line(ctx, scope, 35)

	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 37)

	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	godebug.Line(ctx, scope, 39)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 40)
	if *last == -1 && *nLatest {
		godebug.Line(ctx, scope, 41)
		*last = 1
	}
	{
		scope := scope.EnteringNewChildScope()

		for _, f := range flFilter.GetAll() {
			godebug.Line(ctx, scope, 46)
			scope.Declare("f", &f)
			godebug.Line(ctx, scope, 47)
			if psFilterArgs, err = filters.ParseFlag(f, psFilterArgs); err != nil {
				godebug.Line(ctx, scope, 48)
				return err
			}
		}
		godebug.Line(ctx, scope, 46)
	}
	godebug.Line(ctx, scope, 52)

	options := types.ContainerListOptions{
		All:    *all,
		Limit:  *last,
		Since:  *since,
		Before: *before,
		Size:   *size,
		Filter: psFilterArgs,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 61)

	containers, err := cli.client.ContainerList(context.Background(), options)
	scope.Declare("containers", &containers)
	godebug.Line(ctx, scope, 62)
	if err != nil {
		godebug.Line(ctx, scope, 63)
		return err
	}
	godebug.Line(ctx, scope, 66)

	f := *format
	scope.Declare("f", &f)
	godebug.Line(ctx, scope, 67)
	if len(f) == 0 {
		godebug.Line(ctx, scope, 68)
		if len(cli.PsFormat()) > 0 && !*quiet {
			godebug.Line(ctx, scope, 69)
			f = cli.PsFormat()
		} else {
			godebug.Line(ctx, scope, 70)
			godebug.Line(ctx, scope, 71)
			f = "table"
		}
	}
	godebug.Line(ctx, scope, 75)

	psCtx := formatter.ContainerContext{
		Context: formatter.Context{
			Output: cli.out,
			Format: f,
			Quiet:  *quiet,
			Trunc:  !*noTrunc,
		},
		Size:       *size,
		Containers: containers,
	}
	scope.Declare("psCtx", &psCtx)
	godebug.Line(ctx, scope, 86)

	psCtx.Write()
	godebug.Line(ctx, scope, 88)

	return nil
}

var ps_go_contents = `package client

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/client/formatter"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// CmdPs outputs a list of Docker containers.
//
// Usage: docker ps [OPTIONS]
func (cli *DockerCli) CmdPs(args ...string) error {
	var (
		err error

		psFilterArgs = filters.NewArgs()

		cmd      = Cli.Subcmd("ps", nil, Cli.DockerCommands["ps"].Description, true)
		quiet    = cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")
		size     = cmd.Bool([]string{"s", "-size"}, false, "Display total file sizes")
		all      = cmd.Bool([]string{"a", "-all"}, false, "Show all containers (default shows just running)")
		noTrunc  = cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
		nLatest  = cmd.Bool([]string{"l", "-latest"}, false, "Show the latest created container (includes all states)")
		since    = cmd.String([]string{"#-since"}, "", "Show containers created since Id or Name (includes all states)")
		before   = cmd.String([]string{"#-before"}, "", "Only show containers created before Id or Name")
		last     = cmd.Int([]string{"n"}, -1, "Show n last created containers (includes all states)")
		format   = cmd.String([]string{"-format"}, "", "Pretty-print containers using a Go template")
		flFilter = opts.NewListOpts(nil)
	)
	cmd.Require(flag.Exact, 0)

	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.ParseFlags(args, true)
	if *last == -1 && *nLatest {
		*last = 1
	}

	// Consolidate all filter flags, and sanity check them.
	// They'll get processed in the daemon/server.
	for _, f := range flFilter.GetAll() {
		if psFilterArgs, err = filters.ParseFlag(f, psFilterArgs); err != nil {
			return err
		}
	}

	options := types.ContainerListOptions{
		All:    *all,
		Limit:  *last,
		Since:  *since,
		Before: *before,
		Size:   *size,
		Filter: psFilterArgs,
	}

	containers, err := cli.client.ContainerList(context.Background(), options)
	if err != nil {
		return err
	}

	f := *format
	if len(f) == 0 {
		if len(cli.PsFormat()) > 0 && !*quiet {
			f = cli.PsFormat()
		} else {
			f = "table"
		}
	}

	psCtx := formatter.ContainerContext{
		Context: formatter.Context{
			Output: cli.out,
			Format: f,
			Quiet:  *quiet,
			Trunc:  !*noTrunc,
		},
		Size:       *size,
		Containers: containers,
	}

	psCtx.Write()

	return nil
}
`


var pull_go_scope = godebug.EnteringNewFile(client_pkg_scope, pull_go_contents)

func (cli *DockerCli) CmdPull(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdPull(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := pull_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 22)
	cmd := Cli.Subcmd("pull", []string{"NAME[:TAG|@DIGEST]"}, Cli.DockerCommands["pull"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 23)
	allTags := cmd.Bool([]string{"a", "-all-tags"}, false, "Download all tagged images in the repository")
	scope.Declare("allTags", &allTags)
	godebug.Line(ctx, scope, 24)
	addTrustedFlags(cmd, true)
	godebug.Line(ctx, scope, 25)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 27)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 28)
	remote := cmd.Arg(0)
	scope.Declare("remote", &remote)
	godebug.Line(ctx, scope, 30)

	distributionRef, err := reference.ParseNamed(remote)
	scope.Declare("distributionRef", &distributionRef, "err", &err)
	godebug.Line(ctx, scope, 31)
	if err != nil {
		godebug.Line(ctx, scope, 32)
		return err
	}
	godebug.Line(ctx, scope, 34)
	if *allTags && !reference.IsNameOnly(distributionRef) {
		godebug.Line(ctx, scope, 35)
		return errors.New("tag can't be used with --all-tags/-a")
	}
	godebug.Line(ctx, scope, 38)

	if !*allTags && reference.IsNameOnly(distributionRef) {
		godebug.Line(ctx, scope, 39)
		distributionRef = reference.WithDefaultTag(distributionRef)
		godebug.Line(ctx, scope, 40)
		fmt.Fprintf(cli.out, "Using default tag: %s\n", reference.DefaultTag)
	}
	godebug.Line(ctx, scope, 43)

	var tag string
	scope.Declare("tag", &tag)
	godebug.Line(ctx, scope, 44)
	switch x := distributionRef.(type) {
	case reference.Canonical:
		godebug.Line(ctx, scope, 45)
		godebug.Line(ctx, scope, 46)
		tag = x.Digest().String()
	case reference.NamedTagged:
		godebug.Line(ctx, scope, 47)
		godebug.Line(ctx, scope, 48)
		tag = x.Tag()
	}
	godebug.Line(ctx, scope, 51)

	ref := registry.ParseReference(tag)
	scope.Declare("ref", &ref)
	godebug.Line(ctx, scope, 54)

	repoInfo, err := registry.ParseRepositoryInfo(distributionRef)
	scope.Declare("repoInfo", &repoInfo)
	godebug.Line(ctx, scope, 55)
	if err != nil {
		godebug.Line(ctx, scope, 56)
		return err
	}
	godebug.Line(ctx, scope, 59)

	authConfig := cli.resolveAuthConfig(repoInfo.Index)
	scope.Declare("authConfig", &authConfig)
	godebug.Line(ctx, scope, 60)
	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(repoInfo.Index, "pull")
	scope.Declare("requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 62)

	if isTrusted() && !ref.HasDigest() {
		godebug.Line(ctx, scope, 64)

		return cli.trustedPull(repoInfo, ref, authConfig, requestPrivilege)
	}
	godebug.Line(ctx, scope, 67)

	return cli.imagePullPrivileged(authConfig, distributionRef.String(), "", requestPrivilege)
}

func (cli *DockerCli) imagePullPrivileged(authConfig types.AuthConfig, imageID, tag string, requestPrivilege apiclient.RequestPrivilegeFunc) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.imagePullPrivileged(authConfig, imageID, tag, requestPrivilege)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := pull_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "authConfig", &authConfig, "imageID", &imageID, "tag", &tag, "requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 72)

	encodedAuth, err := encodeAuthToBase64(authConfig)
	scope.Declare("encodedAuth", &encodedAuth, "err", &err)
	godebug.Line(ctx, scope, 73)
	if err != nil {
		godebug.Line(ctx, scope, 74)
		return err
	}
	godebug.Line(ctx, scope, 76)
	options := types.ImagePullOptions{
		ImageID:      imageID,
		Tag:          tag,
		RegistryAuth: encodedAuth,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 82)

	responseBody, err := cli.client.ImagePull(context.Background(), options, requestPrivilege)
	scope.Declare("responseBody", &responseBody)
	godebug.Line(ctx, scope, 83)
	if err != nil {
		godebug.Line(ctx, scope, 84)
		return err
	}
	godebug.Line(ctx, scope, 86)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 86)
	godebug.Line(ctx, scope, 88)

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}

var pull_go_contents = `package client

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// CmdPull pulls an image or a repository from the registry.
//
// Usage: docker pull [OPTIONS] IMAGENAME[:TAG|@DIGEST]
func (cli *DockerCli) CmdPull(args ...string) error {
	cmd := Cli.Subcmd("pull", []string{"NAME[:TAG|@DIGEST]"}, Cli.DockerCommands["pull"].Description, true)
	allTags := cmd.Bool([]string{"a", "-all-tags"}, false, "Download all tagged images in the repository")
	addTrustedFlags(cmd, true)
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)
	remote := cmd.Arg(0)

	distributionRef, err := reference.ParseNamed(remote)
	if err != nil {
		return err
	}
	if *allTags && !reference.IsNameOnly(distributionRef) {
		return errors.New("tag can't be used with --all-tags/-a")
	}

	if !*allTags && reference.IsNameOnly(distributionRef) {
		distributionRef = reference.WithDefaultTag(distributionRef)
		fmt.Fprintf(cli.out, "Using default tag: %s\n", reference.DefaultTag)
	}

	var tag string
	switch x := distributionRef.(type) {
	case reference.Canonical:
		tag = x.Digest().String()
	case reference.NamedTagged:
		tag = x.Tag()
	}

	ref := registry.ParseReference(tag)

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(distributionRef)
	if err != nil {
		return err
	}

	authConfig := cli.resolveAuthConfig(repoInfo.Index)
	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(repoInfo.Index, "pull")

	if isTrusted() && !ref.HasDigest() {
		// Check if tag is digest
		return cli.trustedPull(repoInfo, ref, authConfig, requestPrivilege)
	}

	return cli.imagePullPrivileged(authConfig, distributionRef.String(), "", requestPrivilege)
}

func (cli *DockerCli) imagePullPrivileged(authConfig types.AuthConfig, imageID, tag string, requestPrivilege client.RequestPrivilegeFunc) error {

	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}
	options := types.ImagePullOptions{
		ImageID:      imageID,
		Tag:          tag,
		RegistryAuth: encodedAuth,
	}

	responseBody, err := cli.client.ImagePull(context.Background(), options, requestPrivilege)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}
`


var push_go_scope = godebug.EnteringNewFile(client_pkg_scope, push_go_contents)

func (cli *DockerCli) CmdPush(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdPush(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := push_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 22)
	cmd := Cli.Subcmd("push", []string{"NAME[:TAG]"}, Cli.DockerCommands["push"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 23)
	addTrustedFlags(cmd, false)
	godebug.Line(ctx, scope, 24)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 26)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 28)

	ref, err := reference.ParseNamed(cmd.Arg(0))
	scope.Declare("ref", &ref, "err", &err)
	godebug.Line(ctx, scope, 29)
	if err != nil {
		godebug.Line(ctx, scope, 30)
		return err
	}
	godebug.Line(ctx, scope, 33)

	var tag string
	scope.Declare("tag", &tag)
	godebug.Line(ctx, scope, 34)
	switch x := ref.(type) {
	case reference.Canonical:
		godebug.Line(ctx, scope, 35)
		godebug.Line(ctx, scope, 36)
		return errors.New("cannot push a digest reference")
	case reference.NamedTagged:
		godebug.Line(ctx, scope, 37)
		godebug.Line(ctx, scope, 38)
		tag = x.Tag()
	}
	godebug.Line(ctx, scope, 42)

	repoInfo, err := registry.ParseRepositoryInfo(ref)
	scope.Declare("repoInfo", &repoInfo)
	godebug.Line(ctx, scope, 43)
	if err != nil {
		godebug.Line(ctx, scope, 44)
		return err
	}
	godebug.Line(ctx, scope, 47)

	authConfig := cli.resolveAuthConfig(repoInfo.Index)
	scope.Declare("authConfig", &authConfig)
	godebug.Line(ctx, scope, 49)

	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(repoInfo.Index, "push")
	scope.Declare("requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 50)
	if isTrusted() {
		godebug.Line(ctx, scope, 51)
		return cli.trustedPush(repoInfo, tag, authConfig, requestPrivilege)
	}
	godebug.Line(ctx, scope, 54)

	responseBody, err := cli.imagePushPrivileged(authConfig, ref.Name(), tag, requestPrivilege)
	scope.Declare("responseBody", &responseBody)
	godebug.Line(ctx, scope, 55)
	if err != nil {
		godebug.Line(ctx, scope, 56)
		return err
	}
	godebug.Line(ctx, scope, 59)

	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 59)
	godebug.Line(ctx, scope, 61)

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}

func (cli *DockerCli) imagePushPrivileged(authConfig types.AuthConfig, imageID, tag string, requestPrivilege apiclient.RequestPrivilegeFunc) (io.ReadCloser, error) {
	var result1 io.ReadCloser
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = cli.imagePushPrivileged(authConfig, imageID, tag, requestPrivilege)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := push_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "authConfig", &authConfig, "imageID", &imageID, "tag", &tag, "requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 65)
	encodedAuth, err := encodeAuthToBase64(authConfig)
	scope.Declare("encodedAuth", &encodedAuth, "err", &err)
	godebug.Line(ctx, scope, 66)
	if err != nil {
		godebug.Line(ctx, scope, 67)
		return nil, err
	}
	godebug.Line(ctx, scope, 69)
	options := types.ImagePushOptions{
		ImageID:      imageID,
		Tag:          tag,
		RegistryAuth: encodedAuth,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 75)

	return cli.client.ImagePush(context.Background(), options, requestPrivilege)
}

var push_go_contents = `package client

import (
	"errors"
	"io"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// CmdPush pushes an image or repository to the registry.
//
// Usage: docker push NAME[:TAG]
func (cli *DockerCli) CmdPush(args ...string) error {
	cmd := Cli.Subcmd("push", []string{"NAME[:TAG]"}, Cli.DockerCommands["push"].Description, true)
	addTrustedFlags(cmd, false)
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	ref, err := reference.ParseNamed(cmd.Arg(0))
	if err != nil {
		return err
	}

	var tag string
	switch x := ref.(type) {
	case reference.Canonical:
		return errors.New("cannot push a digest reference")
	case reference.NamedTagged:
		tag = x.Tag()
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}
	// Resolve the Auth config relevant for this server
	authConfig := cli.resolveAuthConfig(repoInfo.Index)

	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(repoInfo.Index, "push")
	if isTrusted() {
		return cli.trustedPush(repoInfo, tag, authConfig, requestPrivilege)
	}

	responseBody, err := cli.imagePushPrivileged(authConfig, ref.Name(), tag, requestPrivilege)
	if err != nil {
		return err
	}

	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}

func (cli *DockerCli) imagePushPrivileged(authConfig types.AuthConfig, imageID, tag string, requestPrivilege client.RequestPrivilegeFunc) (io.ReadCloser, error) {
	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return nil, err
	}
	options := types.ImagePushOptions{
		ImageID:      imageID,
		Tag:          tag,
		RegistryAuth: encodedAuth,
	}

	return cli.client.ImagePush(context.Background(), options, requestPrivilege)
}
`


var rename_go_scope = godebug.EnteringNewFile(client_pkg_scope, rename_go_contents)

func (cli *DockerCli) CmdRename(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdRename(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := rename_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 17)
	cmd := Cli.Subcmd("rename", []string{"OLD_NAME NEW_NAME"}, Cli.DockerCommands["rename"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 18)
	cmd.Require(flag.Exact, 2)
	godebug.Line(ctx, scope, 20)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 22)

	oldName := strings.TrimSpace(cmd.Arg(0))
	scope.Declare("oldName", &oldName)
	godebug.Line(ctx, scope, 23)
	newName := strings.TrimSpace(cmd.Arg(1))
	scope.Declare("newName", &newName)
	godebug.Line(ctx, scope, 25)

	if oldName == "" || newName == "" {
		godebug.Line(ctx, scope, 26)
		return fmt.Errorf("Error: Neither old nor new names may be empty")
	}
	godebug.Line(ctx, scope, 29)

	if err := cli.client.ContainerRename(context.Background(), oldName, newName); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 30)
		fmt.Fprintf(cli.err, "%s\n", err)
		godebug.Line(ctx, scope, 31)
		return fmt.Errorf("Error: failed to rename container named %s", oldName)
	}
	godebug.Line(ctx, scope, 33)
	return nil
}

var rename_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdRename renames a container.
//
// Usage: docker rename OLD_NAME NEW_NAME
func (cli *DockerCli) CmdRename(args ...string) error {
	cmd := Cli.Subcmd("rename", []string{"OLD_NAME NEW_NAME"}, Cli.DockerCommands["rename"].Description, true)
	cmd.Require(flag.Exact, 2)

	cmd.ParseFlags(args, true)

	oldName := strings.TrimSpace(cmd.Arg(0))
	newName := strings.TrimSpace(cmd.Arg(1))

	if oldName == "" || newName == "" {
		return fmt.Errorf("Error: Neither old nor new names may be empty")
	}

	if err := cli.client.ContainerRename(context.Background(), oldName, newName); err != nil {
		fmt.Fprintf(cli.err, "%s\n", err)
		return fmt.Errorf("Error: failed to rename container named %s", oldName)
	}
	return nil
}
`


var restart_go_scope = godebug.EnteringNewFile(client_pkg_scope, restart_go_contents)

func (cli *DockerCli) CmdRestart(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdRestart(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := restart_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 17)
	cmd := Cli.Subcmd("restart", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["restart"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 18)
	nSeconds := cmd.Int([]string{"t", "-time"}, 10, "Seconds to wait for stop before killing the container")
	scope.Declare("nSeconds", &nSeconds)
	godebug.Line(ctx, scope, 19)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 21)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 23)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 24)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 25)
			if err := cli.client.ContainerRestart(context.Background(), name, *nSeconds); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 26)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 27)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 28)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 24)
	}
	godebug.Line(ctx, scope, 31)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 32)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 34)
	return nil
}

var restart_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdRestart restarts one or more containers.
//
// Usage: docker restart [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdRestart(args ...string) error {
	cmd := Cli.Subcmd("restart", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["restart"].Description, true)
	nSeconds := cmd.Int([]string{"t", "-time"}, 10, "Seconds to wait for stop before killing the container")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		if err := cli.client.ContainerRestart(context.Background(), name, *nSeconds); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var rm_go_scope = godebug.EnteringNewFile(client_pkg_scope, rm_go_contents)

func (cli *DockerCli) CmdRm(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdRm(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := rm_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 18)
	cmd := Cli.Subcmd("rm", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["rm"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 19)
	v := cmd.Bool([]string{"v", "-volumes"}, false, "Remove the volumes associated with the container")
	scope.Declare("v", &v)
	godebug.Line(ctx, scope, 20)
	link := cmd.Bool([]string{"l", "-link"}, false, "Remove the specified link")
	scope.Declare("link", &link)
	godebug.Line(ctx, scope, 21)
	force := cmd.Bool([]string{"f", "-force"}, false, "Force the removal of a running container (uses SIGKILL)")
	scope.Declare("force", &force)
	godebug.Line(ctx, scope, 22)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 24)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 26)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 27)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 28)
			if name == "" {
				godebug.Line(ctx, scope, 29)
				return fmt.Errorf("Container name cannot be empty")
			}
			godebug.Line(ctx, scope, 31)
			name = strings.Trim(name, "/")
			godebug.Line(ctx, scope, 33)

			if err := cli.removeContainer(name, *v, *link, *force); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 34)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 35)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 36)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 27)
	}
	godebug.Line(ctx, scope, 39)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 40)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 42)
	return nil
}

func (cli *DockerCli) removeContainer(containerID string, removeVolumes, removeLinks, force bool) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.removeContainer(containerID, removeVolumes, removeLinks, force)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := rm_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "containerID", &containerID, "removeVolumes", &removeVolumes, "removeLinks", &removeLinks, "force", &force)
	godebug.Line(ctx, scope, 46)
	options := types.ContainerRemoveOptions{
		ContainerID:   containerID,
		RemoveVolumes: removeVolumes,
		RemoveLinks:   removeLinks,
		Force:         force,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 52)

	if err := cli.client.ContainerRemove(context.Background(), options); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 53)
		return err
	}
	godebug.Line(ctx, scope, 55)
	return nil
}

var rm_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
)

// CmdRm removes one or more containers.
//
// Usage: docker rm [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdRm(args ...string) error {
	cmd := Cli.Subcmd("rm", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["rm"].Description, true)
	v := cmd.Bool([]string{"v", "-volumes"}, false, "Remove the volumes associated with the container")
	link := cmd.Bool([]string{"l", "-link"}, false, "Remove the specified link")
	force := cmd.Bool([]string{"f", "-force"}, false, "Force the removal of a running container (uses SIGKILL)")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		if name == "" {
			return fmt.Errorf("Container name cannot be empty")
		}
		name = strings.Trim(name, "/")

		if err := cli.removeContainer(name, *v, *link, *force); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

func (cli *DockerCli) removeContainer(containerID string, removeVolumes, removeLinks, force bool) error {
	options := types.ContainerRemoveOptions{
		ContainerID:   containerID,
		RemoveVolumes: removeVolumes,
		RemoveLinks:   removeLinks,
		Force:         force,
	}
	if err := cli.client.ContainerRemove(context.Background(), options); err != nil {
		return err
	}
	return nil
}
`


var rmi_go_scope = godebug.EnteringNewFile(client_pkg_scope, rmi_go_contents)

func (cli *DockerCli) CmdRmi(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdRmi(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := rmi_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("rmi", []string{"IMAGE [IMAGE...]"}, Cli.DockerCommands["rmi"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	force := cmd.Bool([]string{"f", "-force"}, false, "Force removal of the image")
	scope.Declare("force", &force)
	godebug.Line(ctx, scope, 21)
	noprune := cmd.Bool([]string{"-no-prune"}, false, "Do not delete untagged parents")
	scope.Declare("noprune", &noprune)
	godebug.Line(ctx, scope, 22)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 24)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 26)

	v := url.Values{}
	scope.Declare("v", &v)
	godebug.Line(ctx, scope, 27)
	if *force {
		godebug.Line(ctx, scope, 28)
		v.Set("force", "1")
	}
	godebug.Line(ctx, scope, 30)
	if *noprune {
		godebug.Line(ctx, scope, 31)
		v.Set("noprune", "1")
	}
	godebug.Line(ctx, scope, 34)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 35)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 36)
			options := types.ImageRemoveOptions{
				ImageID:       name,
				Force:         *force,
				PruneChildren: !*noprune,
			}
			scope := scope.EnteringNewChildScope()
			scope.Declare("options", &options)
			godebug.Line(ctx, scope, 42)

			dels, err := cli.client.ImageRemove(context.Background(), options)
			scope.Declare("dels", &dels, "err", &err)
			godebug.Line(ctx, scope, 43)
			if err != nil {
				godebug.Line(ctx, scope, 44)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 45)
				{
					scope := scope.EnteringNewChildScope()
					for _, del := range dels {
						godebug.Line(ctx, scope, 46)
						scope.Declare("del", &del)
						godebug.Line(ctx, scope, 47)
						if del.Deleted != "" {
							godebug.Line(ctx, scope, 48)
							fmt.Fprintf(cli.out, "Deleted: %s\n", del.Deleted)
						} else {
							godebug.Line(ctx, scope, 49)
							godebug.Line(ctx, scope, 50)
							fmt.Fprintf(cli.out, "Untagged: %s\n", del.Untagged)
						}
					}
					godebug.Line(ctx, scope, 46)
				}
			}
		}
		godebug.Line(ctx, scope, 35)
	}
	godebug.Line(ctx, scope, 55)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 56)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 58)
	return nil
}

var rmi_go_contents = `package client

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
)

// CmdRmi removes all images with the specified name(s).
//
// Usage: docker rmi [OPTIONS] IMAGE [IMAGE...]
func (cli *DockerCli) CmdRmi(args ...string) error {
	cmd := Cli.Subcmd("rmi", []string{"IMAGE [IMAGE...]"}, Cli.DockerCommands["rmi"].Description, true)
	force := cmd.Bool([]string{"f", "-force"}, false, "Force removal of the image")
	noprune := cmd.Bool([]string{"-no-prune"}, false, "Do not delete untagged parents")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	v := url.Values{}
	if *force {
		v.Set("force", "1")
	}
	if *noprune {
		v.Set("noprune", "1")
	}

	var errs []string
	for _, name := range cmd.Args() {
		options := types.ImageRemoveOptions{
			ImageID:       name,
			Force:         *force,
			PruneChildren: !*noprune,
		}

		dels, err := cli.client.ImageRemove(context.Background(), options)
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			for _, del := range dels {
				if del.Deleted != "" {
					fmt.Fprintf(cli.out, "Deleted: %s\n", del.Deleted)
				} else {
					fmt.Fprintf(cli.out, "Untagged: %s\n", del.Untagged)
				}
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var run_go_scope = godebug.EnteringNewFile(client_pkg_scope, run_go_contents)

const (
	errCmdNotFound          = "not found or does not exist"
	errCmdCouldNotBeInvoked = "could not be invoked"
)

func (cid *cidFile) Close() error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cid.Close()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := run_go_scope.EnteringNewChildScope()
	scope.Declare("cid", &cid)
	godebug.Line(_ctx, scope, 29)
	cid.file.Close()
	godebug.Line(_ctx, scope, 31)

	if !cid.written {
		godebug.Line(_ctx, scope, 32)
		if err := os.Remove(cid.path); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 33)
			return fmt.Errorf("failed to remove the CID file '%s': %s \n", cid.path, err)
		}
	}
	godebug.Line(_ctx, scope, 37)

	return nil
}

func (cid *cidFile) Write(id string) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cid.Write(id)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := run_go_scope.EnteringNewChildScope()
	scope.Declare("cid", &cid, "id", &id)
	godebug.Line(_ctx, scope, 41)
	if _, err := cid.file.Write([]byte(id)); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(_ctx, scope, 42)
		return fmt.Errorf("Failed to write the container ID to the file: %s", err)
	}
	godebug.Line(_ctx, scope, 44)
	cid.written = true
	godebug.Line(_ctx, scope, 45)
	return nil
}

func runStartContainerErr(err error) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = runStartContainerErr(err)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := run_go_scope.EnteringNewChildScope()
	scope.Declare("err", &err)
	godebug.Line(_ctx, scope, 52)
	trimmedErr := strings.TrimPrefix(err.Error(), "Error response from daemon: ")
	scope.Declare("trimmedErr", &trimmedErr)
	godebug.Line(_ctx, scope, 53)
	statusError := Cli.StatusError{StatusCode: 125}
	scope.Declare("statusError", &statusError)
	godebug.Line(_ctx, scope, 54)
	if strings.HasPrefix(trimmedErr, "Container command") {
		godebug.Line(_ctx, scope, 55)
		if strings.Contains(trimmedErr, errCmdNotFound) {
			godebug.Line(_ctx, scope, 56)
			statusError = Cli.StatusError{StatusCode: 127}
		} else {
			godebug.ElseIfExpr(_ctx, scope, 57)
			if strings.Contains(trimmedErr, errCmdCouldNotBeInvoked) {
				godebug.Line(_ctx, scope, 58)
				statusError = Cli.StatusError{StatusCode: 126}
			}
		}
	}
	godebug.Line(_ctx, scope, 62)

	return statusError
}

func (cli *DockerCli) CmdRun(args ...string) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdRun(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := run_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(_ctx, scope, 69)
	logrus.Debugf("Executing api/client/run.go : CmdRun(%s)", args)
	godebug.Line(_ctx, scope, 70)
	logrus.Debug("Stack trace:")
	godebug.Line(_ctx, scope, 71)
	debug.PrintStack()
	godebug.Line(_ctx, scope, 72)
	cmd := Cli.Subcmd("run", []string{"IMAGE [COMMAND] [ARG...]"}, Cli.DockerCommands["run"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(_ctx, scope, 73)
	addTrustedFlags(cmd, true)
	godebug.Line(_ctx, scope, 76)

	var (
		flAutoRemove = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Run container in background and print container ID")
		flSigProxy   = cmd.Bool([]string{"-sig-proxy"}, true, "Proxy received signals to the process")
		flName       = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		flDetachKeys = cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
		flAttach     *opts.ListOpts

		ErrConflictAttachDetach               = fmt.Errorf("Conflicting options: -a and -d")
		ErrConflictRestartPolicyAndAutoRemove = fmt.Errorf("Conflicting options: --restart and --rm")
		ErrConflictDetachAutoRemove           = fmt.Errorf("Conflicting options: --rm and -d")
	)
	scope.Declare("flAutoRemove", &flAutoRemove, "flDetach", &flDetach, "flSigProxy", &flSigProxy, "flName", &flName, "flDetachKeys", &flDetachKeys, "flAttach", &flAttach, "ErrConflictAttachDetach", &ErrConflictAttachDetach, "ErrConflictRestartPolicyAndAutoRemove", &ErrConflictRestartPolicyAndAutoRemove, "ErrConflictDetachAutoRemove", &ErrConflictDetachAutoRemove)
	godebug.Line(_ctx, scope, 89)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)
	scope.Declare("config", &config, "hostConfig", &hostConfig, "networkingConfig", &networkingConfig, "err", &err)
	godebug.Line(_ctx, scope, 92)

	if err != nil {
		godebug.Line(_ctx, scope, 93)
		cmd.ReportError(err.Error(), true)
		godebug.Line(_ctx, scope, 94)
		os.Exit(125)
	}
	godebug.Line(_ctx, scope, 97)

	if hostConfig.OomKillDisable != nil && *hostConfig.OomKillDisable && hostConfig.Memory == 0 {
		godebug.Line(_ctx, scope, 98)
		fmt.Fprintf(cli.err, "WARNING: Disabling the OOM killer on containers without setting a '-m/--memory' limit may be dangerous.\n")
	}
	godebug.Line(_ctx, scope, 101)

	if len(hostConfig.DNS) > 0 {
		{
			scope := scope.EnteringNewChildScope()

			for _, dnsIP := range hostConfig.DNS {
				godebug.Line(_ctx, scope, 105)
				scope.Declare("dnsIP", &dnsIP)
				godebug.Line(_ctx, scope, 106)
				if dns.IsLocalhost(dnsIP) {
					godebug.Line(_ctx, scope, 107)
					fmt.Fprintf(cli.err, "WARNING: Localhost DNS setting (--dns=%s) may fail in containers.\n", dnsIP)
					godebug.Line(_ctx, scope, 108)
					break
				}
			}
			godebug.Line(_ctx, scope, 105)
		}
	}
	godebug.Line(_ctx, scope, 112)
	if config.Image == "" {
		godebug.Line(_ctx, scope, 113)
		cmd.Usage()
		godebug.Line(_ctx, scope, 114)
		return nil
	}
	godebug.Line(_ctx, scope, 117)

	config.ArgsEscaped = false
	godebug.Line(_ctx, scope, 119)

	if !*flDetach {
		godebug.Line(_ctx, scope, 120)
		if err := cli.CheckTtyInput(config.AttachStdin, config.Tty); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 121)
			return err
		}
	} else {
		godebug.Line(_ctx, scope, 123)
		godebug.Line(_ctx, scope, 124)
		if fl := cmd.Lookup("-attach"); fl != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("fl", &fl)
			godebug.Line(_ctx, scope, 125)
			flAttach = fl.Value.(*opts.ListOpts)
			godebug.Line(_ctx, scope, 126)
			if flAttach.Len() != 0 {
				godebug.Line(_ctx, scope, 127)
				return ErrConflictAttachDetach
			}
		}
		godebug.Line(_ctx, scope, 130)
		if *flAutoRemove {
			godebug.Line(_ctx, scope, 131)
			return ErrConflictDetachAutoRemove
		}
		godebug.Line(_ctx, scope, 134)

		config.AttachStdin = false
		godebug.Line(_ctx, scope, 135)
		config.AttachStdout = false
		godebug.Line(_ctx, scope, 136)
		config.AttachStderr = false
		godebug.Line(_ctx, scope, 137)
		config.StdinOnce = false
	}
	godebug.Line(_ctx, scope, 141)

	sigProxy := *flSigProxy
	scope.Declare("sigProxy", &sigProxy)
	godebug.Line(_ctx, scope, 142)
	if config.Tty {
		godebug.Line(_ctx, scope, 143)
		sigProxy = false
	}
	godebug.Line(_ctx, scope, 149)

	if runtime.GOOS == "windows" {
		godebug.Line(_ctx, scope, 150)
		hostConfig.ConsoleSize[0], hostConfig.ConsoleSize[1] = cli.getTtySize()
	}
	godebug.Line(_ctx, scope, 153)

	createResponse, err := cli.createContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	scope.Declare("createResponse", &createResponse)
	godebug.Line(_ctx, scope, 154)
	if err != nil {
		godebug.Line(_ctx, scope, 155)
		cmd.ReportError(err.Error(), true)
		godebug.Line(_ctx, scope, 156)
		return runStartContainerErr(err)
	}
	godebug.Line(_ctx, scope, 158)
	if sigProxy {
		godebug.Line(_ctx, scope, 159)
		sigc := cli.forwardAllSignals(createResponse.ID)
		scope := scope.EnteringNewChildScope()
		scope.Declare("sigc", &sigc)
		godebug.Line(_ctx, scope, 160)
		defer signal.StopCatch(sigc)
		defer godebug.Defer(_ctx, scope, 160)
	}
	godebug.Line(_ctx, scope, 162)
	var (
		waitDisplayID chan struct{}
		errCh         chan error
		cancelFun     context.CancelFunc
		ctx           context.Context
	)
	scope.Declare("waitDisplayID", &waitDisplayID, "errCh", &errCh, "cancelFun", &cancelFun, "ctx", &ctx)
	godebug.Line(_ctx, scope, 168)

	if !config.AttachStdout && !config.AttachStderr {
		godebug.Line(_ctx, scope, 170)

		waitDisplayID = make(chan struct{})
		godebug.Line(_ctx, scope, 171)
		go func() {
			fn := func(_ctx *godebug.Context) {
				godebug.Line(_ctx, scope, 172)
				defer close(waitDisplayID)
				defer godebug.Defer(_ctx, scope, 172)
				godebug.Line(_ctx, scope, 173)
				fmt.Fprintf(cli.out, "%s\n", createResponse.ID)
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
		}()
	}
	godebug.Line(_ctx, scope, 176)
	if *flAutoRemove && (hostConfig.RestartPolicy.IsAlways() || hostConfig.RestartPolicy.IsOnFailure()) {
		godebug.Line(_ctx, scope, 177)
		return ErrConflictRestartPolicyAndAutoRemove
	}
	godebug.Line(_ctx, scope, 179)
	attach := config.AttachStdin || config.AttachStdout || config.AttachStderr
	scope.Declare("attach", &attach)
	godebug.Line(_ctx, scope, 180)
	if attach {
		godebug.Line(_ctx, scope, 181)
		var (
			out, stderr io.Writer
			in          io.ReadCloser
		)
		scope := scope.EnteringNewChildScope()
		scope.Declare("out", &out, "stderr", &stderr, "in", &in)
		godebug.Line(_ctx, scope, 185)

		if config.AttachStdin {
			godebug.Line(_ctx, scope, 186)
			in = cli.in
		}
		godebug.Line(_ctx, scope, 188)
		if config.AttachStdout {
			godebug.Line(_ctx, scope, 189)
			out = cli.out
		}
		godebug.Line(_ctx, scope, 191)
		if config.AttachStderr {
			godebug.Line(_ctx, scope, 192)
			if config.Tty {
				godebug.Line(_ctx, scope, 193)
				stderr = cli.out
			} else {
				godebug.Line(_ctx, scope, 194)
				godebug.Line(_ctx, scope, 195)
				stderr = cli.err
			}
		}
		godebug.Line(_ctx, scope, 199)

		if *flDetachKeys != "" {
			godebug.Line(_ctx, scope, 200)
			cli.configFile.DetachKeys = *flDetachKeys
		}
		godebug.Line(_ctx, scope, 203)

		options := types.ContainerAttachOptions{
			ContainerID: createResponse.ID,
			Stream:      true,
			Stdin:       config.AttachStdin,
			Stdout:      config.AttachStdout,
			Stderr:      config.AttachStderr,
			DetachKeys:  cli.configFile.DetachKeys,
		}
		scope.Declare("options", &options)
		godebug.Line(_ctx, scope, 212)

		resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
		scope.Declare("resp", &resp, "errAttach", &errAttach)
		godebug.Line(_ctx, scope, 213)
		if errAttach != nil && errAttach != httputil.ErrPersistEOF {
			godebug.Line(_ctx, scope, 217)

			return errAttach
		}
		godebug.Line(_ctx, scope, 219)
		ctx, cancelFun = context.WithCancel(context.Background())
		godebug.Line(_ctx, scope, 220)
		errCh = promise.Go(func() error {
			var result1 error
			fn := func(_ctx *godebug.Context) {
				result1 = func() error {
					godebug.Line(_ctx, scope, 221)
					errHijack := cli.holdHijackedConnection(ctx, config.Tty, in, out, stderr, resp)
					scope := scope.EnteringNewChildScope()
					scope.Declare("errHijack", &errHijack)
					godebug.Line(_ctx, scope, 222)
					if errHijack == nil {
						godebug.Line(_ctx, scope, 223)
						return errAttach
					}
					godebug.Line(_ctx, scope, 225)
					return errHijack
				}()
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
			return result1
		},
		)
	}
	godebug.Line(_ctx, scope, 229)

	if *flAutoRemove {
		godebug.Line(_ctx, scope, 230)
		defer func() {
			fn := func(_ctx *godebug.Context) {
				godebug.Line(_ctx, scope, 231)
				if err := cli.removeContainer(createResponse.ID, true, false, true); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(_ctx, scope, 232)
					fmt.Fprintf(cli.err, "%v\n", err)
				}
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
		}()
		defer godebug.Defer(_ctx, scope, 230)
	}
	godebug.Line(_ctx, scope, 238)

	if err := cli.client.ContainerStart(context.Background(), createResponse.ID); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(_ctx, scope, 242)

		if attach {
			godebug.Line(_ctx, scope, 243)
			cancelFun()
			godebug.Line(_ctx, scope, 244)
			<-errCh
		}
		godebug.Line(_ctx, scope, 247)

		cmd.ReportError(err.Error(), false)
		godebug.Line(_ctx, scope, 248)
		return runStartContainerErr(err)
	}
	godebug.Line(_ctx, scope, 251)

	if (config.AttachStdin || config.AttachStdout || config.AttachStderr) && config.Tty && cli.isTerminalOut {
		godebug.Line(_ctx, scope, 252)
		if err := cli.monitorTtySize(createResponse.ID, false); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 253)
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}
	godebug.Line(_ctx, scope, 257)

	if errCh != nil {
		godebug.Line(_ctx, scope, 258)
		if err := <-errCh; err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 259)
			logrus.Debugf("Error hijack: %s", err)
			godebug.Line(_ctx, scope, 260)
			return err
		}
	}
	godebug.Line(_ctx, scope, 265)

	if !config.AttachStdout && !config.AttachStderr {
		godebug.Line(_ctx, scope, 267)

		<-waitDisplayID
		godebug.Line(_ctx, scope, 268)
		return nil
	}
	godebug.Line(_ctx, scope, 271)

	var status int
	scope.Declare("status", &status)
	godebug.Line(_ctx, scope, 274)

	if *flAutoRemove {
		godebug.Line(_ctx, scope, 277)

		if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
			godebug.Line(_ctx, scope, 278)
			return runStartContainerErr(err)
		}
		godebug.Line(_ctx, scope, 280)
		if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
			godebug.Line(_ctx, scope, 281)
			return err
		}
	} else {
		godebug.Line(_ctx, scope, 283)
		godebug.Line(_ctx, scope, 285)

		if !config.Tty {
			godebug.Line(_ctx, scope, 287)

			if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
				godebug.Line(_ctx, scope, 288)
				return err
			}
		} else {
			godebug.Line(_ctx, scope, 290)
			godebug.Line(_ctx, scope, 293)

			if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
				godebug.Line(_ctx, scope, 294)
				return err
			}
		}
	}
	godebug.Line(_ctx, scope, 298)
	if status != 0 {
		godebug.Line(_ctx, scope, 299)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(_ctx, scope, 301)
	return nil
}

var run_go_contents = `package client

import (
	"fmt"
	"io"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/signal"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types"
	"github.com/docker/libnetwork/resolvconf/dns"
	"golang.org/x/net/context"
	"runtime/debug"
)

const (
	errCmdNotFound          = "not found or does not exist"
	errCmdCouldNotBeInvoked = "could not be invoked"
)

func (cid *cidFile) Close() error {
	cid.file.Close()

	if !cid.written {
		if err := os.Remove(cid.path); err != nil {
			return fmt.Errorf("failed to remove the CID file '%s': %s \n", cid.path, err)
		}
	}

	return nil
}

func (cid *cidFile) Write(id string) error {
	if _, err := cid.file.Write([]byte(id)); err != nil {
		return fmt.Errorf("Failed to write the container ID to the file: %s", err)
	}
	cid.written = true
	return nil
}

// if container start fails with 'command not found' error, return 127
// if container start fails with 'command cannot be invoked' error, return 126
// return 125 for generic docker daemon failures
func runStartContainerErr(err error) error {
	trimmedErr := strings.TrimPrefix(err.Error(), "Error response from daemon: ")
	statusError := Cli.StatusError{StatusCode: 125}
	if strings.HasPrefix(trimmedErr, "Container command") {
		if strings.Contains(trimmedErr, errCmdNotFound) {
			statusError = Cli.StatusError{StatusCode: 127}
		} else if strings.Contains(trimmedErr, errCmdCouldNotBeInvoked) {
			statusError = Cli.StatusError{StatusCode: 126}
		}
	}

	return statusError
}

// CmdRun runs a command in a new container.
//
// Usage: docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
func (cli *DockerCli) CmdRun(args ...string) error {
	logrus.Debugf("Executing api/client/run.go : CmdRun(%s)", args)
	logrus.Debug("Stack trace:")
	debug.PrintStack()
	cmd := Cli.Subcmd("run", []string{"IMAGE [COMMAND] [ARG...]"}, Cli.DockerCommands["run"].Description, true)
	addTrustedFlags(cmd, true)

	// These are flags not stored in Config/HostConfig
	var (
		flAutoRemove = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Run container in background and print container ID")
		flSigProxy   = cmd.Bool([]string{"-sig-proxy"}, true, "Proxy received signals to the process")
		flName       = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		flDetachKeys = cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
		flAttach     *opts.ListOpts

		ErrConflictAttachDetach               = fmt.Errorf("Conflicting options: -a and -d")
		ErrConflictRestartPolicyAndAutoRemove = fmt.Errorf("Conflicting options: --restart and --rm")
		ErrConflictDetachAutoRemove           = fmt.Errorf("Conflicting options: --rm and -d")
	)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)

	// just in case the Parse does not exit
	if err != nil {
		cmd.ReportError(err.Error(), true)
		os.Exit(125)
	}

	if hostConfig.OomKillDisable != nil && *hostConfig.OomKillDisable && hostConfig.Memory == 0 {
		fmt.Fprintf(cli.err, "WARNING: Disabling the OOM killer on containers without setting a '-m/--memory' limit may be dangerous.\n")
	}

	if len(hostConfig.DNS) > 0 {
		// check the DNS settings passed via --dns against
		// localhost regexp to warn if they are trying to
		// set a DNS to a localhost address
		for _, dnsIP := range hostConfig.DNS {
			if dns.IsLocalhost(dnsIP) {
				fmt.Fprintf(cli.err, "WARNING: Localhost DNS setting (--dns=%s) may fail in containers.\n", dnsIP)
				break
			}
		}
	}
	if config.Image == "" {
		cmd.Usage()
		return nil
	}

	config.ArgsEscaped = false

	if !*flDetach {
		if err := cli.CheckTtyInput(config.AttachStdin, config.Tty); err != nil {
			return err
		}
	} else {
		if fl := cmd.Lookup("-attach"); fl != nil {
			flAttach = fl.Value.(*opts.ListOpts)
			if flAttach.Len() != 0 {
				return ErrConflictAttachDetach
			}
		}
		if *flAutoRemove {
			return ErrConflictDetachAutoRemove
		}

		config.AttachStdin = false
		config.AttachStdout = false
		config.AttachStderr = false
		config.StdinOnce = false
	}

	// Disable flSigProxy when in TTY mode
	sigProxy := *flSigProxy
	if config.Tty {
		sigProxy = false
	}

	// Telling the Windows daemon the initial size of the tty during start makes
	// a far better user experience rather than relying on subsequent resizes
	// to cause things to catch up.
	if runtime.GOOS == "windows" {
		hostConfig.ConsoleSize[0], hostConfig.ConsoleSize[1] = cli.getTtySize()
	}

	createResponse, err := cli.createContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	if err != nil {
		cmd.ReportError(err.Error(), true)
		return runStartContainerErr(err)
	}
	if sigProxy {
		sigc := cli.forwardAllSignals(createResponse.ID)
		defer signal.StopCatch(sigc)
	}
	var (
		waitDisplayID chan struct{}
		errCh         chan error
		cancelFun     context.CancelFunc
		ctx           context.Context
	)
	if !config.AttachStdout && !config.AttachStderr {
		// Make this asynchronous to allow the client to write to stdin before having to read the ID
		waitDisplayID = make(chan struct{})
		go func() {
			defer close(waitDisplayID)
			fmt.Fprintf(cli.out, "%s\n", createResponse.ID)
		}()
	}
	if *flAutoRemove && (hostConfig.RestartPolicy.IsAlways() || hostConfig.RestartPolicy.IsOnFailure()) {
		return ErrConflictRestartPolicyAndAutoRemove
	}
	attach := config.AttachStdin || config.AttachStdout || config.AttachStderr
	if attach {
		var (
			out, stderr io.Writer
			in          io.ReadCloser
		)
		if config.AttachStdin {
			in = cli.in
		}
		if config.AttachStdout {
			out = cli.out
		}
		if config.AttachStderr {
			if config.Tty {
				stderr = cli.out
			} else {
				stderr = cli.err
			}
		}

		if *flDetachKeys != "" {
			cli.configFile.DetachKeys = *flDetachKeys
		}

		options := types.ContainerAttachOptions{
			ContainerID: createResponse.ID,
			Stream:      true,
			Stdin:       config.AttachStdin,
			Stdout:      config.AttachStdout,
			Stderr:      config.AttachStderr,
			DetachKeys:  cli.configFile.DetachKeys,
		}

		resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
		if errAttach != nil && errAttach != httputil.ErrPersistEOF {
			// ContainerAttach returns an ErrPersistEOF (connection closed)
			// means server met an error and put it in Hijacked connection
			// keep the error and read detailed error message from hijacked connection later
			return errAttach
		}
		ctx, cancelFun = context.WithCancel(context.Background())
		errCh = promise.Go(func() error {
			errHijack := cli.holdHijackedConnection(ctx, config.Tty, in, out, stderr, resp)
			if errHijack == nil {
				return errAttach
			}
			return errHijack
		})
	}

	if *flAutoRemove {
		defer func() {
			if err := cli.removeContainer(createResponse.ID, true, false, true); err != nil {
				fmt.Fprintf(cli.err, "%v\n", err)
			}
		}()
	}

	//start the container
	if err := cli.client.ContainerStart(context.Background(), createResponse.ID); err != nil {
		// If we have holdHijackedConnection, we should notify
		// holdHijackedConnection we are going to exit and wait
		// to avoid the terminal are not restored.
		if attach {
			cancelFun()
			<-errCh
		}

		cmd.ReportError(err.Error(), false)
		return runStartContainerErr(err)
	}

	if (config.AttachStdin || config.AttachStdout || config.AttachStderr) && config.Tty && cli.isTerminalOut {
		if err := cli.monitorTtySize(createResponse.ID, false); err != nil {
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}

	if errCh != nil {
		if err := <-errCh; err != nil {
			logrus.Debugf("Error hijack: %s", err)
			return err
		}
	}

	// Detached mode: wait for the id to be displayed and return.
	if !config.AttachStdout && !config.AttachStderr {
		// Detached mode
		<-waitDisplayID
		return nil
	}

	var status int

	// Attached mode
	if *flAutoRemove {
		// Autoremove: wait for the container to finish, retrieve
		// the exit code and remove the container
		if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
			return runStartContainerErr(err)
		}
		if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
			return err
		}
	} else {
		// No Autoremove: Simply retrieve the exit code
		if !config.Tty {
			// In non-TTY mode, we can't detach, so we must wait for container exit
			if status, err = cli.client.ContainerWait(context.Background(), createResponse.ID); err != nil {
				return err
			}
		} else {
			// In TTY mode, there is a race: if the process dies too slowly, the state could
			// be updated after the getExitCode call and result in the wrong exit code being reported
			if _, status, err = getExitCode(cli, createResponse.ID); err != nil {
				return err
			}
		}
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}
`


var save_go_scope = godebug.EnteringNewFile(client_pkg_scope, save_go_contents)

func (cli *DockerCli) CmdSave(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdSave(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := save_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("save", []string{"IMAGE [IMAGE...]"}, Cli.DockerCommands["save"].Description+" (streamed to STDOUT by default)", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	outfile := cmd.String([]string{"o", "-output"}, "", "Write to a file, instead of STDOUT")
	scope.Declare("outfile", &outfile)
	godebug.Line(ctx, scope, 21)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 23)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 25)

	if *outfile == "" && cli.isTerminalOut {
		godebug.Line(ctx, scope, 26)
		return errors.New("Cowardly refusing to save to a terminal. Use the -o flag or redirect.")
	}
	godebug.Line(ctx, scope, 29)

	responseBody, err := cli.client.ImageSave(context.Background(), cmd.Args())
	scope.Declare("responseBody", &responseBody, "err", &err)
	godebug.Line(ctx, scope, 30)
	if err != nil {
		godebug.Line(ctx, scope, 31)
		return err
	}
	godebug.Line(ctx, scope, 33)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 33)
	godebug.Line(ctx, scope, 35)

	if *outfile == "" {
		godebug.Line(ctx, scope, 36)
		_, err := io.Copy(cli.out, responseBody)
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 37)
		return err
	}
	godebug.Line(ctx, scope, 40)

	return copyToFile(*outfile, responseBody)

}

var save_go_contents = `package client

import (
	"errors"
	"io"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdSave saves one or more images to a tar archive.
//
// The tar archive is written to STDOUT by default, or written to a file.
//
// Usage: docker save [OPTIONS] IMAGE [IMAGE...]
func (cli *DockerCli) CmdSave(args ...string) error {
	cmd := Cli.Subcmd("save", []string{"IMAGE [IMAGE...]"}, Cli.DockerCommands["save"].Description+" (streamed to STDOUT by default)", true)
	outfile := cmd.String([]string{"o", "-output"}, "", "Write to a file, instead of STDOUT")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	if *outfile == "" && cli.isTerminalOut {
		return errors.New("Cowardly refusing to save to a terminal. Use the -o flag or redirect.")
	}

	responseBody, err := cli.client.ImageSave(context.Background(), cmd.Args())
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if *outfile == "" {
		_, err := io.Copy(cli.out, responseBody)
		return err
	}

	return copyToFile(*outfile, responseBody)

}
`


var search_go_scope = godebug.EnteringNewFile(client_pkg_scope, search_go_contents)

func (cli *DockerCli) CmdSearch(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdSearch(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := search_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 24)
	cmd := Cli.Subcmd("search", []string{"TERM"}, Cli.DockerCommands["search"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 25)
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
	scope.Declare("noTrunc", &noTrunc)
	godebug.Line(ctx, scope, 26)
	automated := cmd.Bool([]string{"-automated"}, false, "Only show automated builds")
	scope.Declare("automated", &automated)
	godebug.Line(ctx, scope, 27)
	stars := cmd.Uint([]string{"s", "-stars"}, 0, "Only displays with at least x stars")
	scope.Declare("stars", &stars)
	godebug.Line(ctx, scope, 28)
	cmd.Require(flag.Exact, 1)
	godebug.Line(ctx, scope, 30)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 32)

	name := cmd.Arg(0)
	scope.Declare("name", &name)
	godebug.Line(ctx, scope, 33)
	v := url.Values{}
	scope.Declare("v", &v)
	godebug.Line(ctx, scope, 34)
	v.Set("term", name)
	godebug.Line(ctx, scope, 36)

	indexInfo, err := registry.ParseSearchIndexInfo(name)
	scope.Declare("indexInfo", &indexInfo, "err", &err)
	godebug.Line(ctx, scope, 37)
	if err != nil {
		godebug.Line(ctx, scope, 38)
		return err
	}
	godebug.Line(ctx, scope, 41)

	authConfig := cli.resolveAuthConfig(indexInfo)
	scope.Declare("authConfig", &authConfig)
	godebug.Line(ctx, scope, 42)
	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(indexInfo, "search")
	scope.Declare("requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 44)

	encodedAuth, err := encodeAuthToBase64(authConfig)
	scope.Declare("encodedAuth", &encodedAuth)
	godebug.Line(ctx, scope, 45)
	if err != nil {
		godebug.Line(ctx, scope, 46)
		return err
	}
	godebug.Line(ctx, scope, 49)

	options := types.ImageSearchOptions{
		Term:         name,
		RegistryAuth: encodedAuth,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 54)

	unorderedResults, err := cli.client.ImageSearch(context.Background(), options, requestPrivilege)
	scope.Declare("unorderedResults", &unorderedResults)
	godebug.Line(ctx, scope, 55)
	if err != nil {
		godebug.Line(ctx, scope, 56)
		return err
	}
	godebug.Line(ctx, scope, 59)

	results := searchResultsByStars(unorderedResults)
	scope.Declare("results", &results)
	godebug.Line(ctx, scope, 60)
	sort.Sort(results)
	godebug.Line(ctx, scope, 62)

	w := tabwriter.NewWriter(cli.out, 10, 1, 3, ' ', 0)
	scope.Declare("w", &w)
	godebug.Line(ctx, scope, 63)
	fmt.Fprintf(w, "NAME\tDESCRIPTION\tSTARS\tOFFICIAL\tAUTOMATED\n")
	{
		scope := scope.EnteringNewChildScope()
		for _, res := range results {
			godebug.Line(ctx, scope, 64)
			scope.Declare("res", &res)
			godebug.Line(ctx, scope, 65)
			if (*automated && !res.IsAutomated) || (int(*stars) > res.StarCount) {
				godebug.Line(ctx, scope, 66)
				continue
			}
			godebug.Line(ctx, scope, 68)
			desc := strings.Replace(res.Description, "\n", " ", -1)
			scope := scope.EnteringNewChildScope()
			scope.Declare("desc", &desc)
			godebug.Line(ctx, scope, 69)
			desc = strings.Replace(desc, "\r", " ", -1)
			godebug.Line(ctx, scope, 70)
			if !*noTrunc && len(desc) > 45 {
				godebug.Line(ctx, scope, 71)
				desc = stringutils.Truncate(desc, 42) + "..."
			}
			godebug.Line(ctx, scope, 73)
			fmt.Fprintf(w, "%s\t%s\t%d\t", res.Name, desc, res.StarCount)
			godebug.Line(ctx, scope, 74)
			if res.IsOfficial {
				godebug.Line(ctx, scope, 75)
				fmt.Fprint(w, "[OK]")

			}
			godebug.Line(ctx, scope, 78)
			fmt.Fprint(w, "\t")
			godebug.Line(ctx, scope, 79)
			if res.IsAutomated || res.IsTrusted {
				godebug.Line(ctx, scope, 80)
				fmt.Fprint(w, "[OK]")
			}
			godebug.Line(ctx, scope, 82)
			fmt.Fprint(w, "\n")
		}
		godebug.Line(ctx, scope, 64)
	}
	godebug.Line(ctx, scope, 84)
	w.Flush()
	godebug.Line(ctx, scope, 85)
	return nil
}

type searchResultsByStars []registrytypes.SearchResult

func (r searchResultsByStars) Len() int {
	var result1 int
	ctx, ok := godebug.EnterFunc(func() {
		result1 = r.Len()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := search_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r)
	godebug.Line(ctx, scope, 91)
	return len(r)
}
func (r searchResultsByStars) Swap(i, j int) {
	ctx, ok := godebug.EnterFunc(func() {
		r.Swap(i, j)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := search_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 92)
	r[i], r[j] = r[j], r[i]
}
func (r searchResultsByStars) Less(i, j int) bool {
	var result1 bool
	ctx, ok := godebug.EnterFunc(func() {
		result1 = r.Less(i, j)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := search_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 93)
	return r[j].StarCount < r[i].StarCount
}

var search_go_contents = `package client

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/stringutils"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/types"
	registrytypes "github.com/docker/engine-api/types/registry"
)

// CmdSearch searches the Docker Hub for images.
//
// Usage: docker search [OPTIONS] TERM
func (cli *DockerCli) CmdSearch(args ...string) error {
	cmd := Cli.Subcmd("search", []string{"TERM"}, Cli.DockerCommands["search"].Description, true)
	noTrunc := cmd.Bool([]string{"-no-trunc"}, false, "Don't truncate output")
	automated := cmd.Bool([]string{"-automated"}, false, "Only show automated builds")
	stars := cmd.Uint([]string{"s", "-stars"}, 0, "Only displays with at least x stars")
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	name := cmd.Arg(0)
	v := url.Values{}
	v.Set("term", name)

	indexInfo, err := registry.ParseSearchIndexInfo(name)
	if err != nil {
		return err
	}

	authConfig := cli.resolveAuthConfig(indexInfo)
	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(indexInfo, "search")

	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	options := types.ImageSearchOptions{
		Term:         name,
		RegistryAuth: encodedAuth,
	}

	unorderedResults, err := cli.client.ImageSearch(context.Background(), options, requestPrivilege)
	if err != nil {
		return err
	}

	results := searchResultsByStars(unorderedResults)
	sort.Sort(results)

	w := tabwriter.NewWriter(cli.out, 10, 1, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tDESCRIPTION\tSTARS\tOFFICIAL\tAUTOMATED\n")
	for _, res := range results {
		if (*automated && !res.IsAutomated) || (int(*stars) > res.StarCount) {
			continue
		}
		desc := strings.Replace(res.Description, "\n", " ", -1)
		desc = strings.Replace(desc, "\r", " ", -1)
		if !*noTrunc && len(desc) > 45 {
			desc = stringutils.Truncate(desc, 42) + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t", res.Name, desc, res.StarCount)
		if res.IsOfficial {
			fmt.Fprint(w, "[OK]")

		}
		fmt.Fprint(w, "\t")
		if res.IsAutomated || res.IsTrusted {
			fmt.Fprint(w, "[OK]")
		}
		fmt.Fprint(w, "\n")
	}
	w.Flush()
	return nil
}

// SearchResultsByStars sorts search results in descending order by number of stars.
type searchResultsByStars []registrytypes.SearchResult

func (r searchResultsByStars) Len() int           { return len(r) }
func (r searchResultsByStars) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r searchResultsByStars) Less(i, j int) bool { return r[j].StarCount < r[i].StarCount }
`


var start_go_scope = godebug.EnteringNewFile(client_pkg_scope, start_go_contents)

func (cli *DockerCli) forwardAllSignals(cid string) chan os.Signal {
	var result1 chan os.Signal
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.forwardAllSignals(cid)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := start_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "cid", &cid)
	godebug.Line(_ctx, scope, 21)
	sigc := make(chan os.Signal, 128)
	scope.Declare("sigc", &sigc)
	godebug.Line(_ctx, scope, 22)
	signal.CatchAll(sigc)
	godebug.Line(_ctx, scope, 23)
	go func() {
		fn := func(_ctx *godebug.Context) {
			{
				scope := scope.EnteringNewChildScope()
				for s := range sigc {
					godebug.Line(_ctx, scope, 24)
					scope.Declare("s", &s)
					godebug.Line(_ctx, scope, 25)
					if s == signal.SIGCHLD || s == signal.SIGPIPE {
						godebug.Line(_ctx, scope, 26)
						continue
					}
					godebug.Line(_ctx, scope, 28)
					var sig string
					scope := scope.EnteringNewChildScope()
					scope.Declare("sig", &sig)
					{
						scope := scope.EnteringNewChildScope()
						for sigStr, sigN := range signal.SignalMap {
							godebug.Line(_ctx, scope, 29)
							scope.Declare("sigStr", &sigStr, "sigN", &sigN)
							godebug.Line(_ctx, scope, 30)
							if sigN == s {
								godebug.Line(_ctx, scope, 31)
								sig = sigStr
								godebug.Line(_ctx, scope, 32)
								break
							}
						}
						godebug.Line(_ctx, scope, 29)
					}
					godebug.Line(_ctx, scope, 35)
					if sig == "" {
						godebug.Line(_ctx, scope, 36)
						fmt.Fprintf(cli.err, "Unsupported signal: %v. Discarding.\n", s)
						godebug.Line(_ctx, scope, 37)
						continue
					}
					godebug.Line(_ctx, scope, 40)
					if err := cli.client.ContainerKill(context.Background(), cid, sig); err != nil {
						scope := scope.EnteringNewChildScope()
						scope.Declare("err", &err)
						godebug.Line(_ctx, scope, 41)
						logrus.Debugf("Error sending signal: %s", err)
					}
				}
				godebug.Line(_ctx, scope, 24)
			}
		}
		if _ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(_ctx)
			fn(_ctx)
		}
	}()
	godebug.Line(_ctx, scope, 45)
	return sigc
}

func (cli *DockerCli) CmdStart(args ...string) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdStart(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := start_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(_ctx, scope, 52)
	logrus.Debugf("Executing api/client/start.go : CmdStart(%s)", args)
	godebug.Line(_ctx, scope, 53)
	cmd := Cli.Subcmd("start", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["start"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(_ctx, scope, 54)
	attach := cmd.Bool([]string{"a", "-attach"}, false, "Attach STDOUT/STDERR and forward signals")
	scope.Declare("attach", &attach)
	godebug.Line(_ctx, scope, 55)
	openStdin := cmd.Bool([]string{"i", "-interactive"}, false, "Attach container's STDIN")
	scope.Declare("openStdin", &openStdin)
	godebug.Line(_ctx, scope, 56)
	detachKeys := cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
	scope.Declare("detachKeys", &detachKeys)
	godebug.Line(_ctx, scope, 57)
	cmd.Require(flag.Min, 1)
	godebug.Line(_ctx, scope, 59)

	cmd.ParseFlags(args, true)
	godebug.Line(_ctx, scope, 61)

	if *attach || *openStdin {
		godebug.Line(_ctx, scope, 64)

		if cmd.NArg() > 1 {
			godebug.Line(_ctx, scope, 65)
			return fmt.Errorf("You cannot start and attach multiple containers at once.")
		}
		godebug.Line(_ctx, scope, 69)

		containerID := cmd.Arg(0)
		scope := scope.EnteringNewChildScope()
		scope.Declare("containerID", &containerID)
		godebug.Line(_ctx, scope, 70)
		c, err := cli.client.ContainerInspect(context.Background(), containerID)
		scope.Declare("c", &c, "err", &err)
		godebug.Line(_ctx, scope, 71)
		if err != nil {
			godebug.Line(_ctx, scope, 72)
			return err
		}
		godebug.Line(_ctx, scope, 75)

		if !c.Config.Tty {
			godebug.Line(_ctx, scope, 76)
			sigc := cli.forwardAllSignals(containerID)
			scope := scope.EnteringNewChildScope()
			scope.Declare("sigc", &sigc)
			godebug.Line(_ctx, scope, 77)
			defer signal.StopCatch(sigc)
			defer godebug.Defer(_ctx, scope, 77)
		}
		godebug.Line(_ctx, scope, 80)

		if *detachKeys != "" {
			godebug.Line(_ctx, scope, 81)
			cli.configFile.DetachKeys = *detachKeys
		}
		godebug.Line(_ctx, scope, 84)

		options := types.ContainerAttachOptions{
			ContainerID: containerID,
			Stream:      true,
			Stdin:       *openStdin && c.Config.OpenStdin,
			Stdout:      true,
			Stderr:      true,
			DetachKeys:  cli.configFile.DetachKeys,
		}
		scope.Declare("options", &options)
		godebug.Line(_ctx, scope, 93)

		var in io.ReadCloser
		scope.Declare("in", &in)
		godebug.Line(_ctx, scope, 95)

		if options.Stdin {
			godebug.Line(_ctx, scope, 96)
			in = cli.in
		}
		godebug.Line(_ctx, scope, 99)

		resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
		scope.Declare("resp", &resp, "errAttach", &errAttach)
		godebug.Line(_ctx, scope, 100)
		if errAttach != nil && errAttach != httputil.ErrPersistEOF {
			godebug.Line(_ctx, scope, 104)

			return errAttach
		}
		godebug.Line(_ctx, scope, 106)
		defer resp.Close()
		defer godebug.Defer(_ctx, scope, 106)
		godebug.Line(_ctx, scope, 107)
		ctx, cancelFun := context.WithCancel(context.Background())
		scope.Declare("ctx", &ctx, "cancelFun", &cancelFun)
		godebug.Line(_ctx, scope, 108)
		cErr := promise.Go(func() error {
			var result1 error
			fn := func(_ctx *godebug.Context) {
				result1 = func() error {
					godebug.Line(_ctx, scope, 109)
					errHijack := cli.holdHijackedConnection(ctx, c.Config.Tty, in, cli.out, cli.err, resp)
					scope := scope.EnteringNewChildScope()
					scope.Declare("errHijack", &errHijack)
					godebug.Line(_ctx, scope, 110)
					if errHijack == nil {
						godebug.Line(_ctx, scope, 111)
						return errAttach
					}
					godebug.Line(_ctx, scope, 113)
					return errHijack
				}()
			}
			if _ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(_ctx)
				fn(_ctx)
			}
			return result1
		},
		)
		scope.Declare("cErr", &cErr)
		godebug.Line(_ctx, scope, 117)

		if err := cli.client.ContainerStart(context.Background(), containerID); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(_ctx, scope, 118)
			cancelFun()
			godebug.Line(_ctx, scope, 119)
			<-cErr
			godebug.Line(_ctx, scope, 120)
			return err
		}
		godebug.Line(_ctx, scope, 124)

		if c.Config.Tty && cli.isTerminalOut {
			godebug.Line(_ctx, scope, 125)
			if err := cli.monitorTtySize(containerID, false); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(_ctx, scope, 126)
				fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
			}
		}
		godebug.Line(_ctx, scope, 129)
		if attchErr := <-cErr; attchErr != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("attchErr", &attchErr)
			godebug.Line(_ctx, scope, 130)
			return attchErr
		}
		godebug.Line(_ctx, scope, 132)
		_, status, err := getExitCode(cli, containerID)
		scope.Declare("status", &status)
		godebug.Line(_ctx, scope, 133)
		if err != nil {
			godebug.Line(_ctx, scope, 134)
			return err
		}
		godebug.Line(_ctx, scope, 136)
		if status != 0 {
			godebug.Line(_ctx, scope, 137)
			return Cli.StatusError{StatusCode: status}
		}
	} else {
		godebug.Line(_ctx, scope, 139)
		godebug.Line(_ctx, scope, 142)

		return cli.startContainersWithoutAttachments(cmd.Args())
	}
	godebug.Line(_ctx, scope, 145)

	return nil
}

func (cli *DockerCli) startContainersWithoutAttachments(containerIDs []string) error {
	var result1 error
	_ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.startContainersWithoutAttachments(containerIDs)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(_ctx)
	scope := start_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "containerIDs", &containerIDs)
	godebug.Line(_ctx, scope, 149)
	var failedContainers []string
	scope.Declare("failedContainers", &failedContainers)
	{
		scope := scope.EnteringNewChildScope()
		for _, containerID := range containerIDs {
			godebug.Line(_ctx, scope, 150)
			scope.Declare("containerID", &containerID)
			godebug.Line(_ctx, scope, 151)
			if err := cli.client.ContainerStart(context.Background(), containerID); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(_ctx, scope, 152)
				fmt.Fprintf(cli.err, "%s\n", err)
				godebug.Line(_ctx, scope, 153)
				failedContainers = append(failedContainers, containerID)
			} else {
				godebug.Line(_ctx, scope, 154)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(_ctx, scope, 155)

				fmt.Fprintf(cli.out, "%s\n", containerID)
			}
		}
		godebug.Line(_ctx, scope, 150)
	}
	godebug.Line(_ctx, scope, 159)

	if len(failedContainers) > 0 {
		godebug.Line(_ctx, scope, 160)
		return fmt.Errorf("Error: failed to start containers: %v", strings.Join(failedContainers, ", "))
	}
	godebug.Line(_ctx, scope, 162)
	return nil
}

var start_go_contents = `package client

import (
	"fmt"
	"io"
	"net/http/httputil"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/engine-api/types"
)

func (cli *DockerCli) forwardAllSignals(cid string) chan os.Signal {
	sigc := make(chan os.Signal, 128)
	signal.CatchAll(sigc)
	go func() {
		for s := range sigc {
			if s == signal.SIGCHLD || s == signal.SIGPIPE {
				continue
			}
			var sig string
			for sigStr, sigN := range signal.SignalMap {
				if sigN == s {
					sig = sigStr
					break
				}
			}
			if sig == "" {
				fmt.Fprintf(cli.err, "Unsupported signal: %v. Discarding.\n", s)
				continue
			}

			if err := cli.client.ContainerKill(context.Background(), cid, sig); err != nil {
				logrus.Debugf("Error sending signal: %s", err)
			}
		}
	}()
	return sigc
}

// CmdStart starts one or more containers.
//
// Usage: docker start [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdStart(args ...string) error {
	logrus.Debugf("Executing api/client/start.go : CmdStart(%s)", args)
	cmd := Cli.Subcmd("start", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["start"].Description, true)
	attach := cmd.Bool([]string{"a", "-attach"}, false, "Attach STDOUT/STDERR and forward signals")
	openStdin := cmd.Bool([]string{"i", "-interactive"}, false, "Attach container's STDIN")
	detachKeys := cmd.String([]string{"-detach-keys"}, "", "Override the key sequence for detaching a container")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	if *attach || *openStdin {
		// We're going to attach to a container.
		// 1. Ensure we only have one container.
		if cmd.NArg() > 1 {
			return fmt.Errorf("You cannot start and attach multiple containers at once.")
		}

		// 2. Attach to the container.
		containerID := cmd.Arg(0)
		c, err := cli.client.ContainerInspect(context.Background(), containerID)
		if err != nil {
			return err
		}

		if !c.Config.Tty {
			sigc := cli.forwardAllSignals(containerID)
			defer signal.StopCatch(sigc)
		}

		if *detachKeys != "" {
			cli.configFile.DetachKeys = *detachKeys
		}

		options := types.ContainerAttachOptions{
			ContainerID: containerID,
			Stream:      true,
			Stdin:       *openStdin && c.Config.OpenStdin,
			Stdout:      true,
			Stderr:      true,
			DetachKeys:  cli.configFile.DetachKeys,
		}

		var in io.ReadCloser

		if options.Stdin {
			in = cli.in
		}

		resp, errAttach := cli.client.ContainerAttach(context.Background(), options)
		if errAttach != nil && errAttach != httputil.ErrPersistEOF {
			// ContainerAttach return an ErrPersistEOF (connection closed)
			// means server met an error and put it in Hijacked connection
			// keep the error and read detailed error message from hijacked connection
			return errAttach
		}
		defer resp.Close()
		ctx, cancelFun := context.WithCancel(context.Background())
		cErr := promise.Go(func() error {
			errHijack := cli.holdHijackedConnection(ctx, c.Config.Tty, in, cli.out, cli.err, resp)
			if errHijack == nil {
				return errAttach
			}
			return errHijack
		})

		// 3. Start the container.
		if err := cli.client.ContainerStart(context.Background(), containerID); err != nil {
			cancelFun()
			<-cErr
			return err
		}

		// 4. Wait for attachment to break.
		if c.Config.Tty && cli.isTerminalOut {
			if err := cli.monitorTtySize(containerID, false); err != nil {
				fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
			}
		}
		if attchErr := <-cErr; attchErr != nil {
			return attchErr
		}
		_, status, err := getExitCode(cli, containerID)
		if err != nil {
			return err
		}
		if status != 0 {
			return Cli.StatusError{StatusCode: status}
		}
	} else {
		// We're not going to attach to anything.
		// Start as many containers as we want.
		return cli.startContainersWithoutAttachments(cmd.Args())
	}

	return nil
}

func (cli *DockerCli) startContainersWithoutAttachments(containerIDs []string) error {
	var failedContainers []string
	for _, containerID := range containerIDs {
		if err := cli.client.ContainerStart(context.Background(), containerID); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			failedContainers = append(failedContainers, containerID)
		} else {
			fmt.Fprintf(cli.out, "%s\n", containerID)
		}
	}

	if len(failedContainers) > 0 {
		return fmt.Errorf("Error: failed to start containers: %v", strings.Join(failedContainers, ", "))
	}
	return nil
}
`


var stats_go_scope = godebug.EnteringNewFile(client_pkg_scope, stats_go_contents)

func (cli *DockerCli) CmdStats(args ...string) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.CmdStats(args...)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 25)
	cmd := Cli.Subcmd("stats", []string{"[CONTAINER...]"}, Cli.DockerCommands["stats"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 26)
	all := cmd.Bool([]string{"a", "-all"}, false, "Show all containers (default shows just running)")
	scope.Declare("all", &all)
	godebug.Line(ctx, scope, 27)
	noStream := cmd.Bool([]string{"-no-stream"}, false, "Disable streaming stats and only pull the first result")
	scope.Declare("noStream", &noStream)
	godebug.Line(ctx, scope, 29)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 31)

	names := cmd.Args()
	scope.Declare("names", &names)
	godebug.Line(ctx, scope, 32)
	showAll := len(names) == 0
	scope.Declare("showAll", &showAll)
	godebug.Line(ctx, scope, 33)
	closeChan := make(chan error)
	scope.Declare("closeChan", &closeChan)
	godebug.Line(ctx, scope, 37)

	monitorContainerEvents := func(started chan<- struct{}, c chan events.Message) {
		fn := func(ctx *godebug.Context) {
			scope := scope.EnteringNewChildScope()
			scope.Declare("started", &started, "c", &c)
			godebug.Line(ctx, scope, 38)
			f := filters.NewArgs()
			scope.Declare("f", &f)
			godebug.Line(ctx, scope, 39)
			f.Add("type", "container")
			godebug.Line(ctx, scope, 40)
			options := types.EventsOptions{Filters: f}
			scope.Declare("options", &options)
			godebug.Line(ctx, scope, 43)
			resBody, err := cli.client.Events(context.Background(), options)
			scope.Declare("resBody", &resBody, "err", &err)
			godebug.Line(ctx, scope, 46)
			close(started)
			godebug.Line(ctx, scope, 47)
			if err != nil {
				godebug.Line(ctx, scope, 48)
				closeChan <- err
				godebug.Line(ctx, scope, 49)
				return
			}
			godebug.Line(ctx, scope, 51)
			defer resBody.Close()
			defer godebug.Defer(ctx, scope, 51)
			godebug.Line(ctx, scope, 53)
			decodeEvents(resBody, func(event events.Message, err error) error {
				var result1 error
				fn := func(ctx *godebug.Context) {
					result1 = func() error {
						scope := scope.EnteringNewChildScope()
						scope.Declare("event", &event, "err", &err)
						godebug.Line(ctx, scope, 54)
						if err != nil {
							godebug.Line(ctx, scope, 55)
							closeChan <- err
							godebug.Line(ctx, scope, 56)
							return nil
						}
						godebug.Line(ctx, scope, 58)
						c <- event
						godebug.Line(ctx, scope, 59)
						return nil
					}()
				}
				if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
					defer godebug.ExitFunc(ctx)
					fn(ctx)
				}
				return result1
			})
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	scope.Declare("monitorContainerEvents", &monitorContainerEvents)
	godebug.Line(ctx, scope, 64)

	waitFirst := &sync.WaitGroup{}
	scope.Declare("waitFirst", &waitFirst)
	godebug.Line(ctx, scope, 66)

	cStats := stats{}
	scope.Declare("cStats", &cStats)
	godebug.Line(ctx, scope, 69)

	getContainerList := func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 70)
			options := types.ContainerListOptions{All: *all}
			scope := scope.EnteringNewChildScope()
			scope.Declare("options", &options)
			godebug.Line(ctx, scope, 73)
			cs, err := cli.client.ContainerList(context.Background(), options)
			scope.Declare("cs", &cs, "err", &err)
			godebug.Line(ctx, scope, 74)
			if err != nil {
				godebug.Line(ctx, scope, 75)
				closeChan <- err
			}
			{
				scope := scope.EnteringNewChildScope()
				for _, container := range cs {
					godebug.Line(ctx, scope, 77)
					scope.Declare("container", &container)
					godebug.Line(ctx, scope, 78)
					s := &containerStats{Name: container.ID[:12]}
					scope := scope.EnteringNewChildScope()
					scope.Declare("s", &s)
					godebug.Line(ctx, scope, 79)
					if cStats.add(s) {
						godebug.Line(ctx, scope, 80)
						waitFirst.Add(1)
						godebug.Line(ctx, scope, 81)
						go s.Collect(cli.client, !*noStream, waitFirst)
					}
				}
				godebug.Line(ctx, scope, 77)
			}
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	scope.Declare("getContainerList", &getContainerList)
	godebug.Line(ctx, scope, 86)

	if showAll {
		godebug.Line(ctx, scope, 91)

		started := make(chan struct{})
		scope := scope.EnteringNewChildScope()
		scope.Declare("started", &started)
		godebug.Line(ctx, scope, 92)
		eh := eventHandler{handlers: make(map[string]func(events.Message))}
		scope.Declare("eh", &eh)
		godebug.Line(ctx, scope, 93)
		eh.Handle("create", func(e events.Message) {
			fn := func(ctx *godebug.Context) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("e", &e)
				godebug.Line(ctx, scope, 94)
				if *all {
					godebug.Line(ctx, scope, 95)
					s := &containerStats{Name: e.ID[:12]}
					scope := scope.EnteringNewChildScope()
					scope.Declare("s", &s)
					godebug.Line(ctx, scope, 96)
					if cStats.add(s) {
						godebug.Line(ctx, scope, 97)
						waitFirst.Add(1)
						godebug.Line(ctx, scope, 98)
						go s.Collect(cli.client, !*noStream, waitFirst)
					}
				}
			}
			if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
				defer godebug.ExitFunc(ctx)
				fn(ctx)
			}
		},
		)
		godebug.Line(ctx, scope, 103)

		eh.Handle("start", func(e events.Message) {
			fn := func(ctx *godebug.Context) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("e", &e)
				godebug.Line(ctx, scope, 104)
				s := &containerStats{Name: e.ID[:12]}
				scope.Declare("s", &s)
				godebug.Line(ctx, scope, 105)
				if cStats.add(s) {
					godebug.Line(ctx, scope, 106)
					waitFirst.Add(1)
					godebug.Line(ctx, scope, 107)
					go s.Collect(cli.client, !*noStream, waitFirst)
				}
			}
			if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
				defer godebug.ExitFunc(ctx)
				fn(ctx)
			}
		},
		)
		godebug.Line(ctx, scope, 111)

		eh.Handle("die", func(e events.Message) {
			fn := func(ctx *godebug.Context) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("e", &e)
				godebug.Line(ctx, scope, 112)
				if !*all {
					godebug.Line(ctx, scope, 113)
					cStats.remove(e.ID[:12])
				}
			}
			if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
				defer godebug.ExitFunc(ctx)
				fn(ctx)
			}
		},
		)
		godebug.Line(ctx, scope, 117)

		eventChan := make(chan events.Message)
		scope.Declare("eventChan", &eventChan)
		godebug.Line(ctx, scope, 118)
		go eh.Watch(eventChan)
		godebug.Line(ctx, scope, 119)
		go monitorContainerEvents(started, eventChan)
		godebug.Line(ctx, scope, 120)
		defer close(eventChan)
		defer godebug.Defer(ctx, scope, 120)
		godebug.Line(ctx, scope, 121)
		<-started
		godebug.Line(ctx, scope, 125)

		getContainerList()
	} else {
		godebug.Line(ctx, scope, 126)
		{
			scope := scope.EnteringNewChildScope()

			for _, name := range names {
				godebug.Line(ctx, scope, 129)
				scope.Declare("name", &name)
				godebug.Line(ctx, scope, 130)
				s := &containerStats{Name: name}
				scope := scope.EnteringNewChildScope()
				scope.Declare("s", &s)
				godebug.Line(ctx, scope, 131)
				if cStats.add(s) {
					godebug.Line(ctx, scope, 132)
					waitFirst.Add(1)
					godebug.Line(ctx, scope, 133)
					go s.Collect(cli.client, !*noStream, waitFirst)
				}
			}
			godebug.Line(ctx, scope, 129)
		}
		godebug.Line(ctx, scope, 138)

		close(closeChan)
		godebug.Line(ctx, scope, 142)

		time.Sleep(1500 * time.Millisecond)
		godebug.Line(ctx, scope, 143)
		var errs []string
		scope := scope.EnteringNewChildScope()
		scope.Declare("errs", &errs)
		godebug.Line(ctx, scope, 144)
		cStats.mu.Lock()
		{
			scope := scope.EnteringNewChildScope()
			for _, c := range cStats.cs {
				godebug.Line(ctx, scope, 145)
				scope.Declare("c", &c)
				godebug.Line(ctx, scope, 146)
				c.mu.Lock()
				godebug.Line(ctx, scope, 147)
				if c.err != nil {
					godebug.Line(ctx, scope, 148)
					errs = append(errs, fmt.Sprintf("%s: %v", c.Name, c.err))
				}
				godebug.Line(ctx, scope, 150)
				c.mu.Unlock()
			}
			godebug.Line(ctx, scope, 145)
		}
		godebug.Line(ctx, scope, 152)
		cStats.mu.Unlock()
		godebug.Line(ctx, scope, 153)
		if len(errs) > 0 {
			godebug.Line(ctx, scope, 154)
			return fmt.Errorf("%s", strings.Join(errs, ", "))
		}
	}
	godebug.Line(ctx, scope, 159)

	waitFirst.Wait()
	godebug.Line(ctx, scope, 161)

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	scope.Declare("w", &w)
	godebug.Line(ctx, scope, 162)
	printHeader := func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 163)
			if !*noStream {
				godebug.Line(ctx, scope, 164)
				fmt.Fprint(cli.out, "\033[2J")
				godebug.Line(ctx, scope, 165)
				fmt.Fprint(cli.out, "\033[H")
			}
			godebug.Line(ctx, scope, 167)
			io.WriteString(w, "CONTAINER\tCPU %\tMEM USAGE / LIMIT\tMEM %\tNET I/O\tBLOCK I/O\tPIDS\n")
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	scope.Declare("printHeader", &printHeader)
	godebug.Line(ctx, scope, 170)

	for range time.Tick(500 * time.Millisecond) {
		godebug.Line(ctx, scope, 171)
		printHeader()
		godebug.Line(ctx, scope, 172)
		toRemove := []int{}
		scope := scope.EnteringNewChildScope()
		scope.Declare("toRemove", &toRemove)
		godebug.Line(ctx, scope, 173)
		cStats.mu.Lock()
		{
			scope := scope.EnteringNewChildScope()
			for i, s := range cStats.cs {
				godebug.Line(ctx, scope, 174)
				scope.Declare("i", &i, "s", &s)
				godebug.Line(ctx, scope, 175)
				if err := s.Display(w); err != nil && !*noStream {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(ctx, scope, 176)
					toRemove = append(toRemove, i)
				}
			}
			godebug.Line(ctx, scope, 174)
		}
		{
			scope := scope.EnteringNewChildScope()
			for j := len(toRemove) - 1; j >= 0; j-- {
				godebug.Line(ctx, scope, 179)
				scope.Declare("j", &j)
				godebug.Line(ctx, scope, 180)
				i := toRemove[j]
				scope := scope.EnteringNewChildScope()
				scope.Declare("i", &i)
				godebug.Line(ctx, scope, 181)
				cStats.cs = append(cStats.cs[:i], cStats.cs[i+1:]...)
			}
			godebug.Line(ctx, scope, 179)
		}
		godebug.Line(ctx, scope, 183)
		if len(cStats.cs) == 0 && !showAll {
			godebug.Line(ctx, scope, 184)
			return nil
		}
		godebug.Line(ctx, scope, 186)
		cStats.mu.Unlock()
		godebug.Line(ctx, scope, 187)
		w.Flush()
		godebug.Line(ctx, scope, 188)
		if *noStream {
			godebug.Line(ctx, scope, 189)
			break
		}
		godebug.Select(ctx, scope, 191)
		select {
		case <-godebug.Comm(ctx, scope, 192):
			panic("impossible")
		case err, ok := <-closeChan:
			godebug.Line(ctx, scope, 192)
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err, "ok", &ok)
			godebug.Line(ctx, scope, 193)
			if ok {
				if err != nil {

					if err == io.ErrUnexpectedEOF {
						return nil
					}
					return err
				}
			}
		default:
			godebug.Line(ctx, scope, 203)
		case <-godebug.EndSelect(ctx, scope):
			panic("impossible")

		}
		godebug.Line(ctx, scope, 170)
	}
	godebug.Line(ctx, scope, 207)
	return nil
}

var stats_go_contents = `package client

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
)

// CmdStats displays a live stream of resource usage statistics for one or more containers.
//
// This shows real-time information on CPU usage, memory usage, and network I/O.
//
// Usage: docker stats [OPTIONS] [CONTAINER...]
func (cli *DockerCli) CmdStats(args ...string) error {
	cmd := Cli.Subcmd("stats", []string{"[CONTAINER...]"}, Cli.DockerCommands["stats"].Description, true)
	all := cmd.Bool([]string{"a", "-all"}, false, "Show all containers (default shows just running)")
	noStream := cmd.Bool([]string{"-no-stream"}, false, "Disable streaming stats and only pull the first result")

	cmd.ParseFlags(args, true)

	names := cmd.Args()
	showAll := len(names) == 0
	closeChan := make(chan error)

	// monitorContainerEvents watches for container creation and removal (only
	// used when calling ` + "`" + `docker stats` + "`" + ` without arguments).
	monitorContainerEvents := func(started chan<- struct{}, c chan events.Message) {
		f := filters.NewArgs()
		f.Add("type", "container")
		options := types.EventsOptions{
			Filters: f,
		}
		resBody, err := cli.client.Events(context.Background(), options)
		// Whether we successfully subscribed to events or not, we can now
		// unblock the main goroutine.
		close(started)
		if err != nil {
			closeChan <- err
			return
		}
		defer resBody.Close()

		decodeEvents(resBody, func(event events.Message, err error) error {
			if err != nil {
				closeChan <- err
				return nil
			}
			c <- event
			return nil
		})
	}

	// waitFirst is a WaitGroup to wait first stat data's reach for each container
	waitFirst := &sync.WaitGroup{}

	cStats := stats{}
	// getContainerList simulates creation event for all previously existing
	// containers (only used when calling ` + "`" + `docker stats` + "`" + ` without arguments).
	getContainerList := func() {
		options := types.ContainerListOptions{
			All: *all,
		}
		cs, err := cli.client.ContainerList(context.Background(), options)
		if err != nil {
			closeChan <- err
		}
		for _, container := range cs {
			s := &containerStats{Name: container.ID[:12]}
			if cStats.add(s) {
				waitFirst.Add(1)
				go s.Collect(cli.client, !*noStream, waitFirst)
			}
		}
	}

	if showAll {
		// If no names were specified, start a long running goroutine which
		// monitors container events. We make sure we're subscribed before
		// retrieving the list of running containers to avoid a race where we
		// would "miss" a creation.
		started := make(chan struct{})
		eh := eventHandler{handlers: make(map[string]func(events.Message))}
		eh.Handle("create", func(e events.Message) {
			if *all {
				s := &containerStats{Name: e.ID[:12]}
				if cStats.add(s) {
					waitFirst.Add(1)
					go s.Collect(cli.client, !*noStream, waitFirst)
				}
			}
		})

		eh.Handle("start", func(e events.Message) {
			s := &containerStats{Name: e.ID[:12]}
			if cStats.add(s) {
				waitFirst.Add(1)
				go s.Collect(cli.client, !*noStream, waitFirst)
			}
		})

		eh.Handle("die", func(e events.Message) {
			if !*all {
				cStats.remove(e.ID[:12])
			}
		})

		eventChan := make(chan events.Message)
		go eh.Watch(eventChan)
		go monitorContainerEvents(started, eventChan)
		defer close(eventChan)
		<-started

		// Start a short-lived goroutine to retrieve the initial list of
		// containers.
		getContainerList()
	} else {
		// Artificially send creation events for the containers we were asked to
		// monitor (same code path than we use when monitoring all containers).
		for _, name := range names {
			s := &containerStats{Name: name}
			if cStats.add(s) {
				waitFirst.Add(1)
				go s.Collect(cli.client, !*noStream, waitFirst)
			}
		}

		// We don't expect any asynchronous errors: closeChan can be closed.
		close(closeChan)

		// Do a quick pause to detect any error with the provided list of
		// container names.
		time.Sleep(1500 * time.Millisecond)
		var errs []string
		cStats.mu.Lock()
		for _, c := range cStats.cs {
			c.mu.Lock()
			if c.err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", c.Name, c.err))
			}
			c.mu.Unlock()
		}
		cStats.mu.Unlock()
		if len(errs) > 0 {
			return fmt.Errorf("%s", strings.Join(errs, ", "))
		}
	}

	// before print to screen, make sure each container get at least one valid stat data
	waitFirst.Wait()

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	printHeader := func() {
		if !*noStream {
			fmt.Fprint(cli.out, "\033[2J")
			fmt.Fprint(cli.out, "\033[H")
		}
		io.WriteString(w, "CONTAINER\tCPU %\tMEM USAGE / LIMIT\tMEM %\tNET I/O\tBLOCK I/O\tPIDS\n")
	}

	for range time.Tick(500 * time.Millisecond) {
		printHeader()
		toRemove := []int{}
		cStats.mu.Lock()
		for i, s := range cStats.cs {
			if err := s.Display(w); err != nil && !*noStream {
				toRemove = append(toRemove, i)
			}
		}
		for j := len(toRemove) - 1; j >= 0; j-- {
			i := toRemove[j]
			cStats.cs = append(cStats.cs[:i], cStats.cs[i+1:]...)
		}
		if len(cStats.cs) == 0 && !showAll {
			return nil
		}
		cStats.mu.Unlock()
		w.Flush()
		if *noStream {
			break
		}
		select {
		case err, ok := <-closeChan:
			if ok {
				if err != nil {
					// this is suppressing "unexpected EOF" in the cli when the
					// daemon restarts so it shutdowns cleanly
					if err == io.ErrUnexpectedEOF {
						return nil
					}
					return err
				}
			}
		default:
			// just skip
		}
	}
	return nil
}
`


var stats_helpers_go_scope = godebug.EnteringNewFile(client_pkg_scope, stats_helpers_go_contents)

type containerStats struct {
	Name             string
	CPUPercentage    float64
	Memory           float64
	MemoryLimit      float64
	MemoryPercentage float64
	NetworkRx        float64
	NetworkTx        float64
	BlockRead        float64
	BlockWrite       float64
	PidsCurrent      uint64
	mu               sync.RWMutex
	err              error
}

type stats struct {
	mu sync.Mutex
	cs []*containerStats
}

func (s *stats) add(cs *containerStats) bool {
	var result1 bool
	ctx, ok := godebug.EnterFunc(func() {
		result1 = s.add(cs)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("s", &s, "cs", &cs)
	godebug.Line(ctx, scope, 38)
	s.mu.Lock()
	godebug.Line(ctx, scope, 39)
	defer s.mu.Unlock()
	defer godebug.Defer(ctx, scope, 39)
	godebug.Line(ctx, scope, 40)
	if _, exists := s.isKnownContainer(cs.Name); !exists {
		scope := scope.EnteringNewChildScope()
		scope.Declare("exists", &exists)
		godebug.Line(ctx, scope, 41)
		s.cs = append(s.cs, cs)
		godebug.Line(ctx, scope, 42)
		return true
	}
	godebug.Line(ctx, scope, 44)
	return false
}

func (s *stats) remove(id string) {
	ctx, ok := godebug.EnterFunc(func() {
		s.remove(id)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("s", &s, "id", &id)
	godebug.Line(ctx, scope, 48)
	s.mu.Lock()
	godebug.Line(ctx, scope, 49)
	if i, exists := s.isKnownContainer(id); exists {
		scope := scope.EnteringNewChildScope()
		scope.Declare("i", &i, "exists", &exists)
		godebug.Line(ctx, scope, 50)
		s.cs = append(s.cs[:i], s.cs[i+1:]...)
	}
	godebug.Line(ctx, scope, 52)
	s.mu.Unlock()
}

func (s *stats) isKnownContainer(cid string) (int, bool) {
	var result1 int
	var result2 bool
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = s.isKnownContainer(cid)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("s", &s, "cid", &cid)
	{
		scope := scope.EnteringNewChildScope()
		for i, c := range s.cs {
			godebug.Line(ctx, scope, 56)
			scope.Declare("i", &i, "c", &c)
			godebug.Line(ctx, scope, 57)
			if c.Name == cid {
				godebug.Line(ctx, scope, 58)
				return i, true
			}
		}
		godebug.Line(ctx, scope, 56)
	}
	godebug.Line(ctx, scope, 61)
	return -1, false
}

func (s *containerStats) Collect(cli apiclient.APIClient, streamStats bool, waitFirst *sync.WaitGroup) {
	ctx, ok := godebug.EnterFunc(func() {
		s.Collect(cli, streamStats, waitFirst)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("s", &s, "cli", &cli, "streamStats", &streamStats, "waitFirst", &waitFirst)
	godebug.Line(ctx, scope, 65)
	var (
		getFirst       bool
		previousCPU    uint64
		previousSystem uint64
		u              = make(chan error, 1)
	)
	scope.Declare("getFirst", &getFirst, "previousCPU", &previousCPU, "previousSystem", &previousSystem, "u", &u)
	godebug.Line(ctx, scope, 72)

	defer func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 74)
			if !getFirst {
				godebug.Line(ctx, scope, 75)
				getFirst = true
				godebug.Line(ctx, scope, 76)
				waitFirst.Done()
			}
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}()
	defer godebug.Defer(ctx, scope, 72)
	godebug.Line(ctx, scope, 80)

	responseBody, err := cli.ContainerStats(context.Background(), s.Name, streamStats)
	scope.Declare("responseBody", &responseBody, "err", &err)
	godebug.Line(ctx, scope, 81)
	if err != nil {
		godebug.Line(ctx, scope, 82)
		s.mu.Lock()
		godebug.Line(ctx, scope, 83)
		s.err = err
		godebug.Line(ctx, scope, 84)
		s.mu.Unlock()
		godebug.Line(ctx, scope, 85)
		return
	}
	godebug.Line(ctx, scope, 87)
	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 87)
	godebug.Line(ctx, scope, 89)

	dec := json.NewDecoder(responseBody)
	scope.Declare("dec", &dec)
	godebug.Line(ctx, scope, 90)
	go func() {
		fn := func(ctx *godebug.Context) {
			godebug.Line(ctx, scope, 91)
			for {
				godebug.Line(ctx, scope, 92)
				var v *types.StatsJSON
				scope := scope.EnteringNewChildScope()
				scope.Declare("v", &v)
				godebug.Line(ctx, scope, 93)
				if err := dec.Decode(&v); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(ctx, scope, 94)
					u <- err
					godebug.Line(ctx, scope, 95)
					return
				}
				godebug.Line(ctx, scope, 98)
				var memPercent = 0.0
				scope.Declare("memPercent", &memPercent)
				godebug.Line(ctx, scope, 99)
				var cpuPercent = 0.0
				scope.Declare("cpuPercent", &cpuPercent)
				godebug.Line(ctx, scope, 103)
				if v.MemoryStats.Limit != 0 {
					godebug.Line(ctx, scope, 104)
					memPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
				}
				godebug.Line(ctx, scope, 107)
				previousCPU = v.PreCPUStats.CPUUsage.TotalUsage
				godebug.Line(ctx, scope, 108)
				previousSystem = v.PreCPUStats.SystemUsage
				godebug.Line(ctx, scope, 109)
				cpuPercent = calculateCPUPercent(previousCPU, previousSystem, v)
				godebug.Line(ctx, scope, 110)
				blkRead, blkWrite := calculateBlockIO(v.BlkioStats)
				scope.Declare("blkRead", &blkRead, "blkWrite", &blkWrite)
				godebug.Line(ctx, scope, 111)
				s.mu.Lock()
				godebug.Line(ctx, scope, 112)
				s.CPUPercentage = cpuPercent
				godebug.Line(ctx, scope, 113)
				s.Memory = float64(v.MemoryStats.Usage)
				godebug.Line(ctx, scope, 114)
				s.MemoryLimit = float64(v.MemoryStats.Limit)
				godebug.Line(ctx, scope, 115)
				s.MemoryPercentage = memPercent
				godebug.Line(ctx, scope, 116)
				s.NetworkRx, s.NetworkTx = calculateNetwork(v.Networks)
				godebug.Line(ctx, scope, 117)
				s.BlockRead = float64(blkRead)
				godebug.Line(ctx, scope, 118)
				s.BlockWrite = float64(blkWrite)
				godebug.Line(ctx, scope, 119)
				s.PidsCurrent = v.PidsStats.Current
				godebug.Line(ctx, scope, 120)
				s.mu.Unlock()
				godebug.Line(ctx, scope, 121)
				u <- nil
				godebug.Line(ctx, scope, 122)
				if !streamStats {
					godebug.Line(ctx, scope, 123)
					return
				}
				godebug.Line(ctx, scope, 91)
			}
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}()
	godebug.Line(ctx, scope, 127)
	for {
		godebug.Select(ctx, scope, 128)
		select {
		case <-godebug.Comm(ctx, scope, 129):
			panic("impossible")
		case <-time.After(2 * time.Second):
			godebug.Line(ctx, scope, 129)
			godebug.Line(ctx, scope, 132)

			s.mu.Lock()
			godebug.Line(ctx, scope, 133)
			s.CPUPercentage = 0
			godebug.Line(ctx, scope, 134)
			s.Memory = 0
			godebug.Line(ctx, scope, 135)
			s.MemoryPercentage = 0
			godebug.Line(ctx, scope, 136)
			s.MemoryLimit = 0
			godebug.Line(ctx, scope, 137)
			s.NetworkRx = 0
			godebug.Line(ctx, scope, 138)
			s.NetworkTx = 0
			godebug.Line(ctx, scope, 139)
			s.BlockRead = 0
			godebug.Line(ctx, scope, 140)
			s.BlockWrite = 0
			godebug.Line(ctx, scope, 141)
			s.PidsCurrent = 0
			godebug.Line(ctx, scope, 142)
			s.mu.Unlock()
			godebug.Line(ctx, scope, 144)

			if !getFirst {
				getFirst = true
				waitFirst.Done()
			}
		case <-godebug.Comm(ctx, scope, 148):
			panic("impossible")
		case err := <-u:
			godebug.Line(ctx, scope, 148)
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 149)
			if err != nil {
				s.mu.Lock()
				s.err = err
				s.mu.Unlock()
				return
			}
			godebug.Line(ctx, scope, 156)

			if !getFirst {
				getFirst = true
				waitFirst.Done()
			}
		case <-godebug.EndSelect(ctx, scope):
			panic("impossible")
		}
		godebug.Line(ctx, scope, 161)
		if !streamStats {
			godebug.Line(ctx, scope, 162)
			return
		}
		godebug.Line(ctx, scope, 127)
	}
}

func (s *containerStats) Display(w io.Writer) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = s.Display(w)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("s", &s, "w", &w)
	godebug.Line(ctx, scope, 168)
	s.mu.RLock()
	godebug.Line(ctx, scope, 169)
	defer s.mu.RUnlock()
	defer godebug.Defer(ctx, scope, 169)
	godebug.Line(ctx, scope, 170)
	if s.err != nil {
		godebug.Line(ctx, scope, 171)
		return s.err
	}
	godebug.Line(ctx, scope, 173)
	fmt.Fprintf(w, "%s\t%.2f%%\t%s / %s\t%.2f%%\t%s / %s\t%s / %s\t%d\n",
		s.Name,
		s.CPUPercentage,
		units.BytesSize(s.Memory), units.BytesSize(s.MemoryLimit),
		s.MemoryPercentage,
		units.HumanSize(s.NetworkRx), units.HumanSize(s.NetworkTx),
		units.HumanSize(s.BlockRead), units.HumanSize(s.BlockWrite),
		s.PidsCurrent)
	godebug.Line(ctx, scope, 181)
	return nil
}

func calculateCPUPercent(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var result1 float64
	ctx, ok := godebug.EnterFunc(func() {
		result1 = calculateCPUPercent(previousCPU, previousSystem, v)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("previousCPU", &previousCPU, "previousSystem", &previousSystem, "v", &v)
	godebug.Line(ctx, scope, 185)
	var (
		cpuPercent = 0.0

		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)

		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)
	scope.Declare("cpuPercent", &cpuPercent, "cpuDelta", &cpuDelta, "systemDelta", &systemDelta)
	godebug.Line(ctx, scope, 193)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		godebug.Line(ctx, scope, 194)
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	godebug.Line(ctx, scope, 196)
	return cpuPercent
}

func calculateBlockIO(blkio types.BlkioStats) (blkRead uint64, blkWrite uint64) {
	ctx, ok := godebug.EnterFunc(func() {
		blkRead, blkWrite = calculateBlockIO(blkio)
	})
	if !ok {
		return blkRead, blkWrite
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("blkio", &blkio, "blkRead", &blkRead, "blkWrite", &blkWrite)
	{
		scope := scope.EnteringNewChildScope()
		for _, bioEntry := range blkio.IoServiceBytesRecursive {
			godebug.Line(ctx, scope, 200)
			scope.Declare("bioEntry", &bioEntry)
			godebug.Line(ctx, scope, 201)
			switch strings.ToLower(bioEntry.Op) {
			case godebug.Case(ctx, scope, 202):
				fallthrough
			case "read":
				godebug.Line(ctx, scope, 203)
				blkRead = blkRead + bioEntry.Value
			case godebug.Case(ctx, scope, 204):
				fallthrough
			case "write":
				godebug.Line(ctx, scope, 205)
				blkWrite = blkWrite + bioEntry.Value
			}
		}
		godebug.Line(ctx, scope, 200)
	}
	godebug.Line(ctx, scope, 208)
	return
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var result1 float64
	var result2 float64
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = calculateNetwork(network)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_helpers_go_scope.EnteringNewChildScope()
	scope.Declare("network", &network)
	godebug.Line(ctx, scope, 212)
	var rx, tx float64
	scope.Declare("rx", &rx, "tx", &tx)
	{
		scope := scope.EnteringNewChildScope()

		for _, v := range network {
			godebug.Line(ctx, scope, 214)
			scope.Declare("v", &v)
			godebug.Line(ctx, scope, 215)
			rx += float64(v.RxBytes)
			godebug.Line(ctx, scope, 216)
			tx += float64(v.TxBytes)
		}
		godebug.Line(ctx, scope, 214)
	}
	godebug.Line(ctx, scope, 218)
	return rx, tx
}

var stats_helpers_go_contents = `package client

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/go-units"
	"golang.org/x/net/context"
)

type containerStats struct {
	Name             string
	CPUPercentage    float64
	Memory           float64
	MemoryLimit      float64
	MemoryPercentage float64
	NetworkRx        float64
	NetworkTx        float64
	BlockRead        float64
	BlockWrite       float64
	PidsCurrent      uint64
	mu               sync.RWMutex
	err              error
}

type stats struct {
	mu sync.Mutex
	cs []*containerStats
}

func (s *stats) add(cs *containerStats) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.isKnownContainer(cs.Name); !exists {
		s.cs = append(s.cs, cs)
		return true
	}
	return false
}

func (s *stats) remove(id string) {
	s.mu.Lock()
	if i, exists := s.isKnownContainer(id); exists {
		s.cs = append(s.cs[:i], s.cs[i+1:]...)
	}
	s.mu.Unlock()
}

func (s *stats) isKnownContainer(cid string) (int, bool) {
	for i, c := range s.cs {
		if c.Name == cid {
			return i, true
		}
	}
	return -1, false
}

func (s *containerStats) Collect(cli client.APIClient, streamStats bool, waitFirst *sync.WaitGroup) {
	var (
		getFirst       bool
		previousCPU    uint64
		previousSystem uint64
		u              = make(chan error, 1)
	)

	defer func() {
		// if error happens and we get nothing of stats, release wait group whatever
		if !getFirst {
			getFirst = true
			waitFirst.Done()
		}
	}()

	responseBody, err := cli.ContainerStats(context.Background(), s.Name, streamStats)
	if err != nil {
		s.mu.Lock()
		s.err = err
		s.mu.Unlock()
		return
	}
	defer responseBody.Close()

	dec := json.NewDecoder(responseBody)
	go func() {
		for {
			var v *types.StatsJSON
			if err := dec.Decode(&v); err != nil {
				u <- err
				return
			}

			var memPercent = 0.0
			var cpuPercent = 0.0

			// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
			// got any data from cgroup
			if v.MemoryStats.Limit != 0 {
				memPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
			}

			previousCPU = v.PreCPUStats.CPUUsage.TotalUsage
			previousSystem = v.PreCPUStats.SystemUsage
			cpuPercent = calculateCPUPercent(previousCPU, previousSystem, v)
			blkRead, blkWrite := calculateBlockIO(v.BlkioStats)
			s.mu.Lock()
			s.CPUPercentage = cpuPercent
			s.Memory = float64(v.MemoryStats.Usage)
			s.MemoryLimit = float64(v.MemoryStats.Limit)
			s.MemoryPercentage = memPercent
			s.NetworkRx, s.NetworkTx = calculateNetwork(v.Networks)
			s.BlockRead = float64(blkRead)
			s.BlockWrite = float64(blkWrite)
			s.PidsCurrent = v.PidsStats.Current
			s.mu.Unlock()
			u <- nil
			if !streamStats {
				return
			}
		}
	}()
	for {
		select {
		case <-time.After(2 * time.Second):
			// zero out the values if we have not received an update within
			// the specified duration.
			s.mu.Lock()
			s.CPUPercentage = 0
			s.Memory = 0
			s.MemoryPercentage = 0
			s.MemoryLimit = 0
			s.NetworkRx = 0
			s.NetworkTx = 0
			s.BlockRead = 0
			s.BlockWrite = 0
			s.PidsCurrent = 0
			s.mu.Unlock()
			// if this is the first stat you get, release WaitGroup
			if !getFirst {
				getFirst = true
				waitFirst.Done()
			}
		case err := <-u:
			if err != nil {
				s.mu.Lock()
				s.err = err
				s.mu.Unlock()
				return
			}
			// if this is the first stat you get, release WaitGroup
			if !getFirst {
				getFirst = true
				waitFirst.Done()
			}
		}
		if !streamStats {
			return
		}
	}
}

func (s *containerStats) Display(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.err != nil {
		return s.err
	}
	fmt.Fprintf(w, "%s\t%.2f%%\t%s / %s\t%.2f%%\t%s / %s\t%s / %s\t%d\n",
		s.Name,
		s.CPUPercentage,
		units.BytesSize(s.Memory), units.BytesSize(s.MemoryLimit),
		s.MemoryPercentage,
		units.HumanSize(s.NetworkRx), units.HumanSize(s.NetworkTx),
		units.HumanSize(s.BlockRead), units.HumanSize(s.BlockWrite),
		s.PidsCurrent)
	return nil
}

func calculateCPUPercent(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateBlockIO(blkio types.BlkioStats) (blkRead uint64, blkWrite uint64) {
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
}
`


var stop_go_scope = godebug.EnteringNewFile(client_pkg_scope, stop_go_contents)

func (cli *DockerCli) CmdStop(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdStop(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := stop_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("stop", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["stop"].Description+".\nSending SIGTERM and then SIGKILL after a grace period", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	nSeconds := cmd.Int([]string{"t", "-time"}, 10, "Seconds to wait for stop before killing it")
	scope.Declare("nSeconds", &nSeconds)
	godebug.Line(ctx, scope, 21)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 23)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 25)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 26)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 27)
			if err := cli.client.ContainerStop(context.Background(), name, *nSeconds); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 28)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 29)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 30)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 26)
	}
	godebug.Line(ctx, scope, 33)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 34)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 36)
	return nil
}

var stop_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdStop stops one or more containers.
//
// A running container is stopped by first sending SIGTERM and then SIGKILL if the container fails to stop within a grace period (the default is 10 seconds).
//
// Usage: docker stop [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdStop(args ...string) error {
	cmd := Cli.Subcmd("stop", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["stop"].Description+".\nSending SIGTERM and then SIGKILL after a grace period", true)
	nSeconds := cmd.Int([]string{"t", "-time"}, 10, "Seconds to wait for stop before killing it")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		if err := cli.client.ContainerStop(context.Background(), name, *nSeconds); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var tag_go_scope = godebug.EnteringNewFile(client_pkg_scope, tag_go_contents)

func (cli *DockerCli) CmdTag(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdTag(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := tag_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 18)
	cmd := Cli.Subcmd("tag", []string{"IMAGE[:TAG] [REGISTRYHOST/][USERNAME/]NAME[:TAG]"}, Cli.DockerCommands["tag"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 19)
	force := cmd.Bool([]string{"#f", "#-force"}, false, "Force the tagging even if there's a conflict")
	scope.Declare("force", &force)
	godebug.Line(ctx, scope, 20)
	cmd.Require(flag.Exact, 2)
	godebug.Line(ctx, scope, 22)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 24)

	ref, err := reference.ParseNamed(cmd.Arg(1))
	scope.Declare("ref", &ref, "err", &err)
	godebug.Line(ctx, scope, 25)
	if err != nil {
		godebug.Line(ctx, scope, 26)
		return err
	}
	godebug.Line(ctx, scope, 29)

	if _, isCanonical := ref.(reference.Canonical); isCanonical {
		scope := scope.EnteringNewChildScope()
		scope.Declare("isCanonical", &isCanonical)
		godebug.Line(ctx, scope, 30)
		return errors.New("refusing to create a tag with a digest reference")
	}
	godebug.Line(ctx, scope, 33)

	var tag string
	scope.Declare("tag", &tag)
	godebug.Line(ctx, scope, 34)
	if tagged, isTagged := ref.(reference.NamedTagged); isTagged {
		scope := scope.EnteringNewChildScope()
		scope.Declare("tagged", &tagged, "isTagged", &isTagged)
		godebug.Line(ctx, scope, 35)
		tag = tagged.Tag()
	}
	godebug.Line(ctx, scope, 38)

	options := types.ImageTagOptions{
		ImageID:        cmd.Arg(0),
		RepositoryName: ref.Name(),
		Tag:            tag,
		Force:          *force,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 45)

	return cli.client.ImageTag(context.Background(), options)
}

var tag_go_contents = `package client

import (
	"errors"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/reference"
	"github.com/docker/engine-api/types"
)

// CmdTag tags an image into a repository.
//
// Usage: docker tag [OPTIONS] IMAGE[:TAG] [REGISTRYHOST/][USERNAME/]NAME[:TAG]
func (cli *DockerCli) CmdTag(args ...string) error {
	cmd := Cli.Subcmd("tag", []string{"IMAGE[:TAG] [REGISTRYHOST/][USERNAME/]NAME[:TAG]"}, Cli.DockerCommands["tag"].Description, true)
	force := cmd.Bool([]string{"#f", "#-force"}, false, "Force the tagging even if there's a conflict")
	cmd.Require(flag.Exact, 2)

	cmd.ParseFlags(args, true)

	ref, err := reference.ParseNamed(cmd.Arg(1))
	if err != nil {
		return err
	}

	if _, isCanonical := ref.(reference.Canonical); isCanonical {
		return errors.New("refusing to create a tag with a digest reference")
	}

	var tag string
	if tagged, isTagged := ref.(reference.NamedTagged); isTagged {
		tag = tagged.Tag()
	}

	options := types.ImageTagOptions{
		ImageID:        cmd.Arg(0),
		RepositoryName: ref.Name(),
		Tag:            tag,
		Force:          *force,
	}

	return cli.client.ImageTag(context.Background(), options)
}
`


var top_go_scope = godebug.EnteringNewFile(client_pkg_scope, top_go_contents)

func (cli *DockerCli) CmdTop(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdTop(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := top_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 18)
	cmd := Cli.Subcmd("top", []string{"CONTAINER [ps OPTIONS]"}, Cli.DockerCommands["top"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 19)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 21)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 23)

	var arguments []string
	scope.Declare("arguments", &arguments)
	godebug.Line(ctx, scope, 24)
	if cmd.NArg() > 1 {
		godebug.Line(ctx, scope, 25)
		arguments = cmd.Args()[1:]
	}
	godebug.Line(ctx, scope, 28)

	procList, err := cli.client.ContainerTop(context.Background(), cmd.Arg(0), arguments)
	scope.Declare("procList", &procList, "err", &err)
	godebug.Line(ctx, scope, 29)
	if err != nil {
		godebug.Line(ctx, scope, 30)
		return err
	}
	godebug.Line(ctx, scope, 33)

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	scope.Declare("w", &w)
	godebug.Line(ctx, scope, 34)
	fmt.Fprintln(w, strings.Join(procList.Titles, "\t"))
	{
		scope := scope.EnteringNewChildScope()

		for _, proc := range procList.Processes {
			godebug.Line(ctx, scope, 36)
			scope.Declare("proc", &proc)
			godebug.Line(ctx, scope, 37)
			fmt.Fprintln(w, strings.Join(proc, "\t"))
		}
		godebug.Line(ctx, scope, 36)
	}
	godebug.Line(ctx, scope, 39)
	w.Flush()
	godebug.Line(ctx, scope, 40)
	return nil
}

var top_go_contents = `package client

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdTop displays the running processes of a container.
//
// Usage: docker top CONTAINER
func (cli *DockerCli) CmdTop(args ...string) error {
	cmd := Cli.Subcmd("top", []string{"CONTAINER [ps OPTIONS]"}, Cli.DockerCommands["top"].Description, true)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var arguments []string
	if cmd.NArg() > 1 {
		arguments = cmd.Args()[1:]
	}

	procList, err := cli.client.ContainerTop(context.Background(), cmd.Arg(0), arguments)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(procList.Titles, "\t"))

	for _, proc := range procList.Processes {
		fmt.Fprintln(w, strings.Join(proc, "\t"))
	}
	w.Flush()
	return nil
}
`


var trust_go_scope = godebug.EnteringNewFile(client_pkg_scope, trust_go_contents)

var (
	releasesRole = path.Join(data.CanonicalTargetsRole, "releases")
	untrusted    bool
)

func addTrustedFlags(fs *flag.FlagSet, verify bool) {
	ctx, _ok := godebug.EnterFunc(func() {
		addTrustedFlags(fs, verify)
	})
	if !_ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("fs", &fs, "verify", &verify)
	godebug.Line(ctx, scope, 48)
	var trusted bool
	scope.Declare("trusted", &trusted)
	godebug.Line(ctx, scope, 49)
	if e := os.Getenv("DOCKER_CONTENT_TRUST"); e != "" {
		scope := scope.EnteringNewChildScope()
		scope.Declare("e", &e)
		godebug.Line(ctx, scope, 50)
		if t, err := strconv.ParseBool(e); t || err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("t", &t, "err", &err)
			godebug.Line(ctx, scope, 52)

			trusted = true
		}
	}
	godebug.Line(ctx, scope, 55)
	message := "Skip image signing"
	scope.Declare("message", &message)
	godebug.Line(ctx, scope, 56)
	if verify {
		godebug.Line(ctx, scope, 57)
		message = "Skip image verification"
	}
	godebug.Line(ctx, scope, 59)
	fs.BoolVar(&untrusted, []string{"-disable-content-trust"}, !trusted, message)
}

func isTrusted() bool {
	var result1 bool
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = isTrusted()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	godebug.Line(ctx, trust_go_scope, 63)
	return !untrusted
}

type target struct {
	reference registry.Reference
	digest    digest.Digest
	size      int64
}

func (cli *DockerCli) trustDirectory() string {
	var result1 string
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.trustDirectory()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 73)
	return filepath.Join(cliconfig.ConfigDir(), "trust")
}

func (cli *DockerCli) certificateDirectory(server string) (string, error) {
	var result1 string
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = cli.certificateDirectory(server)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "server", &server)
	godebug.Line(ctx, scope, 80)
	u, err := url.Parse(server)
	scope.Declare("u", &u, "err", &err)
	godebug.Line(ctx, scope, 81)
	if err != nil {
		godebug.Line(ctx, scope, 82)
		return "", err
	}
	godebug.Line(ctx, scope, 85)

	return filepath.Join(cliconfig.ConfigDir(), "tls", u.Host), nil
}

func trustServer(index *registrytypes.IndexInfo) (string, error) {
	var result1 string
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = trustServer(index)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("index", &index)
	godebug.Line(ctx, scope, 89)
	if s := os.Getenv("DOCKER_CONTENT_TRUST_SERVER"); s != "" {
		scope := scope.EnteringNewChildScope()
		scope.Declare("s", &s)
		godebug.Line(ctx, scope, 90)
		urlObj, err := url.Parse(s)
		scope.Declare("urlObj", &urlObj, "err", &err)
		godebug.Line(ctx, scope, 91)
		if err != nil || urlObj.Scheme != "https" {
			godebug.Line(ctx, scope, 92)
			return "", fmt.Errorf("valid https URL required for trust server, got %s", s)
		}
		godebug.Line(ctx, scope, 95)

		return s, nil
	}
	godebug.Line(ctx, scope, 97)
	if index.Official {
		godebug.Line(ctx, scope, 98)
		return registry.NotaryServer, nil
	}
	godebug.Line(ctx, scope, 100)
	return "https://" + index.Name, nil
}

type simpleCredentialStore struct {
	auth types.AuthConfig
}

func (scs simpleCredentialStore) Basic(u *url.URL) (string, string) {
	var result1 string
	var result2 string
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = scs.Basic(u)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("scs", &scs, "u", &u)
	godebug.Line(ctx, scope, 108)
	return scs.auth.Username, scs.auth.Password
}

func (scs simpleCredentialStore) RefreshToken(u *url.URL, service string) string {
	var result1 string
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = scs.RefreshToken(u, service)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("scs", &scs, "u", &u, "service", &service)
	godebug.Line(ctx, scope, 112)
	return scs.auth.IdentityToken
}

func (scs simpleCredentialStore) SetRefreshToken(*url.URL, string, string) {
}

func (cli *DockerCli) getNotaryRepository(repoInfo *registry.RepositoryInfo, authConfig types.AuthConfig, actions ...string) (*client.NotaryRepository, error) {
	var result1 *client.NotaryRepository
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = cli.getNotaryRepository(repoInfo, authConfig, actions...)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "repoInfo", &repoInfo, "authConfig", &authConfig, "actions", &actions)
	godebug.Line(ctx, scope, 122)
	server, err := trustServer(repoInfo.Index)
	scope.Declare("server", &server, "err", &err)
	godebug.Line(ctx, scope, 123)
	if err != nil {
		godebug.Line(ctx, scope, 124)
		return nil, err
	}
	godebug.Line(ctx, scope, 127)

	var cfg = tlsconfig.ClientDefault
	scope.Declare("cfg", &cfg)
	godebug.Line(ctx, scope, 128)
	cfg.InsecureSkipVerify = !repoInfo.Index.Secure
	godebug.Line(ctx, scope, 131)

	certDir, err := cli.certificateDirectory(server)
	scope.Declare("certDir", &certDir)
	godebug.Line(ctx, scope, 132)
	if err != nil {
		godebug.Line(ctx, scope, 133)
		return nil, err
	}
	godebug.Line(ctx, scope, 135)
	logrus.Debugf("reading certificate directory: %s", certDir)
	godebug.Line(ctx, scope, 137)

	if err := registry.ReadCertsDirectory(&cfg, certDir); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 138)
		return nil, err
	}
	godebug.Line(ctx, scope, 141)

	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &cfg,
		DisableKeepAlives:   true,
	}
	scope.Declare("base", &base)
	godebug.Line(ctx, scope, 154)

	modifiers := registry.DockerHeaders(clientUserAgent(), http.Header{})
	scope.Declare("modifiers", &modifiers)
	godebug.Line(ctx, scope, 155)
	authTransport := transport.NewTransport(base, modifiers...)
	scope.Declare("authTransport", &authTransport)
	godebug.Line(ctx, scope, 156)
	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   5 * time.Second,
	}
	scope.Declare("pingClient", &pingClient)
	godebug.Line(ctx, scope, 160)

	endpointStr := server + "/v2/"
	scope.Declare("endpointStr", &endpointStr)
	godebug.Line(ctx, scope, 161)
	req, err := http.NewRequest("GET", endpointStr, nil)
	scope.Declare("req", &req)
	godebug.Line(ctx, scope, 162)
	if err != nil {
		godebug.Line(ctx, scope, 163)
		return nil, err
	}
	godebug.Line(ctx, scope, 166)

	challengeManager := auth.NewSimpleChallengeManager()
	scope.Declare("challengeManager", &challengeManager)
	godebug.Line(ctx, scope, 168)

	resp, err := pingClient.Do(req)
	scope.Declare("resp", &resp)
	godebug.Line(ctx, scope, 169)
	if err != nil {
		godebug.Line(ctx, scope, 171)

		logrus.Debugf("Error pinging notary server %q: %s", endpointStr, err)
	} else {
		godebug.Line(ctx, scope, 172)
		godebug.Line(ctx, scope, 173)
		defer resp.Body.Close()
		defer godebug.Defer(ctx, scope, 173)
		godebug.Line(ctx, scope, 177)

		if err := challengeManager.AddResponse(resp); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 178)
			return nil, err
		}
	}
	godebug.Line(ctx, scope, 182)

	creds := simpleCredentialStore{auth: authConfig}
	scope.Declare("creds", &creds)
	godebug.Line(ctx, scope, 183)
	tokenHandler := auth.NewTokenHandler(authTransport, creds, repoInfo.FullName(), actions...)
	scope.Declare("tokenHandler", &tokenHandler)
	godebug.Line(ctx, scope, 184)
	basicHandler := auth.NewBasicHandler(creds)
	scope.Declare("basicHandler", &basicHandler)
	godebug.Line(ctx, scope, 185)
	modifiers = append(modifiers, transport.RequestModifier(auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler)))
	godebug.Line(ctx, scope, 186)
	tr := transport.NewTransport(base, modifiers...)
	scope.Declare("tr", &tr)
	godebug.Line(ctx, scope, 188)

	return client.NewNotaryRepository(cli.trustDirectory(), repoInfo.FullName(), server, tr, cli.getPassphraseRetriever())
}

func convertTarget(t client.Target) (target, error) {
	var result1 target
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = convertTarget(t)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 192)
	h, ok := t.Hashes["sha256"]
	scope.Declare("h", &h, "ok", &ok)
	godebug.Line(ctx, scope, 193)
	if !ok {
		godebug.Line(ctx, scope, 194)
		return target{}, errors.New("no valid hash, expecting sha256")
	}
	godebug.Line(ctx, scope, 196)
	return target{
		reference: registry.ParseReference(t.Name),
		digest:    digest.NewDigestFromHex("sha256", hex.EncodeToString(h)),
		size:      t.Length,
	}, nil
}

func (cli *DockerCli) getPassphraseRetriever() passphrase.Retriever {
	var result1 passphrase.Retriever
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.getPassphraseRetriever()
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 204)
	aliasMap := map[string]string{
		"root":     "root",
		"snapshot": "repository",
		"targets":  "repository",
		"default":  "repository",
	}
	scope.Declare("aliasMap", &aliasMap)
	godebug.Line(ctx, scope, 210)

	baseRetriever := passphrase.PromptRetrieverWithInOut(cli.in, cli.out, aliasMap)
	scope.Declare("baseRetriever", &baseRetriever)
	godebug.Line(ctx, scope, 211)
	env := map[string]string{
		"root":     os.Getenv("DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE"),
		"snapshot": os.Getenv("DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE"),
		"targets":  os.Getenv("DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE"),
		"default":  os.Getenv("DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE"),
	}
	scope.Declare("env", &env)
	godebug.Line(ctx, scope, 219)

	if env["root"] == "" {
		godebug.Line(ctx, scope, 220)
		if passphrase := os.Getenv("DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE"); passphrase != "" {
			scope := scope.EnteringNewChildScope()
			scope.Declare("passphrase", &passphrase)
			godebug.Line(ctx, scope, 221)
			env["root"] = passphrase
			godebug.Line(ctx, scope, 222)
			fmt.Fprintf(cli.err, "[DEPRECATED] The environment variable DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE has been deprecated and will be removed in v1.10. Please use DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE\n")
		}
	}
	godebug.Line(ctx, scope, 225)
	if env["snapshot"] == "" || env["targets"] == "" || env["default"] == "" {
		godebug.Line(ctx, scope, 226)
		if passphrase := os.Getenv("DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE"); passphrase != "" {
			scope := scope.EnteringNewChildScope()
			scope.Declare("passphrase", &passphrase)
			godebug.Line(ctx, scope, 227)
			env["snapshot"] = passphrase
			godebug.Line(ctx, scope, 228)
			env["targets"] = passphrase
			godebug.Line(ctx, scope, 229)
			env["default"] = passphrase
			godebug.Line(ctx, scope, 230)
			fmt.Fprintf(cli.err, "[DEPRECATED] The environment variable DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE has been deprecated and will be removed in v1.10. Please use DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE\n")
		}
	}
	godebug.Line(ctx, scope, 234)

	return func(keyName string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		var result1 string
		var result2 bool
		var result3 error
		fn := func(ctx *godebug.Context) {
			result1, result2, result3 = func() (string, bool, error) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("keyName", &keyName, "alias", &alias, "createNew", &createNew, "numAttempts", &numAttempts)
				godebug.Line(ctx, scope, 235)
				if v := env[alias]; v != "" {
					scope := scope.EnteringNewChildScope()
					scope.Declare("v", &v)
					godebug.Line(ctx, scope, 236)
					return v, numAttempts > 1, nil
				}
				godebug.Line(ctx, scope, 239)
				if v := env["default"]; v != "" && alias != data.CanonicalRootRole {
					scope := scope.EnteringNewChildScope()
					scope.Declare("v", &v)
					godebug.Line(ctx, scope, 240)
					return v, numAttempts > 1, nil
				}
				godebug.Line(ctx, scope, 242)
				return baseRetriever(keyName, alias, createNew, numAttempts)
			}()
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2, result3
	}

}

func (cli *DockerCli) trustedReference(ref reference.NamedTagged) (reference.Canonical, error) {
	var result1 reference.Canonical
	var result2 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1, result2 = cli.trustedReference(ref)
	})
	if !_ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "ref", &ref)
	godebug.Line(ctx, scope, 247)
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	scope.Declare("repoInfo", &repoInfo, "err", &err)
	godebug.Line(ctx, scope, 248)
	if err != nil {
		godebug.Line(ctx, scope, 249)
		return nil, err
	}
	godebug.Line(ctx, scope, 253)

	authConfig := cli.resolveAuthConfig(repoInfo.Index)
	scope.Declare("authConfig", &authConfig)
	godebug.Line(ctx, scope, 255)

	notaryRepo, err := cli.getNotaryRepository(repoInfo, authConfig, "pull")
	scope.Declare("notaryRepo", &notaryRepo)
	godebug.Line(ctx, scope, 256)
	if err != nil {
		godebug.Line(ctx, scope, 257)
		fmt.Fprintf(cli.out, "Error establishing connection to trust repository: %s\n", err)
		godebug.Line(ctx, scope, 258)
		return nil, err
	}
	godebug.Line(ctx, scope, 261)

	t, err := notaryRepo.GetTargetByName(ref.Tag(), releasesRole, data.CanonicalTargetsRole)
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 262)
	if err != nil {
		godebug.Line(ctx, scope, 263)
		return nil, err
	}
	godebug.Line(ctx, scope, 267)

	if t.Role != releasesRole && t.Role != data.CanonicalTargetsRole {
		godebug.Line(ctx, scope, 268)
		return nil, notaryError(repoInfo.FullName(), fmt.Errorf("No trust data for %s", ref.Tag()))
	}
	godebug.Line(ctx, scope, 270)
	r, err := convertTarget(t.Target)
	scope.Declare("r", &r)
	godebug.Line(ctx, scope, 271)
	if err != nil {
		godebug.Line(ctx, scope, 272)
		return nil, err

	}
	godebug.Line(ctx, scope, 276)

	return reference.WithDigest(ref, r.digest)
}

func (cli *DockerCli) tagTrusted(trustedRef reference.Canonical, ref reference.NamedTagged) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.tagTrusted(trustedRef, ref)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "trustedRef", &trustedRef, "ref", &ref)
	godebug.Line(ctx, scope, 280)
	fmt.Fprintf(cli.out, "Tagging %s as %s\n", trustedRef.String(), ref.String())
	godebug.Line(ctx, scope, 282)

	options := types.ImageTagOptions{
		ImageID:        trustedRef.String(),
		RepositoryName: trustedRef.Name(),
		Tag:            ref.Tag(),
		Force:          true,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 289)

	return cli.client.ImageTag(context.Background(), options)
}

func notaryError(repoName string, err error) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = notaryError(repoName, err)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("repoName", &repoName, "err", &err)
	godebug.Line(ctx, scope, 293)
	switch err.(type) {
	case *json.SyntaxError:
		godebug.Line(ctx, scope, 294)
		godebug.Line(ctx, scope, 295)
		logrus.Debugf("Notary syntax error: %s", err)
		godebug.Line(ctx, scope, 296)
		return fmt.Errorf("Error: no trust data available for remote repository %s. Try running notary server and setting DOCKER_CONTENT_TRUST_SERVER to its HTTPS address?", repoName)
	case signed.ErrExpired:
		godebug.Line(ctx, scope, 297)
		godebug.Line(ctx, scope, 298)
		return fmt.Errorf("Error: remote repository %s out-of-date: %v", repoName, err)
	case trustmanager.ErrKeyNotFound:
		godebug.Line(ctx, scope, 299)
		godebug.Line(ctx, scope, 300)
		return fmt.Errorf("Error: signing keys for remote repository %s not found: %v", repoName, err)
	case *net.OpError:
		godebug.Line(ctx, scope, 301)
		godebug.Line(ctx, scope, 302)
		return fmt.Errorf("Error: error contacting notary server: %v", err)
	case store.ErrMetaNotFound:
		godebug.Line(ctx, scope, 303)
		godebug.Line(ctx, scope, 304)
		return fmt.Errorf("Error: trust data missing for remote repository %s or remote repository not found: %v", repoName, err)
	case signed.ErrInvalidKeyType:
		godebug.Line(ctx, scope, 305)
		godebug.Line(ctx, scope, 306)
		return fmt.Errorf("Warning: potential malicious behavior - trust data mismatch for remote repository %s: %v", repoName, err)
	case signed.ErrNoKeys:
		godebug.Line(ctx, scope, 307)
		godebug.Line(ctx, scope, 308)
		return fmt.Errorf("Error: could not find signing keys for remote repository %s, or could not decrypt signing key: %v", repoName, err)
	case signed.ErrLowVersion:
		godebug.Line(ctx, scope, 309)
		godebug.Line(ctx, scope, 310)
		return fmt.Errorf("Warning: potential malicious behavior - trust data version is lower than expected for remote repository %s: %v", repoName, err)
	case signed.ErrRoleThreshold:
		godebug.Line(ctx, scope, 311)
		godebug.Line(ctx, scope, 312)
		return fmt.Errorf("Warning: potential malicious behavior - trust data has insufficient signatures for remote repository %s: %v", repoName, err)
	case client.ErrRepositoryNotExist:
		godebug.Line(ctx, scope, 313)
		godebug.Line(ctx, scope, 314)
		return fmt.Errorf("Error: remote trust data does not exist for %s: %v", repoName, err)
	case signed.ErrInsufficientSignatures:
		godebug.Line(ctx, scope, 315)
		godebug.Line(ctx, scope, 316)
		return fmt.Errorf("Error: could not produce valid signature for %s.  If Yubikey was used, was touch input provided?: %v", repoName, err)
	}
	godebug.Line(ctx, scope, 319)

	return err
}

func (cli *DockerCli) trustedPull(repoInfo *registry.RepositoryInfo, ref registry.Reference, authConfig types.AuthConfig, requestPrivilege apiclient.RequestPrivilegeFunc) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.trustedPull(repoInfo, ref, authConfig, requestPrivilege)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "repoInfo", &repoInfo, "ref", &ref, "authConfig", &authConfig, "requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 323)
	var refs []target
	scope.Declare("refs", &refs)
	godebug.Line(ctx, scope, 325)

	notaryRepo, err := cli.getNotaryRepository(repoInfo, authConfig, "pull")
	scope.Declare("notaryRepo", &notaryRepo, "err", &err)
	godebug.Line(ctx, scope, 326)
	if err != nil {
		godebug.Line(ctx, scope, 327)
		fmt.Fprintf(cli.out, "Error establishing connection to trust repository: %s\n", err)
		godebug.Line(ctx, scope, 328)
		return err
	}
	godebug.Line(ctx, scope, 331)

	if ref.String() == "" {
		godebug.Line(ctx, scope, 333)

		targets, err := notaryRepo.ListTargets(releasesRole, data.CanonicalTargetsRole)
		scope := scope.EnteringNewChildScope()
		scope.Declare("targets", &targets, "err", &err)
		godebug.Line(ctx, scope, 334)
		if err != nil {
			godebug.Line(ctx, scope, 335)
			return notaryError(repoInfo.FullName(), err)
		}
		{
			scope := scope.EnteringNewChildScope()
			for _, tgt := range targets {
				godebug.Line(ctx, scope, 337)
				scope.Declare("tgt", &tgt)
				godebug.Line(ctx, scope, 338)
				t, err := convertTarget(tgt.Target)
				scope := scope.EnteringNewChildScope()
				scope.Declare("t", &t, "err", &err)
				godebug.Line(ctx, scope, 339)
				if err != nil {
					godebug.Line(ctx, scope, 340)
					fmt.Fprintf(cli.out, "Skipping target for %q\n", repoInfo.Name())
					godebug.Line(ctx, scope, 341)
					continue
				}
				godebug.Line(ctx, scope, 345)

				if tgt.Role != releasesRole && tgt.Role != data.CanonicalTargetsRole {
					godebug.Line(ctx, scope, 346)
					continue
				}
				godebug.Line(ctx, scope, 348)
				refs = append(refs, t)
			}
			godebug.Line(ctx, scope, 337)
		}
		godebug.Line(ctx, scope, 350)
		if len(refs) == 0 {
			godebug.Line(ctx, scope, 351)
			return notaryError(repoInfo.FullName(), fmt.Errorf("No trusted tags for %s", repoInfo.FullName()))
		}
	} else {
		godebug.Line(ctx, scope, 353)
		godebug.Line(ctx, scope, 354)
		t, err := notaryRepo.GetTargetByName(ref.String(), releasesRole, data.CanonicalTargetsRole)
		scope := scope.EnteringNewChildScope()
		scope.Declare("t", &t, "err", &err)
		godebug.Line(ctx, scope, 355)
		if err != nil {
			godebug.Line(ctx, scope, 356)
			return notaryError(repoInfo.FullName(), err)
		}
		godebug.Line(ctx, scope, 360)

		if t.Role != releasesRole && t.Role != data.CanonicalTargetsRole {
			godebug.Line(ctx, scope, 361)
			return notaryError(repoInfo.FullName(), fmt.Errorf("No trust data for %s", ref.String()))
		}
		godebug.Line(ctx, scope, 364)

		logrus.Debugf("retrieving target for %s role\n", t.Role)
		godebug.Line(ctx, scope, 365)
		r, err := convertTarget(t.Target)
		scope.Declare("r", &r)
		godebug.Line(ctx, scope, 366)
		if err != nil {
			godebug.Line(ctx, scope, 367)
			return err

		}
		godebug.Line(ctx, scope, 370)
		refs = append(refs, r)
	}
	{
		scope := scope.EnteringNewChildScope()

		for i, r := range refs {
			godebug.Line(ctx, scope, 373)
			scope.Declare("i", &i, "r", &r)
			godebug.Line(ctx, scope, 374)
			displayTag := r.reference.String()
			scope := scope.EnteringNewChildScope()
			scope.Declare("displayTag", &displayTag)
			godebug.Line(ctx, scope, 375)
			if displayTag != "" {
				godebug.Line(ctx, scope, 376)
				displayTag = ":" + displayTag
			}
			godebug.Line(ctx, scope, 378)
			fmt.Fprintf(cli.out, "Pull (%d of %d): %s%s@%s\n", i+1, len(refs), repoInfo.Name(), displayTag, r.digest)
			godebug.Line(ctx, scope, 380)

			if err := cli.imagePullPrivileged(authConfig, repoInfo.Name(), r.digest.String(), requestPrivilege); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 381)
				return err
			}
			godebug.Line(ctx, scope, 385)

			if !r.reference.HasDigest() {
				godebug.Line(ctx, scope, 386)
				tagged, err := reference.WithTag(repoInfo, r.reference.String())
				scope := scope.EnteringNewChildScope()
				scope.Declare("tagged", &tagged, "err", &err)
				godebug.Line(ctx, scope, 387)
				if err != nil {
					godebug.Line(ctx, scope, 388)
					return err
				}
				godebug.Line(ctx, scope, 390)
				trustedRef, err := reference.WithDigest(repoInfo, r.digest)
				scope.Declare("trustedRef", &trustedRef)
				godebug.Line(ctx, scope, 391)
				if err != nil {
					godebug.Line(ctx, scope, 392)
					return err
				}
				godebug.Line(ctx, scope, 394)
				if err := cli.tagTrusted(trustedRef, tagged); err != nil {
					scope := scope.EnteringNewChildScope()
					scope.Declare("err", &err)
					godebug.Line(ctx, scope, 395)
					return err
				}
			}
		}
		godebug.Line(ctx, scope, 373)
	}
	godebug.Line(ctx, scope, 399)
	return nil
}

func (cli *DockerCli) trustedPush(repoInfo *registry.RepositoryInfo, tag string, authConfig types.AuthConfig, requestPrivilege apiclient.RequestPrivilegeFunc) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.trustedPush(repoInfo, tag, authConfig, requestPrivilege)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "repoInfo", &repoInfo, "tag", &tag, "authConfig", &authConfig, "requestPrivilege", &requestPrivilege)
	godebug.Line(ctx, scope, 403)
	responseBody, err := cli.imagePushPrivileged(authConfig, repoInfo.Name(), tag, requestPrivilege)
	scope.Declare("responseBody", &responseBody, "err", &err)
	godebug.Line(ctx, scope, 404)
	if err != nil {
		godebug.Line(ctx, scope, 405)
		return err
	}
	godebug.Line(ctx, scope, 408)

	defer responseBody.Close()
	defer godebug.Defer(ctx, scope, 408)
	godebug.Line(ctx, scope, 412)

	target := &client.Target{}
	scope.Declare("target", &target)
	godebug.Line(ctx, scope, 415)

	cnt := 0
	scope.Declare("cnt", &cnt)
	godebug.Line(ctx, scope, 416)
	handleTarget := func(aux *json.RawMessage) {
		fn := func(ctx *godebug.Context) {
			scope := scope.EnteringNewChildScope()
			scope.Declare("aux", &aux)
			godebug.Line(ctx, scope, 417)
			cnt++
			godebug.Line(ctx, scope, 418)
			if cnt > 1 {
				godebug.Line(ctx, scope, 420)
				return
			}
			godebug.Line(ctx, scope, 423)
			var pushResult distribution.PushResult
			scope.Declare("pushResult", &pushResult)
			godebug.Line(ctx, scope, 424)
			err := json.Unmarshal(*aux, &pushResult)
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 425)
			if err == nil && pushResult.Tag != "" && pushResult.Digest.Validate() == nil {
				godebug.Line(ctx, scope, 426)
				h, err := hex.DecodeString(pushResult.Digest.Hex())
				scope := scope.EnteringNewChildScope()
				scope.Declare("h", &h, "err", &err)
				godebug.Line(ctx, scope, 427)
				if err != nil {
					godebug.Line(ctx, scope, 428)
					target = nil
					godebug.Line(ctx, scope, 429)
					return
				}
				godebug.Line(ctx, scope, 431)
				target.Name = registry.ParseReference(pushResult.Tag).String()
				godebug.Line(ctx, scope, 432)
				target.Hashes = data.Hashes{string(pushResult.Digest.Algorithm()): h}
				godebug.Line(ctx, scope, 433)
				target.Length = int64(pushResult.Size)
			}
		}
		if ctx, _ok := godebug.EnterFuncLit(fn); _ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
	}
	scope.Declare("handleTarget", &handleTarget)
	godebug.Line(ctx, scope, 439)

	if tag == "" {
		godebug.Line(ctx, scope, 440)
		if err = jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil); err != nil {
			godebug.Line(ctx, scope, 441)
			return err
		}
		godebug.Line(ctx, scope, 443)
		fmt.Fprintln(cli.out, "No tag specified, skipping trust metadata push")
		godebug.Line(ctx, scope, 444)
		return nil
	}
	godebug.Line(ctx, scope, 447)

	if err = jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, handleTarget); err != nil {
		godebug.Line(ctx, scope, 448)
		return err
	}
	godebug.Line(ctx, scope, 451)

	if cnt > 1 {
		godebug.Line(ctx, scope, 452)
		return fmt.Errorf("internal error: only one call to handleTarget expected")
	}
	godebug.Line(ctx, scope, 455)

	if target == nil {
		godebug.Line(ctx, scope, 456)
		fmt.Fprintln(cli.out, "No targets found, please provide a specific tag in order to sign it")
		godebug.Line(ctx, scope, 457)
		return nil
	}
	godebug.Line(ctx, scope, 460)

	fmt.Fprintln(cli.out, "Signing and pushing trust metadata")
	godebug.Line(ctx, scope, 462)

	repo, err := cli.getNotaryRepository(repoInfo, authConfig, "push", "pull")
	scope.Declare("repo", &repo)
	godebug.Line(ctx, scope, 463)
	if err != nil {
		godebug.Line(ctx, scope, 464)
		fmt.Fprintf(cli.out, "Error establishing connection to notary repository: %s\n", err)
		godebug.Line(ctx, scope, 465)
		return err
	}
	godebug.Line(ctx, scope, 469)

	_, err = repo.Update(false)
	godebug.Line(ctx, scope, 471)

	switch err.(type) {
	case client.ErrRepoNotInitialized, client.ErrRepositoryNotExist:
		godebug.Line(ctx, scope, 472)
		godebug.Line(ctx, scope, 473)
		keys := repo.CryptoService.ListKeys(data.CanonicalRootRole)
		scope := scope.EnteringNewChildScope()
		scope.Declare("keys", &keys)
		godebug.Line(ctx, scope, 474)
		var rootKeyID string
		scope.Declare("rootKeyID", &rootKeyID)
		godebug.Line(ctx, scope, 476)

		if len(keys) > 0 {
			godebug.Line(ctx, scope, 477)
			sort.Strings(keys)
			godebug.Line(ctx, scope, 478)
			rootKeyID = keys[0]
		} else {
			godebug.Line(ctx, scope, 479)
			godebug.Line(ctx, scope, 480)
			rootPublicKey, err := repo.CryptoService.Create(data.CanonicalRootRole, "", data.ECDSAKey)
			scope := scope.EnteringNewChildScope()
			scope.Declare("rootPublicKey", &rootPublicKey, "err", &err)
			godebug.Line(ctx, scope, 481)
			if err != nil {
				godebug.Line(ctx, scope, 482)
				return err
			}
			godebug.Line(ctx, scope, 484)
			rootKeyID = rootPublicKey.ID()
		}
		godebug.Line(ctx, scope, 488)

		if err := repo.Initialize(rootKeyID, data.CanonicalSnapshotRole); err != nil {
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 489)
			return notaryError(repoInfo.FullName(), err)
		}
		godebug.Line(ctx, scope, 491)
		fmt.Fprintf(cli.out, "Finished initializing %q\n", repoInfo.FullName())
		godebug.Line(ctx, scope, 492)
		err = repo.AddTarget(target, data.CanonicalTargetsRole)
	case nil:
		godebug.Line(ctx, scope, 493)
		godebug.Line(ctx, scope, 495)

		err = cli.addTargetToAllSignableRoles(repo, target)
	default:
		godebug.Line(ctx, scope, 496)
		godebug.Line(ctx, scope, 497)
		return notaryError(repoInfo.FullName(), err)
	}
	godebug.Line(ctx, scope, 500)

	if err == nil {
		godebug.Line(ctx, scope, 501)
		err = repo.Publish()
	}
	godebug.Line(ctx, scope, 504)

	if err != nil {
		godebug.Line(ctx, scope, 505)
		fmt.Fprintf(cli.out, "Failed to sign %q:%s - %s\n", repoInfo.FullName(), tag, err.Error())
		godebug.Line(ctx, scope, 506)
		return notaryError(repoInfo.FullName(), err)
	}
	godebug.Line(ctx, scope, 509)

	fmt.Fprintf(cli.out, "Successfully signed %q:%s\n", repoInfo.FullName(), tag)
	godebug.Line(ctx, scope, 510)
	return nil
}

func (cli *DockerCli) addTargetToAllSignableRoles(repo *client.NotaryRepository, target *client.Target) error {
	var result1 error
	ctx, _ok := godebug.EnterFunc(func() {
		result1 = cli.addTargetToAllSignableRoles(repo, target)
	})
	if !_ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "repo", &repo, "target", &target)
	godebug.Line(ctx, scope, 518)
	var signableRoles []string
	scope.Declare("signableRoles", &signableRoles)
	godebug.Line(ctx, scope, 521)

	allCanonicalKeyIDs := make(map[string]struct{})
	scope.Declare("allCanonicalKeyIDs", &allCanonicalKeyIDs)
	{
		scope := scope.EnteringNewChildScope()
		for fullKeyID := range repo.CryptoService.ListAllKeys() {
			godebug.Line(ctx, scope, 522)
			scope.Declare("fullKeyID", &fullKeyID)
			godebug.Line(ctx, scope, 523)
			allCanonicalKeyIDs[path.Base(fullKeyID)] = struct{}{}
		}
		godebug.Line(ctx, scope, 522)
	}
	godebug.Line(ctx, scope, 526)

	allDelegationRoles, err := repo.GetDelegationRoles()
	scope.Declare("allDelegationRoles", &allDelegationRoles, "err", &err)
	godebug.Line(ctx, scope, 527)
	if err != nil {
		godebug.Line(ctx, scope, 528)
		return err
	}
	godebug.Line(ctx, scope, 532)

	if len(allDelegationRoles) == 0 {
		godebug.Line(ctx, scope, 533)
		return repo.AddTarget(target, data.CanonicalTargetsRole)
	}
	{
		scope := scope.EnteringNewChildScope()

		for _, delegationRole := range allDelegationRoles {
			godebug.Line(ctx, scope, 538)
			scope.Declare("delegationRole", &delegationRole)
			godebug.Line(ctx, scope, 542)

			if path.Dir(delegationRole.Name) != data.CanonicalTargetsRole || !delegationRole.CheckPaths(target.Name) {
				godebug.Line(ctx, scope, 543)
				continue
			}
			{
				scope := scope.EnteringNewChildScope()

				for _, canonicalKeyID := range delegationRole.KeyIDs {
					godebug.Line(ctx, scope, 546)
					scope.Declare("canonicalKeyID", &canonicalKeyID)
					godebug.Line(ctx, scope, 547)
					if _, ok := allCanonicalKeyIDs[canonicalKeyID]; ok {
						scope := scope.EnteringNewChildScope()
						scope.Declare("ok", &ok)
						godebug.Line(ctx, scope, 548)
						signableRoles = append(signableRoles, delegationRole.Name)
						godebug.Line(ctx, scope, 549)
						break
					}
				}
				godebug.Line(ctx, scope, 546)
			}
		}
		godebug.Line(ctx, scope, 538)
	}
	godebug.Line(ctx, scope, 554)

	if len(signableRoles) == 0 {
		godebug.Line(ctx, scope, 555)
		return fmt.Errorf("no valid signing keys for delegation roles")
	}
	godebug.Line(ctx, scope, 558)

	return repo.AddTarget(target, signableRoles...)
}

var trust_go_contents = `package client

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/distribution"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	apiclient "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	registrytypes "github.com/docker/engine-api/types/registry"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/notary/client"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/store"
)

var (
	releasesRole = path.Join(data.CanonicalTargetsRole, "releases")
	untrusted    bool
)

func addTrustedFlags(fs *flag.FlagSet, verify bool) {
	var trusted bool
	if e := os.Getenv("DOCKER_CONTENT_TRUST"); e != "" {
		if t, err := strconv.ParseBool(e); t || err != nil {
			// treat any other value as true
			trusted = true
		}
	}
	message := "Skip image signing"
	if verify {
		message = "Skip image verification"
	}
	fs.BoolVar(&untrusted, []string{"-disable-content-trust"}, !trusted, message)
}

func isTrusted() bool {
	return !untrusted
}

type target struct {
	reference registry.Reference
	digest    digest.Digest
	size      int64
}

func (cli *DockerCli) trustDirectory() string {
	return filepath.Join(cliconfig.ConfigDir(), "trust")
}

// certificateDirectory returns the directory containing
// TLS certificates for the given server. An error is
// returned if there was an error parsing the server string.
func (cli *DockerCli) certificateDirectory(server string) (string, error) {
	u, err := url.Parse(server)
	if err != nil {
		return "", err
	}

	return filepath.Join(cliconfig.ConfigDir(), "tls", u.Host), nil
}

func trustServer(index *registrytypes.IndexInfo) (string, error) {
	if s := os.Getenv("DOCKER_CONTENT_TRUST_SERVER"); s != "" {
		urlObj, err := url.Parse(s)
		if err != nil || urlObj.Scheme != "https" {
			return "", fmt.Errorf("valid https URL required for trust server, got %s", s)
		}

		return s, nil
	}
	if index.Official {
		return registry.NotaryServer, nil
	}
	return "https://" + index.Name, nil
}

type simpleCredentialStore struct {
	auth types.AuthConfig
}

func (scs simpleCredentialStore) Basic(u *url.URL) (string, string) {
	return scs.auth.Username, scs.auth.Password
}

func (scs simpleCredentialStore) RefreshToken(u *url.URL, service string) string {
	return scs.auth.IdentityToken
}

func (scs simpleCredentialStore) SetRefreshToken(*url.URL, string, string) {
}

// getNotaryRepository returns a NotaryRepository which stores all the
// information needed to operate on a notary repository.
// It creates a HTTP transport providing authentication support.
func (cli *DockerCli) getNotaryRepository(repoInfo *registry.RepositoryInfo, authConfig types.AuthConfig, actions ...string) (*client.NotaryRepository, error) {
	server, err := trustServer(repoInfo.Index)
	if err != nil {
		return nil, err
	}

	var cfg = tlsconfig.ClientDefault
	cfg.InsecureSkipVerify = !repoInfo.Index.Secure

	// Get certificate base directory
	certDir, err := cli.certificateDirectory(server)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading certificate directory: %s", certDir)

	if err := registry.ReadCertsDirectory(&cfg, certDir); err != nil {
		return nil, err
	}

	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &cfg,
		DisableKeepAlives:   true,
	}

	// Skip configuration headers since request is not going to Docker daemon
	modifiers := registry.DockerHeaders(clientUserAgent(), http.Header{})
	authTransport := transport.NewTransport(base, modifiers...)
	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   5 * time.Second,
	}
	endpointStr := server + "/v2/"
	req, err := http.NewRequest("GET", endpointStr, nil)
	if err != nil {
		return nil, err
	}

	challengeManager := auth.NewSimpleChallengeManager()

	resp, err := pingClient.Do(req)
	if err != nil {
		// Ignore error on ping to operate in offline mode
		logrus.Debugf("Error pinging notary server %q: %s", endpointStr, err)
	} else {
		defer resp.Body.Close()

		// Add response to the challenge manager to parse out
		// authentication header and register authentication method
		if err := challengeManager.AddResponse(resp); err != nil {
			return nil, err
		}
	}

	creds := simpleCredentialStore{auth: authConfig}
	tokenHandler := auth.NewTokenHandler(authTransport, creds, repoInfo.FullName(), actions...)
	basicHandler := auth.NewBasicHandler(creds)
	modifiers = append(modifiers, transport.RequestModifier(auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler)))
	tr := transport.NewTransport(base, modifiers...)

	return client.NewNotaryRepository(cli.trustDirectory(), repoInfo.FullName(), server, tr, cli.getPassphraseRetriever())
}

func convertTarget(t client.Target) (target, error) {
	h, ok := t.Hashes["sha256"]
	if !ok {
		return target{}, errors.New("no valid hash, expecting sha256")
	}
	return target{
		reference: registry.ParseReference(t.Name),
		digest:    digest.NewDigestFromHex("sha256", hex.EncodeToString(h)),
		size:      t.Length,
	}, nil
}

func (cli *DockerCli) getPassphraseRetriever() passphrase.Retriever {
	aliasMap := map[string]string{
		"root":     "root",
		"snapshot": "repository",
		"targets":  "repository",
		"default":  "repository",
	}
	baseRetriever := passphrase.PromptRetrieverWithInOut(cli.in, cli.out, aliasMap)
	env := map[string]string{
		"root":     os.Getenv("DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE"),
		"snapshot": os.Getenv("DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE"),
		"targets":  os.Getenv("DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE"),
		"default":  os.Getenv("DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE"),
	}

	// Backwards compatibility with old env names. We should remove this in 1.10
	if env["root"] == "" {
		if passphrase := os.Getenv("DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE"); passphrase != "" {
			env["root"] = passphrase
			fmt.Fprintf(cli.err, "[DEPRECATED] The environment variable DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE has been deprecated and will be removed in v1.10. Please use DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE\n")
		}
	}
	if env["snapshot"] == "" || env["targets"] == "" || env["default"] == "" {
		if passphrase := os.Getenv("DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE"); passphrase != "" {
			env["snapshot"] = passphrase
			env["targets"] = passphrase
			env["default"] = passphrase
			fmt.Fprintf(cli.err, "[DEPRECATED] The environment variable DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE has been deprecated and will be removed in v1.10. Please use DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE\n")
		}
	}

	return func(keyName string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		if v := env[alias]; v != "" {
			return v, numAttempts > 1, nil
		}
		// For non-root roles, we can also try the "default" alias if it is specified
		if v := env["default"]; v != "" && alias != data.CanonicalRootRole {
			return v, numAttempts > 1, nil
		}
		return baseRetriever(keyName, alias, createNew, numAttempts)
	}
}

func (cli *DockerCli) trustedReference(ref reference.NamedTagged) (reference.Canonical, error) {
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return nil, err
	}

	// Resolve the Auth config relevant for this server
	authConfig := cli.resolveAuthConfig(repoInfo.Index)

	notaryRepo, err := cli.getNotaryRepository(repoInfo, authConfig, "pull")
	if err != nil {
		fmt.Fprintf(cli.out, "Error establishing connection to trust repository: %s\n", err)
		return nil, err
	}

	t, err := notaryRepo.GetTargetByName(ref.Tag(), releasesRole, data.CanonicalTargetsRole)
	if err != nil {
		return nil, err
	}
	// Only list tags in the top level targets role or the releases delegation role - ignore
	// all other delegation roles
	if t.Role != releasesRole && t.Role != data.CanonicalTargetsRole {
		return nil, notaryError(repoInfo.FullName(), fmt.Errorf("No trust data for %s", ref.Tag()))
	}
	r, err := convertTarget(t.Target)
	if err != nil {
		return nil, err

	}

	return reference.WithDigest(ref, r.digest)
}

func (cli *DockerCli) tagTrusted(trustedRef reference.Canonical, ref reference.NamedTagged) error {
	fmt.Fprintf(cli.out, "Tagging %s as %s\n", trustedRef.String(), ref.String())

	options := types.ImageTagOptions{
		ImageID:        trustedRef.String(),
		RepositoryName: trustedRef.Name(),
		Tag:            ref.Tag(),
		Force:          true,
	}

	return cli.client.ImageTag(context.Background(), options)
}

func notaryError(repoName string, err error) error {
	switch err.(type) {
	case *json.SyntaxError:
		logrus.Debugf("Notary syntax error: %s", err)
		return fmt.Errorf("Error: no trust data available for remote repository %s. Try running notary server and setting DOCKER_CONTENT_TRUST_SERVER to its HTTPS address?", repoName)
	case signed.ErrExpired:
		return fmt.Errorf("Error: remote repository %s out-of-date: %v", repoName, err)
	case trustmanager.ErrKeyNotFound:
		return fmt.Errorf("Error: signing keys for remote repository %s not found: %v", repoName, err)
	case *net.OpError:
		return fmt.Errorf("Error: error contacting notary server: %v", err)
	case store.ErrMetaNotFound:
		return fmt.Errorf("Error: trust data missing for remote repository %s or remote repository not found: %v", repoName, err)
	case signed.ErrInvalidKeyType:
		return fmt.Errorf("Warning: potential malicious behavior - trust data mismatch for remote repository %s: %v", repoName, err)
	case signed.ErrNoKeys:
		return fmt.Errorf("Error: could not find signing keys for remote repository %s, or could not decrypt signing key: %v", repoName, err)
	case signed.ErrLowVersion:
		return fmt.Errorf("Warning: potential malicious behavior - trust data version is lower than expected for remote repository %s: %v", repoName, err)
	case signed.ErrRoleThreshold:
		return fmt.Errorf("Warning: potential malicious behavior - trust data has insufficient signatures for remote repository %s: %v", repoName, err)
	case client.ErrRepositoryNotExist:
		return fmt.Errorf("Error: remote trust data does not exist for %s: %v", repoName, err)
	case signed.ErrInsufficientSignatures:
		return fmt.Errorf("Error: could not produce valid signature for %s.  If Yubikey was used, was touch input provided?: %v", repoName, err)
	}

	return err
}

func (cli *DockerCli) trustedPull(repoInfo *registry.RepositoryInfo, ref registry.Reference, authConfig types.AuthConfig, requestPrivilege apiclient.RequestPrivilegeFunc) error {
	var refs []target

	notaryRepo, err := cli.getNotaryRepository(repoInfo, authConfig, "pull")
	if err != nil {
		fmt.Fprintf(cli.out, "Error establishing connection to trust repository: %s\n", err)
		return err
	}

	if ref.String() == "" {
		// List all targets
		targets, err := notaryRepo.ListTargets(releasesRole, data.CanonicalTargetsRole)
		if err != nil {
			return notaryError(repoInfo.FullName(), err)
		}
		for _, tgt := range targets {
			t, err := convertTarget(tgt.Target)
			if err != nil {
				fmt.Fprintf(cli.out, "Skipping target for %q\n", repoInfo.Name())
				continue
			}
			// Only list tags in the top level targets role or the releases delegation role - ignore
			// all other delegation roles
			if tgt.Role != releasesRole && tgt.Role != data.CanonicalTargetsRole {
				continue
			}
			refs = append(refs, t)
		}
		if len(refs) == 0 {
			return notaryError(repoInfo.FullName(), fmt.Errorf("No trusted tags for %s", repoInfo.FullName()))
		}
	} else {
		t, err := notaryRepo.GetTargetByName(ref.String(), releasesRole, data.CanonicalTargetsRole)
		if err != nil {
			return notaryError(repoInfo.FullName(), err)
		}
		// Only get the tag if it's in the top level targets role or the releases delegation role
		// ignore it if it's in any other delegation roles
		if t.Role != releasesRole && t.Role != data.CanonicalTargetsRole {
			return notaryError(repoInfo.FullName(), fmt.Errorf("No trust data for %s", ref.String()))
		}

		logrus.Debugf("retrieving target for %s role\n", t.Role)
		r, err := convertTarget(t.Target)
		if err != nil {
			return err

		}
		refs = append(refs, r)
	}

	for i, r := range refs {
		displayTag := r.reference.String()
		if displayTag != "" {
			displayTag = ":" + displayTag
		}
		fmt.Fprintf(cli.out, "Pull (%d of %d): %s%s@%s\n", i+1, len(refs), repoInfo.Name(), displayTag, r.digest)

		if err := cli.imagePullPrivileged(authConfig, repoInfo.Name(), r.digest.String(), requestPrivilege); err != nil {
			return err
		}

		// If reference is not trusted, tag by trusted reference
		if !r.reference.HasDigest() {
			tagged, err := reference.WithTag(repoInfo, r.reference.String())
			if err != nil {
				return err
			}
			trustedRef, err := reference.WithDigest(repoInfo, r.digest)
			if err != nil {
				return err
			}
			if err := cli.tagTrusted(trustedRef, tagged); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cli *DockerCli) trustedPush(repoInfo *registry.RepositoryInfo, tag string, authConfig types.AuthConfig, requestPrivilege apiclient.RequestPrivilegeFunc) error {
	responseBody, err := cli.imagePushPrivileged(authConfig, repoInfo.Name(), tag, requestPrivilege)
	if err != nil {
		return err
	}

	defer responseBody.Close()

	// If it is a trusted push we would like to find the target entry which match the
	// tag provided in the function and then do an AddTarget later.
	target := &client.Target{}
	// Count the times of calling for handleTarget,
	// if it is called more that once, that should be considered an error in a trusted push.
	cnt := 0
	handleTarget := func(aux *json.RawMessage) {
		cnt++
		if cnt > 1 {
			// handleTarget should only be called one. This will be treated as an error.
			return
		}

		var pushResult distribution.PushResult
		err := json.Unmarshal(*aux, &pushResult)
		if err == nil && pushResult.Tag != "" && pushResult.Digest.Validate() == nil {
			h, err := hex.DecodeString(pushResult.Digest.Hex())
			if err != nil {
				target = nil
				return
			}
			target.Name = registry.ParseReference(pushResult.Tag).String()
			target.Hashes = data.Hashes{string(pushResult.Digest.Algorithm()): h}
			target.Length = int64(pushResult.Size)
		}
	}

	// We want trust signatures to always take an explicit tag,
	// otherwise it will act as an untrusted push.
	if tag == "" {
		if err = jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil); err != nil {
			return err
		}
		fmt.Fprintln(cli.out, "No tag specified, skipping trust metadata push")
		return nil
	}

	if err = jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, handleTarget); err != nil {
		return err
	}

	if cnt > 1 {
		return fmt.Errorf("internal error: only one call to handleTarget expected")
	}

	if target == nil {
		fmt.Fprintln(cli.out, "No targets found, please provide a specific tag in order to sign it")
		return nil
	}

	fmt.Fprintln(cli.out, "Signing and pushing trust metadata")

	repo, err := cli.getNotaryRepository(repoInfo, authConfig, "push", "pull")
	if err != nil {
		fmt.Fprintf(cli.out, "Error establishing connection to notary repository: %s\n", err)
		return err
	}

	// get the latest repository metadata so we can figure out which roles to sign
	_, err = repo.Update(false)

	switch err.(type) {
	case client.ErrRepoNotInitialized, client.ErrRepositoryNotExist:
		keys := repo.CryptoService.ListKeys(data.CanonicalRootRole)
		var rootKeyID string
		// always select the first root key
		if len(keys) > 0 {
			sort.Strings(keys)
			rootKeyID = keys[0]
		} else {
			rootPublicKey, err := repo.CryptoService.Create(data.CanonicalRootRole, "", data.ECDSAKey)
			if err != nil {
				return err
			}
			rootKeyID = rootPublicKey.ID()
		}

		// Initialize the notary repository with a remotely managed snapshot key
		if err := repo.Initialize(rootKeyID, data.CanonicalSnapshotRole); err != nil {
			return notaryError(repoInfo.FullName(), err)
		}
		fmt.Fprintf(cli.out, "Finished initializing %q\n", repoInfo.FullName())
		err = repo.AddTarget(target, data.CanonicalTargetsRole)
	case nil:
		// already initialized and we have successfully downloaded the latest metadata
		err = cli.addTargetToAllSignableRoles(repo, target)
	default:
		return notaryError(repoInfo.FullName(), err)
	}

	if err == nil {
		err = repo.Publish()
	}

	if err != nil {
		fmt.Fprintf(cli.out, "Failed to sign %q:%s - %s\n", repoInfo.FullName(), tag, err.Error())
		return notaryError(repoInfo.FullName(), err)
	}

	fmt.Fprintf(cli.out, "Successfully signed %q:%s\n", repoInfo.FullName(), tag)
	return nil
}

// Attempt to add the image target to all the top level delegation roles we can
// (based on whether we have the signing key and whether the role's path allows
// us to).
// If there are no delegation roles, we add to the targets role.
func (cli *DockerCli) addTargetToAllSignableRoles(repo *client.NotaryRepository, target *client.Target) error {
	var signableRoles []string

	// translate the full key names, which includes the GUN, into just the key IDs
	allCanonicalKeyIDs := make(map[string]struct{})
	for fullKeyID := range repo.CryptoService.ListAllKeys() {
		allCanonicalKeyIDs[path.Base(fullKeyID)] = struct{}{}
	}

	allDelegationRoles, err := repo.GetDelegationRoles()
	if err != nil {
		return err
	}

	// if there are no delegation roles, then just try to sign it into the targets role
	if len(allDelegationRoles) == 0 {
		return repo.AddTarget(target, data.CanonicalTargetsRole)
	}

	// there are delegation roles, find every delegation role we have a key for, and
	// attempt to sign into into all those roles.
	for _, delegationRole := range allDelegationRoles {
		// We do not support signing any delegation role that isn't a direct child of the targets role.
		// Also don't bother checking the keys if we can't add the target
		// to this role due to path restrictions
		if path.Dir(delegationRole.Name) != data.CanonicalTargetsRole || !delegationRole.CheckPaths(target.Name) {
			continue
		}

		for _, canonicalKeyID := range delegationRole.KeyIDs {
			if _, ok := allCanonicalKeyIDs[canonicalKeyID]; ok {
				signableRoles = append(signableRoles, delegationRole.Name)
				break
			}
		}
	}

	if len(signableRoles) == 0 {
		return fmt.Errorf("no valid signing keys for delegation roles")
	}

	return repo.AddTarget(target, signableRoles...)
}
`


var unpause_go_scope = godebug.EnteringNewFile(client_pkg_scope, unpause_go_contents)

func (cli *DockerCli) CmdUnpause(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdUnpause(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := unpause_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 17)
	cmd := Cli.Subcmd("unpause", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["unpause"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 18)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 20)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 22)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 23)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 24)
			if err := cli.client.ContainerUnpause(context.Background(), name); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 25)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 26)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 27)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 23)
	}
	godebug.Line(ctx, scope, 30)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 31)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 33)
	return nil
}

var unpause_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdUnpause unpauses all processes within a container, for one or more containers.
//
// Usage: docker unpause CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdUnpause(args ...string) error {
	cmd := Cli.Subcmd("unpause", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["unpause"].Description, true)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		if err := cli.client.ContainerUnpause(context.Background(), name); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var update_go_scope = godebug.EnteringNewFile(client_pkg_scope, update_go_contents)

func (cli *DockerCli) CmdUpdate(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdUpdate(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := update_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 20)
	cmd := Cli.Subcmd("update", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["update"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 21)
	flBlkioWeight := cmd.Uint16([]string{"-blkio-weight"}, 0, "Block IO (relative weight), between 10 and 1000")
	scope.Declare("flBlkioWeight", &flBlkioWeight)
	godebug.Line(ctx, scope, 22)
	flCPUPeriod := cmd.Int64([]string{"-cpu-period"}, 0, "Limit CPU CFS (Completely Fair Scheduler) period")
	scope.Declare("flCPUPeriod", &flCPUPeriod)
	godebug.Line(ctx, scope, 23)
	flCPUQuota := cmd.Int64([]string{"-cpu-quota"}, 0, "Limit CPU CFS (Completely Fair Scheduler) quota")
	scope.Declare("flCPUQuota", &flCPUQuota)
	godebug.Line(ctx, scope, 24)
	flCpusetCpus := cmd.String([]string{"-cpuset-cpus"}, "", "CPUs in which to allow execution (0-3, 0,1)")
	scope.Declare("flCpusetCpus", &flCpusetCpus)
	godebug.Line(ctx, scope, 25)
	flCpusetMems := cmd.String([]string{"-cpuset-mems"}, "", "MEMs in which to allow execution (0-3, 0,1)")
	scope.Declare("flCpusetMems", &flCpusetMems)
	godebug.Line(ctx, scope, 26)
	flCPUShares := cmd.Int64([]string{"#c", "-cpu-shares"}, 0, "CPU shares (relative weight)")
	scope.Declare("flCPUShares", &flCPUShares)
	godebug.Line(ctx, scope, 27)
	flMemoryString := cmd.String([]string{"m", "-memory"}, "", "Memory limit")
	scope.Declare("flMemoryString", &flMemoryString)
	godebug.Line(ctx, scope, 28)
	flMemoryReservation := cmd.String([]string{"-memory-reservation"}, "", "Memory soft limit")
	scope.Declare("flMemoryReservation", &flMemoryReservation)
	godebug.Line(ctx, scope, 29)
	flMemorySwap := cmd.String([]string{"-memory-swap"}, "", "Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	scope.Declare("flMemorySwap", &flMemorySwap)
	godebug.Line(ctx, scope, 30)
	flKernelMemory := cmd.String([]string{"-kernel-memory"}, "", "Kernel memory limit")
	scope.Declare("flKernelMemory", &flKernelMemory)
	godebug.Line(ctx, scope, 31)
	flRestartPolicy := cmd.String([]string{"-restart"}, "", "Restart policy to apply when a container exits")
	scope.Declare("flRestartPolicy", &flRestartPolicy)
	godebug.Line(ctx, scope, 33)

	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 34)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 35)
	if cmd.NFlag() == 0 {
		godebug.Line(ctx, scope, 36)
		return fmt.Errorf("You must provide one or more flags when using this command.")
	}
	godebug.Line(ctx, scope, 39)

	var err error
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 40)
	var flMemory int64
	scope.Declare("flMemory", &flMemory)
	godebug.Line(ctx, scope, 41)
	if *flMemoryString != "" {
		godebug.Line(ctx, scope, 42)
		flMemory, err = units.RAMInBytes(*flMemoryString)
		godebug.Line(ctx, scope, 43)
		if err != nil {
			godebug.Line(ctx, scope, 44)
			return err
		}
	}
	godebug.Line(ctx, scope, 48)

	var memoryReservation int64
	scope.Declare("memoryReservation", &memoryReservation)
	godebug.Line(ctx, scope, 49)
	if *flMemoryReservation != "" {
		godebug.Line(ctx, scope, 50)
		memoryReservation, err = units.RAMInBytes(*flMemoryReservation)
		godebug.Line(ctx, scope, 51)
		if err != nil {
			godebug.Line(ctx, scope, 52)
			return err
		}
	}
	godebug.Line(ctx, scope, 56)

	var memorySwap int64
	scope.Declare("memorySwap", &memorySwap)
	godebug.Line(ctx, scope, 57)
	if *flMemorySwap != "" {
		godebug.Line(ctx, scope, 58)
		if *flMemorySwap == "-1" {
			godebug.Line(ctx, scope, 59)
			memorySwap = -1
		} else {
			godebug.Line(ctx, scope, 60)
			godebug.Line(ctx, scope, 61)
			memorySwap, err = units.RAMInBytes(*flMemorySwap)
			godebug.Line(ctx, scope, 62)
			if err != nil {
				godebug.Line(ctx, scope, 63)
				return err
			}
		}
	}
	godebug.Line(ctx, scope, 68)

	var kernelMemory int64
	scope.Declare("kernelMemory", &kernelMemory)
	godebug.Line(ctx, scope, 69)
	if *flKernelMemory != "" {
		godebug.Line(ctx, scope, 70)
		kernelMemory, err = units.RAMInBytes(*flKernelMemory)
		godebug.Line(ctx, scope, 71)
		if err != nil {
			godebug.Line(ctx, scope, 72)
			return err
		}
	}
	godebug.Line(ctx, scope, 76)

	var restartPolicy container.RestartPolicy
	scope.Declare("restartPolicy", &restartPolicy)
	godebug.Line(ctx, scope, 77)
	if *flRestartPolicy != "" {
		godebug.Line(ctx, scope, 78)
		restartPolicy, err = runconfigopts.ParseRestartPolicy(*flRestartPolicy)
		godebug.Line(ctx, scope, 79)
		if err != nil {
			godebug.Line(ctx, scope, 80)
			return err
		}
	}
	godebug.Line(ctx, scope, 84)

	resources := container.Resources{
		BlkioWeight:       *flBlkioWeight,
		CpusetCpus:        *flCpusetCpus,
		CpusetMems:        *flCpusetMems,
		CPUShares:         *flCPUShares,
		Memory:            flMemory,
		MemoryReservation: memoryReservation,
		MemorySwap:        memorySwap,
		KernelMemory:      kernelMemory,
		CPUPeriod:         *flCPUPeriod,
		CPUQuota:          *flCPUQuota,
	}
	scope.Declare("resources", &resources)
	godebug.Line(ctx, scope, 97)

	updateConfig := container.UpdateConfig{
		Resources:     resources,
		RestartPolicy: restartPolicy,
	}
	scope.Declare("updateConfig", &updateConfig)
	godebug.Line(ctx, scope, 102)

	names := cmd.Args()
	scope.Declare("names", &names)
	godebug.Line(ctx, scope, 103)
	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range names {
			godebug.Line(ctx, scope, 104)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 105)
			if err := cli.client.ContainerUpdate(context.Background(), name, updateConfig); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 106)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 107)
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 108)

				fmt.Fprintf(cli.out, "%s\n", name)
			}
		}
		godebug.Line(ctx, scope, 104)
	}
	godebug.Line(ctx, scope, 112)

	if len(errs) > 0 {
		godebug.Line(ctx, scope, 113)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 116)

	return nil
}

var update_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/go-units"
)

// CmdUpdate updates resources of one or more containers.
//
// Usage: docker update [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdUpdate(args ...string) error {
	cmd := Cli.Subcmd("update", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["update"].Description, true)
	flBlkioWeight := cmd.Uint16([]string{"-blkio-weight"}, 0, "Block IO (relative weight), between 10 and 1000")
	flCPUPeriod := cmd.Int64([]string{"-cpu-period"}, 0, "Limit CPU CFS (Completely Fair Scheduler) period")
	flCPUQuota := cmd.Int64([]string{"-cpu-quota"}, 0, "Limit CPU CFS (Completely Fair Scheduler) quota")
	flCpusetCpus := cmd.String([]string{"-cpuset-cpus"}, "", "CPUs in which to allow execution (0-3, 0,1)")
	flCpusetMems := cmd.String([]string{"-cpuset-mems"}, "", "MEMs in which to allow execution (0-3, 0,1)")
	flCPUShares := cmd.Int64([]string{"#c", "-cpu-shares"}, 0, "CPU shares (relative weight)")
	flMemoryString := cmd.String([]string{"m", "-memory"}, "", "Memory limit")
	flMemoryReservation := cmd.String([]string{"-memory-reservation"}, "", "Memory soft limit")
	flMemorySwap := cmd.String([]string{"-memory-swap"}, "", "Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	flKernelMemory := cmd.String([]string{"-kernel-memory"}, "", "Kernel memory limit")
	flRestartPolicy := cmd.String([]string{"-restart"}, "", "Restart policy to apply when a container exits")

	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)
	if cmd.NFlag() == 0 {
		return fmt.Errorf("You must provide one or more flags when using this command.")
	}

	var err error
	var flMemory int64
	if *flMemoryString != "" {
		flMemory, err = units.RAMInBytes(*flMemoryString)
		if err != nil {
			return err
		}
	}

	var memoryReservation int64
	if *flMemoryReservation != "" {
		memoryReservation, err = units.RAMInBytes(*flMemoryReservation)
		if err != nil {
			return err
		}
	}

	var memorySwap int64
	if *flMemorySwap != "" {
		if *flMemorySwap == "-1" {
			memorySwap = -1
		} else {
			memorySwap, err = units.RAMInBytes(*flMemorySwap)
			if err != nil {
				return err
			}
		}
	}

	var kernelMemory int64
	if *flKernelMemory != "" {
		kernelMemory, err = units.RAMInBytes(*flKernelMemory)
		if err != nil {
			return err
		}
	}

	var restartPolicy container.RestartPolicy
	if *flRestartPolicy != "" {
		restartPolicy, err = opts.ParseRestartPolicy(*flRestartPolicy)
		if err != nil {
			return err
		}
	}

	resources := container.Resources{
		BlkioWeight:       *flBlkioWeight,
		CpusetCpus:        *flCpusetCpus,
		CpusetMems:        *flCpusetMems,
		CPUShares:         *flCPUShares,
		Memory:            flMemory,
		MemoryReservation: memoryReservation,
		MemorySwap:        memorySwap,
		KernelMemory:      kernelMemory,
		CPUPeriod:         *flCPUPeriod,
		CPUQuota:          *flCPUQuota,
	}

	updateConfig := container.UpdateConfig{
		Resources:     resources,
		RestartPolicy: restartPolicy,
	}

	names := cmd.Args()
	var errs []string
	for _, name := range names {
		if err := cli.client.ContainerUpdate(context.Background(), name, updateConfig); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}

	return nil
}
`


var utils_go_scope = godebug.EnteringNewFile(client_pkg_scope, utils_go_contents)

func (cli *DockerCli) electAuthServer() string {
	var result1 string
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.electAuthServer()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 31)

	serverAddress := registry.IndexServer
	scope.Declare("serverAddress", &serverAddress)
	godebug.Line(ctx, scope, 32)
	if info, err := cli.client.Info(context.Background()); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("info", &info, "err", &err)
		godebug.Line(ctx, scope, 33)
		fmt.Fprintf(cli.out, "Warning: failed to get default registry endpoint from daemon (%v). Using system default: %s\n", err, serverAddress)
	} else {
		godebug.Line(ctx, scope, 34)
		scope := scope.EnteringNewChildScope()
		scope.Declare("info", &info, "err", &err)
		godebug.Line(ctx, scope, 35)

		serverAddress = info.IndexServerAddress
	}
	godebug.Line(ctx, scope, 37)
	return serverAddress
}

func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	var result1 string
	var result2 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = encodeAuthToBase64(authConfig)
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("authConfig", &authConfig)
	godebug.Line(ctx, scope, 42)
	buf, err := json.Marshal(authConfig)
	scope.Declare("buf", &buf, "err", &err)
	godebug.Line(ctx, scope, 43)
	if err != nil {
		godebug.Line(ctx, scope, 44)
		return "", err
	}
	godebug.Line(ctx, scope, 46)
	return base64.URLEncoding.EncodeToString(buf), nil
}

func (cli *DockerCli) registryAuthenticationPrivilegedFunc(index *registrytypes.IndexInfo, cmdName string) apiclient.RequestPrivilegeFunc {
	var result1 apiclient.RequestPrivilegeFunc
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.registryAuthenticationPrivilegedFunc(index, cmdName)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "index", &index, "cmdName", &cmdName)
	godebug.Line(ctx, scope, 50)
	return func() (string, error) {
		var result1 string
		var result2 error
		fn := func(ctx *godebug.Context) {
			result1, result2 = func() (string, error) {
				godebug.Line(ctx, scope, 51)
				fmt.Fprintf(cli.out, "\nPlease login prior to %s:\n", cmdName)
				godebug.Line(ctx, scope, 52)
				indexServer := registry.GetAuthConfigKey(index)
				scope := scope.EnteringNewChildScope()
				scope.Declare("indexServer", &indexServer)
				godebug.Line(ctx, scope, 53)
				authConfig, err := cli.configureAuth("", "", indexServer, false)
				scope.Declare("authConfig", &authConfig, "err", &err)
				godebug.Line(ctx, scope, 54)
				if err != nil {
					godebug.Line(ctx, scope, 55)
					return "", err
				}
				godebug.Line(ctx, scope, 57)
				return encodeAuthToBase64(authConfig)
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2
	}

}

func (cli *DockerCli) resizeTty(id string, isExec bool) {
	ctx, ok := godebug.EnterFunc(func() {
		cli.resizeTty(id, isExec)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "id", &id, "isExec", &isExec)
	godebug.Line(ctx, scope, 62)
	height, width := cli.getTtySize()
	scope.Declare("height", &height, "width", &width)
	godebug.Line(ctx, scope, 63)
	cli.resizeTtyTo(id, height, width, isExec)
}

func (cli *DockerCli) resizeTtyTo(id string, height, width int, isExec bool) {
	ctx, ok := godebug.EnterFunc(func() {
		cli.resizeTtyTo(id, height, width, isExec)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "id", &id, "height", &height, "width", &width, "isExec", &isExec)
	godebug.Line(ctx, scope, 67)
	if height == 0 && width == 0 {
		godebug.Line(ctx, scope, 68)
		return
	}
	godebug.Line(ctx, scope, 71)

	options := types.ResizeOptions{
		ID:     id,
		Height: height,
		Width:  width,
	}
	scope.Declare("options", &options)
	godebug.Line(ctx, scope, 77)

	var err error
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 78)
	if isExec {
		godebug.Line(ctx, scope, 79)
		err = cli.client.ContainerExecResize(context.Background(), options)
	} else {
		godebug.Line(ctx, scope, 80)
		godebug.Line(ctx, scope, 81)
		err = cli.client.ContainerResize(context.Background(), options)
	}
	godebug.Line(ctx, scope, 84)

	if err != nil {
		godebug.Line(ctx, scope, 85)
		logrus.Debugf("Error resize: %s", err)
	}
}

func getExitCode(cli *DockerCli, containerID string) (bool, int, error) {
	var result1 bool
	var result2 int
	var result3 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2, result3 = getExitCode(cli, containerID)
	})
	if !ok {
		return result1, result2, result3
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "containerID", &containerID)
	godebug.Line(ctx, scope, 92)
	c, err := cli.client.ContainerInspect(context.Background(), containerID)
	scope.Declare("c", &c, "err", &err)
	godebug.Line(ctx, scope, 93)
	if err != nil {
		godebug.Line(ctx, scope, 95)

		if err != apiclient.ErrConnectionFailed {
			godebug.Line(ctx, scope, 96)
			return false, -1, err
		}
		godebug.Line(ctx, scope, 98)
		return false, -1, nil
	}
	godebug.Line(ctx, scope, 101)

	return c.State.Running, c.State.ExitCode, nil
}

func getExecExitCode(cli *DockerCli, execID string) (bool, int, error) {
	var result1 bool
	var result2 int
	var result3 error
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2, result3 = getExecExitCode(cli, execID)
	})
	if !ok {
		return result1, result2, result3
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "execID", &execID)
	godebug.Line(ctx, scope, 107)
	resp, err := cli.client.ContainerExecInspect(context.Background(), execID)
	scope.Declare("resp", &resp, "err", &err)
	godebug.Line(ctx, scope, 108)
	if err != nil {
		godebug.Line(ctx, scope, 110)

		if err != apiclient.ErrConnectionFailed {
			godebug.Line(ctx, scope, 111)
			return false, -1, err
		}
		godebug.Line(ctx, scope, 113)
		return false, -1, nil
	}
	godebug.Line(ctx, scope, 116)

	return resp.Running, resp.ExitCode, nil
}

func (cli *DockerCli) monitorTtySize(id string, isExec bool) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.monitorTtySize(id, isExec)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "id", &id, "isExec", &isExec)
	godebug.Line(ctx, scope, 120)
	cli.resizeTty(id, isExec)
	godebug.Line(ctx, scope, 122)

	if runtime.GOOS == "windows" {
		godebug.Line(ctx, scope, 123)
		go func() {
			fn := func(ctx *godebug.Context) {
				godebug.Line(ctx, scope, 124)
				prevH, prevW := cli.getTtySize()
				scope := scope.EnteringNewChildScope()
				scope.Declare("prevH", &prevH, "prevW", &prevW)
				godebug.Line(ctx, scope, 125)
				for {
					godebug.Line(ctx, scope, 126)
					time.Sleep(time.Millisecond * 250)
					godebug.Line(ctx, scope, 127)
					h, w := cli.getTtySize()
					scope := scope.EnteringNewChildScope()
					scope.Declare("h", &h, "w", &w)
					godebug.Line(ctx, scope, 129)
					if prevW != w || prevH != h {
						godebug.Line(ctx, scope, 130)
						cli.resizeTty(id, isExec)
					}
					godebug.Line(ctx, scope, 132)
					prevH = h
					godebug.Line(ctx, scope, 133)
					prevW = w
					godebug.Line(ctx, scope, 125)
				}
			}
			if ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(ctx)
				fn(ctx)
			}
		}()
	} else {
		godebug.Line(ctx, scope, 136)
		godebug.Line(ctx, scope, 137)
		sigchan := make(chan os.Signal, 1)
		scope := scope.EnteringNewChildScope()
		scope.Declare("sigchan", &sigchan)
		godebug.Line(ctx, scope, 138)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		godebug.Line(ctx, scope, 139)
		go func() {
			fn := func(ctx *godebug.Context) {
				godebug.Line(ctx, scope, 140)
				for range sigchan {
					godebug.Line(ctx, scope, 141)
					cli.resizeTty(id, isExec)
					godebug.Line(ctx, scope, 140)
				}
			}
			if ctx, ok := godebug.EnterFuncLit(fn); ok {
				defer godebug.ExitFunc(ctx)
				fn(ctx)
			}
		}()
	}
	godebug.Line(ctx, scope, 145)
	return nil
}

func (cli *DockerCli) getTtySize() (int, int) {
	var result1 int
	var result2 int
	ctx, ok := godebug.EnterFunc(func() {
		result1, result2 = cli.getTtySize()
	})
	if !ok {
		return result1, result2
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 149)
	if !cli.isTerminalOut {
		godebug.Line(ctx, scope, 150)
		return 0, 0
	}
	godebug.Line(ctx, scope, 152)
	ws, err := term.GetWinsize(cli.outFd)
	scope.Declare("ws", &ws, "err", &err)
	godebug.Line(ctx, scope, 153)
	if err != nil {
		godebug.Line(ctx, scope, 154)
		logrus.Debugf("Error getting size: %s", err)
		godebug.Line(ctx, scope, 155)
		if ws == nil {
			godebug.Line(ctx, scope, 156)
			return 0, 0
		}
	}
	godebug.Line(ctx, scope, 159)
	return int(ws.Height), int(ws.Width)
}

func copyToFile(outfile string, r io.Reader) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = copyToFile(outfile, r)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("outfile", &outfile, "r", &r)
	godebug.Line(ctx, scope, 163)
	tmpFile, err := ioutil.TempFile(filepath.Dir(outfile), ".docker_temp_")
	scope.Declare("tmpFile", &tmpFile, "err", &err)
	godebug.Line(ctx, scope, 164)
	if err != nil {
		godebug.Line(ctx, scope, 165)
		return err
	}
	godebug.Line(ctx, scope, 168)

	tmpPath := tmpFile.Name()
	scope.Declare("tmpPath", &tmpPath)
	godebug.Line(ctx, scope, 170)

	_, err = io.Copy(tmpFile, r)
	godebug.Line(ctx, scope, 171)
	tmpFile.Close()
	godebug.Line(ctx, scope, 173)

	if err != nil {
		godebug.Line(ctx, scope, 174)
		os.Remove(tmpPath)
		godebug.Line(ctx, scope, 175)
		return err
	}
	godebug.Line(ctx, scope, 178)

	if err = os.Rename(tmpPath, outfile); err != nil {
		godebug.Line(ctx, scope, 179)
		os.Remove(tmpPath)
		godebug.Line(ctx, scope, 180)
		return err
	}
	godebug.Line(ctx, scope, 183)

	return nil
}

func (cli *DockerCli) resolveAuthConfig(index *registrytypes.IndexInfo) types.AuthConfig {
	var result1 types.AuthConfig
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.resolveAuthConfig(index)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "index", &index)
	godebug.Line(ctx, scope, 190)
	configKey := index.Name
	scope.Declare("configKey", &configKey)
	godebug.Line(ctx, scope, 191)
	if index.Official {
		godebug.Line(ctx, scope, 192)
		configKey = cli.electAuthServer()
	}
	godebug.Line(ctx, scope, 195)

	a, _ := getCredentials(cli.configFile, configKey)
	scope.Declare("a", &a)
	godebug.Line(ctx, scope, 196)
	return a
}

func (cli *DockerCli) retrieveAuthConfigs() map[string]types.AuthConfig {
	var result1 map[string]types.AuthConfig
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.retrieveAuthConfigs()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := utils_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli)
	godebug.Line(ctx, scope, 200)
	acs, _ := getAllCredentials(cli.configFile)
	scope.Declare("acs", &acs)
	godebug.Line(ctx, scope, 201)
	return acs
}

var utils_go_contents = `package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	gosignal "os/signal"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	registrytypes "github.com/docker/engine-api/types/registry"
)

func (cli *DockerCli) electAuthServer() string {
	// The daemon ` + "`" + `/info` + "`" + ` endpoint informs us of the default registry being
	// used. This is essential in cross-platforms environment, where for
	// example a Linux client might be interacting with a Windows daemon, hence
	// the default registry URL might be Windows specific.
	serverAddress := registry.IndexServer
	if info, err := cli.client.Info(context.Background()); err != nil {
		fmt.Fprintf(cli.out, "Warning: failed to get default registry endpoint from daemon (%v). Using system default: %s\n", err, serverAddress)
	} else {
		serverAddress = info.IndexServerAddress
	}
	return serverAddress
}

// encodeAuthToBase64 serializes the auth configuration as JSON base64 payload
func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

func (cli *DockerCli) registryAuthenticationPrivilegedFunc(index *registrytypes.IndexInfo, cmdName string) client.RequestPrivilegeFunc {
	return func() (string, error) {
		fmt.Fprintf(cli.out, "\nPlease login prior to %s:\n", cmdName)
		indexServer := registry.GetAuthConfigKey(index)
		authConfig, err := cli.configureAuth("", "", indexServer, false)
		if err != nil {
			return "", err
		}
		return encodeAuthToBase64(authConfig)
	}
}

func (cli *DockerCli) resizeTty(id string, isExec bool) {
	height, width := cli.getTtySize()
	cli.resizeTtyTo(id, height, width, isExec)
}

func (cli *DockerCli) resizeTtyTo(id string, height, width int, isExec bool) {
	if height == 0 && width == 0 {
		return
	}

	options := types.ResizeOptions{
		ID:     id,
		Height: height,
		Width:  width,
	}

	var err error
	if isExec {
		err = cli.client.ContainerExecResize(context.Background(), options)
	} else {
		err = cli.client.ContainerResize(context.Background(), options)
	}

	if err != nil {
		logrus.Debugf("Error resize: %s", err)
	}
}

// getExitCode perform an inspect on the container. It returns
// the running state and the exit code.
func getExitCode(cli *DockerCli, containerID string) (bool, int, error) {
	c, err := cli.client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != client.ErrConnectionFailed {
			return false, -1, err
		}
		return false, -1, nil
	}

	return c.State.Running, c.State.ExitCode, nil
}

// getExecExitCode perform an inspect on the exec command. It returns
// the running state and the exit code.
func getExecExitCode(cli *DockerCli, execID string) (bool, int, error) {
	resp, err := cli.client.ContainerExecInspect(context.Background(), execID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != client.ErrConnectionFailed {
			return false, -1, err
		}
		return false, -1, nil
	}

	return resp.Running, resp.ExitCode, nil
}

func (cli *DockerCli) monitorTtySize(id string, isExec bool) error {
	cli.resizeTty(id, isExec)

	if runtime.GOOS == "windows" {
		go func() {
			prevH, prevW := cli.getTtySize()
			for {
				time.Sleep(time.Millisecond * 250)
				h, w := cli.getTtySize()

				if prevW != w || prevH != h {
					cli.resizeTty(id, isExec)
				}
				prevH = h
				prevW = w
			}
		}()
	} else {
		sigchan := make(chan os.Signal, 1)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		go func() {
			for range sigchan {
				cli.resizeTty(id, isExec)
			}
		}()
	}
	return nil
}

func (cli *DockerCli) getTtySize() (int, int) {
	if !cli.isTerminalOut {
		return 0, 0
	}
	ws, err := term.GetWinsize(cli.outFd)
	if err != nil {
		logrus.Debugf("Error getting size: %s", err)
		if ws == nil {
			return 0, 0
		}
	}
	return int(ws.Height), int(ws.Width)
}

func copyToFile(outfile string, r io.Reader) error {
	tmpFile, err := ioutil.TempFile(filepath.Dir(outfile), ".docker_temp_")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, r)
	tmpFile.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err = os.Rename(tmpPath, outfile); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// resolveAuthConfig is like registry.ResolveAuthConfig, but if using the
// default index, it uses the default index name for the daemon's platform,
// not the client's platform.
func (cli *DockerCli) resolveAuthConfig(index *registrytypes.IndexInfo) types.AuthConfig {
	configKey := index.Name
	if index.Official {
		configKey = cli.electAuthServer()
	}

	a, _ := getCredentials(cli.configFile, configKey)
	return a
}

func (cli *DockerCli) retrieveAuthConfigs() map[string]types.AuthConfig {
	acs, _ := getAllCredentials(cli.configFile)
	return acs
}
`


var version_go_scope = godebug.EnteringNewFile(client_pkg_scope, version_go_contents)

var versionTemplate = `Client:
 Version:      {{.Client.Version}}
 API version:  {{.Client.APIVersion}}
 Go version:   {{.Client.GoVersion}}
 Git commit:   {{.Client.GitCommit}}
 Built:        {{.Client.BuildTime}}
 OS/Arch:      {{.Client.Os}}/{{.Client.Arch}}{{if .Client.Experimental}}
 Experimental: {{.Client.Experimental}}{{end}}{{if .ServerOK}}

Server:
 Version:      {{.Server.Version}}
 API version:  {{.Server.APIVersion}}
 Go version:   {{.Server.GoVersion}}
 Git commit:   {{.Server.GitCommit}}
 Built:        {{.Server.BuildTime}}
 OS/Arch:      {{.Server.Os}}/{{.Server.Arch}}{{if .Server.Experimental}}
 Experimental: {{.Server.Experimental}}{{end}}{{end}}`

func (cli *DockerCli) CmdVersion(args ...string) (err error) {
	ctx, ok := godebug.EnterFunc(func() {
		err = cli.CmdVersion(args...)
	})
	if !ok {
		return err
	}
	defer godebug.ExitFunc(ctx)
	scope := version_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args, "err", &err)
	godebug.Line(ctx, scope, 42)
	cmd := Cli.Subcmd("version", nil, Cli.DockerCommands["version"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 43)
	tmplStr := cmd.String([]string{"f", "#format", "-format"}, "", "Format the output using the given go template")
	scope.Declare("tmplStr", &tmplStr)
	godebug.Line(ctx, scope, 44)
	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 46)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 48)

	templateFormat := versionTemplate
	scope.Declare("templateFormat", &templateFormat)
	godebug.Line(ctx, scope, 49)
	if *tmplStr != "" {
		godebug.Line(ctx, scope, 50)
		templateFormat = *tmplStr
	}
	godebug.Line(ctx, scope, 53)

	var tmpl *template.Template
	scope.Declare("tmpl", &tmpl)
	godebug.Line(ctx, scope, 54)
	if tmpl, err = templates.Parse(templateFormat); err != nil {
		godebug.Line(ctx, scope, 55)
		return Cli.StatusError{StatusCode: 64,
			Status: "Template parsing error: " + err.Error()}
	}
	godebug.Line(ctx, scope, 59)

	vd := types.VersionResponse{
		Client: &types.Version{
			Version:      dockerversion.Version,
			APIVersion:   cli.client.ClientVersion(),
			GoVersion:    runtime.Version(),
			GitCommit:    dockerversion.GitCommit,
			BuildTime:    dockerversion.BuildTime,
			Os:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Experimental: utils.ExperimentalBuild(),
		},
	}
	scope.Declare("vd", &vd)
	godebug.Line(ctx, scope, 72)

	serverVersion, err := cli.client.ServerVersion(context.Background())
	scope.Declare("serverVersion", &serverVersion)
	godebug.Line(ctx, scope, 73)
	if err == nil {
		godebug.Line(ctx, scope, 74)
		vd.Server = &serverVersion
	}
	godebug.Line(ctx, scope, 78)

	t, errTime := time.Parse(time.RFC3339Nano, vd.Client.BuildTime)
	scope.Declare("t", &t, "errTime", &errTime)
	godebug.Line(ctx, scope, 79)
	if errTime == nil {
		godebug.Line(ctx, scope, 80)
		vd.Client.BuildTime = t.Format(time.ANSIC)
	}
	godebug.Line(ctx, scope, 83)

	if vd.ServerOK() {
		godebug.Line(ctx, scope, 84)
		t, errTime = time.Parse(time.RFC3339Nano, vd.Server.BuildTime)
		godebug.Line(ctx, scope, 85)
		if errTime == nil {
			godebug.Line(ctx, scope, 86)
			vd.Server.BuildTime = t.Format(time.ANSIC)
		}
	}
	godebug.Line(ctx, scope, 90)

	if err2 := tmpl.Execute(cli.out, vd); err2 != nil && err == nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err2", &err2)
		godebug.Line(ctx, scope, 91)
		err = err2
	}
	godebug.Line(ctx, scope, 93)
	cli.out.Write([]byte{'\n'})
	godebug.Line(ctx, scope, 94)
	return err
}

var version_go_contents = `package client

import (
	"runtime"
	"text/template"
	"time"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/dockerversion"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/docker/utils/templates"
	"github.com/docker/engine-api/types"
)

var versionTemplate = ` + "`" + `Client:
 Version:      {{.Client.Version}}
 API version:  {{.Client.APIVersion}}
 Go version:   {{.Client.GoVersion}}
 Git commit:   {{.Client.GitCommit}}
 Built:        {{.Client.BuildTime}}
 OS/Arch:      {{.Client.Os}}/{{.Client.Arch}}{{if .Client.Experimental}}
 Experimental: {{.Client.Experimental}}{{end}}{{if .ServerOK}}

Server:
 Version:      {{.Server.Version}}
 API version:  {{.Server.APIVersion}}
 Go version:   {{.Server.GoVersion}}
 Git commit:   {{.Server.GitCommit}}
 Built:        {{.Server.BuildTime}}
 OS/Arch:      {{.Server.Os}}/{{.Server.Arch}}{{if .Server.Experimental}}
 Experimental: {{.Server.Experimental}}{{end}}{{end}}` + "`" + `

// CmdVersion shows Docker version information.
//
// Available version information is shown for: client Docker version, client API version, client Go version, client Git commit, client OS/Arch, server Docker version, server API version, server Go version, server Git commit, and server OS/Arch.
//
// Usage: docker version
func (cli *DockerCli) CmdVersion(args ...string) (err error) {
	cmd := Cli.Subcmd("version", nil, Cli.DockerCommands["version"].Description, true)
	tmplStr := cmd.String([]string{"f", "#format", "-format"}, "", "Format the output using the given go template")
	cmd.Require(flag.Exact, 0)

	cmd.ParseFlags(args, true)

	templateFormat := versionTemplate
	if *tmplStr != "" {
		templateFormat = *tmplStr
	}

	var tmpl *template.Template
	if tmpl, err = templates.Parse(templateFormat); err != nil {
		return Cli.StatusError{StatusCode: 64,
			Status: "Template parsing error: " + err.Error()}
	}

	vd := types.VersionResponse{
		Client: &types.Version{
			Version:      dockerversion.Version,
			APIVersion:   cli.client.ClientVersion(),
			GoVersion:    runtime.Version(),
			GitCommit:    dockerversion.GitCommit,
			BuildTime:    dockerversion.BuildTime,
			Os:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Experimental: utils.ExperimentalBuild(),
		},
	}

	serverVersion, err := cli.client.ServerVersion(context.Background())
	if err == nil {
		vd.Server = &serverVersion
	}

	// first we need to make BuildTime more human friendly
	t, errTime := time.Parse(time.RFC3339Nano, vd.Client.BuildTime)
	if errTime == nil {
		vd.Client.BuildTime = t.Format(time.ANSIC)
	}

	if vd.ServerOK() {
		t, errTime = time.Parse(time.RFC3339Nano, vd.Server.BuildTime)
		if errTime == nil {
			vd.Server.BuildTime = t.Format(time.ANSIC)
		}
	}

	if err2 := tmpl.Execute(cli.out, vd); err2 != nil && err == nil {
		err = err2
	}
	cli.out.Write([]byte{'\n'})
	return err
}
`


var volume_go_scope = godebug.EnteringNewFile(client_pkg_scope, volume_go_contents)

func (cli *DockerCli) CmdVolume(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdVolume(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 22)
	description := Cli.DockerCommands["volume"].Description + "\n\nCommands:\n"
	scope.Declare("description", &description)
	godebug.Line(ctx, scope, 23)
	commands := [][]string{
		{"create", "Create a volume"},
		{"inspect", "Return low-level information on a volume"},
		{"ls", "List volumes"},
		{"rm", "Remove a volume"},
	}
	scope.Declare("commands", &commands)
	{
		scope := scope.EnteringNewChildScope()

		for _, cmd := range commands {
			godebug.Line(ctx, scope, 30)
			scope.Declare("cmd", &cmd)
			godebug.Line(ctx, scope, 31)
			description += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
		}
		godebug.Line(ctx, scope, 30)
	}
	godebug.Line(ctx, scope, 34)

	description += "\nRun 'docker volume COMMAND --help' for more information on a command"
	godebug.Line(ctx, scope, 35)
	cmd := Cli.Subcmd("volume", []string{"[COMMAND]"}, description, false)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 37)

	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 38)
	err := cmd.ParseFlags(args, true)
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 39)
	cmd.Usage()
	godebug.Line(ctx, scope, 40)
	return err
}

func (cli *DockerCli) CmdVolumeLs(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdVolumeLs(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 47)
	cmd := Cli.Subcmd("volume ls", nil, "List volumes", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 49)

	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display volume names")
	scope.Declare("quiet", &quiet)
	godebug.Line(ctx, scope, 50)
	flFilter := opts.NewListOpts(nil)
	scope.Declare("flFilter", &flFilter)
	godebug.Line(ctx, scope, 51)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Provide filter values (i.e. 'dangling=true')")
	godebug.Line(ctx, scope, 53)

	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 54)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 56)

	volFilterArgs := filters.NewArgs()
	scope.Declare("volFilterArgs", &volFilterArgs)
	{
		scope := scope.EnteringNewChildScope()
		for _, f := range flFilter.GetAll() {
			godebug.Line(ctx, scope, 57)
			scope.Declare("f", &f)
			godebug.Line(ctx, scope, 58)
			var err error
			scope := scope.EnteringNewChildScope()
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 59)
			volFilterArgs, err = filters.ParseFlag(f, volFilterArgs)
			godebug.Line(ctx, scope, 60)
			if err != nil {
				godebug.Line(ctx, scope, 61)
				return err
			}
		}
		godebug.Line(ctx, scope, 57)
	}
	godebug.Line(ctx, scope, 65)

	volumes, err := cli.client.VolumeList(context.Background(), volFilterArgs)
	scope.Declare("volumes", &volumes, "err", &err)
	godebug.Line(ctx, scope, 66)
	if err != nil {
		godebug.Line(ctx, scope, 67)
		return err
	}
	godebug.Line(ctx, scope, 70)

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	scope.Declare("w", &w)
	godebug.Line(ctx, scope, 71)
	if !*quiet {
		{
			scope := scope.EnteringNewChildScope()
			for _, warn := range volumes.Warnings {
				godebug.Line(ctx, scope, 72)
				scope.Declare("warn", &warn)
				godebug.Line(ctx, scope, 73)
				fmt.Fprintln(cli.err, warn)
			}
			godebug.Line(ctx, scope, 72)
		}
		godebug.Line(ctx, scope, 75)
		fmt.Fprintf(w, "DRIVER \tVOLUME NAME")
		godebug.Line(ctx, scope, 76)
		fmt.Fprintf(w, "\n")
	}
	godebug.Line(ctx, scope, 79)

	sort.Sort(byVolumeName(volumes.Volumes))
	{
		scope := scope.EnteringNewChildScope()
		for _, vol := range volumes.Volumes {
			godebug.Line(ctx, scope, 80)
			scope.Declare("vol", &vol)
			godebug.Line(ctx, scope, 81)
			if *quiet {
				godebug.Line(ctx, scope, 82)
				fmt.Fprintln(w, vol.Name)
				godebug.Line(ctx, scope, 83)
				continue
			}
			godebug.Line(ctx, scope, 85)
			fmt.Fprintf(w, "%s\t%s\n", vol.Driver, vol.Name)
		}
		godebug.Line(ctx, scope, 80)
	}
	godebug.Line(ctx, scope, 87)
	w.Flush()
	godebug.Line(ctx, scope, 88)
	return nil
}

type byVolumeName []*types.Volume

func (r byVolumeName) Len() int {
	var result1 int
	ctx, ok := godebug.EnterFunc(func() {
		result1 = r.Len()
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r)
	godebug.Line(ctx, scope, 93)
	return len(r)
}
func (r byVolumeName) Swap(i, j int) {
	ctx, ok := godebug.EnterFunc(func() {
		r.Swap(i, j)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 94)
	r[i], r[j] = r[j], r[i]
}
func (r byVolumeName) Less(i, j int) bool {
	var result1 bool
	ctx, ok := godebug.EnterFunc(func() {
		result1 = r.Less(i, j)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("r", &r, "i", &i, "j", &j)
	godebug.Line(ctx, scope, 96)
	return r[i].Name < r[j].Name
}

func (cli *DockerCli) CmdVolumeInspect(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdVolumeInspect(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 103)
	cmd := Cli.Subcmd("volume inspect", []string{"VOLUME [VOLUME...]"}, "Return low-level information on a volume", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 104)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")
	scope.Declare("tmplStr", &tmplStr)
	godebug.Line(ctx, scope, 106)

	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 107)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 109)

	if err := cmd.Parse(args); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 110)
		return nil
	}
	godebug.Line(ctx, scope, 113)

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		var result1 interface {
		}
		var result2 []byte
		var result3 error
		fn := func(ctx *godebug.Context) {
			result1, result2, result3 = func() (interface {
			}, []byte, error) {
				scope := scope.EnteringNewChildScope()
				scope.Declare("name", &name)
				godebug.Line(ctx, scope, 114)
				i, err := cli.client.VolumeInspect(context.Background(), name)
				scope.Declare("i", &i, "err", &err)
				godebug.Line(ctx, scope, 115)
				return i, nil, err
			}()
		}
		if ctx, ok := godebug.EnterFuncLit(fn); ok {
			defer godebug.ExitFunc(ctx)
			fn(ctx)
		}
		return result1, result2, result3
	}
	scope.Declare("inspectSearcher", &inspectSearcher)
	godebug.Line(ctx, scope, 118)

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}

func (cli *DockerCli) CmdVolumeCreate(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdVolumeCreate(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 125)
	cmd := Cli.Subcmd("volume create", nil, "Create a volume", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 126)
	flDriver := cmd.String([]string{"d", "-driver"}, "local", "Specify volume driver name")
	scope.Declare("flDriver", &flDriver)
	godebug.Line(ctx, scope, 127)
	flName := cmd.String([]string{"-name"}, "", "Specify volume name")
	scope.Declare("flName", &flName)
	godebug.Line(ctx, scope, 129)

	flDriverOpts := opts.NewMapOpts(nil, nil)
	scope.Declare("flDriverOpts", &flDriverOpts)
	godebug.Line(ctx, scope, 130)
	cmd.Var(flDriverOpts, []string{"o", "-opt"}, "Set driver specific options")
	godebug.Line(ctx, scope, 132)

	flLabels := opts.NewListOpts(nil)
	scope.Declare("flLabels", &flLabels)
	godebug.Line(ctx, scope, 133)
	cmd.Var(&flLabels, []string{"-label"}, "Set metadata for a volume")
	godebug.Line(ctx, scope, 135)

	cmd.Require(flag.Exact, 0)
	godebug.Line(ctx, scope, 136)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 138)

	volReq := types.VolumeCreateRequest{
		Driver:     *flDriver,
		DriverOpts: flDriverOpts.GetAll(),
		Name:       *flName,
		Labels:     runconfigopts.ConvertKVStringsToMap(flLabels.GetAll()),
	}
	scope.Declare("volReq", &volReq)
	godebug.Line(ctx, scope, 145)

	vol, err := cli.client.VolumeCreate(context.Background(), volReq)
	scope.Declare("vol", &vol, "err", &err)
	godebug.Line(ctx, scope, 146)
	if err != nil {
		godebug.Line(ctx, scope, 147)
		return err
	}
	godebug.Line(ctx, scope, 150)

	fmt.Fprintf(cli.out, "%s\n", vol.Name)
	godebug.Line(ctx, scope, 151)
	return nil
}

func (cli *DockerCli) CmdVolumeRm(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdVolumeRm(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := volume_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 158)
	cmd := Cli.Subcmd("volume rm", []string{"VOLUME [VOLUME...]"}, "Remove a volume", true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 159)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 160)
	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 162)

	var status = 0
	scope.Declare("status", &status)
	{
		scope := scope.EnteringNewChildScope()

		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 164)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 165)
			if err := cli.client.VolumeRemove(context.Background(), name); err != nil {
				scope := scope.EnteringNewChildScope()
				scope.Declare("err", &err)
				godebug.Line(ctx, scope, 166)
				fmt.Fprintf(cli.err, "%s\n", err)
				godebug.Line(ctx, scope, 167)
				status = 1
				godebug.Line(ctx, scope, 168)
				continue
			}
			godebug.Line(ctx, scope, 170)
			fmt.Fprintf(cli.out, "%s\n", name)
		}
		godebug.Line(ctx, scope, 164)
	}
	godebug.Line(ctx, scope, 173)

	if status != 0 {
		godebug.Line(ctx, scope, 174)
		return Cli.StatusError{StatusCode: status}
	}
	godebug.Line(ctx, scope, 176)
	return nil
}

var volume_go_contents = `package client

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// CmdVolume is the parent subcommand for all volume commands
//
// Usage: docker volume <COMMAND> <OPTS>
func (cli *DockerCli) CmdVolume(args ...string) error {
	description := Cli.DockerCommands["volume"].Description + "\n\nCommands:\n"
	commands := [][]string{
		{"create", "Create a volume"},
		{"inspect", "Return low-level information on a volume"},
		{"ls", "List volumes"},
		{"rm", "Remove a volume"},
	}

	for _, cmd := range commands {
		description += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	description += "\nRun 'docker volume COMMAND --help' for more information on a command"
	cmd := Cli.Subcmd("volume", []string{"[COMMAND]"}, description, false)

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdVolumeLs outputs a list of Docker volumes.
//
// Usage: docker volume ls [OPTIONS]
func (cli *DockerCli) CmdVolumeLs(args ...string) error {
	cmd := Cli.Subcmd("volume ls", nil, "List volumes", true)

	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display volume names")
	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Provide filter values (i.e. 'dangling=true')")

	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	volFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		var err error
		volFilterArgs, err = filters.ParseFlag(f, volFilterArgs)
		if err != nil {
			return err
		}
	}

	volumes, err := cli.client.VolumeList(context.Background(), volFilterArgs)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	if !*quiet {
		for _, warn := range volumes.Warnings {
			fmt.Fprintln(cli.err, warn)
		}
		fmt.Fprintf(w, "DRIVER \tVOLUME NAME")
		fmt.Fprintf(w, "\n")
	}

	sort.Sort(byVolumeName(volumes.Volumes))
	for _, vol := range volumes.Volumes {
		if *quiet {
			fmt.Fprintln(w, vol.Name)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\n", vol.Driver, vol.Name)
	}
	w.Flush()
	return nil
}

type byVolumeName []*types.Volume

func (r byVolumeName) Len() int      { return len(r) }
func (r byVolumeName) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r byVolumeName) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

// CmdVolumeInspect displays low-level information on one or more volumes.
//
// Usage: docker volume inspect [OPTIONS] VOLUME [VOLUME...]
func (cli *DockerCli) CmdVolumeInspect(args ...string) error {
	cmd := Cli.Subcmd("volume inspect", []string{"VOLUME [VOLUME...]"}, "Return low-level information on a volume", true)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")

	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		i, err := cli.client.VolumeInspect(context.Background(), name)
		return i, nil, err
	}

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}

// CmdVolumeCreate creates a new volume.
//
// Usage: docker volume create [OPTIONS]
func (cli *DockerCli) CmdVolumeCreate(args ...string) error {
	cmd := Cli.Subcmd("volume create", nil, "Create a volume", true)
	flDriver := cmd.String([]string{"d", "-driver"}, "local", "Specify volume driver name")
	flName := cmd.String([]string{"-name"}, "", "Specify volume name")

	flDriverOpts := opts.NewMapOpts(nil, nil)
	cmd.Var(flDriverOpts, []string{"o", "-opt"}, "Set driver specific options")

	flLabels := opts.NewListOpts(nil)
	cmd.Var(&flLabels, []string{"-label"}, "Set metadata for a volume")

	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	volReq := types.VolumeCreateRequest{
		Driver:     *flDriver,
		DriverOpts: flDriverOpts.GetAll(),
		Name:       *flName,
		Labels:     runconfigopts.ConvertKVStringsToMap(flLabels.GetAll()),
	}

	vol, err := cli.client.VolumeCreate(context.Background(), volReq)
	if err != nil {
		return err
	}

	fmt.Fprintf(cli.out, "%s\n", vol.Name)
	return nil
}

// CmdVolumeRm removes one or more volumes.
//
// Usage: docker volume rm VOLUME [VOLUME...]
func (cli *DockerCli) CmdVolumeRm(args ...string) error {
	cmd := Cli.Subcmd("volume rm", []string{"VOLUME [VOLUME...]"}, "Remove a volume", true)
	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	var status = 0

	for _, name := range cmd.Args() {
		if err := cli.client.VolumeRemove(context.Background(), name); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
		fmt.Fprintf(cli.out, "%s\n", name)
	}

	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}
`


var wait_go_scope = godebug.EnteringNewFile(client_pkg_scope, wait_go_contents)

func (cli *DockerCli) CmdWait(args ...string) error {
	var result1 error
	ctx, ok := godebug.EnterFunc(func() {
		result1 = cli.CmdWait(args...)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := wait_go_scope.EnteringNewChildScope()
	scope.Declare("cli", &cli, "args", &args)
	godebug.Line(ctx, scope, 19)
	cmd := Cli.Subcmd("wait", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["wait"].Description, true)
	scope.Declare("cmd", &cmd)
	godebug.Line(ctx, scope, 20)
	cmd.Require(flag.Min, 1)
	godebug.Line(ctx, scope, 22)

	cmd.ParseFlags(args, true)
	godebug.Line(ctx, scope, 24)

	var errs []string
	scope.Declare("errs", &errs)
	{
		scope := scope.EnteringNewChildScope()
		for _, name := range cmd.Args() {
			godebug.Line(ctx, scope, 25)
			scope.Declare("name", &name)
			godebug.Line(ctx, scope, 26)
			status, err := cli.client.ContainerWait(context.Background(), name)
			scope := scope.EnteringNewChildScope()
			scope.Declare("status", &status, "err", &err)
			godebug.Line(ctx, scope, 27)
			if err != nil {
				godebug.Line(ctx, scope, 28)
				errs = append(errs, err.Error())
			} else {
				godebug.Line(ctx, scope, 29)
				godebug.Line(ctx, scope, 30)
				fmt.Fprintf(cli.out, "%d\n", status)
			}
		}
		godebug.Line(ctx, scope, 25)
	}
	godebug.Line(ctx, scope, 33)
	if len(errs) > 0 {
		godebug.Line(ctx, scope, 34)
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	godebug.Line(ctx, scope, 36)
	return nil
}

var wait_go_contents = `package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdWait blocks until a container stops, then prints its exit code.
//
// If more than one container is specified, this will wait synchronously on each container.
//
// Usage: docker wait CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdWait(args ...string) error {
	cmd := Cli.Subcmd("wait", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["wait"].Description, true)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var errs []string
	for _, name := range cmd.Args() {
		status, err := cli.client.ContainerWait(context.Background(), name)
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%d\n", status)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
`


var exec_test_go_scope = godebug.EnteringNewFile(client_pkg_scope, exec_test_go_contents)

type arguments struct {
	args []string
}

func TestParseExec(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestParseExec(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := exec_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 17)
	invalids := map[*arguments]error{
		&arguments{[]string{"-unknown"}}: fmt.Errorf("flag provided but not defined: -unknown"),
		&arguments{[]string{"-u"}}:       fmt.Errorf("flag needs an argument: -u"),
		&arguments{[]string{"--user"}}:   fmt.Errorf("flag needs an argument: --user"),
	}
	scope.Declare("invalids", &invalids)
	godebug.Line(ctx, scope, 22)

	valids := map[*arguments]*types.ExecConfig{
		&arguments{
			[]string{"container", "command"},
		}: {
			Container:    "container",
			Cmd:          []string{"command"},
			AttachStdout: true,
			AttachStderr: true,
		},
		&arguments{
			[]string{"container", "command1", "command2"},
		}: {
			Container:    "container",
			Cmd:          []string{"command1", "command2"},
			AttachStdout: true,
			AttachStderr: true,
		},
		&arguments{
			[]string{"-i", "-t", "-u", "uid", "container", "command"},
		}: {
			User:         "uid",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Container:    "container",
			Cmd:          []string{"command"},
		},
		&arguments{
			[]string{"-d", "container", "command"},
		}: {
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Detach:       true,
			Container:    "container",
			Cmd:          []string{"command"},
		},
		&arguments{
			[]string{"-t", "-i", "-d", "container", "command"},
		}: {
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Detach:       true,
			Tty:          true,
			Container:    "container",
			Cmd:          []string{"command"},
		},
	}
	scope.Declare("valids", &valids)
	{
		scope := scope.EnteringNewChildScope()

		for invalid, expectedError := range invalids {
			godebug.Line(ctx, scope, 72)
			scope.Declare("invalid", &invalid, "expectedError", &expectedError)
			godebug.Line(ctx, scope, 73)
			cmd := flag.NewFlagSet("exec", flag.ContinueOnError)
			scope := scope.EnteringNewChildScope()
			scope.Declare("cmd", &cmd)
			godebug.Line(ctx, scope, 74)
			cmd.ShortUsage = func() {
				fn := func(ctx *godebug.Context) {
				}
				if ctx, ok := godebug.EnterFuncLit(fn); ok {
					defer godebug.ExitFunc(ctx)
					fn(ctx)
				}
			}
			godebug.Line(ctx, scope, 75)
			cmd.SetOutput(ioutil.Discard)
			godebug.Line(ctx, scope, 76)
			_, err := ParseExec(cmd, invalid.args)
			scope.Declare("err", &err)
			godebug.Line(ctx, scope, 77)
			if err == nil || err.Error() != expectedError.Error() {
				godebug.Line(ctx, scope, 78)
				t.Fatalf("Expected an error [%v] for %v, got %v", expectedError, invalid, err)
			}

		}
		godebug.Line(ctx, scope, 72)
	}
	{
		scope := scope.EnteringNewChildScope()
		for valid, expectedExecConfig := range valids {
			godebug.Line(ctx, scope, 82)
			scope.Declare("valid", &valid, "expectedExecConfig", &expectedExecConfig)
			godebug.Line(ctx, scope, 83)
			cmd := flag.NewFlagSet("exec", flag.ContinueOnError)
			scope := scope.EnteringNewChildScope()
			scope.Declare("cmd", &cmd)
			godebug.Line(ctx, scope, 84)
			cmd.ShortUsage = func() {
				fn := func(ctx *godebug.Context) {
				}
				if ctx, ok := godebug.EnterFuncLit(fn); ok {
					defer godebug.ExitFunc(ctx)
					fn(ctx)
				}
			}
			godebug.Line(ctx, scope, 85)
			cmd.SetOutput(ioutil.Discard)
			godebug.Line(ctx, scope, 86)
			execConfig, err := ParseExec(cmd, valid.args)
			scope.Declare("execConfig", &execConfig, "err", &err)
			godebug.Line(ctx, scope, 87)
			if err != nil {
				godebug.Line(ctx, scope, 88)
				t.Fatal(err)
			}
			godebug.Line(ctx, scope, 90)
			if !compareExecConfig(expectedExecConfig, execConfig) {
				godebug.Line(ctx, scope, 91)
				t.Fatalf("Expected [%v] for %v, got [%v]", expectedExecConfig, valid, execConfig)
			}
		}
		godebug.Line(ctx, scope, 82)
	}
}

func compareExecConfig(config1 *types.ExecConfig, config2 *types.ExecConfig) bool {
	var result1 bool
	ctx, ok := godebug.EnterFunc(func() {
		result1 = compareExecConfig(config1, config2)
	})
	if !ok {
		return result1
	}
	defer godebug.ExitFunc(ctx)
	scope := exec_test_go_scope.EnteringNewChildScope()
	scope.Declare("config1", &config1, "config2", &config2)
	godebug.Line(ctx, scope, 97)
	if config1.AttachStderr != config2.AttachStderr {
		godebug.Line(ctx, scope, 98)
		return false
	}
	godebug.Line(ctx, scope, 100)
	if config1.AttachStdin != config2.AttachStdin {
		godebug.Line(ctx, scope, 101)
		return false
	}
	godebug.Line(ctx, scope, 103)
	if config1.AttachStdout != config2.AttachStdout {
		godebug.Line(ctx, scope, 104)
		return false
	}
	godebug.Line(ctx, scope, 106)
	if config1.Container != config2.Container {
		godebug.Line(ctx, scope, 107)
		return false
	}
	godebug.Line(ctx, scope, 109)
	if config1.Detach != config2.Detach {
		godebug.Line(ctx, scope, 110)
		return false
	}
	godebug.Line(ctx, scope, 112)
	if config1.Privileged != config2.Privileged {
		godebug.Line(ctx, scope, 113)
		return false
	}
	godebug.Line(ctx, scope, 115)
	if config1.Tty != config2.Tty {
		godebug.Line(ctx, scope, 116)
		return false
	}
	godebug.Line(ctx, scope, 118)
	if config1.User != config2.User {
		godebug.Line(ctx, scope, 119)
		return false
	}
	godebug.Line(ctx, scope, 121)
	if len(config1.Cmd) != len(config2.Cmd) {
		godebug.Line(ctx, scope, 122)
		return false
	}
	{
		scope := scope.EnteringNewChildScope()
		for index, value := range config1.Cmd {
			godebug.Line(ctx, scope, 124)
			scope.Declare("index", &index, "value", &value)
			godebug.Line(ctx, scope, 125)
			if value != config2.Cmd[index] {
				godebug.Line(ctx, scope, 126)
				return false
			}
		}
		godebug.Line(ctx, scope, 124)
	}
	godebug.Line(ctx, scope, 129)
	return true
}

var exec_test_go_contents = `package client

import (
	"fmt"
	"io/ioutil"
	"testing"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
)

type arguments struct {
	args []string
}

func TestParseExec(t *testing.T) {
	invalids := map[*arguments]error{
		&arguments{[]string{"-unknown"}}: fmt.Errorf("flag provided but not defined: -unknown"),
		&arguments{[]string{"-u"}}:       fmt.Errorf("flag needs an argument: -u"),
		&arguments{[]string{"--user"}}:   fmt.Errorf("flag needs an argument: --user"),
	}
	valids := map[*arguments]*types.ExecConfig{
		&arguments{
			[]string{"container", "command"},
		}: {
			Container:    "container",
			Cmd:          []string{"command"},
			AttachStdout: true,
			AttachStderr: true,
		},
		&arguments{
			[]string{"container", "command1", "command2"},
		}: {
			Container:    "container",
			Cmd:          []string{"command1", "command2"},
			AttachStdout: true,
			AttachStderr: true,
		},
		&arguments{
			[]string{"-i", "-t", "-u", "uid", "container", "command"},
		}: {
			User:         "uid",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Container:    "container",
			Cmd:          []string{"command"},
		},
		&arguments{
			[]string{"-d", "container", "command"},
		}: {
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Detach:       true,
			Container:    "container",
			Cmd:          []string{"command"},
		},
		&arguments{
			[]string{"-t", "-i", "-d", "container", "command"},
		}: {
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Detach:       true,
			Tty:          true,
			Container:    "container",
			Cmd:          []string{"command"},
		},
	}
	for invalid, expectedError := range invalids {
		cmd := flag.NewFlagSet("exec", flag.ContinueOnError)
		cmd.ShortUsage = func() {}
		cmd.SetOutput(ioutil.Discard)
		_, err := ParseExec(cmd, invalid.args)
		if err == nil || err.Error() != expectedError.Error() {
			t.Fatalf("Expected an error [%v] for %v, got %v", expectedError, invalid, err)
		}

	}
	for valid, expectedExecConfig := range valids {
		cmd := flag.NewFlagSet("exec", flag.ContinueOnError)
		cmd.ShortUsage = func() {}
		cmd.SetOutput(ioutil.Discard)
		execConfig, err := ParseExec(cmd, valid.args)
		if err != nil {
			t.Fatal(err)
		}
		if !compareExecConfig(expectedExecConfig, execConfig) {
			t.Fatalf("Expected [%v] for %v, got [%v]", expectedExecConfig, valid, execConfig)
		}
	}
}

func compareExecConfig(config1 *types.ExecConfig, config2 *types.ExecConfig) bool {
	if config1.AttachStderr != config2.AttachStderr {
		return false
	}
	if config1.AttachStdin != config2.AttachStdin {
		return false
	}
	if config1.AttachStdout != config2.AttachStdout {
		return false
	}
	if config1.Container != config2.Container {
		return false
	}
	if config1.Detach != config2.Detach {
		return false
	}
	if config1.Privileged != config2.Privileged {
		return false
	}
	if config1.Tty != config2.Tty {
		return false
	}
	if config1.User != config2.User {
		return false
	}
	if len(config1.Cmd) != len(config2.Cmd) {
		return false
	}
	for index, value := range config1.Cmd {
		if value != config2.Cmd[index] {
			return false
		}
	}
	return true
}
`


var stats_unit_test_go_scope = godebug.EnteringNewFile(client_pkg_scope, stats_unit_test_go_contents)

func TestDisplay(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestDisplay(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_unit_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 12)
	c := &containerStats{
		Name:             "app",
		CPUPercentage:    30.0,
		Memory:           100 * 1024 * 1024.0,
		MemoryLimit:      2048 * 1024 * 1024.0,
		MemoryPercentage: 100.0 / 2048.0 * 100.0,
		NetworkRx:        100 * 1024 * 1024,
		NetworkTx:        800 * 1024 * 1024,
		BlockRead:        100 * 1024 * 1024,
		BlockWrite:       800 * 1024 * 1024,
		PidsCurrent:      1,
		mu:               sync.RWMutex{},
	}
	scope.Declare("c", &c)
	godebug.Line(ctx, scope, 25)

	var b bytes.Buffer
	scope.Declare("b", &b)
	godebug.Line(ctx, scope, 26)
	if err := c.Display(&b); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 27)
		t.Fatalf("c.Display() gave error: %s", err)
	}
	godebug.Line(ctx, scope, 29)
	got := b.String()
	scope.Declare("got", &got)
	godebug.Line(ctx, scope, 30)
	want := "app\t30.00%\t100 MiB / 2 GiB\t4.88%\t104.9 MB / 838.9 MB\t104.9 MB / 838.9 MB\t1\n"
	scope.Declare("want", &want)
	godebug.Line(ctx, scope, 31)
	if got != want {
		godebug.Line(ctx, scope, 32)
		t.Fatalf("c.Display() = %q, want %q", got, want)
	}
}

func TestCalculBlockIO(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestCalculBlockIO(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := stats_unit_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 37)
	blkio := types.BlkioStats{
		IoServiceBytesRecursive: []types.BlkioStatEntry{{8, 0, "read", 1234}, {8, 1, "read", 4567}, {8, 0, "write", 123}, {8, 1, "write", 456}},
	}
	scope.Declare("blkio", &blkio)
	godebug.Line(ctx, scope, 40)

	blkRead, blkWrite := calculateBlockIO(blkio)
	scope.Declare("blkRead", &blkRead, "blkWrite", &blkWrite)
	godebug.Line(ctx, scope, 41)
	if blkRead != 5801 {
		godebug.Line(ctx, scope, 42)
		t.Fatalf("blkRead = %d, want 5801", blkRead)
	}
	godebug.Line(ctx, scope, 44)
	if blkWrite != 579 {
		godebug.Line(ctx, scope, 45)
		t.Fatalf("blkWrite = %d, want 579", blkWrite)
	}
}

var stats_unit_test_go_contents = `package client

import (
	"bytes"
	"sync"
	"testing"

	"github.com/docker/engine-api/types"
)

func TestDisplay(t *testing.T) {
	c := &containerStats{
		Name:             "app",
		CPUPercentage:    30.0,
		Memory:           100 * 1024 * 1024.0,
		MemoryLimit:      2048 * 1024 * 1024.0,
		MemoryPercentage: 100.0 / 2048.0 * 100.0,
		NetworkRx:        100 * 1024 * 1024,
		NetworkTx:        800 * 1024 * 1024,
		BlockRead:        100 * 1024 * 1024,
		BlockWrite:       800 * 1024 * 1024,
		PidsCurrent:      1,
		mu:               sync.RWMutex{},
	}
	var b bytes.Buffer
	if err := c.Display(&b); err != nil {
		t.Fatalf("c.Display() gave error: %s", err)
	}
	got := b.String()
	want := "app\t30.00%\t100 MiB / 2 GiB\t4.88%\t104.9 MB / 838.9 MB\t104.9 MB / 838.9 MB\t1\n"
	if got != want {
		t.Fatalf("c.Display() = %q, want %q", got, want)
	}
}

func TestCalculBlockIO(t *testing.T) {
	blkio := types.BlkioStats{
		IoServiceBytesRecursive: []types.BlkioStatEntry{{8, 0, "read", 1234}, {8, 1, "read", 4567}, {8, 0, "write", 123}, {8, 1, "write", 456}},
	}
	blkRead, blkWrite := calculateBlockIO(blkio)
	if blkRead != 5801 {
		t.Fatalf("blkRead = %d, want 5801", blkRead)
	}
	if blkWrite != 579 {
		t.Fatalf("blkWrite = %d, want 579", blkWrite)
	}
}
`


var trust_test_go_scope = godebug.EnteringNewFile(client_pkg_scope, trust_test_go_contents)

func unsetENV() {
	ctx, ok := godebug.EnterFunc(unsetENV)
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	godebug.Line(ctx, trust_test_go_scope, 12)
	os.Unsetenv("DOCKER_CONTENT_TRUST")
	godebug.Line(ctx, trust_test_go_scope, 13)
	os.Unsetenv("DOCKER_CONTENT_TRUST_SERVER")
}

func TestENVTrustServer(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestENVTrustServer(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 17)
	defer unsetENV()
	defer godebug.Defer(ctx, scope, 17)
	godebug.Line(ctx, scope, 18)
	indexInfo := &registrytypes.IndexInfo{Name: "testserver"}
	scope.Declare("indexInfo", &indexInfo)
	godebug.Line(ctx, scope, 19)
	if err := os.Setenv("DOCKER_CONTENT_TRUST_SERVER", "https://notary-test.com:5000"); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 20)
		t.Fatal("Failed to set ENV variable")
	}
	godebug.Line(ctx, scope, 22)
	output, err := trustServer(indexInfo)
	scope.Declare("output", &output, "err", &err)
	godebug.Line(ctx, scope, 23)
	expectedStr := "https://notary-test.com:5000"
	scope.Declare("expectedStr", &expectedStr)
	godebug.Line(ctx, scope, 24)
	if err != nil || output != expectedStr {
		godebug.Line(ctx, scope, 25)
		t.Fatalf("Expected server to be %s, got %s", expectedStr, output)
	}
}

func TestHTTPENVTrustServer(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestHTTPENVTrustServer(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 30)
	defer unsetENV()
	defer godebug.Defer(ctx, scope, 30)
	godebug.Line(ctx, scope, 31)
	indexInfo := &registrytypes.IndexInfo{Name: "testserver"}
	scope.Declare("indexInfo", &indexInfo)
	godebug.Line(ctx, scope, 32)
	if err := os.Setenv("DOCKER_CONTENT_TRUST_SERVER", "http://notary-test.com:5000"); err != nil {
		scope := scope.EnteringNewChildScope()
		scope.Declare("err", &err)
		godebug.Line(ctx, scope, 33)
		t.Fatal("Failed to set ENV variable")
	}
	godebug.Line(ctx, scope, 35)
	_, err := trustServer(indexInfo)
	scope.Declare("err", &err)
	godebug.Line(ctx, scope, 36)
	if err == nil {
		godebug.Line(ctx, scope, 37)
		t.Fatal("Expected error with invalid scheme")
	}
}

func TestOfficialTrustServer(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestOfficialTrustServer(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 42)
	indexInfo := &registrytypes.IndexInfo{Name: "testserver", Official: true}
	scope.Declare("indexInfo", &indexInfo)
	godebug.Line(ctx, scope, 43)
	output, err := trustServer(indexInfo)
	scope.Declare("output", &output, "err", &err)
	godebug.Line(ctx, scope, 44)
	if err != nil || output != registry.NotaryServer {
		godebug.Line(ctx, scope, 45)
		t.Fatalf("Expected server to be %s, got %s", registry.NotaryServer, output)
	}
}

func TestNonOfficialTrustServer(t *testing.T) {
	ctx, ok := godebug.EnterFunc(func() {
		TestNonOfficialTrustServer(t)
	})
	if !ok {
		return
	}
	defer godebug.ExitFunc(ctx)
	scope := trust_test_go_scope.EnteringNewChildScope()
	scope.Declare("t", &t)
	godebug.Line(ctx, scope, 50)
	indexInfo := &registrytypes.IndexInfo{Name: "testserver", Official: false}
	scope.Declare("indexInfo", &indexInfo)
	godebug.Line(ctx, scope, 51)
	output, err := trustServer(indexInfo)
	scope.Declare("output", &output, "err", &err)
	godebug.Line(ctx, scope, 52)
	expectedStr := "https://" + indexInfo.Name
	scope.Declare("expectedStr", &expectedStr)
	godebug.Line(ctx, scope, 53)
	if err != nil || output != expectedStr {
		godebug.Line(ctx, scope, 54)
		t.Fatalf("Expected server to be %s, got %s", expectedStr, output)
	}
}

var trust_test_go_contents = `package client

import (
	"os"
	"testing"

	"github.com/docker/docker/registry"
	registrytypes "github.com/docker/engine-api/types/registry"
)

func unsetENV() {
	os.Unsetenv("DOCKER_CONTENT_TRUST")
	os.Unsetenv("DOCKER_CONTENT_TRUST_SERVER")
}

func TestENVTrustServer(t *testing.T) {
	defer unsetENV()
	indexInfo := &registrytypes.IndexInfo{Name: "testserver"}
	if err := os.Setenv("DOCKER_CONTENT_TRUST_SERVER", "https://notary-test.com:5000"); err != nil {
		t.Fatal("Failed to set ENV variable")
	}
	output, err := trustServer(indexInfo)
	expectedStr := "https://notary-test.com:5000"
	if err != nil || output != expectedStr {
		t.Fatalf("Expected server to be %s, got %s", expectedStr, output)
	}
}

func TestHTTPENVTrustServer(t *testing.T) {
	defer unsetENV()
	indexInfo := &registrytypes.IndexInfo{Name: "testserver"}
	if err := os.Setenv("DOCKER_CONTENT_TRUST_SERVER", "http://notary-test.com:5000"); err != nil {
		t.Fatal("Failed to set ENV variable")
	}
	_, err := trustServer(indexInfo)
	if err == nil {
		t.Fatal("Expected error with invalid scheme")
	}
}

func TestOfficialTrustServer(t *testing.T) {
	indexInfo := &registrytypes.IndexInfo{Name: "testserver", Official: true}
	output, err := trustServer(indexInfo)
	if err != nil || output != registry.NotaryServer {
		t.Fatalf("Expected server to be %s, got %s", registry.NotaryServer, output)
	}
}

func TestNonOfficialTrustServer(t *testing.T) {
	indexInfo := &registrytypes.IndexInfo{Name: "testserver", Official: false}
	output, err := trustServer(indexInfo)
	expectedStr := "https://" + indexInfo.Name
	if err != nil || output != expectedStr {
		t.Fatalf("Expected server to be %s, got %s", expectedStr, output)
	}
}
`
