package api

import (
	"log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewK8SOrDie init k8s client or panic
func NewK8SOrDie(config *rest.Config) *kubernetes.Clientset {
	return kubernetes.NewForConfigOrDie(config)
}

// NewCtrlClientorDie gets a controller client or die
func NewCtrlClientorDie(config *rest.Config, options client.Options) client.Client {
	ctrlClient, error := client.New(config, options)
	if error != nil {
		log.Fatal("Can't create runtime-controller client")
	}

	return ctrlClient
}

// NewK8SDynClientOrDie init k8s client or panic
func NewK8SDynClientOrDie(config *rest.Config) dynamic.Interface {
	return dynamic.NewForConfigOrDie(config)
}
