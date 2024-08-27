// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWorkspaceBundleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_workspaces_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBundleDataSourceConfig_basic("wsb-b0s22j3d7"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "bundle_id", "wsb-b0s22j3d7"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.0.name", "PERFORMANCE"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "Performance with Windows 7"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrOwner, "Amazon"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.0.capacity", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.0.capacity", "100"),
				),
			},
		},
	})
}

func testAccWorkspaceBundleDataSource_byOwnerName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_workspaces_bundle.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBundleDataSourceConfig_byOwnerName("AMAZON", "Value with Windows 10 and Office 2016"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "bundle_id", "wsb-df76rqys9"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.0.name", "VALUE"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "Value with Windows 10 and Office 2016"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrOwner, "Amazon"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.0.capacity", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.0.capacity", acctest.Ct10),
				),
			},
		},
	})
}

func testAccWorkspaceBundleDataSource_bundleIDAndNameConflict(t *testing.T) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccBundleDataSourceConfig_idAndOwnerNameConflict("wsb-df76rqys9", "AMAZON", "Value with Windows 10 and Office 2016"),
				ExpectError: regexache.MustCompile("\"bundle_id\": conflicts with owner"),
			},
		},
	})
}

func testAccWorkspaceBundleDataSource_privateOwner(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_workspaces_bundle.test"
	bundleName := os.Getenv("AWS_WORKSPACES_BUNDLE_NAME")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccBundlePreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBundleDataSourceConfig_privateOwner(bundleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, bundleName),
				),
			},
		},
	})
}

func testAccBundlePreCheck(t *testing.T) {
	if os.Getenv("AWS_WORKSPACES_BUNDLE_NAME") == "" {
		t.Skip("AWS_WORKSPACES_BUNDLE_NAME env var must be set for AWS WorkSpaces private bundle acceptance test. This is required until AWS provides bundle creation API.")
	}
}

func testAccBundleDataSourceConfig_basic(bundleID string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
}
`, bundleID)
}

func testAccBundleDataSourceConfig_byOwnerName(owner, name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  owner = %q
  name  = %q
}
`, owner, name)
}

func testAccBundleDataSourceConfig_idAndOwnerNameConflict(bundleID, owner, name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
  owner     = %q
  name      = %q
}
`, bundleID, owner, name)
}

func testAccBundleDataSourceConfig_privateOwner(name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  name = %q
}
`, name)
}
