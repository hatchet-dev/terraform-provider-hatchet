// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"os"
	"regexp"
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
				Config: testAccOrganizationMemberResourceConfig(testEmail, "OWNER"),
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

func TestAccOrganizationMemberResource_member(t *testing.T) {
	testEmail := os.Getenv("TF_ACC_TEST_EMAIL")
	if testEmail == "" {
		t.Skip("TF_ACC_TEST_EMAIL must be set to run organization member tests (avoids real invite sends)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationMemberResourceConfig(testEmail, "MEMBER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_organization_member.test", "email", testEmail),
					resource.TestCheckResourceAttr("hatchet_organization_member.test", "role", "MEMBER"),
					resource.TestCheckResourceAttr("hatchet_organization_member.test", "status", "PENDING"),
					resource.TestCheckResourceAttrSet("hatchet_organization_member.test", "invite_id"),
				),
			},
		},
	})
}

func TestAccOrganizationMemberResource_invalidRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganizationMemberResourceConfig("invalid-role@example.com", "ADMIN"),
				ExpectError: regexp.MustCompile(`Role must be one of`),
			},
		},
	})
}

func testAccOrganizationMemberResourceConfig(email, role string) string {
	return fmt.Sprintf(`%s
resource "hatchet_organization_member" "test" {
  email = %q
  role  = %q
}`, testAccProviderConfig(), email, role)
}
