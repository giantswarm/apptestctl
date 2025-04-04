package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v8/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/to"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/apptestctl/pkg/crds"
)

const (
	appOperatorVersion             = "6.7.0"
	chartMuseumCatalogHelmIndexURL = "https://chartmuseum.github.io/charts"
	chartMuseumCatalogName         = "apptestctl-chartmuseum"
	chartMuseumCatalogStorageURL   = "http://chartmuseum:8080/"
	chartMuseumName                = "chartmuseum"
	chartMuseumVersion             = "3.9.3"
	chartOperatorVersion           = "2.35.0"
	controlPlaneCatalogStorageURL  = "https://giantswarm.github.io/control-plane-catalog/"
	namespace                      = "giantswarm"
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

	var logger micrologger.Logger
	{
		c := micrologger.ActivationLoggerConfig{
			Underlying: r.logger,

			Activations: map[string]interface{}{
				micrologger.KeyLevel: r.flag.LogLevel,
			},
		}
		logger, err = micrologger.NewActivation(c)
		if err != nil {
			panic(err)
		}
		r.logger = logger
	}

	var kubeConfig string
	var restConfig *rest.Config
	{
		if r.flag.KubeConfig != "" {
			// Set kube config for passing to the apptest library.
			kubeConfig = r.flag.KubeConfig

			bytes := []byte(kubeConfig)
			restConfig, err = clientcmd.RESTConfigFromKubeConfig(bytes)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if r.flag.KubeConfigPath != "" {
			restConfig, err = clientcmd.BuildConfigFromFlags("", r.flag.KubeConfigPath)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if os.Getenv(kubeconfigEnvVar) != "" {
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			mergedConfig, err := loadingRules.Load()
			if err != nil {
				return microerror.Mask(err)
			}

			json, err := runtime.Encode(clientcmdlatest.Codec, mergedConfig)
			if err != nil {
				return microerror.Mask(err)
			}

			bytes, err := yaml.JSONToYAML(json)
			if err != nil {
				return microerror.Mask(err)
			}

			// Set kube config for passing to the apptest library.
			kubeConfig = string(bytes)

			restConfig, err = clientcmd.RESTConfigFromKubeConfig(bytes)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			// Shouldn't happen but returning error just in case.
			return microerror.Maskf(invalidConfigError, "KubeConfig and KubeConfigPath must not be empty at the same time")
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
			K8sClient:  k8sClients.K8sClient(),
			Logger:     r.logger,
			RestClient: k8sClients.RESTClient(),
			RestConfig: k8sClients.RESTConfig(),
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var appTest apptest.Interface
	{
		c := apptest.Config{
			KubeConfig:     kubeConfig,
			KubeConfigPath: r.flag.KubeConfigPath,

			Logger: logger,
		}
		appTest, err = apptest.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	_, _ = fmt.Fprintln(r.stdout, "bootstrapping app platform components")

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

	// If --install-operators is false we stop here. This is useful when we
	// don't want to use the pinned app-operator and chart-operator versions.
	if !r.flag.InstallOperators {
		_, _ = fmt.Fprintln(r.stdout, "skipping installing operators")
		return nil
	}

	err = r.installOperators(ctx, helmClient, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.installCatalogs(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	hasPSP, err := r.hasPSP(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureChartMuseumPSP(ctx, k8sClients, hasPSP)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.installChartMuseum(ctx, appTest)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.waitForChartMuseum(ctx, appTest)
	if err != nil {
		return microerror.Mask(err)
	}

	_, _ = fmt.Fprintln(r.stdout, "app platform components are ready")

	return nil
}

func (r *runner) ensureCRDs(ctx context.Context, k8sClients k8sclient.Interface) error {
	var err error

	{
		for _, crdYAML := range crds.CRDs() {
			var crd apiextensionsv1.CustomResourceDefinition

			err = yaml.Unmarshal([]byte(crdYAML), &crd)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "creating CRD %#q", crd.Name)

			err = k8sClients.CtrlClient().Create(ctx, &crd)
			if apierrors.IsAlreadyExists(err) {
				r.logger.Debugf(ctx, "%#q already exists", crd.Name)
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "created %#q CRD", crd.Name)
		}
	}

	return nil
}

func (r *runner) ensureNamespace(ctx context.Context, k8sClients k8sclient.Interface) error {
	namespace := "giantswarm"

	r.logger.Debugf(ctx, "ensuring namespace %#q", namespace)

	o := func() error {
		{
			n := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, err := k8sClients.K8sClient().CoreV1().Namespaces().Create(ctx, n, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				r.logger.Debugf(ctx, "namespace %#q already exists", namespace)
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
				return microerror.Maskf(executionFailedError, "namespace in status %#q", n.Status.Phase)
			}
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

	err := backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensured namespace %#q", namespace)

	return nil
}

func (r *runner) ensurePriorityClass(ctx context.Context, k8sClients k8sclient.Interface) error {
	priorityClassName := "giantswarm-critical"

	r.logger.Debugf(ctx, "creating priorityclass %#q", priorityClassName)

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
		r.logger.Debugf(ctx, "priorityclass %#q already exists", priorityClassName)
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.Debugf(ctx, "created priorityclass %#q", priorityClassName)
	}

	return nil
}

func (r *runner) installCatalogs(ctx context.Context, k8sClients k8sclient.Interface) error {
	var err error

	catalogs := map[string]string{
		chartMuseumName: chartMuseumCatalogStorageURL,
	}

	for name, url := range catalogs {
		r.logger.Debugf(ctx, "creating %#q catalog cr", name)

		catalogCR := &v1alpha1.Catalog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: metav1.NamespaceDefault,
			},
			Spec: v1alpha1.CatalogSpec{
				Description: name,
				Title:       name,
				Storage: v1alpha1.CatalogSpecStorage{
					Type: "helm",
					URL:  url,
				},
				Repositories: []v1alpha1.CatalogSpecRepository{
					{
						Type: "helm",
						URL:  url,
					},
				},
			},
		}
		err = k8sClients.CtrlClient().Create(ctx, catalogCR)
		if apierrors.IsAlreadyExists(err) {
			r.logger.Debugf(ctx, "%#q catalog CR already exists", catalogCR.Name)
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "created %#q catalog cr", name)
	}

	return nil
}

func (r *runner) hasPSP(ctx context.Context, k8sClients k8sclient.Interface) (bool, error) {

	list, err := k8sClients.K8sClient().Discovery().ServerGroups()
	if err != nil {
		return false, microerror.Mask(err)
	}

	for _, group := range list.Groups {
		if group.Name == "policy" {
			for _, gv := range group.Versions {
				if gv.GroupVersion == "policy/v1beta1" {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (r *runner) ensureChartMuseumPSP(ctx context.Context, k8sClients k8sclient.Interface, installPSP bool) error {
	name := "chartmuseum-psp"
	r.logger.Debugf(ctx, "ensuring additional chartmuseum resources %#q", name)

	o := func() error {
		if installPSP {
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
					r.logger.Debugf(ctx, "clusterRole %#q already exists", name)
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
					r.logger.Debugf(ctx, "clusterRoleBinding %#q already exists", name)
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
					r.logger.Debugf(ctx, "psp %#q already exists", name)
					// fall through
				} else if err != nil {
					return microerror.Mask(err)
				}
			}
		}

		{
			tcp := v1.ProtocolTCP
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
				r.logger.Debugf(ctx, "networkpolicy %#q already exists", name)
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

	r.logger.Debugf(ctx, "ensured additional chartmuseum resources %#q", name)

	return nil
}

func (r *runner) installChartMuseum(ctx context.Context, appTest apptest.Interface) error {
	var err error

	{
		r.logger.Debugf(ctx, "creating %#q app cr", chartMuseumName)

		apps := []apptest.App{
			{
				CatalogName:   chartMuseumCatalogName,
				CatalogURL:    chartMuseumCatalogHelmIndexURL,
				Name:          chartMuseumName,
				Namespace:     namespace,
				ValuesYAML:    chartMuseumValuesYAML,
				Version:       chartMuseumVersion,
				WaitForDeploy: r.flag.Wait,
			},
		}
		err = appTest.InstallApps(ctx, apps)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "created %#q app cr", chartMuseumName)
	}

	return nil
}

func (r *runner) installOperators(ctx context.Context, helmClient helmclient.Interface, k8sClients k8sclient.Interface) error {
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
	var operatorTarballPath string
	{
		r.logger.Debugf(ctx, "getting tarball URL for %#q", name)

		operatorTarballURL, err := appcatalog.GetLatestChart(ctx, controlPlaneCatalogStorageURL, name, version)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "tarball URL is %#q", operatorTarballURL)

		r.logger.Debugf(ctx, "pulling tarball")

		operatorTarballPath, err = helmClient.PullChartTarball(ctx, operatorTarballURL)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "tarball path is %#q", operatorTarballPath)
	}

	{
		defer func() {
			fs := afero.NewOsFs()
			err := fs.Remove(operatorTarballPath)
			if err != nil {
				r.logger.Errorf(ctx, err, "deletion of %#q failed", operatorTarballPath)
			}
		}()

		r.logger.Debugf(ctx, "installing %#q", name)

		var input map[string]interface{}

		err := yaml.Unmarshal([]byte(operatorValuesYAML), &input)
		if err != nil {
			return microerror.Mask(err)
		}

		opts := helmclient.InstallOptions{
			ReleaseName: name,
		}
		err = helmClient.InstallReleaseFromTarball(ctx,
			operatorTarballPath,
			namespace,
			input,
			opts)
		if helmclient.IsCannotReuseRelease(err) {
			r.logger.Debugf(ctx, "%#q already installed", name)
			return nil
		} else if helmclient.IsReleaseAlreadyExists(err) {
			r.logger.Debugf(ctx, "%#q already installed", name)
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "installed %#q", name)
	}

	return nil
}

func (r *runner) waitForChartMuseum(ctx context.Context, appTest apptest.Interface) error {
	var err error

	deployName := chartMuseumName

	r.logger.Debugf(ctx, "waiting for ready %#q deployment", deployName)

	o := func() error {
		deploy, err := appTest.K8sClient().AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		if *deploy.Spec.Replicas != deploy.Status.ReadyReplicas {
			return microerror.Maskf(executionFailedError, "waiting for %d ready pods, current %d", *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
		}

		return nil
	}

	n := func(err error, t time.Duration) {
		r.logger.Errorf(ctx, err, "failed to get ready deployment '%s': retrying in %s", deployName, t)
	}

	b := backoff.NewConstant(5*time.Minute, 10*time.Second)
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "waited for ready %#q deployment", deployName)

	return nil
}
