package k8s

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client ...
type Client struct {
	c *kubernetes.Clientset
}

// NewClient creates new Kubernetes client.
func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: client,
	}, nil
}

// Pods returns a slice of pods.
func (c *Client) Pods() ([]Pod, error) {
	list, err := c.c.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]Pod, len(list.Items))
	for i, pod := range list.Items {
		pods[i] = Pod{
			Name:           pod.ObjectMeta.Name,
			Containers:     make(map[string]*Container, len(pod.Spec.Containers)),
			InitContainers: make(map[string]*Container, len(pod.Spec.InitContainers)),
		}

		for _, container := range pod.Spec.Containers {
			pods[i].Containers[container.Name] = &Container{
				Image: container.Image,
			}
		}

		for _, container := range pod.Spec.InitContainers {
			pods[i].InitContainers[container.Name] = &Container{
				Image: container.Image,
			}
		}

		for _, status := range pod.Status.ContainerStatuses {
			pods[i].Containers[status.Name].Digest = strings.SplitAfterN(status.ImageID, ":", 3)[2]
		}

		for _, status := range pod.Status.InitContainerStatuses {
			pods[i].InitContainers[status.Name].Digest = strings.SplitAfterN(status.ImageID, ":", 3)[2]
		}
	}

	return pods, nil
}

// Pod ...
type Pod struct {
	Name           string
	Containers     map[string]*Container
	InitContainers map[string]*Container
}

// Container ...
type Container struct {
	Image  string
	Digest string
}
