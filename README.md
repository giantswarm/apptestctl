[![CircleCI](https://circleci.com/gh/giantswarm/apptestctl.svg?style=shield)](https://circleci.com/gh/giantswarm/apptestctl)

# apptestctl

Command line tool for using the Giant Swarm app platform in integration tests.

## Installation

This project uses Go modules. Be sure to have it outside your `$GOPATH` or
set `GO111MODULE=on` environment variable. Then regular `go install` should do
the trick. Alternatively the following one-liner may help.

```sh
GO111MODULE=on go install -ldflags "-X 'github.com/giantswarm/apptestctl/pkg/project.gitSHA=$(git rev-parse HEAD)'" .
```

## Usage

After creating a [kind](https://kind.sigs.k8s.io/) cluster on your local machine, type below command. 

```sh
apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)"

{"caller":"github.com/giantswarm/k8sclient/v5/pkg/k8srestconfig/rest_config.go:137","level":"debug","message":"creating REST config from kubeconfig","time":"2020-09-29T11:09:41.587218+00:00"}
{"caller":"github.com/giantswarm/k8sclient/v5/pkg/k8srestconfig/rest_config.go:145","level":"debug","message":"created REST config from kubeconfig","time":"2020-09-29T11:09:41.588999+00:00"}
...
```

It will automatically create all resources such as app-operator, chart-operator and CRDs for app testing.

## Update CRDs

The bootstrap command installs CRDs in the group `application.giantswarm.io`.
These are embedded in `pkg/crds` to avoid hitting GitHub API rate limits.

The CRD manifests can be synced with [apiextensions](https://github.com/giantswarm/apiextensions)
using the Makefile.

```sh
make sync-crds
```
