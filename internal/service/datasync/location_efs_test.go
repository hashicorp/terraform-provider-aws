// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationEFS_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationEfsOutput
	efsFileSystemResourceName := "aws_efs_file_system.test"
	resourceName := "aws_datasync_location_efs.test"
	subnetResourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationEFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ec2_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ec2_config.0.security_group_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_config.0.subnet_arn", subnetResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "efs_file_system_arn", efsFileSystemResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^efs://.+/`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSyncLocationEFS_accessPointARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationEFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_accessPointARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "access_point_arn", "aws_efs_access_point.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "in_transit_encryption", "TLS1_2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSyncLocationEFS_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationEFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationEFS(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationEFS(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationEFS_subdirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationEFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_subdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSyncLocationEFS_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationEFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocationEFSConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationEFSConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckLocationEFSDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_efs" {
				continue
			}

			_, err := tfdatasync.FindLocationEFSByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location EFS %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationEFSExists(ctx context.Context, n string, v *datasync.DescribeLocationEfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationEFSByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationEFSConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName)
}

func testAccLocationEFSConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLocationEFSConfig_base(rName), `
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`)
}

func testAccLocationEFSConfig_subdirectory(rName, subdirectory string) string {
	return acctest.ConfigCompose(testAccLocationEFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn
  subdirectory        = %[1]q

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`, subdirectory))
}

func testAccLocationEFSConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationEFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationEFSConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationEFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccLocationEFSConfig_accessPointARN(rName string) string {
	return acctest.ConfigCompose(testAccLocationEFSConfig_base(rName), fmt.Sprintf(`
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn   = aws_efs_mount_target.test.file_system_arn
  access_point_arn      = aws_efs_access_point.test.arn
  in_transit_encryption = "TLS1_2"

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`, rName))
}
