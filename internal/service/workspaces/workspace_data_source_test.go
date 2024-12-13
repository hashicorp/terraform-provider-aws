// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWorkspaceDataSource_byWorkspaceID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()
	dataSourceName := "data.aws_workspaces_workspace.test"
	resourceName := "aws_workspaces_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole") },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_byID(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bundle_id", resourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddress, resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_volume_encryption_enabled", resourceName, "root_volume_encryption_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrUserName),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_encryption_key", resourceName, "volume_encryption_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.#", resourceName, "workspace_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.compute_type_name", resourceName, "workspace_properties.0.compute_type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.root_volume_size_gib", resourceName, "workspace_properties.0.root_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode", resourceName, "workspace_properties.0.running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.user_volume_size_gib", resourceName, "workspace_properties.0.user_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccWorkspaceDataSource_byDirectoryID_userName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()
	dataSourceName := "data.aws_workspaces_workspace.test"
	resourceName := "aws_workspaces_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole") },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_byDirectoryIDUserName(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bundle_id", resourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddress, resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_volume_encryption_enabled", resourceName, "root_volume_encryption_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrUserName),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_encryption_key", resourceName, "volume_encryption_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.#", resourceName, "workspace_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.compute_type_name", resourceName, "workspace_properties.0.compute_type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.root_volume_size_gib", resourceName, "workspace_properties.0.root_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode", resourceName, "workspace_properties.0.running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.user_volume_size_gib", resourceName, "workspace_properties.0.user_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccWorkspaceDataSource_workspaceIDAndDirectoryIDConflict(t *testing.T) {
	ctx := acctest.Context(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole") },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkspaceDataSourceConfig_idAndDirectoryIDConflict(),
				ExpectError: regexache.MustCompile("\"workspace_id\": conflicts with directory_id"),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig_byID(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccWorkspaceConfig_Prerequisites(rName, domain),
		`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "AWS_WorkSpaces" user is always present in a bare directory.
  user_name = "AWS_WorkSpaces"

  workspace_properties {
    root_volume_size_gib = 80
    user_volume_size_gib = 10
  }

  tags = {
    TerraformProviderAwsTest = true
  }
}

data "aws_workspaces_workspace" "test" {
  workspace_id = aws_workspaces_workspace.test.id
}
`)
}

func testAccWorkspaceDataSourceConfig_byDirectoryIDUserName(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccWorkspaceConfig_Prerequisites(rName, domain),
		`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
    root_volume_size_gib = 80
    user_volume_size_gib = 10
  }

  tags = {
    TerraformProviderAwsTest = true
  }
}

data "aws_workspaces_workspace" "test" {
  directory_id = aws_workspaces_workspace.test.directory_id
  user_name    = aws_workspaces_workspace.test.user_name
}
`)
}

func testAccWorkspaceDataSourceConfig_idAndDirectoryIDConflict() string {
	return `
data "aws_workspaces_workspace" "test" {
  workspace_id = "ws-cj5xcxsz5"
  directory_id = "d-9967252f57"
  user_name    = "Administrator"
}
`
}
