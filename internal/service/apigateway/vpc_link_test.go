// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

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
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayVPCLink_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_vpc_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_basic(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCLinkExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCLinkConfig_basic(rName, "test update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCLinkExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test update"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAPIGatewayVPCLink_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_vpc_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_basic(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceVPCLink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCLinkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_vpc_link" {
				continue
			}

			_, err := tfapigateway.FindVPCLinkByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway VPC Link %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCLinkExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		_, err := tfapigateway.FindVPCLinkByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccVPCLinkConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id
}
`, rName))
}

func testAccVPCLinkConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(testAccVPCLinkConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = %[1]q
  description = %[2]q
  target_arns = [aws_lb.test.arn]
}
`, rName, description))
}
