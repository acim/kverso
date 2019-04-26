package registry

import (
	"net/url"
	"strings"
	"sync"

	"github.com/docker/distribution/reference"
	"github.com/nokia/docker-registry-client/registry"
)

// Client ...
type Client struct {
	registries *sync.Map
	tags       *sync.Map
	digests    *sync.Map
}

// NewClient creates new Docker registries client.
func NewClient() *Client {
	return &Client{
		registries: &sync.Map{},
		tags:       &sync.Map{},
		digests:    &sync.Map{},
	}
}

// Tags returns slice of available tags.
func (c *Client) Tags(image string) ([]string, string, error) {
	info, err := parseImage(image)
	if err != nil {
		return nil, "", err
	}

	if tags, ok := c.tags.Load(info.registryURL + info.image); ok {
		return tags.([]string), info.tag, nil
	}

	r, err := c.registry(info.registryURL)
	if err != nil {
		return nil, "", err
	}

	tags, err := r.Tags(info.image)
	if err != nil {
		return nil, "", err
	}

	c.tags.Store(info.registryURL+info.image, tags)

	return tags, info.tag, nil
}

// Digest returns digest of the image.
func (c *Client) Digest(image string) (string, error) {
	info, err := parseImage(image)
	if err != nil {
		return "", err
	}

	if digest, ok := c.digests.Load(info.registryURL + info.image); ok {
		return digest.(string), nil
	}

	r, err := c.registry(info.registryURL)
	if err != nil {
		return "", err
	}

	digest, err := r.ManifestV2Digest(info.image, info.tag)
	if err != nil {
		return "", err
	}

	c.digests.Store(info.registryURL+info.image, string(digest))

	return string(digest), nil
}

func (c *Client) registry(url string) (*registry.Registry, error) {
	if r, ok := c.registries.Load(url); ok {
		return r.(*registry.Registry), nil
	}

	r, err := registry.NewCustom(url, registry.Options{})
	if err != nil {
		return nil, err
	}
	c.registries.Store(url, r)

	return r, nil
}

func parseImage(image string) (*info, error) {
	parsed, err := reference.ParseAnyReference(image)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(parsed.String(), ":")
	u, err := url.Parse("https://" + parts[0])
	if err != nil {
		return nil, err
	}

	info := &info{
		image: strings.TrimPrefix(u.Path, "/"),
		tag:   "latest",
	}

	if len(parts) > 1 {
		info.tag = parts[1]
	}

	if u.Host == "docker.io" {
		u.Host = "registry.hub.docker.com"
	}
	u.Path = ""
	info.registryURL = u.String()

	return info, nil
}

type info struct {
	registryURL string
	image       string
	tag         string
}
