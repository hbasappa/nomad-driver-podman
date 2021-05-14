package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

type API struct {
	baseUrl    string
	httpClient *http.Client
	logger     hclog.Logger
}

type ClientConfig struct {
	SocketPath  string
	HttpTimeout time.Duration
}

var (
	podmanTransientErrs = []string{
		"Client.Timeout exceeded while awaiting headers",
		"EOF",
		"API error (500)",
		"no such image",
		"unknown error",
	}
	BackOffDelay    = time.Duration(200)
	BackoffAttempts = 2
)

func DefaultClientConfig() ClientConfig {
	cfg := ClientConfig{
		HttpTimeout: 90 * time.Second,
	}
	uid := os.Getuid()
	// are we root?
	if uid == 0 {
		cfg.SocketPath = "unix:/run/podman/podman.sock"
	} else {
		// not? then let's try the default per-user socket location
		cfg.SocketPath = fmt.Sprintf("unix:/run/user/%d/podman/podman.sock", uid)
	}
	return cfg
}

func NewClient(logger hclog.Logger, config ClientConfig) *API {
	ac := &API{
		logger: logger,
	}

	baseUrl := config.SocketPath
	ac.logger.Debug("http baseurl", "url", baseUrl)
	ac.httpClient = &http.Client{
		Timeout: config.HttpTimeout,
	}
	if strings.HasPrefix(baseUrl, "unix:") {
		ac.baseUrl = "http://u"
		path := strings.TrimPrefix(baseUrl, "unix:")
		ac.httpClient.Transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.DialTimeout("unix", path, 10*time.Second)
			},
		}
	} else {
		ac.baseUrl = baseUrl
	}

	return ac
}

func (c *API) Do(req *http.Request) (*http.Response, error) {
	res, err := c.httpClient.Do(req)
	return res, err
}

func (c *API) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *API) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseUrl+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(req)
}

func (c *API) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func isPodmanTransientError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, te := range podmanTransientErrs {
		if strings.Contains(errMsg, te) {
			return true
		}
	}

	return false
}

func nextBackoff(attempted int) time.Duration {
	// attempts in 200ms, 800ms, 3.2s, 12.8s, 51.2s
	// TODO: add randomization factor and 0extract to a helper
	return 1 << (2 * uint64(attempted)) * 50 * time.Millisecond
}
