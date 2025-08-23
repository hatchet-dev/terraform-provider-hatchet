# hatchetcloud_tenant (Resource)

Manages a Hatchet tenant within an organization.

## Example Usage

```terraform
# Create a new tenant
resource "hatchetcloud_tenant" "production" {
  name = "Production Environment"
  slug = "prod"
}

# Create multiple tenants using for_each
variable "environments" {
  description = "Map of environments to create"
  type = map(object({
    name = string
    slug = string
  }))
  default = {
    "staging" = {
      name = "Staging Environment"
      slug = "staging"
    }
    "development" = {
      name = "Development Environment"
      slug = "dev"
    }
  }
}

resource "hatchetcloud_tenant" "environments" {
  for_each = var.environments

  name = each.value.name
  slug = each.value.slug
}
```

## Schema

### Required

- `name` (String) The name of the tenant.
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

- The organization is automatically determined from the JWT token provided to the provider.
- The `slug` cannot be changed after creation. Changing this value will force a new resource to be created.
- Tenant updates (name changes) are not currently supported through the API.
- When a tenant is deleted, it will be permanently removed from the organization.