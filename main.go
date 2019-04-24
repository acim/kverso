package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/acim/kverso/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", handler(client))
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handler(c *kubernetes.Clientset) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		list, err := c.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "text/html")
		for _, p := range list.Items {
			w.Write([]byte("Pod: " + p.ObjectMeta.Name + "<br>"))
			for _, cs := range p.Spec.Containers {
				w.Write([]byte("Container: " + cs.Name + " Image: " + cs.Image + "<br>"))
				client := registry.NewClient(cs.Image)
				tags, _ := client.Tags()
				w.Write([]byte("Available tags: " + strings.Join(tags, ", ")))
			}
			for _, ics := range p.Spec.InitContainers {
				w.Write([]byte("Init container: " + ics.Name + " Image: " + ics.Image + "<br>"))
			}
		}
	})
}
