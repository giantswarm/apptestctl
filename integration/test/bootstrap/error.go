//go:build k8srequired
// +build k8srequired

package bootstrap

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}
