// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroupResource_basic(t *testing.T) {
	name := testAccUniqueName("user-group-basic")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig(name, "MEMBER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_user_group.test", "name", name),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "role", "MEMBER"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "member_count", "0"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.#", "0"),
					resource.TestCheckResourceAttrSet("hatchet_user_group.test", "id"),
				),
			},
			{
				ResourceName:      "hatchet_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccUserGroupResource_update verifies that name, role, and tags can all
// be changed in place without replacing the resource.
func TestAccUserGroupResource_update(t *testing.T) {
	name := testAccUniqueName("user-group-update")
	updatedName := name + "-renamed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfigWithTags(name, "MEMBER", []string{"team-a"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_user_group.test", "name", name),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "role", "MEMBER"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.0", "team-a"),
				),
			},
			{
				Config: testAccUserGroupResourceConfigWithTags(updatedName, "ADMIN", []string{"team-a", "team-b"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_user_group.test", "name", updatedName),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "role", "ADMIN"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.0", "team-a"),
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.1", "team-b"),
				),
			},
			{
				Config: testAccUserGroupResourceConfigWithTags(updatedName, "ADMIN", []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_user_group.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccUserGroupResourceConfig(name, role string) string {
	return fmt.Sprintf(`%s
resource "hatchet_user_group" "test" {
  name = %q
  role = %q
}`, testAccProviderConfig(), name, role)
}

func testAccUserGroupResourceConfigWithTags(name, role string, tags []string) string {
	tagsHCL := "[]"
	if len(tags) > 0 {
		inner := ""
		for _, tag := range tags {
			inner += fmt.Sprintf("%q, ", tag)
		}
		tagsHCL = fmt.Sprintf("[%s]", inner[:len(inner)-2])
	}
	return fmt.Sprintf(`%s
resource "hatchet_user_group" "test" {
  name = %q
  role = %q
  tags = %s
}`, testAccProviderConfig(), name, role, tagsHCL)
}
