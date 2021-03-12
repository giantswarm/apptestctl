module github.com/giantswarm/apptestctl

go 1.16

require (
	github.com/giantswarm/apiextensions/v3 v3.19.0
	github.com/giantswarm/appcatalog v0.4.0
	github.com/giantswarm/apptest v0.10.2
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/helmclient/v4 v4.3.0
	github.com/giantswarm/k8sclient/v5 v5.11.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/to v0.3.0
	github.com/spf13/afero v1.5.1
	github.com/spf13/cobra v1.1.3
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/containerd/containerd v1.3.4 => github.com/containerd/containerd v1.4.4
	// Vulnerabilities in etcd
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	// Use moby v20.10.5 to fix build issue on darwin.
	github.com/docker/docker => github.com/moby/moby v20.10.5+incompatible
	// Use go-logr/logr v0.1.0 due to breaking changes in v0.2.0 that can't be applied.
	github.com/go-logr/logr v0.2.0 => github.com/go-logr/logr v0.1.0
	// Use mergo 0.3.11 due to bug in 0.3.9 merging Go structs.
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.11
	github.com/opencontainers/runc v0.1.1 => github.com/opencontainers/runc v1.0.0-rc93
	// Same as go-logr/logr, klog/v2 is using logr v0.2.0
	k8s.io/klog/v2 v2.2.0 => k8s.io/klog/v2 v2.0.0
	// Use fork of CAPI with Kubernetes 1.18 support.
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
)
