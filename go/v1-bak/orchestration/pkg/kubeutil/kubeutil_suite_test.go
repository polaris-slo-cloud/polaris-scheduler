package kubeutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubeutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeutil Suite")
}
