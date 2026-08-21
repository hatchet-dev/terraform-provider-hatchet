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

# Members of this group are granted the ADMIN role on any tenant tagged "platform"
resource "hatchet_user_group" "platform_admins" {
  name = "platform-admins"
  role = "ADMIN"
  tags = ["platform"]
}
