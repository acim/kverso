package registry

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Client ...
type Client struct {
	baseURL string
	image   string
	tag     string
	auth    bool
}

// NewClient ...
func NewClient(image string) (*Client, error) {
	parsed, err := reference.ParseAnyReference(image)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(parsed.String(), ":")
	u, err := url.Parse("https://" + parts[0])
	if err != nil {
		return nil, err
	}

	c := &Client{
		image: strings.TrimPrefix(u.Path, "/"),
	}
	if len(parts) > 1 {
		c.tag = parts[1]
	}
	if u.Host == "docker.io" {
		u.Host = "registry.hub.docker.com"
		c.auth = true
	}
	u.Path = ""
	c.baseURL = u.String()

	return c, nil
}

// Tags ...
func (c *Client) Tags() ([]string, error) {
	url := c.baseURL + "/" + path.Join("v2", c.image, "tags", "list")
	fmt.Println("URL: ", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if c.auth {
		token, err := c.token()
		if err != nil {
			return nil, errors.Wrap(err, "auth failed")
		}
		req.Header.Add("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	tags := &tags{}
	err = json.NewDecoder(resp.Body).Decode(tags)
	if err != nil {
		return nil, err
	}

	v, err := semver.NewVersion(strings.TrimPrefix(c.tag, "v"))
	if err != nil {
		return tags.Tags, nil
	}

	ts := []string{}
	for _, t := range tags.Tags {
		vn, err := semver.NewVersion(strings.TrimPrefix(t, "v"))
		if err != nil {
			continue
		}
		if v.LessThan(*vn) {
			ts = append(ts, t)
		}
	}

	return ts, nil
}

func (c *Client) token() (string, error) {
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", c.image)
	fmt.Println("Auth URL: ", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth status code %d", resp.StatusCode)
	}

	auth := &auth{}
	err = json.NewDecoder(resp.Body).Decode(auth)
	if err != nil {
		return "", err
	}

	return auth.Token, nil
}

type auth struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

type tags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
