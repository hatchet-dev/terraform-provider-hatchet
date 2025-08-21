.PHONY: build install test clean

# Build the provider
build:
	go build -o ./bin/terraform-provider-hatchetcloud

# Install the provider locally
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchetcloud/0.1.0/darwin_arm64
	cp ./bin/terraform-provider-hatchetcloud ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchetcloud/0.1.0/darwin_arm64/

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
