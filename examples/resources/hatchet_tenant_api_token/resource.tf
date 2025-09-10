terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchet" {
  # Token is read from HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable
}

data "hatchet_tenant" "example" {
  id = "707d0855-80ab-4e1f-a156-f1c4546cbf52"
}

# Create a temporary API token with expiration
resource "hatchet_tenant_api_token" "temp_token" {
  tenant_id  = data.hatchet_tenant.example.id
  name       = "Temporary Access Token"
  expires_at = "720h" // 30 days
}

# Output the token value so it can be used
output "api_token" {
  value     = hatchet_tenant_api_token.temp_token.token
  sensitive = true
}
