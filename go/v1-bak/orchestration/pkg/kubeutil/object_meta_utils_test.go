package kubeutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("object_meta_utils", func() {

	var resourceObj *core.Pod

	BeforeEach(func() {
		resourceObj = &core.Pod{
			ObjectMeta: meta.ObjectMeta{
				Name:      "TestName",
				Namespace: "TestNamespace",
				Labels: map[string]string{
					"LabelA": "ValueA",
					"LabelB": "ValueB",
				},
				Annotations: map[string]string{
					"AnnotationA": "ValueA",
					"AnnotationB": "ValueB",
				},
			},
		}
	})

	Describe("GetLabel", func() {

		Context("Existing Labels field", func() {
			It("returns an existing label", func() {
				value, found := kubeutil.GetLabel(resourceObj, "LabelA")
				Expect(found).To(Equal(true))
				Expect(value).To(Equal("ValueA"))
			})

			It("reports false on a non-existing label", func() {
				_, found := kubeutil.GetLabel(resourceObj, "DoesNotExist")
				Expect(found).To(Equal(false))
			})
		})

		Context("Missing Labels field", func() {
			It("reports false", func() {
				resourceObj.ObjectMeta.Labels = nil
				_, found := kubeutil.GetLabel(resourceObj, "LabelA")
				Expect(found).To(Equal(false))
			})
		})

	})

	Describe("GetNamespace", func() {

		It("Returns the set Namespace", func() {
			Expect(kubeutil.GetNamespace(resourceObj)).To(Equal("TestNamespace"))
		})

		It("Returns the default namespace if none is set", func() {
			resourceObj.ObjectMeta.Namespace = ""
			Expect(kubeutil.GetNamespace(resourceObj)).To(Equal("default"))
		})

	})

	Describe("GetAnnotation", func() {

		Context("Existing Annotations field", func() {
			It("returns an existing annotation", func() {
				value, found := kubeutil.GetAnnotation(resourceObj, "AnnotationA")
				Expect(found).To(Equal(true))
				Expect(value).To(Equal("ValueA"))
			})

			It("reports false on a non-existing annotation", func() {
				_, found := kubeutil.GetAnnotation(resourceObj, "DoesNotExist")
				Expect(found).To(Equal(false))
			})
		})

		Context("Missing Annotations field", func() {
			It("reports false", func() {
				resourceObj.ObjectMeta.Annotations = nil
				_, found := kubeutil.GetAnnotation(resourceObj, "LabelA")
				Expect(found).To(Equal(false))
			})
		})

	})

	Describe("SetAnnotation", func() {

		Context("Existing Annotations field", func() {
			It("updates an existing annotation", func() {
				kubeutil.SetAnnotation(resourceObj, "AnnotationA", "newValue")
				Expect(resourceObj.ObjectMeta.Annotations["AnnotationA"]).To(Equal("newValue"))
			})

			It("creates a new annotation", func() {
				kubeutil.SetAnnotation(resourceObj, "NewAnnotation", "newValue")
				Expect(resourceObj.ObjectMeta.Annotations["NewAnnotation"]).To(Equal("newValue"))
			})
		})

		Context("Missing Annotations field", func() {
			It("creates an Annotations field", func() {
				resourceObj.ObjectMeta.Annotations = nil
				kubeutil.SetAnnotation(resourceObj, "NewAnnotation", "newValue")
				Expect(resourceObj.ObjectMeta.Annotations).NotTo(Equal(nil))
				Expect(resourceObj.ObjectMeta.Annotations["NewAnnotation"]).To(Equal("newValue"))
			})
		})

	})

})
