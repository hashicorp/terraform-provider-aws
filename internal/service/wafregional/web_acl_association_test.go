// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalWebACLAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_web_acl_association.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLAssociationExists(ctx, resourceName),
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

func TestAccWAFRegionalWebACLAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_web_acl_association.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceWebACLAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalWebACLAssociation_multipleAssociations(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_web_acl_association.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLAssociationConfig_multiples,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLAssociationExists(ctx, resourceName),
					testAccCheckWebACLAssociationExists(ctx, "aws_wafregional_web_acl_association.bar"),
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

func TestAccWAFRegionalWebACLAssociation_ResourceARN_apiGatewayStage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_web_acl_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID)
			acctest.PreCheckAPIGatewayTypeEDGE(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLAssociationConfig_checkResourceARNAPIGatewayStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLAssociationExists(ctx, resourceName),
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

func testAccCheckWebACLAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_web_acl_association" {
				continue
			}

			_, err := tfwafregional.FindWebACLByResourceARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional WebACL Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebACLAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		_, err := tfwafregional.FindWebACLByResourceARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])

		return err
	}
}

const testAccWebACLAssociationConfig_basic = `
resource "aws_wafregional_rule" "foo" {
  name        = "foo"
  metric_name = "foo"
}

resource "aws_wafregional_web_acl" "foo" {
  name        = "foo"
  metric_name = "foo"

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "COUNT"
    }

    priority = 100
    rule_id  = aws_wafregional_rule.foo.id
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "bar" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

resource "aws_alb" "foo" {
  internal = true
  subnets  = [aws_subnet.foo.id, aws_subnet.bar.id]
}

resource "aws_wafregional_web_acl_association" "foo" {
  resource_arn = aws_alb.foo.arn
  web_acl_id   = aws_wafregional_web_acl.foo.id
}
`

const testAccWebACLAssociationConfig_multiples = testAccWebACLAssociationConfig_basic + `
resource "aws_alb" "bar" {
  internal = true
  subnets  = [aws_subnet.foo.id, aws_subnet.bar.id]
}

resource "aws_wafregional_web_acl_association" "bar" {
  resource_arn = aws_alb.bar.arn
  web_acl_id   = aws_wafregional_web_acl.foo.id
}
`

func testAccWebACLAssociationConfig_checkResourceARNAPIGatewayStage(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method" "test" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method_response" "test" {
  http_method = aws_api_gateway_method.test.http_method
  resource_id = aws_api_gateway_resource.test.id
  rest_api_id = aws_api_gateway_rest_api.test.id
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "GET"
  resource_id             = aws_api_gateway_resource.test.id
  rest_api_id             = aws_api_gateway_rest_api.test.id
  type                    = "HTTP"
  uri                     = "http://www.example.com"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration_response.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "test"
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = "test"

  default_action {
    type = "ALLOW"
  }
}

resource "aws_wafregional_web_acl_association" "test" {
  resource_arn = aws_api_gateway_stage.test.arn
  web_acl_id   = aws_wafregional_web_acl.test.id
}
`, rName)
}
