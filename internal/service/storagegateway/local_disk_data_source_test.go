package storagegateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccStorageGatewayLocalDiskDataSource_diskNode(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccLocalDiskDataSourceConfig_nodeNonExistent(rName),
				ExpectError: regexp.MustCompile(`no results found`),
			},
			{
				Config: testAccLocalDiskDataSourceConfig_node(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccLocalDiskExistsDataSource(dataSourceName),
					resource.TestMatchResourceAttr(dataSourceName, "disk_id", regexp.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_node", regexp.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_path", regexp.MustCompile(`.+`)),
				),
			},
		},
	})
}

func TestAccStorageGatewayLocalDiskDataSource_diskPath(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccLocalDiskDataSourceConfig_pathNonExistent(rName),
				ExpectError: regexp.MustCompile(`no results found`),
			},
			{
				Config: testAccLocalDiskDataSourceConfig_path(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccLocalDiskExistsDataSource(dataSourceName),
					resource.TestMatchResourceAttr(dataSourceName, "disk_id", regexp.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_node", regexp.MustCompile(`.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "disk_path", regexp.MustCompile(`.+`)),
				),
			},
		},
	})
}

func testAccLocalDiskExistsDataSource(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("not found: %s", dataSourceName)
		}

		return nil
	}
}

func testAccLocalDiskBaseDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayConfig_typeFileS3(rName),
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
		testAccLocalDiskBaseDataSourceConfig(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccLocalDiskDataSourceConfig_nodeNonExistent(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskBaseDataSourceConfig(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_node   = replace(aws_volume_attachment.test.device_name, "xvdb", "nonexistent")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccLocalDiskDataSourceConfig_path(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskBaseDataSourceConfig(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_path   = split(".", aws_instance.test.instance_type)[0] == "m4" ? aws_volume_attachment.test.device_name : replace(aws_volume_attachment.test.device_name, "xvdb", "nvme1n1")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccLocalDiskDataSourceConfig_pathNonExistent(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalDiskBaseDataSourceConfig(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_path   = replace(aws_volume_attachment.test.device_name, "xvdb", "nonexistent")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}
