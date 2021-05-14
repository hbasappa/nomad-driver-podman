package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	nstructs "github.com/hashicorp/nomad/nomad/structs"
)

// ContainerStart starts a container via id or name
func (c *API) ContainerStart(ctx context.Context, name string) error {

	attempted := 0
START:
	res, err := c.Post(ctx, fmt.Sprintf("/v3.0.0/libpod/containers/%s/start", name), nil)
	if err != nil {
		c.logger.Error("failed to start container", "error", err)
		if isPodmanTransientError(err) {
			if attempted < 2 {
				attempted++
				time.Sleep(nextBackoff(attempted))
				goto START
			}
			c.logger.Error("failed to start container after 5 attempts", "error", err)
			return nstructs.NewRecoverableError(err, true)
		}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}
	c.logger.Error("started container", "error", err)
    /*
	// wait max 60 seconds for running state
	// TODO: make timeout configurable
	timeout, _ := context.WithTimeout(ctx, time.Second*30)
	//defer cancel()
WAIT:
	err = c.ContainerWait(timeout, name, "running")
	attempted = 0
	if err != nil {
		if attempted < 2 {
			attempted++
			time.Sleep(nextBackoff(attempted))
			goto WAIT
		}
		return nstructs.NewRecoverableError(err, true)
	}*/
	return nil
}
