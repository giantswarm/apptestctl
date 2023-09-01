#!/bin/bash

crds=( appcatalogentries appcatalogs apps catalogs charts )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/giantswarm/apiextensions-application/master/config/crd/application.giantswarm.io_${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done

crds=( ciliumnetworkpolicies ciliumclusterwidenetworkpolicies )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/cilium/cilium/main/pkg/k8s/apis/cilium.io/client/crds/v2/${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done

crds=( servicemonitors podmonitors )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done

crds=( verticalpodautoscalers )

for crd in "${crds[@]}"; do
    curl -s "https://raw.githubusercontent.com/giantswarm/vertical-pod-autoscaler-crd/master/helm/vertical-pod-autoscaler-crd/templates/${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done
