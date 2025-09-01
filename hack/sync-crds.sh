#!/bin/bash

crds=( appcatalogentries appcatalogs apps catalogs charts )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/giantswarm/apiextensions-application/master/config/crd/application.giantswarm.io_${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done

crds=( ciliumnetworkpolicies ciliumclusterwidenetworkpolicies )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/cilium/cilium/main/pkg/k8s/apis/cilium.io/client/crds/v2/${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done

crds=( servicemonitors podmonitors prometheuses prometheusrules )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done

crds=( verticalpodautoscalers )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/FairwindsOps/charts/master/stable/vpa/crds/vpa-v1-crd.yaml" > "../pkg/crds/${crd}.yaml"
done

# Kyverno

curl -s "https://raw.githubusercontent.com/kyverno/kyverno/refs/tags/v1.14.4/config/crds/kyverno/kyverno.io_policyexceptions.yaml" > "../pkg/crds/policyexception.yaml"
curl -s "https://raw.githubusercontent.com/kyverno/kyverno/refs/tags/v1.14.4/config/crds/kyverno/kyverno.io_clusterpolicies.yaml" > "../pkg/crds/clusterpolicies.yaml"

# RemoteWrite

curl -s "https://raw.githubusercontent.com/giantswarm/prometheus-meta-operator/main/config/crd/monitoring.giantswarm.io_remotewrites.yaml" > "../pkg/crds/remotewrites.yaml"
