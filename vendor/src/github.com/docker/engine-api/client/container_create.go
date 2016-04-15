package client

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"golang.org/x/net/context"
	"net/url"
	"strings"
)

const debug_level int = 1

type configWrapper struct {
	*container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

// ContainerCreate creates a new container based in the given configuration.
// It can be associated with a name, but it's not mandatory.
func (cli *Client) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (types.ContainerCreateResponse, error) {
	if debug_level > 0 {
		logrus.Debugf("Called vendor/src/github.com/docker/engine-api/client/container_create.go:ContainerCreate with config.Image=%s", config.Image)
	}
	var response types.ContainerCreateResponse
	query := url.Values{}
	if containerName != "" {
		query.Set("name", containerName)
	}

	body := configWrapper{
		Config:           config,
		HostConfig:       hostConfig,
		NetworkingConfig: networkingConfig,
	}

	serverResp, err := cli.post(ctx, "/containers/create", query, body, nil)
	if err != nil {
		if serverResp != nil && serverResp.statusCode == 404 && strings.Contains(err.Error(), "No such image") {
			return response, imageNotFoundError{config.Image}
		}
		return response, err
	}

	err = json.NewDecoder(serverResp.body).Decode(&response)
	ensureReaderClosed(serverResp)
	return response, err
}
