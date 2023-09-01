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

//go:embed verticalpodautoscalers.yaml
var verticalPodAutoscalers string

func CRDs() []string {
	return []string{
		appCatalogEntries,
		appCatalogs,
		apps,
		catalogs,
		charts,
		ciliumClusterwideNetworkPolicies,
		ciliumNetworkPolicies,
		serviceMonitors,
		podMonitors,
		verticalPodAutoscalers,
	}
}
