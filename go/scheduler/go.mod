module k8s.rainbow-h2020.eu/rainbow/scheduler

go 1.16

require (
	gonum.org/v1/gonum v0.9.3
	k8s.io/api v0.21.9
	k8s.io/apimachinery v0.21.9
	k8s.io/client-go v0.21.9
	k8s.io/code-generator v0.21.9
	k8s.io/component-base v0.21.9
	k8s.io/klog/v2 v2.9.0
	// k8s.io/kube-openapi should be the same version as the one used by k8s.io/apiserver.
	k8s.io/kube-openapi v0.0.0-20211110012726-3cc51fd1e909
	k8s.io/kube-scheduler v0.21.9 // indirect
	k8s.io/kubernetes v1.21.9
	k8s.rainbow-h2020.eu/rainbow/orchestration v0.0.1
	sigs.k8s.io/controller-runtime v0.9.7
)

replace (
	k8s.io/api => k8s.io/api v0.21.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.9
	k8s.io/apiserver => k8s.io/apiserver v0.21.9
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.9
	k8s.io/client-go => k8s.io/client-go v0.21.9
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.9
	k8s.io/code-generator => k8s.io/code-generator v0.21.9
	k8s.io/component-base => k8s.io/component-base v0.21.9
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.9
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.9
	k8s.io/cri-api => k8s.io/cri-api v0.21.9
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.9
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.9
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.9
	k8s.io/kubectl => k8s.io/kubectl v0.21.9
	k8s.io/kubelet => k8s.io/kubelet v0.21.9
	k8s.io/kubernetes => k8s.io/kubernetes v1.21.9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.9
	k8s.io/metrics => k8s.io/metrics v0.21.9
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.9
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.9
	k8s.rainbow-h2020.eu/rainbow/orchestration => ../orchestration
)

// controller-runtime v0.9.x uses a version of github.com/googleapis/gnostic that predates a breaking change in its function `openapi_v2.NewDocument()`.
// K8s 1.21.x already uses the newer version that includes this breaking change.
// controller-runtime v0.10.x would use the newer version of that library, but it targets K8s v1.22.
// Since we need to target K8s v1.21.x we could rely on the replace statements to force K8s v1.21 dependencies, but this caused a compiler error
// with `wait.PollImmediateUntilWithContext()` being undefined at some point.
// Thus, we exclude the problematic github.com/googleapis/gnostic version. So far, there are no problems.
exclude github.com/googleapis/gnostic v0.5.5
