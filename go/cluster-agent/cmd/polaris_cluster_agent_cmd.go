package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/clusteragent"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
)

type commandLineArgs struct {
	// The path to the cluster-agent config.
	config string

	// The path to the KUBECONFIG file.
	kubeconfig string
}

// Creates a new polaris-cluster-agent command.
func NewPolarisClusterAgentCmd(ctx context.Context, pluginRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices]) *cobra.Command {
	cmdLineArgs := commandLineArgs{}

	logger := initLogger()

	cmd := cobra.Command{
		Use: "polaris-cluster-agent",
		// ToDo: Extend long description.
		Long: "The Polaris Cluster Agent is a component of the Polaris Scheduler that allows the scheduler to interact with the local cluster. Agents in multiple clusters allow interaction with multiple clusters.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("polaris-cluster-agent")

			samplerConfig, err := loadConfigWithDefaults(cmdLineArgs.config, logger)
			if err != nil {
				logger.Error(err, "Error loading config.")
				os.Exit(1)
			}

			if err := runNodeSampler(ctx, samplerConfig, pluginRegistry, logger, &cmdLineArgs); err != nil {
				logger.Error(err, "Error starting polaris-cluster-agent")
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&cmdLineArgs.config, "config", "c", "", "The path to the polaris-cluster-agent configuration file.")
	cmd.MarkFlagFilename("config")
	cmd.PersistentFlags().StringVar(&cmdLineArgs.kubeconfig, "kubeconfig", "", "The path to the KUBECONFIG file.")
	cmd.MarkFlagFilename("kubeconfig")

	return &cmd
}

func initLogger() *logr.Logger {
	logger := stdr.New(nil)
	return &logger
}

// Loads the ClusterAgentConfig from the specified path and fills empty fields with default values.
func loadConfigWithDefaults(configPath string, logger *logr.Logger) (*config.ClusterAgentConfig, error) {
	clusterAgentConfig := &config.ClusterAgentConfig{}

	if configPath != "" {
		if err := util.ParseYamlFile(configPath, clusterAgentConfig); err != nil {
			return nil, err
		}
	}

	config.SetDefaultsClusterAgentConfig(clusterAgentConfig)
	return clusterAgentConfig, nil
}

func setUpClusterClient(k8sConfig *rest.Config, logger *logr.Logger) (kubernetes.KubernetesClusterClient, error) {
	k8sConfigs := map[string]*rest.Config{
		k8sConfig.ServerName: k8sConfig,
	}

	// We only need a single cluster client in the sampler, but we reuse the ClusterClientsManager abstraction.
	clusterClientsMgr, err := kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, "polaris-cluster-agent", logger)
	if err != nil {
		return nil, err
	}

	clusterClient, err := clusterClientsMgr.GetClusterClient(k8sConfig.ServerName)
	if err != nil {
		return nil, err
	}
	k8sClusterClient, ok := clusterClient.(kubernetes.KubernetesClusterClient)
	if !ok {
		return nil, fmt.Errorf("KubernetesClusterClientsManager does not return KubernetesClusterClients")
	}

	return k8sClusterClient, nil
}

func setUpNodesCache(clusterAgentConfig *config.ClusterAgentConfig, clusterClient kubernetes.KubernetesClusterClient) (client.NodesCache, error) {
	updateInterval, err := time.ParseDuration(fmt.Sprintf("%vms", clusterAgentConfig.NodesCacheUpdateIntervalMs))
	if err != nil {
		return nil, fmt.Errorf("error parsing nodesCacheUpdateIntervalMs: %v", err)
	}

	nodesCache := kubernetes.NewKubernetesNodesCache(clusterClient, updateInterval, int(clusterAgentConfig.NodesCacheUpdateQueueSize))
	return nodesCache, nil
}

func startNodeSampler(
	ctx context.Context,
	clusterAgentConfig *config.ClusterAgentConfig,
	k8sClusterClient kubernetes.KubernetesClusterClient,
	ginEngine *gin.Engine,
	pluginRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices],
	logger *logr.Logger,
) (pipeline.PolarisNodeSampler, error) {
	nodesCache, err := setUpNodesCache(clusterAgentConfig, k8sClusterClient)
	if err != nil {
		return nil, err
	}

	nodeSampler := runtime.NewDefaultPolarisNodeSampler(clusterAgentConfig, ginEngine, k8sClusterClient, nodesCache, pluginRegistry, logger)
	err = nodeSampler.Start(ctx)
	if err != nil {
		return nil, err
	}

	return nodeSampler, nil
}

func startClusterAgent(
	ctx context.Context,
	clusterAgentConfig *config.ClusterAgentConfig,
	k8sClusterClient kubernetes.KubernetesClusterClient,
	ginEngine *gin.Engine,
	logger *logr.Logger,
) (clusteragent.PolarisClusterAgent, error) {
	clusterAgent := clusteragent.NewDefaultPolarisClusterAgent(clusterAgentConfig, ginEngine, k8sClusterClient, logger)

	if err := clusterAgent.Start(ctx); err != nil {
		return nil, err
	}

	return clusterAgent, nil
}

func runNodeSampler(
	ctx context.Context,
	clusterAgentConfig *config.ClusterAgentConfig,
	pluginRegistry *pipeline.PluginsRegistry[pipeline.ClusterAgentServices],
	logger *logr.Logger,
	cmdLineArgs *commandLineArgs,
) error {
	k8sConfig, err := kubernetes.LoadKubeconfig(cmdLineArgs.kubeconfig, logger)
	if err != nil {
		return err
	}
	k8sClusterClient, err := setUpClusterClient(k8sConfig, logger)
	if err != nil {
		return err
	}

	ginEngine := gin.Default()
	ginEngine.SetTrustedProxies(nil)

	if _, err := startNodeSampler(ctx, clusterAgentConfig, k8sClusterClient, ginEngine, pluginRegistry, logger); err != nil {
		return err
	}

	if _, err := startClusterAgent(ctx, clusterAgentConfig, k8sClusterClient, ginEngine, logger); err != nil {
		return err
	}

	go func() {
		if err := ginEngine.Run(clusterAgentConfig.ListenOn...); err != nil {
			logger.Error(err, "Error executing HTTP server.")
		}
	}()
	return nil
}
