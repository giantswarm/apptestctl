module github.com/giantswarm/apptestctl

go 1.15

require (
	github.com/giantswarm/apiextensions/v3 v3.4.0
	github.com/giantswarm/appcatalog v0.2.7
	github.com/giantswarm/apptest v0.4.1-0.20201030140104-d7a74a9bb5a5
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/helmclient/v3 v3.0.0
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/micrologger v0.3.3
	github.com/giantswarm/to v0.3.0
	github.com/moby/term v0.0.0-20200915141129-7f0af18e79f2 // indirect
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.0.0
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
)

replace (
	// Use moby v20.10.0-beta1 to fix build issue on darwin.
	github.com/docker/docker => github.com/moby/moby v20.10.0-beta1+incompatible
	// Use fork of CAPI with Kubernetes 1.18 support.
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
)
