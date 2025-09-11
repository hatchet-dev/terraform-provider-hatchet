default: fmt lint install generate

build:
	go build -o terraform-provider-hatchet

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchet/0.2.1/darwin_arm64
	mv terraform-provider-hatchet ~/.terraform.d/plugins/registry.terraform.io/hatchet-dev/hatchet/0.2.1/darwin_arm64/terraform-provider-hatchet
lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofumpt -l -w .
	terraform fmt -recursive .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
