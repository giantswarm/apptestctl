package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/crd"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v2/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v4/pkg/k8srestconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/to"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
)

const (
	appOperatorVersion            = "2.3.2"
	chartMuseumName               = "chartmuseum"
	chartMuseumCatalogStorageURL  = "http://chartmuseum-chartmuseum:8080/charts/"
	chartMuseumVersion            = "2.13.3"
	chartOperatorVersion          = "2.3.3"
	controlPlaneCatalogStorageURL = "https://giantswarm.github.io/control-plane-catalog/"
	helmStableCatalogName         = "helm-stable"
	helmStableCatalogStorageURL   = "https://kubernetes-charts.storage.googleapis.com/"
	namespace                     = "giantswarm"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value,omitempty"`
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

	var helmClient helmclient.Interface
	{
		c := helmclient.Config{
			K8sClient: k8sClients,
			Logger:    r.logger,
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var appTest apptest.Interface
	{
		c := apptest.Config{
			KubeConfig: r.flag.KubeConfig,

			Logger: r.logger,
		}
		appTest, err = apptest.New(c)
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

	if r.flag.InstallOperators {
		err = r.installOperators(ctx, helmClient)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.patchChartOperatorDeployment(ctx, k8sClients)
		if err != nil {
			return microerror.Mask(err)
		}

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "skipping installing operators")
	}

	err = r.installAppCatalogs(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureChartMuseumPSP(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.installChartMuseum(ctx, appTest)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) ensureCRDs(ctx context.Context, k8sClients k8sclient.Interface) error {
	// Ensure Application group CRDs are created.
	crds := []string{
		"AppCatalogEntry",
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

func (r *runner) installAppCatalogs(ctx context.Context, k8sClients k8sclient.Interface) error {
	var err error

	catalogs := map[string]string{
		chartMuseumName: chartMuseumCatalogStorageURL,
	}

	for name, url := range catalogs {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q appcatalog cr", name))

		appCatalogCR := &v1alpha1.AppCatalog{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					// Processed by app-operator-unique.
					label.AppOperatorVersion: "0.0.0",
				},
			},
			Spec: v1alpha1.AppCatalogSpec{
				Description: name,
				Title:       name,
				Storage: v1alpha1.AppCatalogSpecStorage{
					Type: "helm",
					URL:  url,
				},
			},
		}
		_, err = k8sClients.G8sClient().ApplicationV1alpha1().AppCatalogs().Create(ctx, appCatalogCR, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q appcatalog CR already exists", appCatalogCR.Name))
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created %#q appcatalog cr", name))
	}

	return nil
}

func (r *runner) ensureChartMuseumPSP(ctx context.Context, k8sClients k8sclient.Interface) error {
	name := "chartmuseum-psp"
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring psp %#q", name))

	o := func() error {
		{
			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups:     []string{"extensions"},
						Resources:     []string{"podsecuritypolicies"},
						ResourceNames: []string{name},
						Verbs:         []string{"use"},
					},
				},
			}
			_, err := k8sClients.K8sClient().RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("clusterRole %#q already exists", name))
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}
		{
			clusterRoleBinding := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "chartmuseum",
						Namespace: namespace,
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind:     "ClusterRole",
					Name:     name,
					APIGroup: "rbac.authorization.k8s.io",
				},
			}
			_, err := k8sClients.K8sClient().RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("clusterRoleBinding %#q already exists", name))
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}
		{
			psp := &policyv1beta1.PodSecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: policyv1beta1.PodSecurityPolicySpec{
					AllowPrivilegeEscalation: to.BoolP(true),
					Volumes: []policyv1beta1.FSType{
						policyv1beta1.All,
					},
					RunAsUser: policyv1beta1.RunAsUserStrategyOptions{
						Rule: policyv1beta1.RunAsUserStrategyRunAsAny,
					},
					SupplementalGroups: policyv1beta1.SupplementalGroupsStrategyOptions{
						Rule: policyv1beta1.SupplementalGroupsStrategyRunAsAny,
					},
					FSGroup: policyv1beta1.FSGroupStrategyOptions{
						Rule: policyv1beta1.FSGroupStrategyRunAsAny,
					},
					SELinux: policyv1beta1.SELinuxStrategyOptions{
						Rule: policyv1beta1.SELinuxStrategyRunAsAny,
					},
				},
			}
			_, err := k8sClients.K8sClient().PolicyV1beta1().PodSecurityPolicies().Create(ctx, psp, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("psp %#q already exists", name))
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			tcp := corev1.ProtocolTCP
			chartmuseumPort := intstr.FromInt(8080)

			np := &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chartmuseum",
					Namespace: namespace,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":     "chartmuseum",
							"release": "chartmuseum",
						},
					},
					Egress: []networkingv1.NetworkPolicyEgressRule{},
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: &tcp,
									Port:     &chartmuseumPort,
								},
							},
						},
					},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
						networkingv1.PolicyTypeEgress,
					},
				},
			}
			_, err := k8sClients.K8sClient().NetworkingV1().NetworkPolicies(namespace).Create(ctx, np, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("networkpolicy %#q already exists", name))
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

	err := backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured psp %#q", name))

	return nil
}

func (r *runner) installChartMuseum(ctx context.Context, appTest apptest.Interface) error {
	var err error

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q app cr", chartMuseumName))

		valuesYAML := `persistence:
  enabled: "true"
serviceAccount:
  name: "chartmuseum"
  create: "true"
env:
  open:
    DISABLE_API: false`

		apps := []apptest.App{
			{
				CatalogName: helmStableCatalogName,
				CatalogURL:  helmStableCatalogStorageURL,
				Name:        chartMuseumName,
				Namespace:   namespace,
				ValuesYAML:  valuesYAML,
				Version:     chartMuseumVersion,
			},
		}
		err = appTest.InstallApps(ctx, apps)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created %#q app cr", chartMuseumName))
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

		operatorTarballURL, err := appcatalog.GetLatestChart(ctx, controlPlaneCatalogStorageURL, name, version)
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

		// ReleaseName has unique suffix like in the control plane so the test
		// app CRs need to use 0.0.0 for the version label.
		opts := helmclient.InstallOptions{
			ReleaseName: fmt.Sprintf("%s-unique", name),
		}
		err = helmClient.InstallReleaseFromTarball(ctx,
			operatorTarballPath,
			namespace,
			nil,
			opts)
		if helmclient.IsCannotReuseRelease(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q already installed", name))
			return nil
		} else if helmclient.IsReleaseAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q already installed", name))
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installed %#q", name))
	}

	return nil
}

func (r *runner) patchChartOperatorDeployment(ctx context.Context, k8sClients k8sclient.Interface) error {
	labelSelector := "app.kubernetes.io/instance=chart-operator-unique"
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for chart-operator with label selector %#q", labelSelector))

	o := func() error {
		list, err := k8sClients.K8sClient().AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if apierrors.IsNotFound(err) || len(list.Items) != 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator deployment not created yet")
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		patches := []Patch{
			{
				Op:   "remove",
				Path: "/spec/template/spec/dnsConfig",
			},
			{
				Op:    "replace",
				Path:  "/spec/template/spec/dnsPolicy",
				Value: "ClusterFirst",
			},
		}

		bytes, err := json.Marshal(patches)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = k8sClients.K8sClient().AppsV1().Deployments(namespace).Patch(ctx, list.Items[0].Name, types.JSONPatchType, bytes, metav1.PatchOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

	err := backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("patching deployment for chart-operator with name %#q done", ""))

	return nil
}
