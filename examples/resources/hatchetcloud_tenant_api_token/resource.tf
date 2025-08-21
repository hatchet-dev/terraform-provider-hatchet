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
}

# Reference existing resources
data "hatchetcloud_organization" "existing" {
  id = "12345678-1234-1234-1234-123456789012"
}

resource "hatchetcloud_tenant" "example" {
  organization_id = data.hatchetcloud_organization.existing.id
  name            = "Example Tenant"
  slug            = "example"
}

# Create an API token for the tenant
resource "hatchetcloud_tenant_api_token" "production_token" {
  tenant_id = hatchetcloud_tenant.example.id
  name      = "Production API Token"
}

# Create a temporary API token with expiration
resource "hatchetcloud_tenant_api_token" "temp_token" {
  tenant_id  = hatchetcloud_tenant.example.id
  name       = "Temporary Access Token"
  expires_at = "24h" # Expires in 24 hours
}

# Create multiple API tokens for different purposes
variable "api_tokens" {
  description = "Map of API tokens to create"
  type = map(object({
    name       = string
    expires_at = optional(string)
  }))
  default = {
    "ci_cd" = {
      name       = "CI/CD Pipeline Token"
      expires_at = "30d"
    }
    "monitoring" = {
      name = "Monitoring Service Token"
      # No expiration for monitoring token
    }
  }
}

resource "hatchetcloud_tenant_api_token" "service_tokens" {
  for_each = var.api_tokens

  tenant_id  = hatchetcloud_tenant.example.id
  name       = each.value.name
  expires_at = each.value.expires_at
}

# Output the token values (marked as sensitive)
output "production_api_token" {
  description = "The production API token"
  value       = hatchetcloud_tenant_api_token.production_token.token
  sensitive   = true
}

output "temp_token_id" {
  description = "The ID of the temporary token"
  value       = hatchetcloud_tenant_api_token.temp_token.id
}

# Example of storing tokens securely in local files (for development only)
resource "local_sensitive_file" "api_tokens" {
  for_each = hatchetcloud_tenant_api_token.service_tokens

  content  = each.value.token
  filename = "${path.module}/.tokens/${each.key}_token.txt"
}

# Note: In production, consider using a secret management solution instead
# of local files for storing API tokens