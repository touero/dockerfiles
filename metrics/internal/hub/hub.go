// Package hub is a tiny client for the public Docker Hub v2 API.
package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://hub.docker.com/v2/repositories"

// Repository is the subset of the Docker Hub repository payload we care about.
type Repository struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	PullCount int64  `json:"pull_count"`
	StarCount int64  `json:"star_count"`
}

// Client fetches repository metadata. The zero value is ready to use.
type Client struct {
	HTTP *http.Client
}

// New returns a Client with a sensible default timeout.
func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// Repository fetches metadata for namespace/repo. Public repositories need no
// authentication.
func (c *Client) Repository(ctx context.Context, namespace, repo string) (Repository, error) {
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	u := fmt.Sprintf("%s/%s/%s/", apiBase, url.PathEscape(namespace), url.PathEscape(repo))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Repository{}, err
	}
	req.Header.Set("User-Agent", "dockerfiles-metrics")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Repository{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Repository{}, fmt.Errorf("docker hub: %s/%s: HTTP %d", namespace, repo, resp.StatusCode)
	}

	var r Repository
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Repository{}, fmt.Errorf("docker hub: decode %s/%s: %w", namespace, repo, err)
	}
	return r, nil
}
