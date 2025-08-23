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

# Fetch user information by email
data "hatchetcloud_user" "example_user" {
  email = "admin@example.com"
}

# Output the user ID for reference in other resources
output "user_id" {
  value = data.hatchetcloud_user.example_user.id
}
