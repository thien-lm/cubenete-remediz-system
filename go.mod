module github.com/planetlabs/draino

go 1.22

toolchain go1.22.2

// Kube 1.15.3
replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77

// Kube 1.15.3
replace k8s.io/api => k8s.io/api v0.0.0-20190819141258-3544db3b9e44

// Kube 1.15.3
replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/antonmedv/expr v1.8.8
	github.com/go-logr/logr v1.4.1
	github.com/go-logr/zapr v1.3.0
	github.com/go-test/deep v1.0.1
	github.com/julienschmidt/httprouter v1.1.0
	github.com/oklog/run v1.0.0
	github.com/pkg/errors v0.8.0
	github.com/stretchr/testify v1.9.0
	github.com/vmware/go-vcloud-director/v2 v2.25.0
	go.opencensus.io v0.21.0
	go.uber.org/zap v1.26.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.1
	k8s.io/klog/v2 v2.130.1
)

require (
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc // indirect
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf // indirect
	github.com/araddon/dateparse v0.0.0-20190622164848-0fb0a474d195 // indirect
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550 // indirect
	github.com/gogo/protobuf v0.0.0-20171007142547-342cbe0a0415 // indirect
	github.com/golang/groupcache v0.0.0-20160516000752-02826c3e7903 // indirect
	github.com/golang/protobuf v1.2.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/json-iterator/go v0.0.0-20180701071628-ab8a2e0c74be // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/peterhellberg/link v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910 // indirect
	github.com/prometheus/common v0.0.0-20181126121408-4724e9255275 // indirect
	github.com/prometheus/procfs v0.0.0-20181204211112-1dc9a6cbc91a // indirect
	github.com/spf13/pflag v1.0.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/term v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.0.0-20161028155119-f51c12702a4d // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/inf.v0 v0.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30 // indirect
	k8s.io/utils v0.0.0-20190221042446-c2654d5206da // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
