# hatchetcloud_tenant (Resource)

Manages a Hatchet tenant within an organization.

## Example Usage

```terraform
# Create a new tenant
resource "hatchetcloud_tenant" "example" {
  organization_id = hatchetcloud_organization.example.id
  name           = "Production Environment"
  slug           = "prod"
}

# Reference an existing organization
data "hatchetcloud_organization" "existing" {
  id = "12345678-1234-1234-1234-123456789012"
}

resource "hatchetcloud_tenant" "staging" {
  organization_id = data.hatchetcloud_organization.existing.id
  name           = "Staging Environment"
  slug           = "staging"
}
```

## Schema

### Required

- `name` (String) The name of the tenant.
- `organization_id` (String) The ID of the organization this tenant belongs to.
- `slug` (String) The slug of the tenant. This must be unique within the organization.

### Read-Only

- `archived_at` (String) The timestamp when the tenant was archived.
- `id` (String) The ID of the tenant.
- `status` (String) The status of the tenant (active, archived).

## Import

Import is supported using the following syntax:

```shell
terraform import hatchetcloud_tenant.example 12345678-1234-1234-1234-123456789012
```

Where `12345678-1234-1234-1234-123456789012` is the tenant ID.

## Notes

- The `slug` and `organization_id` cannot be changed after creation. Changing these values will force a new resource to be created.
- Tenant updates (name changes) are not currently supported through the API.
- When a tenant is deleted, it will be permanently removed from the organization.