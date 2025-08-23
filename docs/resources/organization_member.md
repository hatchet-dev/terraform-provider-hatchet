# hatchetcloud_organization_member (Resource)

Manages a Hatchet organization member by adding multiple users via their email addresses.

## Example Usage

```terraform
# Add multiple users to an organization using email addresses
resource "hatchetcloud_organization_member" "team_members" {
  emails = [
    "admin@company.com",
    "developer1@company.com", 
    "developer2@company.com"
  ]
}

# Add members using variables
variable "organization_member_emails" {
  description = "List of email addresses to add as organization members"
  type        = list(string)
  default = [
    "user1@example.com",
    "user2@example.com"
  ]
}

resource "hatchetcloud_organization_member" "variable_members" {
  emails = var.organization_member_emails
}

# Conditional email addition
locals {
  base_emails = [
    "employee1@company.com",
    "employee2@company.com"
  ]
  
  contractor_emails = [
    "contractor1@external.com",
    "contractor2@external.com"
  ]
  
  all_emails = var.include_contractors ? concat(local.base_emails, local.contractor_emails) : local.base_emails
}

resource "hatchetcloud_organization_member" "conditional_members" {
  emails = local.all_emails
}
```

## Schema

### Required

- `emails` (List of String) List of email addresses of users to add as members.

## Import

Import is supported using the following syntax:

```shell
terraform import hatchetcloud_organization_member.example organization
```

The organization ID is automatically determined from the JWT token.

## Notes

- The organization is automatically determined from the JWT token provided to the provider.
- The resource manages all the specified email addresses as a group.
- Users will be invited via email if they don't already exist in the Hatchet Cloud system.
- When the resource is destroyed, the specified users will be removed from the organization.
- Email addresses must be valid email format.