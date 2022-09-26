package cmd

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/cluster-broker/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/cluster-broker/runtime"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/cluster-broker/sampling"
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
	samplerConfig := &config.ClusterBrokerConfig{}

	if configPath != "" {
		if err := util.ParseYamlFile(configPath, samplerConfig); err != nil {
			return nil, err
		}
	}

	config.SetDefaultsClusterBrokerConfig(samplerConfig)
	return samplerConfig, nil
}

func runNodeSampler(
	ctx context.Context,
	samplerConfig *config.ClusterBrokerConfig,
	samplingStrategies []sampling.SamplingStrategyFactoryFunc,
	logger *logr.Logger,
	cmdLineArgs *commandLineArgs,
) error {
	k8sConfig, err := kubernetes.LoadKubeconfig(cmdLineArgs.kubeconfig, logger)
	if err != nil {
		return err
	}

	k8sConfigs := map[string]*rest.Config{
		k8sConfig.ServerName: k8sConfig,
	}

	// We only need a single cluster client in the sampler, but we reuse the ClusterClientsManager abstraction.
	clusterClientsMgr, err := kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, "polaris-cluster-broker", logger)
	if err != nil {
		return err
	}

	clusterClient, err := clusterClientsMgr.GetClusterClient(k8sConfig.ServerName)
	if err != nil {
		return err
	}
	k8sClusterClient, ok := clusterClient.(kubernetes.KubernetesClusterClient)
	if !ok {
		panic("KubernetesClusterClientsManager does not return KubernetesClusterClients")
	}

	nodeSampler := runtime.NewDefaultPolarisNodeSampler(samplerConfig, k8sClusterClient, samplingStrategies, logger)
	return nodeSampler.Start(ctx)
}
