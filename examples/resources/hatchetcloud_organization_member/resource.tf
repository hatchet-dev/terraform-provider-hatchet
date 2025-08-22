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

# Add multiple users to the organization using email addresses
resource "hatchetcloud_organization_member" "team_members" {
  org_id = data.hatchetcloud_organization.existing.id
  emails = [
    "admin@company.com",
    "developer1@company.com",
    "developer2@company.com",
    "manager@company.com"
  ]
}

# Add members from variables
variable "organization_member_emails" {
  description = "List of email addresses to add as organization members"
  type        = list(string)
  default = [
    "user1@example.com",
    "user2@example.com",
    "user3@example.com"
  ]
}

resource "hatchetcloud_organization_member" "variable_members" {
  org_id = data.hatchetcloud_organization.existing.id
  emails = var.organization_member_emails
}

# Example with conditional emails
variable "include_contractors" {
  description = "Whether to include contractor emails"
  type        = bool
  default     = true
}

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
  org_id = data.hatchetcloud_organization.existing.id
  emails = local.all_emails
}

# Data source to check current members
data "hatchetcloud_organization_members" "current_members" {
  org_id = data.hatchetcloud_organization.existing.id
}

# Outputs
output "current_members_count" {
  description = "Total number of current members in the organization"
  value       = length(data.hatchetcloud_organization_members.current_members.members)
}

output "added_emails_count" {
  description = "Number of email addresses added to the organization"
  value       = length(hatchetcloud_organization_member.team_members.emails)
}