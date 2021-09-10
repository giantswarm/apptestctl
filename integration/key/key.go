//go:build k8srequired
// +build k8srequired

package key

func ChartMuseumAppName() string {
	return "chartmuseum"
}

func Namespace() string {
	return "giantswarm"
}
