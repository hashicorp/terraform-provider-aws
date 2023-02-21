package efs_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEFSMountTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mount efs.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"
	resourceName2 := "aws_efs_mount_target.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, efs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_name"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "file_system_arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalHostname(resourceName2, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
				),
			},
		},
	})
}

func TestAccEFSMountTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mount efs.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, efs.EndpointsID),
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
	var mount efs.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, efs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddress(rName, "10.0.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.100"),
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
	var mount efs.MountTargetDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, efs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_ipAddressNullIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMountTargetExists(ctx, resourceName, &mount),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn()
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

func testAccCheckMountTargetExists(ctx context.Context, n string, v *efs.MountTargetDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EFS Mount Target ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn()

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
