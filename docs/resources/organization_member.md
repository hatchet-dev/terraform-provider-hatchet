# hatchetcloud_organization_member (Resource)

Manages a Hatchet organization member.

## Example Usage

```terraform
# Add a user to an organization
resource "hatchetcloud_organization_member" "example" {
  organization_id = hatchetcloud_organization.example.id
  user_id         = "87654321-4321-4321-4321-210987654321"
}

# Add multiple members using for_each
variable "member_user_ids" {
  description = "List of user IDs to add as organization members"
  type        = list(string)
  default = [
    "87654321-4321-4321-4321-210987654321",
    "11111111-2222-3333-4444-555555555555"
  ]
}

resource "hatchetcloud_organization_member" "members" {
  for_each = toset(var.member_user_ids)
  
  organization_id = data.hatchetcloud_organization.existing.id
  user_id         = each.value
}
```

## Schema

### Required

- `organization_id` (String) The ID of the organization.
- `user_id` (String) The ID of the user to add as a member.

### Read-Only

- `id` (String) The ID of the organization member.
- `member_type` (String) The type of member (typically "OWNER").

## Import

Import is supported using the following syntax:

```shell
terraform import hatchetcloud_organization_member.example 12345678-1234-1234-1234-123456789012
```

Where `12345678-1234-1234-1234-123456789012` is the organization member ID.

## Notes

- The `organization_id` and `user_id` cannot be changed after creation. Changing these values will force a new resource to be created.
- Member type and permissions cannot be modified through this resource - they are determined by the Hatchet Cloud system.
- When a member is removed, they will lose access to the organization and all its tenants.
- Users must already exist in the Hatchet Cloud system before they can be added as organization members.