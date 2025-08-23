# hatchetcloud_organization_members (Resource)

Manages a Hatchet organization's members by adding multiple users via their user IDs.

## Example Usage

```terraform
# Add multiple users to an organization using user IDs
resource "hatchetcloud_organization_members" "team_members" {
  user_ids = [
    "12345678-1234-1234-1234-123456789012",
    "87654321-4321-4321-4321-210987654321", 
    "11111111-2222-3333-4444-555555555555"
  ]
}

# Add members using user data source lookups
data "hatchetcloud_user" "admin" {
  email = "admin@company.com"
}

data "hatchetcloud_user" "developer1" {
  email = "developer1@company.com"
}

data "hatchetcloud_user" "developer2" {
  email = "developer2@company.com"
}

resource "hatchetcloud_organization_members" "team_from_emails" {
  user_ids = [
    data.hatchetcloud_user.admin.id,
    data.hatchetcloud_user.developer1.id,
    data.hatchetcloud_user.developer2.id
  ]
}

# Add members using variables
variable "organization_member_ids" {
  description = "List of user IDs to add as organization members"
  type        = list(string)
  default = [
    "12345678-1234-1234-1234-123456789012",
    "87654321-4321-4321-4321-210987654321"
  ]
}

resource "hatchetcloud_organization_members" "variable_members" {
  user_ids = var.organization_member_ids
}

# Lookup multiple users and add them all
variable "user_emails" {
  description = "List of user emails to look up and add"
  type        = list(string)
  default = [
    "employee1@company.com",
    "employee2@company.com",
    "contractor1@external.com"
  ]
}

data "hatchetcloud_user" "team_members" {
  for_each = toset(var.user_emails)
  email    = each.value
}

resource "hatchetcloud_organization_members" "all_team_members" {
  user_ids = [for user in data.hatchetcloud_user.team_members : user.id]
}

# Conditional user addition
locals {
  base_user_ids = [
    "12345678-1234-1234-1234-123456789012",
    "87654321-4321-4321-4321-210987654321"
  ]
  
  contractor_user_ids = [
    "11111111-2222-3333-4444-555555555555",
    "99999999-8888-7777-6666-555555555555"
  ]
  
  all_user_ids = var.include_contractors ? concat(local.base_user_ids, local.contractor_user_ids) : local.base_user_ids
}

resource "hatchetcloud_organization_members" "conditional_members" {
  user_ids = local.all_user_ids
}
```

## Schema

### Required

- `user_ids` (List of String) List of user IDs to add as members to the organization.

## Import

Import is supported using the following syntax:

```shell
terraform import hatchetcloud_organization_members.example organization
```

The organization ID is automatically determined from the JWT token.

## Notes

- The organization is automatically determined from the JWT token provided to the provider.
- The resource manages all the specified user IDs as a group.
- Users must already exist in the Hatchet Cloud system before they can be added to an organization.
- Use the `hatchetcloud_user` data source to look up user IDs by email address.
- When the resource is destroyed, the specified users will be removed from the organization.
- User IDs must be valid UUIDs.