// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2VPCLink_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCLinkExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVPCLinkConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
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

func TestAccAPIGatewayV2VPCLink_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceVPCLink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// func TestAccAPIGatewayV2VPCLink_tags(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	var v apigatewayv2.GetVpcLinkOutput
// 	resourceName := "aws_apigatewayv2_vpc_link.test"
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
// 		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckVPCLinkDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccVPCLinkConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckVPCLinkExists(ctx, resourceName, &v),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 			{
// 				Config: testAccVPCLinkConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckVPCLinkExists(ctx, resourceName, &v),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
// 				),
// 			},
// 			{
// 				Config: testAccVPCLinkConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckVPCLinkExists(ctx, resourceName, &v),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
// 					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
// 				),
// 			},
// 		},
// 	})
// }

func testAccCheckVPCLinkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_vpc_link" {
				continue
			}

			_, err := tfapigatewayv2.FindVPCLinkByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 VPC Link %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCLinkExists(ctx context.Context, n string, v *apigatewayv2.GetVpcLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindVPCLinkByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCLinkConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCLinkConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCLinkConfig_base(rName), fmt.Sprintf(`
resource "aws_apigatewayv2_vpc_link" "test" {
  name               = %[1]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id
}
`, rName))
}

// func testAccVPCLinkConfig_tags1(rName, tagKey1, tagValue1 string) string {
// 	return acctest.ConfigCompose(testAccVPCLinkConfig_base(rName), fmt.Sprintf(`
// resource "aws_apigatewayv2_vpc_link" "test" {
//   name               = %[1]q
//   security_group_ids = [aws_security_group.test.id]
//   subnet_ids         = aws_subnet.test[*].id

//   tags = {
//     %[2]q = %[3]q
//   }
// }
// `, rName, tagKey1, tagValue1))
// }

// func testAccVPCLinkConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
// 	return acctest.ConfigCompose(testAccVPCLinkConfig_base(rName), fmt.Sprintf(`
// resource "aws_apigatewayv2_vpc_link" "test" {
//   name               = %[1]q
//   security_group_ids = [aws_security_group.test.id]
//   subnet_ids         = aws_subnet.test[*].id

//   tags = {
//     %[2]q = %[3]q
//     %[4]q = %[5]q
//   }
// }
// `, rName, tagKey1, tagValue1, tagKey2, tagValue2))
// }
