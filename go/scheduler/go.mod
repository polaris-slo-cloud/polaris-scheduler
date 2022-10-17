module polaris-slo-cloud.github.io/polaris-scheduler/v1/scheduler

go 1.16

require (
	gonum.org/v1/gonum v0.9.3
	k8s.io/api v0.22.9
	k8s.io/apimachinery v0.22.9
	k8s.io/client-go v0.22.9
	k8s.io/code-generator v0.22.9
	k8s.io/component-base v0.22.9
	k8s.io/klog/v2 v2.9.0
	// k8s.io/kube-openapi should be the same version as the one used by k8s.io/apiserver.
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
	k8s.io/kube-scheduler v0.22.9 // indirect
	k8s.io/kubernetes v1.22.9
	k8s.rainbow-h2020.eu/rainbow/orchestration v0.0.1
	sigs.k8s.io/controller-runtime v0.10.3
)

replace (
	k8s.io/api => k8s.io/api v0.22.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.9
	k8s.io/apiserver => k8s.io/apiserver v0.22.9
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.9
	k8s.io/client-go => k8s.io/client-go v0.22.9
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.9
	k8s.io/code-generator => k8s.io/code-generator v0.22.9
	k8s.io/component-base => k8s.io/component-base v0.22.9
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.9
	k8s.io/controller-manager => k8s.io/controller-manager v0.22.9
	k8s.io/cri-api => k8s.io/cri-api v0.22.9
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.9
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.9
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.9
	k8s.io/kubectl => k8s.io/kubectl v0.22.9
	k8s.io/kubelet => k8s.io/kubelet v0.22.9
	k8s.io/kubernetes => k8s.io/kubernetes v1.22.9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.9
	k8s.io/metrics => k8s.io/metrics v0.22.9
	k8s.io/mount-utils => k8s.io/mount-utils v0.22.9
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.9
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.9
	k8s.rainbow-h2020.eu/rainbow/orchestration => ../orchestration
)
