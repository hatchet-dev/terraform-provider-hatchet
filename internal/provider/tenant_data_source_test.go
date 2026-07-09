// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTenantDataSource_basic(t *testing.T) {
	tenantName := testAccUniqueName("ds-tenant")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfig(tenantName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.hatchet_tenant.test", "id",
						"hatchet_tenant.source", "id",
					),
					resource.TestCheckResourceAttr("data.hatchet_tenant.test", "status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccTenantDataSourceConfig(tenantName string) string {
	return fmt.Sprintf(`%s
resource "hatchet_tenant" "source" {
  name = %q
}

data "hatchet_tenant" "test" {
  id = hatchet_tenant.source.id
}`, testAccProviderConfig(), tenantName)
}
