package bootstrap

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/giantswarm/apiextensions/v2/pkg/crd"
	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v2/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v4/pkg/k8srestconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	appOperatorVersion                = "2.2.0"
	chartOperatorVersion              = "2.3.1"
	controlPlaneTestCatalogStorageURL = "https://giantswarm.github.io/control-plane-catalog/"
	namespace                         = "giantswarm"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: r.logger,

			KubeConfig: r.flag.KubeConfig,
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var k8sClients k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger:     r.logger,
			RestConfig: restConfig,
		}
		k8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = r.ensureCRDs(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensurePriorityClass(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureNamespace(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.installOperators(ctx, helmClient)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) ensureCRDs(ctx context.Context, k8sClients k8sclient.Interface) error {
	// Ensure Application group CRDs are created.
	crds := []string{
		"AppCatalog",
		"App",
		"Chart",
	}

	{
		for _, crdName := range crds {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring %#q CRD", crdName))

			err := k8sClients.CRDClient().EnsureCreated(ctx, crd.LoadV1("application.giantswarm.io", crdName), backoff.NewMaxRetries(7, 1*time.Second))
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured %#q CRD exists", crdName))
		}
	}

	return nil
}

func (r *runner) ensureNamespace(ctx context.Context, k8sClients k8sclient.Interface) error {
	namespace := "giantswarm"

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring namespace %#q", namespace))

	o := func() error {
		{
			n := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, err := k8sClients.K8sClient().CoreV1().Namespaces().Create(ctx, n, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("namespace %#q already exists", namespace))
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			n, err := k8sClients.K8sClient().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			if n.Status.Phase != v1.NamespaceActive {
				return microerror.Maskf(executionFailedError, fmt.Sprintf("namespace in status %#q", n.Status.Phase))
			}
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

	err := backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured namespace %#q", namespace))

	return nil
}

func (r *runner) ensurePriorityClass(ctx context.Context, k8sClients k8sclient.Interface) error {
	priorityClassName := "giantswarm-critical"

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating priorityclass %#q", priorityClassName))

	pc := &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: priorityClassName,
		},
		Value:         1000000000,
		GlobalDefault: false,
		Description:   "This priority class is used by giantswarm kubernetes components.",
	}

	_, err := k8sClients.K8sClient().SchedulingV1().PriorityClasses().Create(ctx, pc, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("priorityclass %#q already exists", priorityClassName))
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created priorityclass %#q", priorityClassName))
	}

	return nil
}

func (r *runner) installOperators(ctx context.Context, helmClient helmclient.Interface) error {
	var err error

	operators := map[string]string{
		"app-operator":   appOperatorVersion,
		"chart-operator": chartOperatorVersion,
	}

	for name, version := range operators {
		err = r.installOperator(ctx, helmClient, name, version)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *runner) installOperator(ctx context.Context, helmClient helmclient.Interface, name, version string) error {
	var err error

	var operatorTarballPath string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting tarball URL for %#q", name))

		operatorTarballURL, err := appcatalog.GetLatestChart(ctx, controlPlaneTestCatalogStorageURL, name, version)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("tarball URL is %#q", operatorTarballURL))

		r.logger.LogCtx(ctx, "level", "debug", "message", "pulling tarball")

		operatorTarballPath, err = helmClient.PullChartTarball(ctx, operatorTarballURL)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("tarball path is %#q", operatorTarballPath))
	}

	{
		defer func() {
			fs := afero.NewOsFs()
			err := fs.Remove(operatorTarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", operatorTarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installing %#q", name))

		opts := helmclient.InstallOptions{
			ReleaseName: fmt.Sprintf("%s-unique", name),
		}

		err = helmClient.InstallReleaseFromTarball(ctx,
			operatorTarballPath,
			namespace,
			nil,
			opts)
		if helmclient.IsReleaseAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q already installed", name))
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installed %#q", name))
	}

	return nil
}
