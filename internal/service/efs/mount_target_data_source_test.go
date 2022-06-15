package efs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEFSMountTargetDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetDataSourceConfig_byID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups", resourceName, "security_groups"),
				),
			},
		},
	})
}

func TestAccEFSMountTargetDataSource_byAccessPointID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetDataSourceConfig_byAccessPointID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups", resourceName, "security_groups"),
				),
			},
		},
	})
}

func TestAccEFSMountTargetDataSource_byFileSystemID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetDataSourceConfig_byFileSystemID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups", resourceName, "security_groups"),
				),
			},
		},
	})
}

func testAccMountTargetBaseDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccMountTargetDataSourceConfig_byID(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetBaseDataSourceConfig(rName), `
data "aws_efs_mount_target" "test" {
  mount_target_id = aws_efs_mount_target.test.id
}
`)
}

func testAccMountTargetDataSourceConfig_byAccessPointID(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetBaseDataSourceConfig(rName), `
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

data "aws_efs_mount_target" "test" {
  access_point_id = aws_efs_access_point.test.id
}
`)
}

func testAccMountTargetDataSourceConfig_byFileSystemID(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetBaseDataSourceConfig(rName), `
data "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id

  depends_on = [aws_efs_mount_target.test]
}
`)
}
