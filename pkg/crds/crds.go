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

func CRDs() []string {
	return []string{
		appCatalogEntries,
		appCatalogs,
		apps,
		catalogs,
		charts,
	}
}
