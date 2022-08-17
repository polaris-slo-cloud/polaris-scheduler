package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/scheduler/kubernetes"
)

type commandLineArgs struct {
	// The path to the scheduler config.
	config string

	// The path to the KUBECONFIG file.
	kubeconfig string
}

// Creates a new polaris-scheduler command.
func NewPolarisSchedulerCmd() *cobra.Command {
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

			if err := runScheduler(schedConfig, logger, &cmdLineArgs); err != nil {
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
	schedConfig, err := loadConfig(configPath, logger)
	if err != nil {
		return nil, err
	}
	fillConfigWithDefaults(schedConfig)
	return schedConfig, nil
}

// Loads the SchedulerConfig from the specified path or returns an empty config, if configPath is empty.
func loadConfig(configPath string, logger *logr.Logger) (*config.SchedulerConfig, error) {
	schedConfig := &config.SchedulerConfig{}

	if configPath == "" {
		return schedConfig, nil
	}

	logger.Info("Loading configuration file", "configPath", configPath)

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("the specified path is not a file, but a directory: %s", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(schedConfig); err != nil {
		return nil, err
	}

	return schedConfig, nil
}

// Fills empty fields in the SchedulerConfig with default values.
func fillConfigWithDefaults(schedConfig *config.SchedulerConfig) {
	config.SetDefaultsSchedulerConfig(schedConfig)
}

// Loads the Kubernetes config.
// First we attempt to load it from a pod environment (i.e., when operating inside a cluster).
// If this fails, we use the local KUBECONFIG file.
func loadKubeconfig(args *commandLineArgs, logger *logr.Logger) (*rest.Config, error) {
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
	kubeconfigPath := getKubeconfigPath(args.kubeconfig)
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

func runScheduler(schedConfig *config.SchedulerConfig, logger *logr.Logger, cmdLineArgs *commandLineArgs) error {
	k8sConfig, err := loadKubeconfig(cmdLineArgs, logger)
	if err != nil {
		return err
	}

	k8sConfigs := map[string]*rest.Config{
		k8sConfig.ServerName: k8sConfig,
	}

	clusterClientsMgr, err := kubernetes.NewKubernetesClusterClientsManager(k8sConfigs, schedConfig, logger)
	if err != nil {
		return err
	}

	// ToDo PodSource

	// ToDo Plugins (in separate files and registry creation in main.go)

	return nil
}
