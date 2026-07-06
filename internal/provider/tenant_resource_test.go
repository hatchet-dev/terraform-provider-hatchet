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

func TestAccTenantResource_basic(t *testing.T) {
	name := testAccUniqueName("tenant-basic")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_tenant.test", "name", name),
					resource.TestCheckResourceAttrSet("hatchet_tenant.test", "id"),
					resource.TestCheckResourceAttrSet("hatchet_tenant.test", "slug"),
					resource.TestCheckResourceAttr("hatchet_tenant.test", "status", "ACTIVE"),
				),
			},
			{
				ResourceName:            "hatchet_tenant.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"archived_at"},
			},
		},
	})
}

func TestAccTenantResource_withSlug(t *testing.T) {
	name := testAccUniqueName("tenant-slug")
	slug := "tf-acc-custom-slug-001"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantResourceConfigWithSlug(name, slug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_tenant.test", "slug", slug),
				),
			},
		},
	})
}

// TestAccTenantResource_withTags verifies tags round-trip through Create and Read,
// and that in-place tag updates work via the SetTags endpoint.
func TestAccTenantResource_withTags(t *testing.T) {
	name := testAccUniqueName("tenant-tags")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantResourceConfigWithTags(name, []string{"team-a", "team-b"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_tenant.test", "name", name),
					resource.TestCheckResourceAttr("hatchet_tenant.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("hatchet_tenant.test", "tags.0", "team-a"),
					resource.TestCheckResourceAttr("hatchet_tenant.test", "tags.1", "team-b"),
				),
			},
			// In-place tag update: remove one tag, no replacement expected.
			{
				Config: testAccTenantResourceConfigWithTags(name, []string{"team-a"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_tenant.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("hatchet_tenant.test", "tags.0", "team-a"),
				),
			},
			// Clear tags entirely.
			{
				Config: testAccTenantResourceConfigWithTags(name, []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_tenant.test", "tags.#", "0"),
				),
			},
		},
	})
}

// TestAccTenantTagScoping_basic verifies that a scoped management token can create
// and manage a tenant whose tags are a subset of the token's own tags.
// Requires HATCHET_SCOPED_TOKEN set to a management token with tags ["tf-acc-scoped"].
func TestAccTenantTagScoping_basic(t *testing.T) {
	scopedToken := os.Getenv("HATCHET_SCOPED_TOKEN")
	if scopedToken == "" {
		t.Skip("HATCHET_SCOPED_TOKEN not set (needs a management token with tags [\"tf-acc-scoped\"])")
	}
	name := testAccUniqueName("tenant-scoped")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantScopedProviderConfig(scopedToken, name, []string{"tf-acc-scoped"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatchet_tenant.scoped", "tags.#", "1"),
					resource.TestCheckResourceAttr("hatchet_tenant.scoped", "tags.0", "tf-acc-scoped"),
				),
			},
		},
	})
}

// TestAccTenantTagScoping_accessDenied verifies that a scoped management token cannot
// create a tenant with tags outside the token's own tag set.
// Requires HATCHET_SCOPED_TOKEN set to a management token with tags ["tf-acc-scoped"].
func TestAccTenantTagScoping_accessDenied(t *testing.T) {
	scopedToken := os.Getenv("HATCHET_SCOPED_TOKEN")
	if scopedToken == "" {
		t.Skip("HATCHET_SCOPED_TOKEN not set (needs a management token with tags [\"tf-acc-scoped\"])")
	}
	name := testAccUniqueName("tenant-denied")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTenantScopedProviderConfig(scopedToken, name, []string{"other-team"}),
				ExpectError: regexp.MustCompile(`40[03]`),
			},
		},
	})
}

func testAccTenantResourceConfig(name string) string {
	return fmt.Sprintf(`%s
resource "hatchet_tenant" "test" {
  name = %q
}`, testAccProviderConfig(), name)
}

func testAccTenantResourceConfigWithSlug(name, slug string) string {
	return fmt.Sprintf(`%s
resource "hatchet_tenant" "test" {
  name = %q
  slug = %q
}`, testAccProviderConfig(), name, slug)
}

func testAccTenantResourceConfigWithTags(name string, tags []string) string {
	tagsHCL := "[]"
	if len(tags) > 0 {
		inner := ""
		for _, tag := range tags {
			inner += fmt.Sprintf("%q, ", tag)
		}
		tagsHCL = fmt.Sprintf("[%s]", inner[:len(inner)-2])
	}
	return fmt.Sprintf(`%s
resource "hatchet_tenant" "test" {
  name = %q
  tags = %s
}`, testAccProviderConfig(), name, tagsHCL)
}

// testAccTenantScopedProviderConfig generates HCL that uses an aliased scoped provider
// so that tag-based access control can be exercised without touching the main provider config.
func testAccTenantScopedProviderConfig(scopedToken, name string, tags []string) string {
	inner := ""
	for _, tag := range tags {
		inner += fmt.Sprintf("%q, ", tag)
	}
	tagsHCL := "[]"
	if len(inner) > 0 {
		tagsHCL = fmt.Sprintf("[%s]", inner[:len(inner)-2])
	}
	return fmt.Sprintf(`%s

provider "hatchet" {
  alias = "scoped"
  token = %q
}

resource "hatchet_tenant" "scoped" {
  provider = hatchet.scoped
  name     = %q
  tags     = %s
}`, testAccProviderConfig(), scopedToken, name, tagsHCL)
}
