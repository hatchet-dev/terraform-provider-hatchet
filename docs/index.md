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
  # endpoint is optional and defaults to "cloud.onhatchet.run"
  endpoint = "cloud.onhatchet.run"
  
  # Management token for accessing the Hatchet Cloud API
  token = var.hatchet_management_token
}
```

## Authentication

The provider requires a management token to authenticate with the Hatchet Cloud API. This token can be provided in several ways:

1. **Provider configuration** (not recommended for production):
   ```terraform
   provider "hatchetcloud" {
     token = "your-management-token"
   }
   ```

2. **Environment variable** (recommended):
   ```bash
   export HATCHET_TOKEN="your-management-token"
   ```
   ```terraform
   provider "hatchetcloud" {
     # token will be read from HATCHET_TOKEN environment variable
   }
   ```

3. **Terraform variables**:
   ```terraform
   variable "hatchet_management_token" {
     description = "Hatchet Cloud management token"
     type        = string
     sensitive   = true
   }
   
   provider "hatchetcloud" {
     token = var.hatchet_management_token
   }
   ```

## Schema

### Optional

- `endpoint` (String) The Hatchet Cloud API endpoint. Defaults to `cloud.onhatchet.run`.
- `token` (String, Sensitive) The management token for authenticating with the Hatchet Cloud API. Can also be provided via the `HATCHET_TOKEN` environment variable.

## Resources

- [`hatchetcloud_organization`](resources/organization) - Manages a Hatchet organization (read-only)
- [`hatchetcloud_tenant`](resources/tenant) - Manages a Hatchet tenant
- [`hatchetcloud_tenant_api_token`](resources/tenant_api_token) - Manages a tenant API token
- [`hatchetcloud_organization_member`](resources/organization_member) - Manages organization membership

## Data Sources

- [`hatchetcloud_organization`](data-sources/organization) - Fetches organization information
- [`hatchetcloud_tenant`](data-sources/tenant) - Fetches tenant information
- [`hatchetcloud_organization_members`](data-sources/organization_members) - Fetches organization members