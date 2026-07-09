// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationMemberResource_basic(t *testing.T) {
	testEmail := os.Getenv("TF_ACC_TEST_EMAIL")
	if testEmail == "" {
		t.Skip("TF_ACC_TEST_EMAIL must be set to run organization member tests (avoids real invite sends)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationMemberResourceConfig(testEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_organization_member.test", "email", testEmail),
					resource.TestCheckResourceAttr("hatchet_organization_member.test", "role", "OWNER"),
					resource.TestCheckResourceAttr("hatchet_organization_member.test", "status", "PENDING"),
					resource.TestCheckResourceAttrSet("hatchet_organization_member.test", "invite_id"),
				),
			},
		},
	})
}

func testAccOrganizationMemberResourceConfig(email string) string {
	return fmt.Sprintf(`%s
resource "hatchet_organization_member" "test" {
  email = %q
  role  = "OWNER"
}`, testAccProviderConfig(), email)
}
