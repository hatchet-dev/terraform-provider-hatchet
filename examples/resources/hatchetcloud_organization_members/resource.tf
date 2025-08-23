terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # Token is read from HATCHET_TOKEN environment variable
}

# Look up users by email first
data "hatchetcloud_user" "admin" {
  email = "admin@example.com"
}

data "hatchetcloud_user" "developer" {
  email = "developer@example.com"
}

# Add multiple users to the organization using their user IDs
resource "hatchetcloud_organization_members" "team_members" {
  user_ids = [
    data.hatchetcloud_user.admin.id,
    data.hatchetcloud_user.developer.id
  ]
}

# Alternative: Add users directly using known user IDs
resource "hatchetcloud_organization_members" "direct_members" {
  user_ids = [
    "12345678-1234-1234-1234-123456789012",
    "87654321-4321-4321-4321-210987654321"
  ]
}