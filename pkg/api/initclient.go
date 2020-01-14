package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	migv1alpha1 "github.com/fusor/mig-controller/pkg/apis/migration/v1alpha1"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// KubeConfig represents kubeconfig
	KubeConfig *clientcmdapi.Config
	// ClusterNames contains names of contexts and cluster
	ClusterNames = make(map[string]string)
	// CtrlClient k8s controller client for migration cluster
	CtrlClient client.Client
	// K8sSrcClient k8s api client for source cluster
	K8sSrcClient *kubernetes.Clientset
	// K8sDstClient k8s api client for target cluster
	K8sDstClient *kubernetes.Clientset

	kubeConfigGetter = func() (*clientcmdapi.Config, error) {
		return KubeConfig, nil
	}
)

// ParseKubeConfig parse kubeconfig
func ParseKubeConfig() error {
	kubeConfigPath, err := getKubeConfigPath()
	if err != nil {
		return err
	}

	kubeConfigFile, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		return err
	}

	KubeConfig, err = clientcmd.Load(kubeConfigFile)
	if err != nil {
		return err
	}
	// Map context clusters and name for easier access in future
	for name, context := range KubeConfig.Contexts {
		ClusterNames[context.Cluster] = name
	}

	return nil
}

func getKubeConfigPath() (string, error) {
	// Get kubeconfig using $KUBECONFIG, if not try ~/.kube/config
	var kubeConfigPath string

	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if kubeconfigEnv != "" {
		kubeConfigPath = kubeconfigEnv
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return "", errors.Wrap(err, "Can't detect home user directory")
		}
		kubeConfigPath = fmt.Sprintf("%s/.kube/config", home)
	}

	return kubeConfigPath, nil
}

// CreateCtrlClient creates a k8s runtime-controller client for given context
func CreateCtrlClient(contextCluster string) error {
	if CtrlClient == nil {
		config, err := buildConfig(contextCluster)
		if err != nil {
			return err
		}
		crScheme := k8sruntime.NewScheme()
		migv1alpha1.AddToScheme(crScheme)
		CtrlClient = NewCtrlClientorDie(config, client.Options{Scheme: crScheme})
		logrus.Debugf("Kubernetes Controller client initialized for %s", contextCluster)
	}

	return nil
}

// CreateK8sDstClient create api client using cluster from kubeconfig context
func CreateK8sDstClient(contextCluster string) error {
	if K8sDstClient == nil {
		config, err := buildConfig(contextCluster)
		if err != nil {
			return err
		}

		K8sDstClient = NewK8SOrDie(config)
		logrus.Debugf("Kubernetes API client initialized for %s", contextCluster)
	}

	return nil
}

// CreateK8sSrcClient create api client using cluster from kubeconfig context
func CreateK8sSrcClient(contextCluster string) error {
	if K8sSrcClient == nil {
		config, err := buildConfig(contextCluster)
		if err != nil {
			return err
		}

		K8sSrcClient = NewK8SOrDie(config)
		logrus.Debugf("Kubernetes API client initialized for %s", contextCluster)
	}

	return nil
}

func buildConfig(contextCluster string) (*rest.Config, error) {
	// Check if context is present in kubeconfig
	if err := validateConfig(contextCluster); err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", kubeConfigGetter)
	if err != nil {
		return nil, errors.Wrap(err, "Error in KUBECONFIG")
	}

	config.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	config.UserAgent = fmt.Sprintf(
		"cpma/v1.0 (%s/%s) kubernetes/v1.0",
		runtime.GOOS, runtime.GOARCH,
	)

	return config, nil
}

func validateConfig(contextCluster string) error {
	for context := range ClusterNames {
		if context == contextCluster {
			return nil
		}
	}

	return errors.New("Not valid context")
}
