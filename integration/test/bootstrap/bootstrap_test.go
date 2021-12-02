//go:build k8srequired
// +build k8srequired

package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/apptestctl/integration/key"
)

const (
	statusDeployed = "deployed"
)

// TestBootstrap ensures that the chartmuseum app CR is deployed. This confirms
// that the app platform is installed and can deploy app CRs.
//
func TestBootstrap(t *testing.T) {
	ctx := context.Background()

	{
		config.Logger.Debugf(ctx, "ensuring %#q app CR is deployed", key.ChartMuseumAppName())

		var app v1alpha.App

		o := func() error {
			app, err := config.K8sClients.CtrlClient().Get(
				ctx,
				types.NamespacedName{name: key.ChartMuseumAppName(), namespace: key.Namespace},
				&app)
			if err != nil {
				return microerror.Mask(err)
			}
			if app.Status.Release.Status != statusDeployed {
				return microerror.Maskf(executionFailedError, "waiting for %#q, current %#q", statusDeployed, app.Status.Release.Status)
			}
			return nil
		}

		n := func(err error, t time.Duration) {
			config.Logger.Errorf(ctx, err, "failed to get app CR status '%s': retrying in %s", statusDeployed, t)
		}

		b := backoff.NewConstant(20*time.Minute, 60*time.Second)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "ensured %#q app CR is deployed", key.ChartMuseumAppName())
	}
}
