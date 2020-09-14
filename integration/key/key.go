// +build k8srequired

package key

func DefaultCatalogName() string {
	return "default"
}

func DefaultCatalogStorageURL() string {
	return "https://giantswarm.github.com/default-catalog"
}

func Namespace() string {
	return "giantswarm"
}

func TestAppReleaseName() string {
	return "test-app"
}
