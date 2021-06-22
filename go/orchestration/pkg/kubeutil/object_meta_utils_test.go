package kubeutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("object_meta_utils", func() {

	var metaObj *metav1.ObjectMeta

	BeforeEach(func() {
		metaObj = &metav1.ObjectMeta{
			Name:      "TestName",
			Namespace: "TestNamespace",
			Labels: map[string]string{
				"LabelA": "ValueA",
				"LabelB": "ValueB",
			},
		}
	})

	Describe("GetLabel", func() {

		Context("Existing Labels field", func() {
			It("returns an existing label", func() {
				value, found := kubeutil.GetLabel(metaObj, "LabelA")
				Expect(found).To(Equal(true))
				Expect(value).To(Equal("ValueA"))
			})

			It("reports false on a non-existing label", func() {
				_, found := kubeutil.GetLabel(metaObj, "DoesNotExist")
				Expect(found).To(Equal(false))
			})
		})

		Context("Missing Labels field", func() {
			It("reports false", func() {
				metaObj.Labels = nil
				_, found := kubeutil.GetLabel(metaObj, "LabelA")
				Expect(found).To(Equal(false))
			})
		})

	})

	Describe("GetNamespace", func() {

		It("Returns the set Namespace", func() {
			Expect(kubeutil.GetNamespace(metaObj)).To(Equal("TestNamespace"))
		})

		It("Returns the default namespace if none is set", func() {
			metaObj.Namespace = ""
			Expect(kubeutil.GetNamespace(metaObj)).To(Equal("default"))
		})

	})

})
