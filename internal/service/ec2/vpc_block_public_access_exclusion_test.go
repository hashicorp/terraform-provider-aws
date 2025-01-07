// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCBlockPublicAccessExclusion_basicVPC(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_block_public_access_exclusion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	internetGatewayExclusionMode := string(awstypes.InternetGatewayExclusionModeAllowEgress)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basicVPC(rName, internetGatewayExclusionMode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCBlockPublicAccessExclusionExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion_mode"), knownvalue.StringExact(internetGatewayExclusionMode)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrSubnetID), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVPCID), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCBlockPublicAccessExclusion_basicSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_block_public_access_exclusion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	internetGatewayExclusionMode := string(awstypes.InternetGatewayExclusionModeAllowEgress)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basicSubnet(rName, internetGatewayExclusionMode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCBlockPublicAccessExclusionExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion_mode"), knownvalue.StringExact(internetGatewayExclusionMode)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrSubnetID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVPCID), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCBlockPublicAccessExclusion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_block_public_access_exclusion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	internetGatewayExclusionMode := string(awstypes.InternetGatewayExclusionModeAllowEgress)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basicVPC(rName, internetGatewayExclusionMode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCBlockPublicAccessExclusionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCBlockPublicAccessExclusion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCBlockPublicAccessExclusion_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_block_public_access_exclusion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	internetGatewayExclusionMode1 := string(awstypes.InternetGatewayExclusionModeAllowBidirectional)
	internetGatewayExclusionMode2 := string(awstypes.InternetGatewayExclusionModeAllowEgress)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basicVPC(rName, internetGatewayExclusionMode1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCBlockPublicAccessExclusionExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion_mode"), knownvalue.StringExact(internetGatewayExclusionMode1)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basicVPC(rName, internetGatewayExclusionMode2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCBlockPublicAccessExclusionExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_exclusion_mode"), knownvalue.StringExact(internetGatewayExclusionMode2)),
				},
			},
		},
	})
}

func testAccCheckVPCBlockPublicAccessExclusionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_block_public_access_exclusion" {
				continue
			}

			_, err := tfec2.FindVPCBlockPublicAccessExclusionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Block Public Access Exclusion %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCBlockPublicAccessExclusionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindVPCBlockPublicAccessExclusionByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

// For generated tests.
var (
	testAccCheckBlockPublicAccessExclusionDestroy = testAccCheckVPCBlockPublicAccessExclusionDestroy
	testAccCheckBlockPublicAccessExclusionExists  = testAccCheckVPCBlockPublicAccessExclusionExists
)

func testAccVPCBlockPublicAccessExclusionConfig_basicVPC(rName, internetGatewayExclusionMode string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_vpc_block_public_access_exclusion" "test" {
  internet_gateway_exclusion_mode = %[1]q
  vpc_id                          = aws_vpc.test.id
}
`, internetGatewayExclusionMode))
}

func testAccVPCBlockPublicAccessExclusionConfig_basicSubnet(rName, internetGatewayExclusionMode string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_vpc_block_public_access_exclusion" "test" {
  internet_gateway_exclusion_mode = %[1]q
  subnet_id                       = aws_subnet.test[0].id
}
`, internetGatewayExclusionMode))
}
