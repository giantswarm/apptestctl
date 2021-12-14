package bootstrap

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	installOperators = "install-operators"
	kubeconfig       = "kubeconfig"
	kubeconfigEnvVar = "KUBECONFIG"
	kubeconfigPath   = "kubeconfig-path"
	logLevel         = "log-level"
	wait             = "wait"
)

type flag struct {
	InstallOperators bool
	KubeConfig       string
	KubeConfigPath   string
	LogLevel         string
	Wait             bool
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.InstallOperators, installOperators, "o", true, "Install app-operator and chart-operator")
	cmd.Flags().StringVarP(&f.KubeConfig, kubeconfig, "k", "", "Explicit kubeconfig for the target cluster")
	cmd.Flags().StringVarP(&f.KubeConfigPath, kubeconfigPath, "p", "", "Path to a kubeconfig file for the target cluster")
	cmd.Flags().StringVarP(&f.LogLevel, logLevel, "l", "error", "Log level to be used for debug logging. Either debug, info, warning or error.")
	cmd.Flags().BoolVarP(&f.Wait, wait, "w", true, "Wait for all components to be ready")
}

func (f *flag) Validate() error {
	if f.KubeConfig == "" && f.KubeConfigPath == "" && os.Getenv(kubeconfigEnvVar) == "" {
		return microerror.Maskf(invalidFlagError, "either --%s or --%s or KUBECONFIG must be set", kubeconfig, kubeconfigPath)
	} else if f.KubeConfig != "" && f.KubeConfigPath != "" {
		return microerror.Maskf(invalidFlagError, "both --%s or --%s must not be set", kubeconfig, kubeconfigPath)
	}
	if !containsString([]string{"", "debug", "info", "warning", "error"}, f.LogLevel) {
		return microerror.Maskf(invalidFlagError, "Log level must be either debug, info, warning or error.")
	}

	return nil
}

func containsString(list []string, s string) bool {
	for _, l := range list {
		if l == s {
			return true
		}
	}

	return false
}
