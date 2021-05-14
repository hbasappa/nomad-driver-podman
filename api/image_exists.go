package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ImageExists checks if image exists in local store
func (c *API) ImageExists(ctx context.Context, nameWithTag string) (bool, error) {

	attempted := 0
START:
	res, err := c.Get(ctx, fmt.Sprintf("/v3.0.0/libpod/images/%s/exists", nameWithTag))
	if err != nil {
		if attempted < BackoffAttempts {
			attempted++
			time.Sleep(BackOffDelay * time.Millisecond)
			goto START
		}
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode == http.StatusNoContent {
		return true, nil
	}
	return false, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}
