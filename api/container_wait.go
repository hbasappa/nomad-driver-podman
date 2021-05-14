package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ContainerWait waits on a container to met a given condition
func (c *API) ContainerWait(ctx context.Context, name string, condition string) error {
	attempted := 0
START:
	res, err := c.Post(ctx, fmt.Sprintf("/v3.0.0/libpod/containers/%s/wait?condition=%s", name, condition), nil)
	if err != nil {
		if attempted < BackoffAttempts + 1{
			attempted++
			time.Sleep(BackOffDelay * time.Millisecond)
			goto START
		}
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}
