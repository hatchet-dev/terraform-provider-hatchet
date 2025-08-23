# hatchetcloud_user (Data Source)

Fetches information about a Hatchet user by email address.

## Example Usage

```terraform
# Fetch user information by email
data "hatchetcloud_user" "example_user" {
  email = "user@example.com"
}

# Use user data when managing organization members
resource "hatchetcloud_organization_member" "members" {
  emails = [data.hatchetcloud_user.example_user.email]
}

# Multiple user lookups
variable "user_emails" {
  description = "List of user emails to look up"
  type        = list(string)
  default = [
    "admin@example.com",
    "dev@example.com"
  ]
}

data "hatchetcloud_user" "team_members" {
  for_each = toset(var.user_emails)
  email    = each.value
}

# Add all team members to organization
resource "hatchetcloud_organization_member" "team" {
  emails = [for user in data.hatchetcloud_user.team_members : user.email]
}

# Output user information
output "user_id" {
  description = "The ID of the example user"
  value       = data.hatchetcloud_user.example_user.id
}

output "team_member_ids" {
  description = "Map of email to user ID for all team members"
  value = {
    for k, v in data.hatchetcloud_user.team_members : k => v.id
  }
}
```

## Schema

### Required

- `email` (String) The email address of the user.

### Read-Only

- `id` (String) The ID of the user.

## Notes

- The user must exist in the Hatchet Cloud system and be accessible with the provided management token.
- Use this data source to look up user IDs when managing organization memberships or other user-specific operations.
- The email address is case-insensitive for lookups.