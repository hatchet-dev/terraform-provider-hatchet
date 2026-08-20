// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUserGroupMemberResource_basic adds an existing organization member to a
// user group by email and verifies it can be imported back.
//
// Requires TF_ACC_TEST_MEMBER_EMAIL set to the email of an organization member
// that has already accepted their invite (hatchet_organization_member only ever
// creates PENDING invites, so there's no way to mint an accepted member within
// the test itself).
func TestAccUserGroupMemberResource_basic(t *testing.T) {
	memberEmail := os.Getenv("TF_ACC_TEST_MEMBER_EMAIL")
	if memberEmail == "" {
		t.Skip("TF_ACC_TEST_MEMBER_EMAIL must be set to an accepted organization member's email to run user group membership tests")
	}

	groupName := testAccUniqueName("user-group-member")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupMemberResourceConfig(groupName, memberEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_user_group_member.test", "email", memberEmail),
					resource.TestCheckResourceAttrPair("hatchet_user_group_member.test", "user_group_id", "hatchet_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("hatchet_user_group_member.test", "organization_member_id"),
					resource.TestCheckResourceAttrSet("hatchet_user_group_member.test", "id"),
				),
			},
			{
				// hatchet_user_group.member_count is computed from the server-side
				// membership count, which changed as a side effect of the sibling
				// resource applied in the previous step. The group resource's own
				// state isn't re-read until the next plan's refresh, so re-applying
				// the same config here (which triggers that refresh) is what surfaces
				// the updated count — Terraform absorbs the drift into refreshed
				// state silently, so the resulting plan is empty, not a diff.
				Config: testAccUserGroupMemberResourceConfig(groupName, memberEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_user_group.test", "member_count", "1"),
				),
			},
			{
				ResourceName:      "hatchet_user_group_member.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserGroupMemberResourceConfig(groupName, email string) string {
	return fmt.Sprintf(`%s
resource "hatchet_user_group" "test" {
  name = %q
  role = "MEMBER"
}

resource "hatchet_user_group_member" "test" {
  user_group_id = hatchet_user_group.test.id
  email         = %q
}`, testAccProviderConfig(), groupName, email)
}
