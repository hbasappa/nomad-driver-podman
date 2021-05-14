package api

import (
	"context"
	"fmt"
	"time"
	"net/http"
)

// ContainerDelete deletes a container.
// It takes the name or ID of a container.
func (c *API) ContainerDelete(ctx context.Context, name string, force bool, deleteVolumes bool) error {
	attempted := 0
START:
	res, err := c.Delete(ctx, fmt.Sprintf("/v3.0.0/libpod/containers/%s?force=%t&v=%t", name, force, deleteVolumes))
	if err != nil {
		if attempted < 2 * BackoffAttempts {
			attempted++
			time.Sleep(BackOffDelay * time.Millisecond)
			goto START
		}
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}
