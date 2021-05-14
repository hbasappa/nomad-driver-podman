package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	nstructs "github.com/hashicorp/nomad/nomad/structs"
)

// ContainerCreate creates a new container
func (c *API) ContainerCreate(ctx context.Context, create SpecGenerator) (ContainerCreateResponse, error) {

	attempted := 0
	create.CgroupParent = "ifc-scheduler_allocations.slice"
	response := ContainerCreateResponse{}

	jsonString, err := json.Marshal(create)
	if err != nil {
		return response, err
	}
START:
	res, err := c.Post(ctx, "/v3.0.0/libpod/containers/create", bytes.NewBuffer(jsonString))
	if err != nil {
		c.logger.Error("failed to start container", "error", err)
		//if isPodmanTransientError(err) {
		if attempted < 2 {
			attempted++
			time.Sleep(nextBackoff(attempted))
			goto START
		}
		c.logger.Error("failed to start container after 5 attempts", "error", err)
		//}
		return response, nstructs.NewRecoverableError(err, true)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		return response, fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, err
}

type ContainerCreateRequest struct {
	// Name is the name the container will be given.
	// If no name is provided, one will be randomly generated.
	// Optional.
	Name string `json:"name,omitempty"`

	// Command is the container's command.
	// If not given and Image is specified, this will be populated by the
	// image's configuration.
	// Optional.
	Command []string `json:"command,omitempty"`

	// Entrypoint is the container's entrypoint.
	// If not given and Image is specified, this will be populated by the
	// image's configuration.
	// Optional.
	Entrypoint []string `json:"entrypoint,omitempty"`

	// WorkDir is the container's working directory.
	// If unset, the default, /, will be used.
	// Optional.
	WorkDir string `json:"work_dir,omitempty"`
	// Env is a set of environment variables that will be set in the
	// container.
	// Optional.
	Env map[string]string `json:"env,omitempty"`
}

type ContainerCreateResponse struct {
	Id       string
	Warnings []string
}
