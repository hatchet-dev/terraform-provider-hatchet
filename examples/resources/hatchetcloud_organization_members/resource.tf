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
data "hatchetcloud_user" "user1" {
  email = "admin@example.com"
}

data "hatchetcloud_user" "user2" {
  email = "testadmin@example.com"
}

# Add multiple users to the organization using their user IDs
resource "hatchetcloud_organization_members" "team_members" {
  user_ids = [
    data.hatchetcloud_user.user1.id,
    data.hatchetcloud_user.user2.id
  ]
}
