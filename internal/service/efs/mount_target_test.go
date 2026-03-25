// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSMountTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"
	resourceName2 := "aws_efs_mount_target.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_name"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, names.AttrDNSName, "efs", regexache.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "file_system_arn", "elasticfilesystem", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrIPAddress, regexache.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, string(awstypes.IpAddressTypeIpv4Only)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address", ""),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
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
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					testAccCheckMountTargetExists(ctx, t, resourceName2, &mount),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					acctest.CheckSDKResourceDisappears(ctx, t, tfefs.ResourceMountTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSMountTarget_ipAddress(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddress(rName, "10.0.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
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

func TestAccEFSMountTarget_ipAddressTypeIPv6Only(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressTypeIPv6Only(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddress, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, string(awstypes.IpAddressTypeIpv6Only)),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_address"),
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

func TestAccEFSMountTarget_ipAddressTypeIPv6OnlyWithIPv6Address(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressTypeIPv6OnlyWithIPv6Address(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddress, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, string(awstypes.IpAddressTypeIpv6Only)),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_address"),
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

func TestAccEFSMountTarget_ipAddressTypeDualStack(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressTypeDualStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, string(awstypes.IpAddressTypeDualStack)),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_address"),
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

func TestAccEFSMountTarget_ipAddressTypeDualStackWithIPv6Address(t *testing.T) {
	ctx := acctest.Context(t)
	var mount awstypes.MountTargetDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressTypeDualStackWithIPv6Address(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, string(awstypes.IpAddressTypeDualStack)),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_address"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressNullIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &mount),
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

func testAccCheckMountTargetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EFSClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_mount_target" {
				continue
			}

			_, err := tfefs.FindMountTargetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckMountTargetExists(ctx context.Context, t *testing.T, n string, v *awstypes.MountTargetDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EFSClient(ctx)

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

func testAccMountTargetConfig_withDualStackSubnet(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 2), fmt.Sprintf(`
resource "aws_subnet" "test_ipv6_only" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 7)

  enable_resource_name_dns_aaaa_record_on_launch = true

  assign_ipv6_address_on_creation = true

  ipv6_native = true

  tags = {
    Name = %[1]q
  }
}

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

func testAccMountTargetConfig_ipAddressTypeIPv6Only(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_withDualStackSubnet(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id  = aws_efs_file_system.test.id
  ip_address_type = "IPV6_ONLY"
  subnet_id       = aws_subnet.test_ipv6_only.id
}
`)
}

func testAccMountTargetConfig_ipAddressTypeIPv6OnlyWithIPv6Address(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_withDualStackSubnet(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id  = aws_efs_file_system.test.id
  ip_address_type = "IPV6_ONLY"
  ipv6_address    = cidrhost(aws_subnet.test_ipv6_only.ipv6_cidr_block, 10)
  subnet_id       = aws_subnet.test_ipv6_only.id
}
`)
}

func testAccMountTargetConfig_ipAddressTypeDualStack(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_withDualStackSubnet(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id  = aws_efs_file_system.test.id
  ip_address_type = "DUAL_STACK"
  subnet_id       = aws_subnet.test[0].id
}
`)
}

func testAccMountTargetConfig_ipAddressTypeDualStackWithIPv6Address(rName string) string {
	return acctest.ConfigCompose(testAccMountTargetConfig_withDualStackSubnet(rName), `
resource "aws_efs_mount_target" "test" {
  file_system_id  = aws_efs_file_system.test.id
  ip_address_type = "DUAL_STACK"
  ipv6_address    = cidrhost(aws_subnet.test[0].ipv6_cidr_block, 10)
  subnet_id       = aws_subnet.test[0].id
}
`)
}
