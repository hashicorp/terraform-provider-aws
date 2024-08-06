// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDirectoryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()

	resourceName := "aws_workspaces_directory.test"
	dataSourceName := "data.aws_workspaces_directory.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDirectory(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "workspaces_DefaultRole")
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAlias, resourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_name", resourceName, "directory_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_type", resourceName, "directory_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_ip_addresses.#", resourceName, "dns_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_role_id", resourceName, "iam_role_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_group_ids", resourceName, "ip_group_ids"),
					resource.TestCheckResourceAttrPair(dataSourceName, "registration_code", resourceName, "registration_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.#", resourceName, "self_service_permissions.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.change_compute_type", resourceName, "self_service_permissions.0.change_compute_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.increase_volume_size", resourceName, "self_service_permissions.0.increase_volume_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.rebuild_workspace", resourceName, "self_service_permissions.0.rebuild_workspace"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.restart_workspace", resourceName, "self_service_permissions.0.restart_workspace"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.switch_running_mode", resourceName, "self_service_permissions.0.switch_running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.#", resourceName, "workspace_access_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_android", resourceName, "workspace_access_properties.0.device_type_android"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_chromeos", resourceName, "workspace_access_properties.0.device_type_chromeos"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_ios", resourceName, "workspace_access_properties.0.device_type_ios"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_linux", resourceName, "workspace_access_properties.0.device_type_linux"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_osx", resourceName, "workspace_access_properties.0.device_type_osx"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_web", resourceName, "workspace_access_properties.0.device_type_web"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_windows", resourceName, "workspace_access_properties.0.device_type_windows"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_access_properties.0.device_type_zeroclient", resourceName, "workspace_access_properties.0.device_type_zeroclient"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.#", resourceName, "workspace_creation_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.custom_security_group_id", resourceName, "workspace_creation_properties.0.custom_security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.default_ou", resourceName, "workspace_creation_properties.0.default_ou"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.enable_internet_access", resourceName, "workspace_creation_properties.0.enable_internet_access"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.enable_maintenance_mode", resourceName, "workspace_creation_properties.0.enable_maintenance_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_security_group_id", resourceName, "workspace_security_group_id"),
				),
			},
		},
	})
}

func testAccDirectoryDataSourceConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccDirectoryConfig_Prerequisites(rName, domain),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = "tf-testacc-workspaces-directory-%[1]s"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}

resource "aws_workspaces_directory" "test" {
  directory_id = aws_directory_service_directory.main.id

  self_service_permissions {
    change_compute_type  = false
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = false
    switch_running_mode  = true
  }

  workspace_access_properties {
    device_type_android    = "ALLOW"
    device_type_chromeos   = "ALLOW"
    device_type_ios        = "ALLOW"
    device_type_linux      = "DENY"
    device_type_osx        = "ALLOW"
    device_type_web        = "DENY"
    device_type_windows    = "DENY"
    device_type_zeroclient = "DENY"
  }

  workspace_creation_properties {
    custom_security_group_id            = aws_security_group.test.id
    default_ou                          = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access              = true
    enable_maintenance_mode             = false
    user_enabled_as_local_administrator = false
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}

data "aws_workspaces_directory" "test" {
  directory_id = aws_workspaces_directory.test.id
}

data "aws_iam_role" "workspaces-default" {
  name = "workspaces_DefaultRole"
}
`, rName))
}
