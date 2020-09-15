// +build k8srequired

package bootstrap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v2/pkg/helmclient"
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
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring %#q app CR is deployed", key.ChartMuseumAppName()))

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
			config.Logger.Log("level", "debug", "message", fmt.Sprintf("failed to get app CR status '%s': retrying in %s", helmclient.StatusDeployed, t), "stack", fmt.Sprintf("%v", err))
		}

		b := backoff.NewExponential(20*time.Minute, 60*time.Second)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured %#q app CR is deployed", key.ChartMuseumAppName()))
	}
}
