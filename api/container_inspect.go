package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ContainerInspect data takes a name or ID of a container returns the inspection data
func (c *API) ContainerInspect(ctx context.Context, name string) (InspectContainerData, error) {

	var inspectData InspectContainerData
	attempted := 0
START:
	res, err := c.Get(ctx, fmt.Sprintf("/v3.0.0/libpod/containers/%s/json", name))
	if err != nil {
		if attempted < BackoffAttempts {
			attempted++
			time.Sleep(BackOffDelay * time.Millisecond)
			goto START
		}
		return inspectData, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return inspectData, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return inspectData, err
	}
	err = json.Unmarshal(body, &inspectData)
	if err != nil {
		return inspectData, err
	}

	return inspectData, nil
}
