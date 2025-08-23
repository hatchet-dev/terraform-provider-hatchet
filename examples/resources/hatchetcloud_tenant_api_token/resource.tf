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

data "hatchetcloud_tenant" "example" {
  id = "707d0855-80ab-4e1f-a156-f1c4546cbf52"
}

# Create a temporary API token with expiration
resource "hatchetcloud_tenant_api_token" "temp_token" {
  tenant_id  = data.hatchetcloud_tenant.example.id
  name       = "Temporary Access Token"
  expires_at = "3m"
}

# Output the token value so it can be used
output "api_token" {
  value     = hatchetcloud_tenant_api_token.temp_token.token
  sensitive = true
}
