// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hatchet-dev/terraform-provider-hatchet/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hatchet": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("HATCHET_CLOUD_MANAGEMENT_TOKEN") == "" {
		t.Fatal("HATCHET_CLOUD_MANAGEMENT_TOKEN must be set for acceptance tests")
	}
}

func testAccProviderConfig() string {
	return `provider "hatchet" {}`
}

func testAccUniqueName(base string) string {
	suffix := os.Getenv("TF_ACC_SUFFIX")
	if suffix == "" {
		suffix = "default"
	}
	return fmt.Sprintf("%s-tf-acc-%s", base, suffix)
}
