// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInterfacePermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_interface_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckNetworkInterfacePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfacePermissionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfacePermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAWSAccountID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrSet(resourceName, "permission"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccNetworkInterfacePermissionImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "network_interface_permission_id",
			},
		},
	})
}

func TestAccVPCNetworkInterfacePermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_interface_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckNetworkInterfacePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfacePermissionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfacePermissionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkInterfacePermission, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInterfacePermission_ownerExpectError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInterfacePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkInterfacePermissionConfig_accountOwner(rName),
				ExpectError: regexache.MustCompile(`OperationNotPermitted`),
			},
		},
	})
}

func testAccCheckNetworkInterfacePermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_network_interface_permission" {
				continue
			}

			_, err := tfec2.FindNetworkInterfacePermissionByID(ctx, conn, rs.Primary.Attributes["network_interface_permission_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Network Interface Permission %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkInterfacePermissionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindNetworkInterfacePermissionByID(ctx, conn, rs.Primary.Attributes["network_interface_permission_id"])

		return err
	}
}

func testAccNetworkInterfacePermissionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["network_interface_permission_id"], nil
	}
}

func testAccNetworkInterfacePermissionConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccNetworkInterfacePermissionConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), testAccNetworkInterfacePermissionConfig_base(rName), `
data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

resource "aws_network_interface_permission" "test" {
  network_interface_id = aws_network_interface.test.id
  aws_account_id       = data.aws_caller_identity.peer.account_id
  permission           = "INSTANCE-ATTACH"
}
`)
}

func testAccNetworkInterfacePermissionConfig_accountOwner(rName string) string {
	return acctest.ConfigCompose(testAccNetworkInterfacePermissionConfig_base(rName), `
data "aws_caller_identity" "test" {}

resource "aws_network_interface_permission" "test" {
  network_interface_id = aws_network_interface.test.id
  aws_account_id       = data.aws_caller_identity.test.account_id
  permission           = "INSTANCE-ATTACH"
}
`)
}
