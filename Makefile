.PHONY: build install test clean

VERSION ?= 0.1.0

# Build the provider
build:
	go build -o ./bin/terraform-provider-hatchetcloud

# Install the provider locally using goreleaser
install:
	goreleaser build --single-target --snapshot --clean
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchetcloud/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)
	cp ./dist/terraform-provider-hatchetcloud_$$(go env GOOS)_$$(go env GOARCH)*/terraform-provider-hatchetcloud ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchetcloud/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f terraform-provider-hatchetcloud
	go clean

# Format code
fmt:
	gofumpt -w .
	terraform fmt -recursive ./examples/

# Generate documentation
docs:
	go generate

# Initialize go modules
init:
	go mod tidy
