package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/podsubmission"
	polarisRuntime "polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/runtime"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
)

type commandLineArgs struct {
	// The path to the scheduler config.
	config string

	// The path to the KUBECONFIG file.
	kubeconfig string
}

// Creates a new polaris-scheduler command.
func NewPolarisSchedulerCmd(ctx context.Context, pluginsRegistry *pipeline.PluginsRegistry) *cobra.Command {
	cmdLineArgs := commandLineArgs{}

	logger := initLogger()

	cmd := cobra.Command{
		Use: "polaris-scheduler",
		// ToDo: Extend long description.
		Long: "The Polaris Scheduler is a distributed, Edge-aware scheduler for Kubernetes.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("polaris-scheduler")

			schedConfig, err := loadConfigWithDefaults(cmdLineArgs.config, logger)
			if err != nil {
				logger.Error(err, "Error loading config.")
				os.Exit(1)
			}

			if err := runScheduler(ctx, schedConfig, pluginsRegistry, logger, &cmdLineArgs); err != nil {
				logger.Error(err, "Error starting polaris-scheduler")
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&cmdLineArgs.config, "config", "c", "", "The path to the polaris-scheduler configuration file.")
	cmd.MarkFlagFilename("config")
	cmd.PersistentFlags().StringVar(&cmdLineArgs.kubeconfig, "kubeconfig", "", "The path to the KUBECONFIG file.")
	cmd.MarkFlagFilename("kubeconfig")

	return &cmd
}

func initLogger() *logr.Logger {
	logger := stdr.New(nil)
	return &logger
}

// Loads the SchedulerConfig from the specified path and fills empty fields with default values.
func loadConfigWithDefaults(configPath string, logger *logr.Logger) (*config.SchedulerConfig, error) {
	schedConfig := &config.SchedulerConfig{}

	if configPath != "" {
		if err := util.ParseYamlFile(configPath, schedConfig); err != nil {
			return nil, err
		}
	}

	config.SetDefaultsSchedulerConfig(schedConfig)
	return schedConfig, nil
}

func setUpPodSource(schedConfig *config.SchedulerConfig, clusterClientsMgr client.ClusterClientsManager, logger *logr.Logger) (pipeline.PodSource, error) {
	switch schedConfig.OperatingMode {
	case config.SingleCluster:
		return setUpLocalClusterPodSource(schedConfig, clusterClientsMgr, logger)
	case config.MultiCluster:
		return setUpSubmitPodApiPodSource(schedConfig, logger)
	default:
		return nil, fmt.Errorf("invalid \"operatingMode\": %s", schedConfig.OperatingMode)
	}
}

func setUpLocalClusterPodSource(schedConfig *config.SchedulerConfig, clusterClientsMgr client.ClusterClientsManager, logger *logr.Logger) (pipeline.PodSource, error) {
	logger.Info("Setting up local cluster PodSource.")

	podSource := kubernetes.NewKubernetesPodSource(clusterClientsMgr, schedConfig)
	if err := podSource.StartWatching(); err != nil {
		return nil, err
	}
	return podSource, nil
}

func setUpSubmitPodApiPodSource(schedConfig *config.SchedulerConfig, logger *logr.Logger) (pipeline.PodSource, error) {
	logger.Info("Setting up Submit Pod API PodSource.")
	ginEngine := gin.Default()
	ginEngine.SetTrustedProxies(nil)

	// ToDo: Add status endpoint
	// ginEngine.GET("/status", func(c *gin.Context) {
	// 	sampler.handleStatusRequest(c)
	// })

	submitPodApi := podsubmission.NewPodSubmissionApi(schedConfig)
	if err := submitPodApi.RegisterSubmitPodEndpoint("/pods", ginEngine); err != nil {
		return nil, err
	}

	go func() {
		if err := ginEngine.Run(schedConfig.SubmitPodListenOn...); err != nil {
			logger.Error(err, "Error executing HTTP server.")
		}
	}()

	return submitPodApi, nil
}

func runScheduler(
	ctx context.Context,
	schedConfig *config.SchedulerConfig,
	pluginsRegistry *pipeline.PluginsRegistry,
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

	clusterClientsMgr, err := kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, schedConfig.SchedulerName, logger)
	if err != nil {
		return err
	}

	podSource, err := setUpPodSource(schedConfig, clusterClientsMgr, logger)
	if err != nil {
		return err
	}

	polarisScheduler := polarisRuntime.NewDefaultPolarisScheduler(schedConfig, pluginsRegistry, podSource, clusterClientsMgr, logger)
	return polarisScheduler.Start(ctx)
}
