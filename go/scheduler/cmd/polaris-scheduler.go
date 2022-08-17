package cmd

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

// Creates a new polaris-scheduler command.
func NewPolarisSchedulerCmd() *cobra.Command {
	var configPath string

	logger := initLogger()

	cmd := cobra.Command{
		Use: "polaris-scheduler",
		// ToDo: Extend long description.
		Long: "The Polaris Scheduler is a distributed, Edge-aware scheduler for Kubernetes.",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("polaris-scheduler")

			schedConfig, err := loadConfigWithDefaults(configPath, logger)
			if err != nil {
				logger.Error(err, "Error loading config.")
				os.Exit(1)
			}

			runScheduler(schedConfig, logger)
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "The path of the polaris-scheduler configuration file.")
	cmd.MarkFlagFilename("config")

	return &cmd
}

func initLogger() *logr.Logger {
	logger := stdr.New(nil)
	return &logger
}

func loadConfigWithDefaults(configPath string, logger *logr.Logger) (*config.SchedulerConfig, error) {
	schedConfig, err := loadConfig(configPath, logger)
	if err != nil {
		return nil, err
	}
	fillConfigWithDefaults(schedConfig)
	return schedConfig, nil
}

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

func fillConfigWithDefaults(schedConfig *config.SchedulerConfig) {
	config.SetDefaultsSchedulerConfig(schedConfig)
}

func runScheduler(schedConfig *config.SchedulerConfig, logger *logr.Logger) error {
	return nil
}
