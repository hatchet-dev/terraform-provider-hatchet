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
  # endpoint = "cloud.onhatchet.run"  # Optional, defaults to cloud.onhatchet.run
}

# Import an existing organization
resource "hatchetcloud_organization" "example" {
  name = "My Organization"
}