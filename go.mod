module github.com/webner/leo-operator

require (
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cpu/goacmedns v0.0.1
	github.com/go-acme/lego v2.7.1+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/miekg/dns v1.1.15
	github.com/operator-framework/operator-sdk v0.9.1-0.20190724001845-d6e1aba9fa51
	github.com/spf13/pflag v1.0.3
	k8s.io/api v0.0.0-20190612125737-db0771252981
	k8s.io/apimachinery v0.0.0-20190612125636-6a5db36e93ad
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.10 // indirect
)

// Pinned to kubernetes-1.13.4
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190222131558-5cb15d344471
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221124651-86fb29eff628
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v1.6.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde
)
