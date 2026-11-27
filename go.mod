module github.com/spin-stack/runbooks

go 1.27.3

require (
	github.com/yuin/goldmark v1.8.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/docker/cli v29.7.2+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.3 // indirect
	github.com/google/go-containerregistry v0.22.1 // indirect
	github.com/klauspost/compress v1.20.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/rogpeppe/go-internal v1.15.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spin-stack/go-tools v0.20261127.15 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

tool (
	github.com/spin-stack/go-tools/cmd/sfw
	github.com/spin-stack/go-tools/cmd/task
	github.com/spin-stack/go-tools/cmd/taskgate
	github.com/spin-stack/go-tools/cmd/testidentity
)
