# hatchetcloud_organization (Data Source)

Fetches information about a Hatchet organization.

## Example Usage

```terraform
# Fetch organization information by ID
data "hatchetcloud_organization" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Use the organization data in other resources
resource "hatchetcloud_tenant" "new_tenant" {
  organization_id = data.hatchetcloud_organization.example.id
  name           = "New Tenant"
  slug           = "new-tenant"
}

# Output organization information
output "organization_name" {
  value = data.hatchetcloud_organization.example.name
}
```

## Schema

### Required

- `id` (String) The ID of the organization.

### Read-Only

- `name` (String) The name of the organization.

## Notes

- This data source is useful for referencing existing organizations in your Terraform configuration.
- The organization must exist and be accessible with the provided management token.
- Use this data source when you need to create tenants or manage members within an existing organization.