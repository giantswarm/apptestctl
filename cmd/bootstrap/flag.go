package bootstrap

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	installOperators = "install-operators"
	kubeconfig       = "kubeconfig"
	kubeconfigEnvVar = "KUBECONFIG"
	kubeconfigPath   = "kubeconfig-path"
	wait             = "wait"
)

type flag struct {
	InstallOperators bool
	KubeConfig       string
	KubeConfigPath   string
	Wait             bool
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.InstallOperators, installOperators, "o", true, "Install app-operator and chart-operator")
	cmd.Flags().StringVarP(&f.KubeConfig, kubeconfig, "k", "", "Explicit kubeconfig for the target cluster")
	cmd.Flags().StringVarP(&f.KubeConfigPath, kubeconfigPath, "p", "", "Path to a kubeconfig file for the target cluster")
	cmd.Flags().BoolVarP(&f.Wait, wait, "w", true, "Wait for all components to be ready")
}

func (f *flag) Validate() error {
	if f.KubeConfig == "" && f.KubeConfigPath == "" && os.Getenv(kubeconfigEnvVar) == "" {
		return microerror.Maskf(invalidFlagError, "either --%s or --%s or KUBECONFIG must be set", kubeconfig, kubeconfigPath)
	} else if f.KubeConfig != "" && f.KubeConfigPath != "" {
		return microerror.Maskf(invalidFlagError, "both --%s or --%s must not be set", kubeconfig, kubeconfigPath)
	}

	return nil
}
