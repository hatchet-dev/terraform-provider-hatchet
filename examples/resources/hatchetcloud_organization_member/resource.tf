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

# Add multiple users to the organization using email addresses
resource "hatchetcloud_organization_member" "team_members" {
  emails = [
    "admin@example.com"
  ]
}