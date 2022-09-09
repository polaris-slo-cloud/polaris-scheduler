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
// First we attempt to load it from a pod environment (i.e., when operating inside a cluster).
// If this fails, we try the path in kubeconfigPathCmdLineArg or use the local KUBECONFIG file.
func LoadKubeconfig(kubeconfigPathCmdLineArg string, logger *logr.Logger) (*rest.Config, error) {
	// Try loading the config from the in-cluster environment.
	k8sConfig, err := rest.InClusterConfig()
	if err == nil {
		logger.Info("Using in-cluster KUBECONFIG")
		return k8sConfig, nil
	}
	// If an unexpected error occurred, return it.
	if err != rest.ErrNotInCluster {
		return nil, err
	}

	// Try loading the config from a file.
	kubeconfigPath := getKubeconfigPath(kubeconfigPathCmdLineArg)
	if k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath); err != nil {
		return nil, err
	} else {
		logger.Info("Using KUBECONFIG", "path", kubeconfigPath)
		return k8sConfig, nil
	}
}

// Returns the path of the KUBECONFIG file.
// It uses the following order of preference:
//
// 1. --kubeconfig command line flag
// 2. $KUBECONFIG environment variable
// 3. $HOME/.kube/config
func getKubeconfigPath(cmdLineFlagValue string) string {
	if cmdLineFlagValue != "" {
		return cmdLineFlagValue
	}

	if kubeconfigPath, ok := os.LookupEnv("KUBECONFIG"); ok && kubeconfigPath != "" {
		return kubeconfigPath
	}

	home := homedir.HomeDir()
	return filepath.Join(home, ".kube", "config")
}
