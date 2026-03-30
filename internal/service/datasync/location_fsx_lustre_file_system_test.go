// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationFSxLustreFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"
	fsResourceName := "aws_fsx_lustre_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxforLustreFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxforLustreFileSystemExists(ctx, t, resourceName, &locationFsxLustre1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "fsx_filesystem_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^fsxl://.+/`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
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
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxforLustreFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxforLustreFileSystemExists(ctx, t, resourceName, &locationFsxLustre1),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdatasync.ResourceLocationFSxLustreFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationFSxLustreFileSystem_subdirectory(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxforLustreFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_subdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxforLustreFileSystemExists(ctx, t, resourceName, &locationFsxLustre1),
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
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var locationFsxLustre1 datasync.DescribeLocationFsxLustreOutput
	resourceName := "aws_datasync_location_fsx_lustre_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxforLustreFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxLustreFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxforLustreFileSystemExists(ctx, t, resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxLustreImportStateID(resourceName),
			},
			{
				Config: testAccLocationFSxLustreFileSystemConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxforLustreFileSystemExists(ctx, t, resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationFSxLustreFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxforLustreFileSystemExists(ctx, t, resourceName, &locationFsxLustre1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckLocationFSxforLustreFileSystemDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_fsx_lustre_file_system" {
				continue
			}

			_, err := tfdatasync.FindLocationFSxLustreByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location FSx for Lustre File System %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationFSxforLustreFileSystemExists(ctx context.Context, t *testing.T, n string, v *datasync.DescribeLocationFsxLustreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationFSxLustreByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

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

func testAccFSxLustreFileSystemConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = [aws_security_group.test.id]
  storage_capacity   = 1200
  subnet_ids         = aws_subnet.test[*].id
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLocationFSxLustreFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemConfig_base(rName), `
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]
}
`)
}

func testAccLocationFSxLustreFileSystemConfig_subdirectory(rName, subdirectory string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]
  subdirectory        = %[1]q
}
`, subdirectory))
}

func testAccLocationFSxLustreFileSystemConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_fsx_lustre_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationFSxLustreFileSystemConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccFSxLustreFileSystemConfig_base(rName), fmt.Sprintf(`
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
