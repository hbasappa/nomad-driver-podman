package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// SystemInfo returns information on the system and libpod configuration
func (c *API) SystemInfo(ctx context.Context) (Info, error) {

	var infoData Info
	attempted := 0
START:
	res, err := c.Get(ctx, "/v3.0.0/libpod/info")
	if err != nil {
		if attempted < BackoffAttempts {
			attempted++
			time.Sleep(BackOffDelay * time.Millisecond)
			goto START
		}
		return infoData, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return infoData, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return infoData, err
	}
	err = json.Unmarshal(body, &infoData)
	if err != nil {
		return infoData, err
	}

	return infoData, nil
}
