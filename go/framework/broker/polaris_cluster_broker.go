package broker

import (
	"context"

	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// The cluster broker service is responsible for providing a remote polaris-scheduler access to the local cluster.
type PolarisClusterBroker interface {

	// Starts the cluster broker.
	//
	// The context can be used to stop the broker.
	// Returns nil if the broker has started successfully.
	Start(ctx context.Context) error

	// Gets the config used by this broker.
	Config() *config.ClusterBrokerConfig

	// Gets the ClusterClient used by this broker.
	ClusterClient() client.ClusterClient

	// Gets the logger used by this broker.
	Logger() *logr.Logger
}
