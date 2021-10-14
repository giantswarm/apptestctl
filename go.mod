module github.com/giantswarm/apptestctl

go 1.16

require (
	github.com/giantswarm/apiextensions/v3 v3.32.0
	github.com/giantswarm/appcatalog v0.6.0
	github.com/giantswarm/apptest v0.12.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/helmclient/v4 v4.8.0
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/to v0.3.0
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	k8s.io/api v0.20.11
	k8s.io/apiextensions-apiserver v0.20.11
	k8s.io/apimachinery v0.20.11
	k8s.io/client-go v0.20.11
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/containerd/containerd v1.3.4 => github.com/containerd/containerd v1.4.4
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	// Use moby v20.10.5 to fix build issue on darwin.
	github.com/docker/docker => github.com/moby/moby v20.10.9+incompatible
	// Use go-logr/logr v0.1.0 due to breaking changes in v0.2.0 that can't be applied.
	github.com/go-logr/logr v0.2.0 => github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	// Use mergo 0.3.11 due to bug in 0.3.9 merging Go structs.
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.11
	github.com/opencontainers/runc v0.1.1 => github.com/opencontainers/runc v1.0.0-rc93
	github.com/ulikunitz/xz => github.com/ulikunitz/xz v0.5.10
	// Same as go-logr/logr, klog/v2 is using logr v0.2.0
	k8s.io/klog/v2 v2.2.0 => k8s.io/klog/v2 v2.0.0
	// Use fork of CAPI with Kubernetes 1.18 support.
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)
