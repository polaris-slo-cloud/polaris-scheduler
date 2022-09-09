package cmd

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/k8s-connector/kubernetes"
)

type commandLineArgs struct {
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

			if err := runNodeSampler(ctx, logger, &cmdLineArgs); err != nil {
				logger.Error(err, "Error starting polaris-node-sampler")
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cmdLineArgs.kubeconfig, "kubeconfig", "", "The path to the KUBECONFIG file.")
	cmd.MarkFlagFilename("kubeconfig")

	return &cmd
}

func initLogger() *logr.Logger {
	logger := stdr.New(nil)
	return &logger
}

func runNodeSampler(
	ctx context.Context,
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

	_ = clusterClientsMgr
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	go func() {
		if err := r.Run(); err != nil { // listen and serve on 0.0.0.0:8080
			logger.Error(err, "Error executing HTTP server.")
		}
	}()
	return nil
}
