package registry

import (
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/docker/distribution/reference"
	"github.com/hashicorp/go-version"
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

// FilteredTags returns slice of available tags that look similar to the currect tag.
func (c *Client) FilteredTags(image string) ([]string, string, error) {
	tags, currTag, err := c.Tags(image)
	if err != nil {
		return nil, "", err
	}

	t := reDigit.ReplaceAllString(currTag, "\\d")
	t = reDot.ReplaceAllString(t, "\\.")
	r, err := regexp.Compile("^" + t + "$")
	if err != nil {
		return nil, "", err
	}

	var fTags []string
	for _, tag := range tags {
		if r.Match([]byte(tag)) {
			fTags = append(fTags, tag)
		}
	}

	currV, err := version.NewVersion(currTag)
	if err != nil {
		var ffTags []string
		for _, tag := range fTags {
			if tag > currTag {
				ffTags = append(ffTags, tag)
			}
		}
		return ffTags, currTag, nil
	}

	var ffTags []string
	for _, tag := range fTags {
		v, err := version.NewVersion(tag)
		if err != nil {
			continue
		}
		if currV.LessThan(v) {
			ffTags = append(ffTags, tag)
		}
	}

	return ffTags, currTag, nil

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

	d := strings.SplitN(string(digest), ":", 2)[1]
	c.digests.Store(info.registryURL+info.image, d)

	return d, nil
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

var reDigit = regexp.MustCompile(`\d`)
var reDot = regexp.MustCompile(`\.`)
