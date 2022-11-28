package pipeline

import (
	"github.com/go-logr/logr"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Provides ClusterAgent plugins access to services that they may use.
type ClusterAgentServices interface {
	// Gets the config used by this ClusterAgent.
	Config() *config.ClusterAgentConfig

	// Gets the LocalClusterClient used by this ClusterAgent.
	ClusterClient() client.LocalClusterClient

	// The nodes cache used by this ClusterAgent.
	NodesCache() client.NodesCache

	// Gets the logger used by this ClusterAgent.
	Logger() *logr.Logger
}
