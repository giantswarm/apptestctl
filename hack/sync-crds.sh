#!/bin/bash

crds=( appcatalogentries appcatalogs apps catalogs charts )

for crd in "${crds[@]}"; do
        curl -s "https://raw.githubusercontent.com/giantswarm/apiextension-application/master/config/crd/application.giantswarm.io_${crd}.yaml" > "../pkg/crds/${crd}.yaml"
done
