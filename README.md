[![CircleCI](https://circleci.com/gh/giantswarm/apptestctl.svg?style=shield)](https://circleci.com/gh/giantswarm/apptestctl)

# apptestctl

Command line tool for using the Giant Swarm app platform in integration tests.

## Installation

This project uses Go modules. Be sure to have it outside your `$GOPATH` or
set `GO111MODULE=on` environment variable. Then regular `go install` should do
the trick. Alternatively the following one-liner may help.

```sh
GO111MODULE=on go install -ldflags "-X 'github.com/giantswarm/apptestctl/pkg/project.gitSHA=$(git rev-parse HEAD)'" .

# How to use this 

After creating the `kind` cluster on your local machine, type below command. 
```
 apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"

{"caller":"github.com/giantswarm/k8sclient/v4/pkg/k8srestconfig/rest_config.go:137","level":"debug","message":"creating REST config from kubeconfig","time":"2020-09-29T11:09:41.587218+00:00"}
{"caller":"github.com/giantswarm/k8sclient/v4/pkg/k8srestconfig/rest_config.go:145","level":"debug","message":"created REST config from kubeconfig","time":"2020-09-29T11:09:41.588999+00:00"}
{"caller":"github.com/giantswarm/apptestctl/cmd/bootstrap/runner.go:148","level":"debug","message":"ensuring `AppCatalog` CRD","time":"2020-09-29T11:09:41.651762+00:00"}
{"caller":"github.com/giantswarm/k8sclient/v4/pkg/k8scrdclient/crd_client.go:89","level":"debug","message":"creating CRD `appcatalogs.application.giantswarm.io`","time":"2020-09-29T11:09:41.726147+00:00"}
{"caller":"github.com/giantswarm/k8sclient/v4/pkg/k8scrdclient/crd_client.go:100","level":"debug","message":"created CRD `appcatalogs.application.giantswarm.io`","time":"2020-09-29T11:09:41.745538+00:00"}
...
```

It would automatically create all resources such as app-operator, chart-operator and CRDs for app testing.
  