// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSMountTargetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetDataSourceConfig_byID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrFileSystemID, resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddress, resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSubnetID, resourceName, names.AttrSubnetID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrNetworkInterfaceID, resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSecurityGroups, resourceName, names.AttrSecurityGroups),
				),
			},
		},
	})
}

func TestAccEFSMountTargetDataSource_byAccessPointID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetDataSourceConfig_byAccessPointID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrFileSystemID, resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddress, resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSubnetID, resourceName, names.AttrSubnetID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrNetworkInterfaceID, resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSecurityGroups, resourceName, names.AttrSecurityGroups),
				),
			},
		},
	})
}

func TestAccEFSMountTargetDataSource_byFileSystemID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetDataSourceConfig_byFileSystemID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrFileSystemID, resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddress, resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSubnetID, resourceName, names.AttrSubnetID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrNetworkInterfaceID, resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSecurityGroups, resourceName, names.AttrSecurityGroups),
				),
			},
		},
	})
}

func testAccMountTargetDataSourceConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}
`, rName))
}

func testAccMountTargetDataSourceConfig_byID(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetDataSourceConfig_base(rName), `
data "aws_efs_mount_target" "test" {
  mount_target_id = aws_efs_mount_target.test.id
}
`)
}

func testAccMountTargetDataSourceConfig_byAccessPointID(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetDataSourceConfig_base(rName), `
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

data "aws_efs_mount_target" "test" {
  access_point_id = aws_efs_access_point.test.id
}
`)
}

func testAccMountTargetDataSourceConfig_byFileSystemID(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetDataSourceConfig_base(rName), `
data "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id

  depends_on = [aws_efs_mount_target.test]
}
`)
}
