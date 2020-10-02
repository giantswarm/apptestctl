package bootstrap

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
)

const (
	installOperators = "install-operators"
	kubeconfig       = "kubeconfig"
)

type flag struct {
	InstallOperators bool
	KubeConfig       string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.InstallOperators, installOperators, "o", true, "Install app-operator and chart-operator")
	cmd.Flags().StringVarP(&f.KubeConfig, kubeconfig, "k", "", "Explicit kubeconfig for the target cluster")
}

func (f *flag) Validate() error {
	if f.KubeConfig == "" {
		return microerror.Maskf(invalidFlagError, "--%s must not be empty", kubeconfig)
	}

	return nil
}
