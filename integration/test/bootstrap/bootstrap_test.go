//go:build k8srequired
// +build k8srequired

package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/apptestctl/integration/key"
)

// TestBootstrap ensures that the chartmuseum app CR is deployed. This confirms
// that the app platform is installed and can deploy app CRs.
//
func TestBootstrap(t *testing.T) {
	ctx := context.Background()

	{
		config.Logger.Debugf(ctx, "ensuring %#q app CR is deployed", key.ChartMuseumAppName())

		o := func() error {
			app, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Apps(key.Namespace()).Get(ctx, key.ChartMuseumAppName(), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			if app.Status.Release.Status != helmclient.StatusDeployed {
				return microerror.Maskf(executionFailedError, "waiting for %#q, current %#q", helmclient.StatusDeployed, app.Status.Release.Status)
			}
			return nil
		}

		n := func(err error, t time.Duration) {
			config.Logger.Errorf(ctx, err, "failed to get app CR status '%s': retrying in %s", helmclient.StatusDeployed, t)
		}

		b := backoff.NewConstant(20*time.Minute, 60*time.Second)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "ensured %#q app CR is deployed", key.ChartMuseumAppName())
	}
}
