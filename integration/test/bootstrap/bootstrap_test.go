// +build k8srequired

package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/apptestctl/integration/key"
)

// TestBootstrap TODO
//
func TestBootstrap(t *testing.T) {
	ctx := context.Background()

	var err error

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q appcatalog cr", key.DefaultCatalogName()))

		appCatalogCR := &v1alpha1.AppCatalog{
			ObjectMeta: metav1.ObjectMeta{
				Name: key.DefaultCatalogName(),
				Labels: map[string]string{
					// Processed by app-operator-unique.
					label.AppOperatorVersion: "0.0.0",
				},
			},
			Spec: v1alpha1.AppCatalogSpec{
				Description: key.DefaultCatalogName(),
				Title:       key.DefaultCatalogName(),
				Storage: v1alpha1.AppCatalogSpecStorage{
					Type: "helm",
					URL:  key.DefaultCatalogStorageURL(),
				},
			},
		}
		_, err = config.K8sClients.G8sClient().ApplicationV1alpha1().AppCatalogs().Create(ctx, appCatalogCR, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created %#q appcatalog cr", key.DefaultCatalogName()))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q app cr", key.TestAppReleaseName()))

		appCR := &v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.TestAppReleaseName(),
				Namespace: key.Namespace(),
				Labels: map[string]string{
					// Processed by app-operator-unique.
					label.AppOperatorVersion: "0.0.0",
				},
			},
			Spec: v1alpha1.AppSpec{
				Catalog: key.DefaultCatalogName(),
				KubeConfig: v1alpha1.AppSpecKubeConfig{
					InCluster: true,
				},
				Name:      key.TestAppReleaseName(),
				Namespace: key.Namespace(),
				Version:   "0.1.0",
			},
		}
		_, err = config.K8sClients.G8sClient().ApplicationV1alpha1().Apps(key.Namespace()).Create(ctx, appCR, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q app cr", key.TestAppReleaseName()))
	}
}
