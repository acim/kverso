package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Client ...
type Client struct {
	baseURL string
	image   string
	auth    bool
}

// NewClient ...
func NewClient(image string) *Client {
	parts := strings.Split(image, ":")
	image = parts[0]
	c := &Client{}
	switch strings.Count(image, "/") {
	case 0:
		c.baseURL = "https://registry.hub.docker.com"
		c.image = path.Join("library", image)
		c.auth = true
	case 1:
		c.baseURL = "https://registry.hub.docker.com"
		c.image = image
		c.auth = true
	default:
		ut, _ := url.Parse("https://" + image)
		if ut.Host == "docker.io" {
			ut.Host = "registry.hub.docker.com"
			c.auth = true
		}
		c.baseURL = ut.Scheme + "://" + ut.Host
		c.image = strings.TrimPrefix(ut.Path, "/")
	}

	return c
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
			return nil, err
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

	return tags.Tags, nil
}

func (c *Client) token() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", c.image))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code %d", resp.StatusCode)
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
