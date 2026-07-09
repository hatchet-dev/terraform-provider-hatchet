// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTenantAPITokenResource_basic(t *testing.T) {
	tenantName := testAccUniqueName("api-token-tenant")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantAPITokenResourceConfig(tenantName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hatchet_tenant_api_token.test", "id"),
					resource.TestCheckResourceAttrSet("hatchet_tenant_api_token.test", "token"),
					resource.TestCheckResourceAttr("hatchet_tenant_api_token.test", "name", "tf-acc-token"),
				),
			},
		},
	})
}

func testAccTenantAPITokenResourceConfig(tenantName string) string {
	return fmt.Sprintf(`%s
resource "hatchet_tenant" "parent" {
  name = %q
}

resource "hatchet_tenant_api_token" "test" {
  tenant_id  = hatchet_tenant.parent.id
  name       = "tf-acc-token"
  expires_at = "24h"
}`, testAccProviderConfig(), tenantName)
}
