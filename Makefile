.PHONY: build install test clean

# Build the provider
build:
	go build -o terraform-provider-hatchetcloud

# Install the provider locally
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchetcloud/1.0.0/darwin_amd64
	cp terraform-provider-hatchetcloud ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchetcloud/1.0.0/darwin_amd64/

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f terraform-provider-hatchetcloud
	go clean

# Format code
fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

# Generate documentation
docs:
	go generate

# Initialize go modules
init:
	go mod tidy
