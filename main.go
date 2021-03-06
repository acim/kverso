package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/acim/kverso/pkg/k8s"
	"github.com/acim/kverso/pkg/registry"
)

func main() {
	client, err := k8s.NewClient()
	if err != nil {
		panic(err)
	}

	http.Handle("/", handler(client, registry.NewClient()))
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handler(c *k8s.Client, r *registry.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		pods, err := c.Pods()
		if err != nil {
			w.Write([]byte("Error: " + err.Error()))
			return
		}
		for _, pod := range pods {
			w.Write([]byte("Pod name: " + pod.Name + "<br>"))
			for n, c := range pod.Containers {
				w.Write([]byte("Container name: " + n + " Image: " + c.Image + " Current digest: " + c.Digest + "<br>"))
				tags, tag, err := r.FilteredTags(c.Image)
				if err != nil {
					w.Write([]byte("Tags error: " + err.Error() + "<br>"))
					continue
				}
				w.Write([]byte("Current tag: " + tag + " Newer tags: " + strings.Join(tags, ", ") + "<br>"))
				if tag == "latest" {
					digest, err := r.Digest(c.Image)
					if err != nil {
						w.Write([]byte("Remote digest error: " + err.Error() + "<br>"))
						continue
					}
					w.Write([]byte("Remote digest: " + digest + "<br>"))
				}
			}
			for n, c := range pod.InitContainers {
				w.Write([]byte("Init container name: " + n + " Image: " + c.Image + " Digest: " + c.Digest + "<br>"))
				tags, tag, err := r.FilteredTags(c.Image)
				if err != nil {
					w.Write([]byte("Tags error: " + err.Error() + "<br>"))
					continue
				}
				w.Write([]byte("Current tag: " + tag + " Newer tags: " + strings.Join(tags, ", ") + "<br>"))
				if tag == "latest" {
					digest, err := r.Digest(c.Image)
					if err != nil {
						w.Write([]byte("Remote digest error: " + err.Error() + "<br>"))
						continue
					}
					w.Write([]byte("Remote digest: " + digest + "<br>"))
				}
			}
			w.Write([]byte("<br>"))
		}
	})
}
