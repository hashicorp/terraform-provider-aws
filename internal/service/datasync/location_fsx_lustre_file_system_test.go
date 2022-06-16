package datasync_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDataSyncLocationFSxLustreFileSystem_basic(t *testing.T) {
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"
	fsResourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxLustreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxLustreExists(resourceName, &locationFsxLustre1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "fsx_filesystem_arn", fsResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^fsxl://.+/`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxLustreImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxLustreFileSystem_disappears(t *testing.T) {
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxLustreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxLustreExists(resourceName, &locationFsxLustre1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationFSxLustreFileSystem(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationFSxLustreFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationFSxLustreFileSystem_subdirectory(t *testing.T) {
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxLustreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_subdirectory("/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxLustreExists(resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxLustreImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxLustreFileSystem_tags(t *testing.T) {
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxLustreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxLustreExists(resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxLustreImportStateID(resourceName),
			},
			{
				Config: testAccLocationFSxLustreFileSystemConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxLustreExists(resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationFSxLustreFileSystemConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxLustreExists(resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationFSxLustreDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_fsx_lustre_file_system" {
			continue
		}

		_, err := tfdatasync.FindFSxLustreLocationByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataSync Task %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckLocationFSxLustreExists(resourceName string, locationFsxLustre *datasync.DescribeLocationFsxLustreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		output, err := tfdatasync.FindFSxLustreLocationByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationFsxLustre = *output

		return nil
	}
}

func testAccLocationFSxLustreImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s#%s", rs.Primary.ID, rs.Primary.Attributes["fsx_filesystem_arn"]), nil
	}
}

func testAccLocationFSxLustreFileSystemConfig_basic() string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemBaseConfig(), `
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]
}
`)
}

func testAccLocationFSxLustreFileSystemConfig_subdirectory(subdirectory string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]
  subdirectory        = %[1]q
}
`, subdirectory))
}

func testAccLocationFSxLustreFileSystemConfig_tags1(key1, value1 string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationFSxLustreFileSystemConfig_tags2(key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccFSxLustreFileSystemBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_security_group" "test" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = [aws_security_group.test.id]
  storage_capacity   = 1200
  subnet_ids         = [aws_subnet.test.id]
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}
`)
}
