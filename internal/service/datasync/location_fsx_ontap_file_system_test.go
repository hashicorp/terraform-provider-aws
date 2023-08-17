// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
)

func TestAccDataSyncLocationFSxOntapFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var locationFsxOntap1 datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"
	fsResourceName := "aws_fsx_ontap_file_system.test"
	svmResourceName := "aws_fsx_ontap_storage_virtual_machine.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxOntapDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxOntapFileSystemConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxOntapExists(ctx, resourceName, &locationFsxOntap1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "fsx_filesystem_arn", fsResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_virtual_machine_arn", svmResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^fsxn-(nfs|smb)://.+/`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxOntapImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxOntapFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var locationFsxOntap1 datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxOntapDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxOntapFileSystemConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxOntapExists(ctx, resourceName, &locationFsxOntap1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationFSxOntapFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationFSxOntapFileSystem_subdirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var locationFsxOntap1 datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxOntapDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxOntapFileSystemConfig_subdirectory("/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxOntapExists(ctx, resourceName, &locationFsxOntap1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxOntapImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxOntapFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var locationFsxOntap1 datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxOntapDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxOntapFileSystemConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxOntapExists(ctx, resourceName, &locationFsxOntap1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxOntapImportStateID(resourceName),
			},
			{
				Config: testAccLocationFSxOntapFileSystemConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxOntapExists(ctx, resourceName, &locationFsxOntap1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationFSxOntapFileSystemConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxOntapExists(ctx, resourceName, &locationFsxOntap1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationFSxOntapDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_fsx_ontap_file_system" {
				continue
			}

			input := &datasync.DescribeLocationFsxOntapInput{
				LocationArn: aws.String(rs.Primary.ID),
			}

			_, err := conn.DescribeLocationFsxOntapWithContext(ctx, input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				return nil
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckLocationFSxOntapExists(ctx context.Context, resourceName string, locationFsxOntap *datasync.DescribeLocationFsxOntapOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn(ctx)

		input := &datasync.DescribeLocationFsxOntapInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationFsxOntapWithContext(ctx, input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationFsxOntap = *output

		return nil
	}
}

func testAccLocationFSxOntapImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s#%s", rs.Primary.ID, rs.Primary.Attributes["fsx_filesystem_arn"]), nil
	}
}

func testAccLocationFSxOntapFileSystemConfig_basic() string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`))
}

func testAccLocationFSxOntapFileSystemConfig_subdirectory(subdirectory string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn
  subdirectory                = %[1]q

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`, subdirectory))
}

func testAccLocationFSxOntapFileSystemConfig_tags1(key1, value1 string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  tags = {
    %[1]q = %[2]q
  }

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`, key1, value1))
}

func testAccLocationFSxOntapFileSystemConfig_tags2(key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`, key1, value1, key2, value2))
}

func testAccFSxOntapFileSystemBaseConfig() string {
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

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test.id
}

resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = "test"
}
`)
}
