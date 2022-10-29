package kubeutil

import (
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Returns a Config object for creating a Kubernetes REST client.
func GetRestClientConfig() *rest.Config {
	return ctrl.GetConfigOrDie()
}
