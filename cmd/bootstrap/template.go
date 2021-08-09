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

	operatorValuesYAML string = `operatorkit:
  resyncPeriod: "20s"

provider:
  kind: "aws"`
)
