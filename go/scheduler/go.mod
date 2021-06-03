module rainbow-h2020.eu/gomod/rainbow-scheduler

go 1.15

require (
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	gonum.org/v1/gonum v0.8.2
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/code-generator v0.20.7
	k8s.io/component-base v0.20.7
	k8s.io/klog/v2 v2.4.0
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kube-scheduler v0.20.7 // indirect
	k8s.io/kubernetes v1.20.7
	k8s.rainbow-h2020.eu/rainbow/orchestration v0.0.1
)

replace (
	k8s.io/api => k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.7
	k8s.io/apiserver => k8s.io/apiserver v0.20.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.7
	k8s.io/client-go => k8s.io/client-go v0.20.7
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.7
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.7
	k8s.io/code-generator => k8s.io/code-generator v0.20.7
	k8s.io/component-base => k8s.io/component-base v0.20.7
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.7
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.7
	k8s.io/cri-api => k8s.io/cri-api v0.20.7
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.7
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.7
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.7
	k8s.io/kubectl => k8s.io/kubectl v0.20.7
	k8s.io/kubelet => k8s.io/kubelet v0.20.7
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.7
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.7
	k8s.io/metrics => k8s.io/metrics v0.20.7
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.7
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.7
	k8s.rainbow-h2020.eu/rainbow/orchestration => ../orchestration
)
