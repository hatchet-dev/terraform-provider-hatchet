# hatchetcloud_tenant (Data Source)

Fetches information about a Hatchet tenant.

## Example Usage

```terraform
# Fetch tenant information by ID and organization ID
data "hatchetcloud_tenant" "example" {
  id              = "87654321-4321-4321-4321-210987654321"
  organization_id = "12345678-1234-1234-1234-123456789012"
}

# Use tenant data to create API tokens
resource "hatchetcloud_tenant_api_token" "api_token" {
  tenant_id = data.hatchetcloud_tenant.example.id
  name      = "API Token for ${data.hatchetcloud_tenant.example.id}"
}

# Output tenant information
output "tenant_status" {
  value = data.hatchetcloud_tenant.example.status
}

output "tenant_archived_at" {
  value = data.hatchetcloud_tenant.example.archived_at
}
```

## Schema

### Required

- `id` (String) The ID of the tenant.
- `organization_id` (String) The ID of the organization this tenant belongs to.

### Read-Only

- `archived_at` (String) The timestamp when the tenant was archived.
- `status` (String) The status of the tenant (active, archived).

## Notes

- This data source requires both the tenant ID and organization ID to locate the tenant.
- The tenant must exist within the specified organization and be accessible with the provided management token.
- Use this data source when you need to reference existing tenants for creating API tokens or other tenant-specific resources.