// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightVPCConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcConnection quicksight.VPCConnection
	resourceName := "aws_quicksight_vpc_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &vpcConnection),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight", fmt.Sprintf("vpcConnection/%[1]s", rId)),
					resource.TestCheckResourceAttr(resourceName, "vpc_connection_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
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
	var vpcConnection quicksight.VPCConnection
	resourceName := "aws_quicksight_vpc_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &vpcConnection),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceVPCConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightVPCConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpcConnection quicksight.VPCConnection
	resourceName := "aws_quicksight_vpc_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_tags1(rId, rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &vpcConnection),
					resource.TestCheckResourceAttr(resourceName, "vpc_connection_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccVPCConnectionConfig_tags2(rId, rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &vpcConnection),
					resource.TestCheckResourceAttr(resourceName, "vpc_connection_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCConnectionConfig_tags1(rId, rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &vpcConnection),
					resource.TestCheckResourceAttr(resourceName, "vpc_connection_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckVPCConnectionExists(ctx context.Context, resourceName string, vpcConnection *quicksight.VPCConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		output, err := tfquicksight.FindVPCConnectionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameVPCConnection, rs.Primary.ID, err)
		}

		*vpcConnection = *output

		return nil
	}
}

func testAccCheckVPCConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_vpc_connection" {
				continue
			}

			output, err := tfquicksight.FindVPCConnectionByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			if output != nil && aws.StringValue(output.Status) == quicksight.VPCConnectionResourceStatusDeleted {
				return nil
			}

			return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameVPCConnection, rs.Primary.ID, err)
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

func testAccVPCConnectionConfig_tags1(rId, rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[3]q = %[4]q
  }
}
`, rId, rName, tagKey1, tagValue1))
}

func testAccVPCConnectionConfig_tags2(rId, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rId, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
