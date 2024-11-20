// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCBlockPublicAccessOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_vpc_block_public_access_options.test"

	rMode := string(awstypes.InternetGatewayBlockModeBlockBidirectional)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCBlockPublicAccessOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_block_mode", rMode),
					resource.TestCheckResourceAttrSet(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, "reason"),
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

func TestAccVPCBlockPublicAccessOptions_updates(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_vpc_block_public_access_options.test"

	rMode1 := string(awstypes.InternetGatewayBlockModeBlockBidirectional)
	rMode2 := string(awstypes.InternetGatewayBlockModeBlockIngress)
	rMode3 := string(awstypes.InternetGatewayBlockModeOff)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCBlockPublicAccessOptionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_block_mode", rMode1),
					resource.TestCheckResourceAttrSet(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, "reason"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_block_mode", rMode2),
					resource.TestCheckResourceAttrSet(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, "reason"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "internet_gateway_block_mode", rMode3),
					resource.TestCheckResourceAttrSet(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, "reason"),
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

func testAccCheckVPCBlockPublicAccessOptionsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_block_public_access_options" {
				continue
			}

			out, err := conn.DescribeVpcBlockPublicAccessOptions(ctx, &ec2.DescribeVpcBlockPublicAccessOptionsInput{})
			if out.VpcBlockPublicAccessOptions.InternetGatewayBlockMode == awstypes.InternetGatewayBlockModeOff {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCBlockPublicAccessOptions, rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCBlockPublicAccessOptions, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccPreCheckVPCBlockPublicAccess(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpcBlockPublicAccessOptionsInput{}
	_, err := conn.DescribeVpcBlockPublicAccessOptions(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVPCBlockPublicAccessOptionsConfig_basic(rMode string) string {
	return fmt.Sprintf(`
resource "aws_vpc_block_public_access_options" "test" {
  internet_gateway_block_mode = %[1]q
}
`, rMode)
}
