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

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/broker"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/sampling"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
)

type commandLineArgs struct {
	// The path to the cluster-broker config.
	config string

	// The path to the KUBECONFIG file.
	kubeconfig string
}

// Creates a new polaris-cluster-broker command.
func NewPolarisClusterBrokerCmd(ctx context.Context, samplingStrategies []sampling.SamplingStrategyFactoryFunc) *cobra.Command {
	cmdLineArgs := commandLineArgs{}

	logger := initLogger()

	cmd := cobra.Command{
		Use: "polaris-cluster-broker",
		// ToDo: Extend long description.
		Long: "The Polaris Cluster Broker is a component of the Polaris Scheduler that allows the scheduler to interact with the local cluster. Brokers in multiple clusters allow interaction with multiple clusters.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("polaris-cluster-broker")

			samplerConfig, err := loadConfigWithDefaults(cmdLineArgs.config, logger)
			if err != nil {
				logger.Error(err, "Error loading config.")
				os.Exit(1)
			}

			if err := runNodeSampler(ctx, samplerConfig, samplingStrategies, logger, &cmdLineArgs); err != nil {
				logger.Error(err, "Error starting polaris-cluster-broker")
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&cmdLineArgs.config, "config", "c", "", "The path to the polaris-cluster-broker configuration file.")
	cmd.MarkFlagFilename("config")
	cmd.PersistentFlags().StringVar(&cmdLineArgs.kubeconfig, "kubeconfig", "", "The path to the KUBECONFIG file.")
	cmd.MarkFlagFilename("kubeconfig")

	return &cmd
}

func initLogger() *logr.Logger {
	logger := stdr.New(nil)
	return &logger
}

// Loads the ClusterBrokerConfig from the specified path and fills empty fields with default values.
func loadConfigWithDefaults(configPath string, logger *logr.Logger) (*config.ClusterBrokerConfig, error) {
	clusterBrokerConfig := &config.ClusterBrokerConfig{}

	if configPath != "" {
		if err := util.ParseYamlFile(configPath, clusterBrokerConfig); err != nil {
			return nil, err
		}
	}

	config.SetDefaultsClusterBrokerConfig(clusterBrokerConfig)
	return clusterBrokerConfig, nil
}

func setUpClusterClient(k8sConfig *rest.Config, logger *logr.Logger) (kubernetes.KubernetesClusterClient, error) {
	k8sConfigs := map[string]*rest.Config{
		k8sConfig.ServerName: k8sConfig,
	}

	// We only need a single cluster client in the sampler, but we reuse the ClusterClientsManager abstraction.
	clusterClientsMgr, err := kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, "polaris-cluster-broker", logger)
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

func setUpNodesCache(clusterBrokerConfig *config.ClusterBrokerConfig, clusterClient kubernetes.KubernetesClusterClient) (client.NodesCache, error) {
	updateInterval, err := time.ParseDuration(fmt.Sprintf("%vms", clusterBrokerConfig.NodesCacheUpdateIntervalMs))
	if err != nil {
		return nil, fmt.Errorf("error parsing nodesCacheUpdateIntervalMs: %v", err)
	}

	nodesCache := kubernetes.NewKubernetesNodesCache(clusterClient, updateInterval, int(clusterBrokerConfig.NodesCacheUpdateQueueSize))
	return nodesCache, nil
}

func startNodeSampler(
	ctx context.Context,
	clusterBrokerConfig *config.ClusterBrokerConfig,
	k8sClusterClient kubernetes.KubernetesClusterClient,
	ginEngine *gin.Engine,
	samplingStrategies []sampling.SamplingStrategyFactoryFunc,
	logger *logr.Logger,
) (sampling.PolarisNodeSampler, error) {
	nodesCache, err := setUpNodesCache(clusterBrokerConfig, k8sClusterClient)
	if err != nil {
		return nil, err
	}

	nodeSampler := sampling.NewDefaultPolarisNodeSampler(clusterBrokerConfig, ginEngine, k8sClusterClient, nodesCache, samplingStrategies, logger)
	err = nodeSampler.Start(ctx)
	if err != nil {
		return nil, err
	}

	return nodeSampler, nil
}

func startClusterBroker(
	ctx context.Context,
	clusterBrokerConfig *config.ClusterBrokerConfig,
	k8sClusterClient kubernetes.KubernetesClusterClient,
	ginEngine *gin.Engine,
	logger *logr.Logger,
) (broker.PolarisClusterBroker, error) {
	clusterBroker := broker.NewDefaultPolarisClusterBroker(clusterBrokerConfig, ginEngine, k8sClusterClient, logger)

	if err := clusterBroker.Start(ctx); err != nil {
		return nil, err
	}

	return clusterBroker, nil
}

func runNodeSampler(
	ctx context.Context,
	clusterBrokerConfig *config.ClusterBrokerConfig,
	samplingStrategies []sampling.SamplingStrategyFactoryFunc,
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

	if _, err := startNodeSampler(ctx, clusterBrokerConfig, k8sClusterClient, ginEngine, samplingStrategies, logger); err != nil {
		return err
	}

	if _, err := startClusterBroker(ctx, clusterBrokerConfig, k8sClusterClient, ginEngine, logger); err != nil {
		return err
	}

	go func() {
		if err := ginEngine.Run(clusterBrokerConfig.ListenOn...); err != nil {
			logger.Error(err, "Error executing HTTP server.")
		}
	}()
	return nil
}
