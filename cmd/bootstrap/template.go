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
    DISABLE_API: false`

	operatorValuesYAML string = `Installation:
  V1:
    Helm:
      HTTP:
        ClientTimeout: "30s"
      Kubernetes:
        WaitTimeout: "180s"
    Registry:
      Domain: "quay.io"`
)
