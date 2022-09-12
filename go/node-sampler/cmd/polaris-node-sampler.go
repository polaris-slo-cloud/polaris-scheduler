package cmd

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/node-sampler/runtime"
)

type commandLineArgs struct {
	// The path to the sampler config.
	config string

	// The path to the KUBECONFIG file.
	kubeconfig string
}

// Creates a new polaris-node-sampler command.
func NewPolarisNodeSamplerCmd(ctx context.Context) *cobra.Command {
	cmdLineArgs := commandLineArgs{}

	logger := initLogger()

	cmd := cobra.Command{
		Use: "polaris-node-sampler",
		// ToDo: Extend long description.
		Long: "The Polaris Node Sampler is a component of the Polaris Scheduler that allows sampling nodes from a cluster.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("polaris-node-sampler")

			samplerConfig, err := loadConfigWithDefaults(cmdLineArgs.config, logger)
			if err != nil {
				logger.Error(err, "Error loading config.")
				os.Exit(1)
			}

			if err := runNodeSampler(ctx, samplerConfig, logger, &cmdLineArgs); err != nil {
				logger.Error(err, "Error starting polaris-node-sampler")
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&cmdLineArgs.config, "config", "c", "", "The path to the polaris-node-sampler configuration file.")
	cmd.MarkFlagFilename("config")
	cmd.PersistentFlags().StringVar(&cmdLineArgs.kubeconfig, "kubeconfig", "", "The path to the KUBECONFIG file.")
	cmd.MarkFlagFilename("kubeconfig")

	return &cmd
}

func initLogger() *logr.Logger {
	logger := stdr.New(nil)
	return &logger
}

// Loads the NodeSamplerConfig from the specified path and fills empty fields with default values.
func loadConfigWithDefaults(configPath string, logger *logr.Logger) (*config.NodeSamplerConfig, error) {
	samplerConfig := &config.NodeSamplerConfig{}

	if configPath != "" {
		if err := util.ParseYamlFile(configPath, samplerConfig); err != nil {
			return nil, err
		}
	}

	config.SetDefaultsNodeSamplerConfig(samplerConfig)
	return samplerConfig, nil
}

func runNodeSampler(
	ctx context.Context,
	samplerConfig *config.NodeSamplerConfig,
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
	clusterClientsMgr, err := kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, "polaris-node-sampler", logger)
	if err != nil {
		return err
	}

	clusterClient, err := clusterClientsMgr.GetClusterClient(k8sConfig.ServerName)
	if err != nil {
		return err
	}

	nodeSampler := runtime.NewDefaultPolarisNodeSampler(samplerConfig, clusterClient, logger)
	return nodeSampler.Start(ctx)
}
