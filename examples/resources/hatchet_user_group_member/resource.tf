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

resource "hatchet_user_group" "platform_admins" {
  name = "platform-admins"
  role = "ADMIN"
  tags = ["platform"]
}

# email must belong to a member who has already accepted their organization
# invite (see hatchet_organization_member); pending invites cannot be added
# to a group.
resource "hatchet_user_group_member" "john" {
  user_group_id = hatchet_user_group.platform_admins.id
  email         = "john@example.com"
}
