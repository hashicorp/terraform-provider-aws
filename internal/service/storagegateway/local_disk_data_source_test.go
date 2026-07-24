// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccStorageGatewayLocalDiskDataSource_diskNode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLocalDiskDataSourceConfig_nodeNonExistent(rName),
				ExpectError: regexache.MustCompile(`no matching Storage Gateway Local Disk found`),
			},
			{
				Config: testAccLocalDiskDataSourceConfig_node(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "disk_id", regexache.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_node", regexache.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_path", regexache.MustCompile(`.+`)),
				),
			},
		},
	})
}

func TestAccStorageGatewayLocalDiskDataSource_diskPath(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLocalDiskDataSourceConfig_pathNonExistent(rName),
				ExpectError: regexache.MustCompile(`no matching Storage Gateway Local Disk found`),
			},
			{
				Config: testAccLocalDiskDataSourceConfig_path(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "disk_id", regexache.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_node", regexache.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_path", regexache.MustCompile(`.+`)),
				),
			},
		},
	})
}

func testAccLocalDiskDataSourceConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayConfig_typeFileS3(rName, rName),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}
`, rName))
}

func testAccLocalDiskDataSourceConfig_node(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskDataSourceConfig_base(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccLocalDiskDataSourceConfig_nodeNonExistent(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskDataSourceConfig_base(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_node   = replace(aws_volume_attachment.test.device_name, "xvdb", "nonexistent")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccLocalDiskDataSourceConfig_path(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskDataSourceConfig_base(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_path   = split(".", aws_instance.test.instance_type)[0] == "m4" ? aws_volume_attachment.test.device_name : replace(aws_volume_attachment.test.device_name, "xvdb", "nvme1n1")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccLocalDiskDataSourceConfig_pathNonExistent(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskDataSourceConfig_base(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_path   = replace(aws_volume_attachment.test.device_name, "xvdb", "nonexistent")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}
