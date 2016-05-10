package client

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/reference"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	networktypes "github.com/docker/engine-api/types/network"
	"github.com/docker/libnetwork/resolvconf/dns"
	"golang.org/x/net/context"
	"io"
	"net/http/httputil"
	"os"
	"runtime"
	"runtime/debug"
)

// CmdMerge will runs a command in a new container from two images.
//
// Usage: docker merge [OPTIONS] IMAGE1 IMAGE2 [COMMAND] [ARG...]
func (cli *DockerCli) CmdMerge(args ...string) error {
	// _ = "breakpoint"
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
		logrus.Debug("Calling cli.mergeContainer(config,... ")
	}
	createResponse, err := cli.mergeContainer(config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
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

func (cli *DockerCli) mergeContainer(config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, cidfile, name string) (*types.ContainerCreateResponse, error) {
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
	response, err := cli.client.ContainerMerge(context.Background(), config, hostConfig, networkingConfig, name)

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
			response, retryErr = cli.client.ContainerMerge(context.Background(), config, hostConfig, networkingConfig, name)
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
