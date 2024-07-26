// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSMountTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"
	resourceName2 := "aws_efs_mount_target.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_name"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, names.AttrDNSName, "efs", regexache.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "file_system_arn", "elasticfilesystem", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrIPAddress, regexache.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMountTargetConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					testAccCheckMountTargetExists(ctx, resourceName2, &mount),
					acctest.MatchResourceAttrRegionalHostname(resourceName, names.AttrDNSName, "efs", regexache.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalHostname(resourceName2, names.AttrDNSName, "efs", regexache.MustCompile(`fs-[^.]+`)),
				),
			},
		},
	})
}

func TestAccEFSMountTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfefs.ResourceMountTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSMountTarget_ipAddress(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddress(rName, "10.0.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddress, "10.0.0.100"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13845
func TestAccEFSMountTarget_IPAddress_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressNullIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					resource.TestMatchResourceAttr(resourceName, names.AttrIPAddress, regexache.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func testAccCheckMountTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_mount_target" {
				continue
			}

			_, err := tfefs.FindMountTargetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EFS Mount Target %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMountTargetExists(ctx context.Context, n string, v *awstypes.MountTargetDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		output, err := tfefs.FindMountTargetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccMountTargetConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccMountTargetConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_base(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}
`)
}

func testAccMountTargetConfig_modified(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_base(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[0].id
}

resource "aws_efs_mount_target" "test2" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test[1].id
}
`)
}

func testAccMountTargetConfig_ipAddress(rName, ipAddress string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_base(rName), fmt.Sprintf(`
resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  ip_address     = %[1]q
  subnet_id      = aws_subnet.test[0].id
}
`, ipAddress))
}

func testAccMountTargetConfig_ipAddressNullIP(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_base(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  ip_address     = null
  subnet_id      = aws_subnet.test[0].id
}
`)
}
