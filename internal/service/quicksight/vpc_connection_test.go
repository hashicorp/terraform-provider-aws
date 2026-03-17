// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightVPCConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcConnection awstypes.VPCConnection
	resourceName := "aws_quicksight_vpc_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, t, resourceName, &vpcConnection),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("vpcConnection/%[1]s", rId)),
					resource.TestCheckResourceAttr(resourceName, "vpc_connection_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
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

func TestAccQuickSightVPCConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcConnection awstypes.VPCConnection
	resourceName := "aws_quicksight_vpc_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, t, resourceName, &vpcConnection),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceVPCConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCConnectionExists(ctx context.Context, t *testing.T, n string, v *awstypes.VPCConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindVPCConnectionByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["vpc_connection_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPCConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_vpc_connection" {
				continue
			}

			_, err := tfquicksight.FindVPCConnectionByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["vpc_connection_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight VPC Connection (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBaseVPCConnectionConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
      }
    ]
  })
  inline_policy {
    name = "QuicksightVPCConnectionRolePolicy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "ec2:CreateNetworkInterface",
            "ec2:ModifyNetworkInterfaceAttribute",
            "ec2:DeleteNetworkInterface",
            "ec2:DescribeSubnets",
            "ec2:DescribeSecurityGroups"
          ]
          Resource = ["*"]
        }
      ]
    })
  }
}
`)
}

func testAccVPCConnectionConfig_basic(rId string, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseVPCConnectionConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_vpc_connection" "test" {
  vpc_connection_id = %[1]q
  name              = %[2]q
  role_arn          = aws_iam_role.test.arn
  security_group_ids = [
    aws_security_group.test.id,
  ]
  subnet_ids = aws_subnet.test[*].id
}
`, rId, rName))
}
