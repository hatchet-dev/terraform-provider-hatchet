# Hatchet Cloud Provider

The Hatchet Cloud provider is used to interact with the Hatchet Cloud API for managing organizations, tenants, API tokens, and members.

## Example Usage

```terraform
terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # JWT management token for accessing the Hatchet Cloud API
  # The token automatically contains the endpoint and organization ID
  token = var.hatchet_management_token
}
```

## Authentication

The provider requires a JWT management token to authenticate with the Hatchet Cloud API. The JWT token automatically contains the endpoint and organization ID, so no additional configuration is needed.

The token can be provided in several ways:

1. **Environment variable** (recommended):
   ```bash
   export HATCHET_TOKEN="eyJhbGciOiJFUzI1NiIsImtpZCI6IjRMOVhBQSJ9..."
   ```
   ```terraform
   provider "hatchetcloud" {
     # token will be read from HATCHET_TOKEN environment variable
   }
   ```

2. **Provider configuration** (not recommended for production):
   ```terraform
   provider "hatchetcloud" {
     token = "eyJhbGciOiJFUzI1NiIsImtpZCI6IjRMOVhBQSJ9..."
   }
   ```

3. **Terraform variables**:
   ```terraform
   variable "hatchet_management_token" {
     description = "Hatchet Cloud JWT management token"
     type        = string
     sensitive   = true
   }
   
   provider "hatchetcloud" {
     token = var.hatchet_management_token
   }
   ```

## JWT Token Format

The management token is a JWT that contains:
- `sub`: Organization ID 
- `server_url`: Hatchet Cloud endpoint
- Standard JWT claims (exp, iat, etc.)

The provider automatically extracts these values from the token.

## Schema

### Optional

- `token` (String, Sensitive) The JWT management token for authenticating with the Hatchet Cloud API. Can also be provided via the `HATCHET_TOKEN` environment variable.

## Resources

- [`hatchetcloud_tenant`](resources/tenant) - Manages a Hatchet tenant
- [`hatchetcloud_tenant_api_token`](resources/tenant_api_token) - Manages a tenant API token
- [`hatchetcloud_organization_member`](resources/organization_member) - Manages organization membership

## Data Sources

- [`hatchetcloud_organization`](data-sources/organization) - Fetches organization information
- [`hatchetcloud_tenant`](data-sources/tenant) - Fetches tenant information