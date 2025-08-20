# hatchetcloud_organization_members (Data Source)

Fetches all members of a Hatchet organization.

## Example Usage

```terraform
# Fetch all members of an organization
data "hatchetcloud_organization_members" "example" {
  organization_id = "12345678-1234-1234-1234-123456789012"
}

# Output member information
output "member_count" {
  value = length(data.hatchetcloud_organization_members.example.members)
}

output "member_user_ids" {
  value = [for member in data.hatchetcloud_organization_members.example.members : member.user_id]
}

# Use member data in conditionals
locals {
  is_user_member = contains(
    [for member in data.hatchetcloud_organization_members.example.members : member.user_id],
    var.target_user_id
  )
}

# Create a tenant only if the user is a member
resource "hatchetcloud_tenant" "conditional_tenant" {
  count = local.is_user_member ? 1 : 0
  
  organization_id = data.hatchetcloud_organization_members.example.organization_id
  name           = "User-specific Tenant"
  slug           = "user-tenant"
}
```

## Schema

### Required

- `organization_id` (String) The ID of the organization.

### Read-Only

- `members` (List of Object) List of organization members. Each member has the following attributes:
  - `id` (String) The ID of the organization member.
  - `member_type` (String) The type of member (e.g., "OWNER").
  - `user_id` (String) The ID of the user.

## Notes

- This data source retrieves all members of the specified organization.
- The organization must exist and be accessible with the provided management token.
- Use this data source when you need to check membership, count members, or make decisions based on the current member list.
- The member list includes all active members of the organization.