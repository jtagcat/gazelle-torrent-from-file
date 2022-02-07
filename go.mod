module github.com/jtagcat/gazelle-torrent-from-file/v0

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/charles-haynes/whatapi v0.3.0
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a // indirect
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v0.23.3
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/utils v0.0.0-20220127004650-9b3446523e65 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
)

replace github.com/charles-haynes/whatapi => /f/git/tmp/whatapi
