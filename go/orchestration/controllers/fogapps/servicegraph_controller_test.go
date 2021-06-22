package fogapps_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
)

var _ = Describe("ServiceGraphController", func() {

	var createServiceGraph = func(replicaSetType fogappsCRDs.ReplicaSetType) *fogappsCRDs.ServiceGraph {
		return nil
	}

	Context("SimpleReplicaSet", func() {

		It("Creates the deployment resources", func() {
			graph := createServiceGraph(fogappsCRDs.SimpleReplicaSet)
			Expect(graph).To(BeNil())
		})
	})

	Context("StatefulReplicaSet", func() {

	})

})
