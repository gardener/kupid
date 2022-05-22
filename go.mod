module github.com/gardener/kupid

go 1.16

require (
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/gardener/gardener v1.42.6
	github.com/go-logr/logr v1.2.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.0
	github.com/prometheus/client_golang v1.11.0
	go.uber.org/zap v1.19.1
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/tools v0.1.10 // indirect
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.11.1

)

replace (
	github.com/gardener/gardener => github.com/gardener/gardener v1.42.6
	k8s.io/api => k8s.io/api v0.23.3 // 1.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.3 // 1.16.8
	k8s.io/client-go => k8s.io/client-go v0.23.3 // 1.16.8

)
