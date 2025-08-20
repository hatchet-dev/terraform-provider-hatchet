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

# Fetch tenant information by ID and organization ID
data "hatchetcloud_tenant" "production" {
  id              = "87654321-4321-4321-4321-210987654321"
  organization_id = data.hatchetcloud_organization.existing.id
}

# Use tenant data to create API tokens
resource "hatchetcloud_tenant_api_token" "production_api_token" {
  tenant_id = data.hatchetcloud_tenant.production.id
  name      = "API Token for Production Tenant"
}

# Conditional resource creation based on tenant status
resource "hatchetcloud_tenant_api_token" "active_tenant_token" {
  count = data.hatchetcloud_tenant.production.status == "active" ? 1 : 0
  
  tenant_id = data.hatchetcloud_tenant.production.id
  name      = "Active Tenant Token"
}

# Example with multiple tenant lookups
variable "tenant_configs" {
  description = "Map of tenant configurations to look up"
  type = map(object({
    tenant_id       = string
    organization_id = string
  }))
  default = {
    "prod" = {
      tenant_id       = "87654321-4321-4321-4321-210987654321"
      organization_id = "12345678-1234-1234-1234-123456789012"
    }
    "staging" = {
      tenant_id       = "11111111-2222-3333-4444-555555555555"
      organization_id = "12345678-1234-1234-1234-123456789012"
    }
  }
}

data "hatchetcloud_tenant" "environments" {
  for_each = var.tenant_configs
  
  id              = each.value.tenant_id
  organization_id = each.value.organization_id
}

# Create API tokens for active tenants only
resource "hatchetcloud_tenant_api_token" "environment_tokens" {
  for_each = {
    for k, v in data.hatchetcloud_tenant.environments : k => v
    if v.status == "active"
  }
  
  tenant_id = each.value.id
  name      = "Token for ${each.key} environment"
}

# Output tenant information
output "production_tenant_status" {
  description = "The status of the production tenant"
  value       = data.hatchetcloud_tenant.production.status
}

output "production_tenant_archived_at" {
  description = "When the production tenant was archived (if applicable)"
  value       = data.hatchetcloud_tenant.production.archived_at
}

output "active_environments" {
  description = "List of active environment names"
  value = [
    for k, v in data.hatchetcloud_tenant.environments : k
    if v.status == "active"
  ]
}

output "environment_statuses" {
  description = "Map of environment statuses"
  value = {
    for k, v in data.hatchetcloud_tenant.environments : k => v.status
  }
}