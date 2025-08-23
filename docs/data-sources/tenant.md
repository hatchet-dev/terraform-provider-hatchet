# hatchetcloud_tenant (Data Source)

Fetches information about a Hatchet tenant.

## Example Usage

```terraform
# Fetch tenant information by ID
data "hatchetcloud_tenant" "production" {
  id = "87654321-4321-4321-4321-210987654321"
}

# Use tenant data to create API tokens
resource "hatchetcloud_tenant_api_token" "production_api_token" {
  tenant_id = data.hatchetcloud_tenant.production.id
  name      = "API Token for Production Tenant"
}

# Conditional resource creation based on tenant status
resource "hatchetcloud_tenant_api_token" "active_tenant_token" {
  count = data.hatchetcloud_tenant.production.status == "active" ? 1 : 0

  tenant_id = data.hatchetcloud_tenant.production.id
  name      = "Active Tenant Token"
}

# Multiple tenant lookups
variable "tenant_ids" {
  description = "List of tenant IDs to look up"
  type        = list(string)
  default = [
    "87654321-4321-4321-4321-210987654321",
    "11111111-2222-3333-4444-555555555555"
  ]
}

data "hatchetcloud_tenant" "environments" {
  for_each = toset(var.tenant_ids)
  id       = each.value
}

# Output information
output "production_tenant_status" {
  description = "The status of the production tenant"
  value       = data.hatchetcloud_tenant.production.status
}

output "active_environments" {
  description = "List of active tenant IDs"
  value = [
    for k, v in data.hatchetcloud_tenant.environments : k
    if v.status == "active"
  ]
}
```

## Schema

### Required

- `id` (String) The ID of the tenant.

### Read-Only

- `archived_at` (String) The timestamp when the tenant was archived.
- `status` (String) The status of the tenant (active, archived).

## Notes

- The organization is automatically determined from the JWT token provided to the provider.
- The tenant must exist within the organization and be accessible with the provided management token.
- Use this data source when you need to reference existing tenants for creating API tokens or other tenant-specific resources.