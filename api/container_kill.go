package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ContainerKill sends a signal to a container
func (c *API) ContainerKill(ctx context.Context, name string, signal string) error {
	attempted := 0
START:
	res, err := c.Post(ctx, fmt.Sprintf("/v3.0.0/libpod/containers/%s/kill?signal=%s", name, signal), nil)
	if err != nil {
		if attempted < BackoffAttempts {
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
