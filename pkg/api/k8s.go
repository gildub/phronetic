package api

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewK8SOrDie init k8s client or panic
func NewK8SOrDie(config *rest.Config) *kubernetes.Clientset {
	return kubernetes.NewForConfigOrDie(config)
}
