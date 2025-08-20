# hatchetcloud_tenant_api_token (Resource)

Manages a Hatchet tenant API token.

## Example Usage

```terraform
# Create an API token for a tenant
resource "hatchetcloud_tenant_api_token" "example" {
  tenant_id = hatchetcloud_tenant.example.id
  name      = "Production API Token"
}

# Create an API token with expiration
resource "hatchetcloud_tenant_api_token" "temp_token" {
  tenant_id  = hatchetcloud_tenant.example.id
  name       = "Temporary Token"
  expires_at = "24h"  # Expires in 24 hours
}

# Output the token value (sensitive)
output "api_token" {
  value     = hatchetcloud_tenant_api_token.example.token
  sensitive = true
}
```

## Schema

### Required

- `name` (String) The name of the API token.
- `tenant_id` (String) The ID of the tenant this API token belongs to.

### Optional

- `expires_at` (String) The expiration duration of the API token (e.g., "24h", "7d", "30d"). If not specified, the token will not expire.

### Read-Only

- `id` (String) The ID of the API token.
- `token` (String, Sensitive) The API token value. This is only available immediately after creation.

## Import

Import is supported using the following syntax:

```shell
terraform import hatchetcloud_tenant_api_token.example 12345678-1234-1234-1234-123456789012
```

Where `12345678-1234-1234-1234-123456789012` is the API token ID.

## Notes

- The `tenant_id`, `name`, and `expires_at` cannot be changed after creation. Changing these values will force a new resource to be created.
- The `token` value is only returned during the initial creation and is not retrievable afterwards for security reasons.
- Store the token value securely immediately after creation, as it cannot be retrieved again.
- When a token is deleted, it will be immediately revoked and can no longer be used for API access.