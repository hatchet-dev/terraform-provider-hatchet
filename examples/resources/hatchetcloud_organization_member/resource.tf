terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # Token can be provided via HATCHET_TOKEN environment variable
}

# Reference an existing organization
data "hatchetcloud_organization" "existing" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Add a single user to the organization
resource "hatchetcloud_organization_member" "admin_user" {
  organization_id = data.hatchetcloud_organization.existing.id
  user_id         = "87654321-4321-4321-4321-210987654321"
}

# Add multiple members using for_each
variable "organization_members" {
  description = "List of user IDs to add as organization members"
  type        = list(string)
  default = [
    "87654321-4321-4321-4321-210987654321",
    "11111111-2222-3333-4444-555555555555",
    "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
  ]
}

resource "hatchetcloud_organization_member" "team_members" {
  for_each = toset(var.organization_members)
  
  organization_id = data.hatchetcloud_organization.existing.id
  user_id         = each.value
}

# Add members from a data source
data "hatchetcloud_organization_members" "current_members" {
  organization_id = data.hatchetcloud_organization.existing.id
}

# Example: Conditionally add a member only if they're not already in the org
variable "potential_member_id" {
  description = "User ID to potentially add as a member"
  type        = string
  default     = "99999999-8888-7777-6666-555555555555"
}

locals {
  current_member_ids = [for member in data.hatchetcloud_organization_members.current_members.members : member.user_id]
  should_add_member  = !contains(local.current_member_ids, var.potential_member_id)
}

resource "hatchetcloud_organization_member" "conditional_member" {
  count = local.should_add_member ? 1 : 0
  
  organization_id = data.hatchetcloud_organization.existing.id
  user_id         = var.potential_member_id
}

# Output member information
output "admin_member_id" {
  description = "The organization member ID for the admin user"
  value       = hatchetcloud_organization_member.admin_user.id
}

output "team_member_ids" {
  description = "Map of organization member IDs for team members"
  value       = { for k, v in hatchetcloud_organization_member.team_members : k => v.id }
}

output "total_members_count" {
  description = "Total number of members in the organization"
  value       = length(data.hatchetcloud_organization_members.current_members.members)
}