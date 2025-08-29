# Terraform Provider for Hatchet Cloud

A Terraform provider for managing Hatchet Cloud resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `go build` command:

```shell
go build -o terraform-provider-hatchet
```

## Installing the Provider

### Local Installation

To install the provider locally for development:

```shell
make install
```

This will build the provider and install it in the correct location for Terraform to find it.

### Using the Provider

Create a Terraform configuration file (e.g., `main.tf`) with the following content:

```hcl
terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 1.0"
    }
  }
}

provider "hatchet" {
  # optionally can set the "token" but for production environments please use the HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable
}
```

You can also use environment variables for configuration:

```bash
export HATCHET_CLOUD_MANAGEMENT_TOKEN="your-api-token-here"
```

## Development

### Running Tests

```shell
make test
```

### Formatting Code

```shell
make fmt
```

### Generating Documentation

```shell
make docs
```

### Cleaning Build Artifacts

```shell
make clean
```

## Provider Configuration

The Hatchet Cloud provider supports the following configuration options:

- `token` (Sensitive): Your Hatchet Cloud API token for authentication. Can also be set via the `HATCHET_CLOUD_MANAGEMENT_TOKEN` environment variable.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License.
