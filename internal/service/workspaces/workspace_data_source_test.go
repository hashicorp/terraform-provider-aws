package workspaces_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccWorkspaceDataSource_byWorkspaceID(t *testing.T) {
	rName := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()

	dataSourceName := "data.aws_workspaces_workspace.test"
	resourceName := "aws_workspaces_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_byID(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bundle_id", resourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_volume_encryption_enabled", resourceName, "root_volume_encryption_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_encryption_key", resourceName, "volume_encryption_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.#", resourceName, "workspace_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.compute_type_name", resourceName, "workspace_properties.0.compute_type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.root_volume_size_gib", resourceName, "workspace_properties.0.root_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode", resourceName, "workspace_properties.0.running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.user_volume_size_gib", resourceName, "workspace_properties.0.user_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSource_byDirectoryID_userName(t *testing.T) {
	rName := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()

	dataSourceName := "data.aws_workspaces_workspace.test"
	resourceName := "aws_workspaces_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_byDirectoryIDUserName(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bundle_id", resourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_volume_encryption_enabled", resourceName, "root_volume_encryption_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_encryption_key", resourceName, "volume_encryption_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.#", resourceName, "workspace_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.compute_type_name", resourceName, "workspace_properties.0.compute_type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.root_volume_size_gib", resourceName, "workspace_properties.0.root_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode", resourceName, "workspace_properties.0.running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.user_volume_size_gib", resourceName, "workspace_properties.0.user_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSource_workspaceIDAndDirectoryIDConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkspaceDataSourceConfig_idAndDirectoryIDConflict(),
				ExpectError: regexp.MustCompile("\"workspace_id\": conflicts with directory_id"),
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
