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

# Fetch organization information by ID
data "hatchetcloud_organization" "example" {
  id = "17015348-d740-45f2-b23d-ea284c6eb3ee"
}

# Output the organization slug
output "organization_slug" {
  value = data.hatchetcloud_organization.example.slug
}