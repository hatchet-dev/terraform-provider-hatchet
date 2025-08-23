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

# Fetch tenant information by ID
data "hatchetcloud_tenant" "production" {
  id = "707d0855-80ab-4e1f-a156-f1c4546cbf52"
}

# Output the tenant slug for reference
output "tenant_status" {
  value = data.hatchetcloud_tenant.production.status
}
