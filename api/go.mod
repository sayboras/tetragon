module github.com/cilium/tetragon/api

// renovate: datasource=golang-version depName=go
go 1.23.0

toolchain go1.23.1

require (
	github.com/cilium/tetragon v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.34.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240924160255-9d4c2d233b61 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cilium/tetragon => ../
