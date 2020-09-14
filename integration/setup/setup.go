// +build k8srequired

package setup

import (
	"os"
	"testing"
)

func Setup(m *testing.M, config Config) {
	var v int

	// Add any setup tasks that should execute for all tests here.

	if v == 0 {
		v = m.Run()
	}

	os.Exit(v)
}
