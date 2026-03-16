// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCInternetGatewayAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	igwResourceName := "aws_internet_gateway.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInternetGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCInternetGatewayAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "internet_gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
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

func TestAccVPCInternetGatewayAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInternetGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCInternetGatewayAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceInternetGatewayAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCInternetGatewayAttachment_Disappears_internetGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInternetGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCInternetGatewayAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceInternetGatewayAttachment(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceInternetGateway(), igwResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInternetGatewayAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_internet_gateway_attachment" {
				continue
			}

			igwID, vpcID, err := tfec2.InternetGatewayAttachmentParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfec2.FindInternetGatewayAttachment(ctx, conn, igwID, vpcID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Internet Gateway Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInternetGatewayAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.InternetGatewayAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		igwID, vpcID, err := tfec2.InternetGatewayAttachmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfec2.FindInternetGatewayAttachment(ctx, conn, igwID, vpcID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCInternetGatewayAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway_attachment" "test" {
  internet_gateway_id = aws_internet_gateway.test.id
  vpc_id              = aws_vpc.test.id
}
`, rName)
}
