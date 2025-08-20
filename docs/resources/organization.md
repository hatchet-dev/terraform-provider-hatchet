# hatchetcloud_organization (Resource)

Manages a Hatchet organization. This resource is read-only as organization creation and modification must be done through the Hatchet Cloud interface.

## Example Usage

```terraform
# Import an existing organization
resource "hatchetcloud_organization" "example" {
  id   = "12345678-1234-1234-1234-123456789012"
  name = "My Organization"
}
```

## Schema

### Required

- `name` (String) The name of the organization.

### Read-Only

- `id` (String) The ID of the organization.

## Import

Import is supported using the following syntax:

```shell
terraform import hatchetcloud_organization.example 12345678-1234-1234-1234-123456789012
```

Where `12345678-1234-1234-1234-123456789012` is the organization ID.

## Notes

- Organizations cannot be created or deleted through Terraform. They must be managed through the Hatchet Cloud interface.
- This resource is primarily used for importing existing organizations into Terraform state for reference in other resources.