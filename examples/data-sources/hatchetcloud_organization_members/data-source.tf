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

# Fetch all members of an organization
data "hatchetcloud_organization_members" "example" {
  organization_id = "12345678-1234-1234-1234-123456789012"
}

# Use member data for conditional logic
variable "required_user_id" {
  description = "User ID that must be a member of the organization"
  type        = string
  default     = "87654321-4321-4321-4321-210987654321"
}

locals {
  member_user_ids = [for member in data.hatchetcloud_organization_members.example.members : member.user_id]
  is_user_member  = contains(local.member_user_ids, var.required_user_id)
  owner_members   = [for member in data.hatchetcloud_organization_members.example.members : member if member.member_type == "OWNER"]
}

# Create resources only if required user is a member
resource "hatchetcloud_tenant" "member_restricted_tenant" {
  count = local.is_user_member ? 1 : 0

  organization_id = data.hatchetcloud_organization_members.example.organization_id
  name            = "Member Restricted Tenant"
  slug            = "member-restricted"
}

# Add new members based on current member count
variable "max_members" {
  description = "Maximum number of members allowed"
  type        = number
  default     = 10
}

variable "new_member_ids" {
  description = "List of new member user IDs to add"
  type        = list(string)
  default = [
    "11111111-2222-3333-4444-555555555555",
    "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
  ]
}

locals {
  current_member_count = length(data.hatchetcloud_organization_members.example.members)
  can_add_members      = local.current_member_count < var.max_members
  members_to_add       = local.can_add_members ? var.new_member_ids : []
}

resource "hatchetcloud_organization_member" "conditional_members" {
  for_each = toset(local.members_to_add)

  organization_id = data.hatchetcloud_organization_members.example.organization_id
  user_id         = each.value
}

# Example: Create tenant-specific API tokens for each member
resource "hatchetcloud_tenant" "member_tenants" {
  for_each = {
    for idx, member in data.hatchetcloud_organization_members.example.members :
    "member-${idx}" => member
  }

  organization_id = data.hatchetcloud_organization_members.example.organization_id
  name            = "Tenant for Member ${each.value.user_id}"
  slug            = "member-${substr(each.value.user_id, 0, 8)}"
}

# Create API tokens for member tenants
resource "hatchetcloud_tenant_api_token" "member_tokens" {
  for_each = hatchetcloud_tenant.member_tenants

  tenant_id = each.value.id
  name      = "Token for ${each.key}"
}

# Output member information
output "total_members" {
  description = "Total number of organization members"
  value       = length(data.hatchetcloud_organization_members.example.members)
}

output "member_user_ids" {
  description = "List of all member user IDs"
  value       = local.member_user_ids
}

output "owner_count" {
  description = "Number of owner members"
  value       = length(local.owner_members)
}

output "is_required_user_member" {
  description = "Whether the required user is a member"
  value       = local.is_user_member
}

output "can_add_more_members" {
  description = "Whether more members can be added based on the limit"
  value       = local.can_add_members
}

output "members_by_type" {
  description = "Members grouped by member type"
  value = {
    for member_type in distinct([for member in data.hatchetcloud_organization_members.example.members : member.member_type]) :
    member_type => [
      for member in data.hatchetcloud_organization_members.example.members :
      member.user_id if member.member_type == member_type
    ]
  }
}