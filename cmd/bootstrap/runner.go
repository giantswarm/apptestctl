package bootstrap

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/giantswarm/apiextensions/v2/pkg/crd"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v4/pkg/k8srestconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
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

	k8sClientConfig := k8sclient.ClientsConfig{
		Logger:     r.logger,
		RestConfig: restConfig,
	}
	k8sClients, err := k8sclient.NewClients(k8sClientConfig)
	if err != nil {
		return microerror.Mask(err)
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
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring %#q CRD exists", crdName))

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
