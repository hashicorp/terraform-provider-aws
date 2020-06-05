package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsWorkspacesDirectory_basic(t *testing.T) {
	rName := acctest.RandString(8)

	iamRoleDataSourceName := "data.aws_iam_role.workspaces-default"
	directoryResourceName := "aws_directory_service_directory.main"
	dataSourceName := "data.aws_workspaces_directory.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsWorkspacesDirectoryConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "self_service_permissions.0.increase_volume_size", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "self_service_permissions.0.rebuild_workspace", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "self_service_permissions.0.restart_workspace", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "self_service_permissions.0.switch_running_mode", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "dns_ip_addresses.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "directory_type", "SIMPLE_AD"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_name", directoryResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "alias", directoryResourceName, "alias"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", directoryResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_role_id", iamRoleDataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "workspace_security_group_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "registration_code"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWorkspacesDirectoryConfig(rName string) string {
	return testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName) + fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"

  self_service_permissions {
    change_compute_type  = false
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = false
    switch_running_mode  = true
  }
}

data "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"
}

data "aws_iam_role" "workspaces-default" {
  name = "workspaces_DefaultRole"
}
`)
}
