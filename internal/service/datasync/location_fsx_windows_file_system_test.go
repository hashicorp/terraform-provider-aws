// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationFSxWindowsFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"
	fsResourceName := "aws_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxWindowsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsFileSystemConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "fsx_filesystem_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^fsxw://.+/`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccLocationFSxWindowsImportStateID(resourceName),
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDataSyncLocationFSxWindowsFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxWindowsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsFileSystemConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationFSxWindowsFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationFSxWindowsFileSystem_subdirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxWindowsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsFileSystemConfig_subdirectory(rName, domainName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccLocationFSxWindowsImportStateID(resourceName),
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
		},
	})
}

func TestAccDataSyncLocationFSxWindowsFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxWindowsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsFileSystemConfig_tags1(rName, domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccLocationFSxWindowsImportStateID(resourceName),
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccLocationFSxWindowsFileSystemConfig_tags2(rName, domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationFSxWindowsFileSystemConfig_tags1(rName, domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckLocationFSxWindowsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_fsx_windows_file_system" {
				continue
			}

			_, err := tfdatasync.FindLocationFSxWindowsByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location FSx for Windows File Server File System %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationFSxWindowsExists(ctx context.Context, n string, v *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationFSxWindowsByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationFSxWindowsImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s#%s", rs.Primary.ID, rs.Primary.Attributes["fsx_filesystem_arn"]), nil
	}
}

func testAccLocationFSxWindowsFileSystemConfig_base(rName, domain string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}
`, domain))
}

func testAccLocationFSxWindowsFileSystemConfig_baseFS(rName, domain string) string {
	return acctest.ConfigCompose(testAccLocationFSxWindowsFileSystemConfig_base(rName, domain), fmt.Sprintf(`
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

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLocationFSxWindowsFileSystemConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccLocationFSxWindowsFileSystemConfig_baseFS(rName, domain), `
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]
}
`)
}

func testAccLocationFSxWindowsFileSystemConfig_subdirectory(rName, domain, subdirectory string) string {
	return acctest.ConfigCompose(testAccLocationFSxWindowsFileSystemConfig_baseFS(rName, domain), fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]
  subdirectory        = %[1]q
}
`, subdirectory))
}

func testAccLocationFSxWindowsFileSystemConfig_tags1(rName, domain, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationFSxWindowsFileSystemConfig_baseFS(rName, domain), fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationFSxWindowsFileSystemConfig_tags2(rName, domain, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationFSxWindowsFileSystemConfig_baseFS(rName, domain), fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}
