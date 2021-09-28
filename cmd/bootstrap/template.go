package bootstrap

const (
	chartMuseumValuesYAML string = `persistence:
  enabled: "true"
serviceAccount:
  name: "chartmuseum"
  create: "true"
env:
  open:
    ALLOW_OVERWRITE: true
    DISABLE_API: false
probes:
  readiness:
    initialDelaySeconds: 10`

	// Set isManagementCluster to true so we ClusterFirst for chart-operator
	// DNS settings.
	operatorValuesYAML string = `isManagementCluster: "true"
operatorkit:
  resyncPeriod: "20s"

provider:
  kind: "aws"`
)
