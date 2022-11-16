package kubernetes

import (
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Loads the Kubernetes config.
// If the pathOverride is set, we load the KUBECONFIG from there, otherwise
// we attempt to load it from a pod environment (i.e., when operating inside a cluster).
// If this fails, we try to use the local KUBECONFIG file.
func LoadKubeconfig(pathOverride string, logger *logr.Logger) (*rest.Config, error) {
	var kubeconfigPath string

	if pathOverride == "" {
		// Try loading the config from the in-cluster environment.
		k8sConfig, err := rest.InClusterConfig()
		if err == nil {
			logger.Info("Loaded in-cluster KUBECONFIG")
			return k8sConfig, nil
		}
		// If an unexpected error occurred, return it.
		if err != rest.ErrNotInCluster {
			return nil, err
		}

		// Try getting the path of the local KUBECONFIG file
		kubeconfigPath = getLocalKubeconfigPath()
	} else {
		kubeconfigPath = pathOverride
	}

	// Load the config from a file.
	if k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath); err != nil {
		return nil, err
	} else {
		logger.Info("Loaded KUBECONFIG", "path", kubeconfigPath)
		return k8sConfig, nil
	}
}

// Returns the path of the local KUBECONFIG file.
// It uses the following order of preference:
//
// 1. $KUBECONFIG environment variable
// 2. $HOME/.kube/config
func getLocalKubeconfigPath() string {
	if kubeconfigPath, ok := os.LookupEnv("KUBECONFIG"); ok && kubeconfigPath != "" {
		return kubeconfigPath
	}

	home := homedir.HomeDir()
	return filepath.Join(home, ".kube", "config")
}
