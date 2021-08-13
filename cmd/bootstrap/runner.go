package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/apptestctl/pkg/project"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"
)

const (
	appOperatorName    = "app-operator"
	appOperatorVersion = "5.1.1-260404d1d7df9e58a7daa3c1b22ee574d13a7c8f"
	// TODO Fix name
	appPlatformName               = "apptestlctl"
	chartMuseumName               = "chartmuseum"
	controlPlaneCatalogStorageURL = "https://giantswarm.github.io/control-plane-test-catalog/"
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

	err = r.ensureNamespace(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureChartMuseumPSP(ctx, k8sClients)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.installAppPlatform(ctx, helmClient)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.installAppOperator(ctx, helmClient)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.waitForChartMuseum(ctx, k8sClients.K8sClient())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) ensureNamespace(ctx context.Context, k8sClients k8sclient.Interface) error {
	namespace := "giantswarm"

	r.logger.Debugf(ctx, "ensuring namespace %#q", namespace)

	o := func() error {
		{
			n := &corev1.Namespace{
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

	r.logger.Debugf(ctx, "ensured namespace %#q", namespace)

	return nil
}

func (r *runner) ensureChartMuseumPSP(ctx context.Context, k8sClients k8sclient.Interface) error {
	name := "chartmuseum-psp"
	r.logger.Debugf(ctx, "ensuring psp %#q", name)

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

	r.logger.Debugf(ctx, "ensured psp %#q", name)

	return nil
}

func (r *runner) installAppPlatform(ctx context.Context, helmClient helmclient.Interface) error {
	err := r.installHelmRelease(ctx, helmClient, appPlatformName, project.Version(), "")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) installAppOperator(ctx context.Context, helmClient helmclient.Interface) error {
	err := r.installHelmRelease(ctx, helmClient, appOperatorName, appOperatorVersion, appOperatorValuesYAML)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) installHelmRelease(ctx context.Context, helmClient helmclient.Interface, name, version, valuesYAML string) error {
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

		err := yaml.Unmarshal([]byte(valuesYAML), &input)
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

func (r *runner) waitForChartMuseum(ctx context.Context, k8sClient kubernetes.Interface) error {
	var err error

	deployName := fmt.Sprintf("%s-%s", chartMuseumName, chartMuseumName)

	r.logger.Debugf(ctx, "waiting for ready %#q deployment", deployName)

	o := func() error {
		deploy, err := k8sClient.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
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
