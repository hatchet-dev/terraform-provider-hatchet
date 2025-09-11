terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 0.2.1"
    }
  }
}

provider "hatchet" {
  # Token is read from HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable
}

# Invite a user to join the organization as an owner
resource "hatchet_organization_member" "john" {
  email = "john@example.com"
  role  = "OWNER"
}

# Invite another user to the organization
resource "hatchet_organization_member" "jane" {
  email = "jane@example.com"
  role  = "OWNER"
}
