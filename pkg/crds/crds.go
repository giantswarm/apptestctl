package crds

import (
	_ "embed"
)

//go:embed appcatalogentries.yaml
var appCatalogEntries string

//go:embed appcatalogs.yaml
var appCatalogs string

//go:embed apps.yaml
var apps string

//go:embed catalogs.yaml
var catalogs string

//go:embed charts.yaml
var charts string

//go:embed ciliumclusterwidenetworkpolicies.yaml
var ciliumClusterwideNetworkPolicies string

//go:embed ciliumnetworkpolicies.yaml
var ciliumNetworkPolicies string

//go:embed servicemonitors.yaml
var serviceMonitors string

//go:embed podmonitors.yaml
var podMonitors string

//go:embed prometheuses.yaml
var prometheuses string

//go:embed prometheusrules.yaml
var prometheusrules string

//go:embed verticalpodautoscalers.yaml
var verticalPodAutoscalers string

//go:embed policyexception.yaml
var policyException string

//go:embed clusterpolicies.yaml
var clusterPolicies string

//go:embed remotewrites.yaml
var remotewrites string

func CRDs() []string {
	return []string{
		appCatalogEntries,
		appCatalogs,
		apps,
		catalogs,
		charts,
		ciliumClusterwideNetworkPolicies,
		ciliumNetworkPolicies,
		clusterPolicies,
		serviceMonitors,
		podMonitors,
		prometheuses,
		prometheusrules,
		verticalPodAutoscalers,
		policyException,
		remotewrites,
	}
}
