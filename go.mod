module github.com/servicemeshinterface/sparrow

go 1.13

require (
	github.com/servicemeshinterface/smi-sdk-go v0.3.0
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/heapster v1.5.4 // indirect
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/servicemeshinterface/smi-sdk-go => /go/src/github.com/servicemeshinterface/smi-sdk-go
