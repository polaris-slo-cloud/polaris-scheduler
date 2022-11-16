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
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/clusteragent"
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
func NewPolarisSchedulerCmd(ctx context.Context, pluginsRegistry *pipeline.PluginsRegistry[pipeline.PolarisScheduler]) *cobra.Command {
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

func setUpK8sClusterClientsMgr(schedConfig *config.SchedulerConfig, cmdLineArgs *commandLineArgs, logger *logr.Logger) (*kubernetes.KubernetesClusterClientsManager, error) {
	k8sConfig, err := kubernetes.LoadKubeconfig(cmdLineArgs.kubeconfig, logger)
	if err != nil {
		return nil, err
	}

	k8sConfigs := map[string]*rest.Config{
		k8sConfig.ServerName: k8sConfig,
	}

	return kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, schedConfig.SchedulerName, logger)
}

func setUpClusterAgentClientsMgr(schedConfig *config.SchedulerConfig, logger *logr.Logger) (client.ClusterClientsManager, error) {
	if len(schedConfig.RemoteClusters) == 0 {
		return nil, fmt.Errorf("no remoteClusters configured in scheduler config")
	}
	clientsMap := make(map[string]*clusteragent.RemoteClusterAgentClient, len(schedConfig.RemoteClusters))

	for clusterName, clusterConfig := range schedConfig.RemoteClusters {
		clientsMap[clusterName] = clusteragent.NewRemoteClusterAgentClient(clusterName, clusterConfig, logger)
	}

	clientsMgr := client.NewGenericClusterClientsManager(clientsMap)
	return clientsMgr, nil
}

func setUpLocalClusterPodSource(schedConfig *config.SchedulerConfig, clusterClientsMgr *kubernetes.KubernetesClusterClientsManager, logger *logr.Logger) (pipeline.PodSource, error) {
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
	pluginsRegistry *pipeline.PluginsRegistry[pipeline.PolarisScheduler],
	logger *logr.Logger,
	cmdLineArgs *commandLineArgs,
) error {
	var clusterClientsMgr client.ClusterClientsManager
	var podSource pipeline.PodSource
	var err error

	switch schedConfig.OperatingMode {
	case config.SingleCluster:
		k8sClusterClientsMgr, err := setUpK8sClusterClientsMgr(schedConfig, cmdLineArgs, logger)
		if err != nil {
			return err
		}
		clusterClientsMgr = k8sClusterClientsMgr
		podSource, err = setUpLocalClusterPodSource(schedConfig, k8sClusterClientsMgr, logger)
		if err != nil {
			return err
		}
	case config.MultiCluster:
		clusterClientsMgr, err = setUpClusterAgentClientsMgr(schedConfig, logger)
		if err != nil {
			return err
		}
		podSource, err = setUpSubmitPodApiPodSource(schedConfig, logger)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid \"operatingMode\": %s", schedConfig.OperatingMode)
	}

	polarisScheduler := polarisRuntime.NewDefaultPolarisScheduler(schedConfig, pluginsRegistry, podSource, clusterClientsMgr, logger)
	return polarisScheduler.Start(ctx)
}
