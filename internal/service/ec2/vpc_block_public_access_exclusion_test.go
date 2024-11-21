// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Acceptance test access AWS and cost money to run.
func TestAccVPCBlockPublicAccessExclusion_basic_vpc(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	blockMode := string(awstypes.InternetGatewayBlockModeBlockBidirectional)
	exclusionModeBidrectional := string(awstypes.InternetGatewayExclusionModeAllowBidirectional)
	internetGatewayExclusionModeAllowEgress := string(awstypes.InternetGatewayExclusionModeAllowEgress)

	resourceName := "aws_vpc_block_public_access_exclusion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basic_vpc(blockMode, exclusionModeBidrectional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessExclusionExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "exclusion_id", regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_exclusion_mode", exclusionModeBidrectional),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrResourceARN, "ec2", regexache.MustCompile(`vpc/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basic_vpc(blockMode, internetGatewayExclusionModeAllowEgress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessExclusionExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "exclusion_id", regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_exclusion_mode", internetGatewayExclusionModeAllowEgress),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(resourceName, "reason"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrResourceARN, "ec2", regexache.MustCompile(`vpc/+.`)),
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

func TestAccVPCBlockPublicAccessExclusion_basic_subnet(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	blockMode := string(awstypes.InternetGatewayBlockModeBlockBidirectional)
	exclusionModeBidrectional := string(awstypes.InternetGatewayExclusionModeAllowBidirectional)
	internetGatewayExclusionModeAllowEgress := string(awstypes.InternetGatewayExclusionModeAllowEgress)

	resourceName := "aws_vpc_block_public_access_exclusion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basic_subnet(blockMode, exclusionModeBidrectional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessExclusionExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "exclusion_id", regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_exclusion_mode", exclusionModeBidrectional),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrSubnetID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrResourceARN, "ec2", regexache.MustCompile(`subnet/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basic_subnet(blockMode, internetGatewayExclusionModeAllowEgress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessExclusionExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestMatchResourceAttr(resourceName, "exclusion_id", regexache.MustCompile(`vpcbpa-exclude-([0-9a-fA-F]+)$`)),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_exclusion_mode", internetGatewayExclusionModeAllowEgress),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrSubnetID),
					resource.TestCheckResourceAttrSet(resourceName, "reason"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrResourceARN, "ec2", regexache.MustCompile(`subnet/+.`)),
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

func TestAccVPCBlockPublicAccessExclusion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	blockMode := string(awstypes.InternetGatewayBlockModeBlockBidirectional)
	exclusionMode := string(awstypes.InternetGatewayExclusionModeAllowBidirectional)
	resourceName := "aws_vpc_block_public_access_exclusion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessExclusionConfig_basic_vpc(blockMode, exclusionMode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessExclusionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCBlockPublicAccessExclusion, resourceName),
				),
			},
		},
	})
}

func testAccCheckBlockPublicAccessExclusionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_block_public_access_exclusion" {
				continue
			}

			id := rs.Primary.Attributes[names.AttrID]

			out, err := tfec2.FindVPCBlockPublicAccessExclusionByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCBlockPublicAccessExclusion, id, err)
			}

			//If the status is Delete Complete, that indicates the resource has been destroyed
			if out.State == awstypes.VpcBlockPublicAccessExclusionStateDeleteComplete {
				return nil
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCBlockPublicAccessExclusion, id, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBlockPublicAccessExclusionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCBlockPublicAccessExclusion, name, errors.New("not found"))
		}

		id := rs.Primary.Attributes[names.AttrID]

		if id == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCBlockPublicAccessExclusion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		_, err := tfec2.FindVPCBlockPublicAccessExclusionByID(ctx, conn, id)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCBlockPublicAccessExclusion, rs.Primary.Attributes[names.AttrID], err)
		}

		return nil
	}
}

const testAccVPCBlockPublicAccessExclusionConfig_base = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
`

func testAccVPCBlockPublicAccessExclusionConfig_basic_vpc(rBlockMode, rExclusionMode string) string {
	return acctest.ConfigCompose(testAccVPCBlockPublicAccessExclusionConfig_base, fmt.Sprintf(`

resource "aws_vpc_block_public_access_exclusion" "test" {
  internet_gateway_exclusion_mode = %[2]q
  vpc_id                          = aws_vpc.test.id
}
`, rBlockMode, rExclusionMode))
}

func testAccVPCBlockPublicAccessExclusionConfig_basic_subnet(rBlockMode, rExclusionMode string) string {
	return acctest.ConfigCompose(testAccVPCBlockPublicAccessExclusionConfig_base, fmt.Sprintf(`

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_vpc_block_public_access_exclusion" "test" {
  internet_gateway_exclusion_mode = %[2]q
  subnet_id                       = aws_subnet.test.id
}
`, rBlockMode, rExclusionMode))
}
