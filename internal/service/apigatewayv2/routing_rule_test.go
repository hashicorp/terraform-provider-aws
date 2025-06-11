// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2RoutingRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var routingrule apigatewayv2.GetRoutingRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)
	resourceName := "aws_apigatewayv2_routing_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.APIGatewayV2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoutingRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingRuleConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingRuleExists(ctx, resourceName, &routingrule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/domainnames/.+/routingrules/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccAPIGatewayV2RoutingRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var routingrule apigatewayv2.GetRoutingRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apigatewayv2_routing_rule.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.APIGatewayV2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoutingRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingRuleConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingRuleExists(ctx, resourceName, &routingrule),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceRoutingRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoutingRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_routing_rule" {
				continue
			}

			_, err := tfapigatewayv2.FindRoutingRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.APIGatewayV2, create.ErrActionCheckingDestroyed, tfapigatewayv2.ResNameRoutingRule, rs.Primary.Attributes[names.AttrARN], err)
			}

			return create.Error(names.APIGatewayV2, create.ErrActionCheckingDestroyed, tfapigatewayv2.ResNameRoutingRule, rs.Primary.Attributes[names.AttrARN], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRoutingRuleExists(ctx context.Context, name string, routingrule *apigatewayv2.GetRoutingRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.APIGatewayV2, create.ErrActionCheckingExistence, tfapigatewayv2.ResNameRoutingRule, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.APIGatewayV2, create.ErrActionCheckingExistence, tfapigatewayv2.ResNameRoutingRule, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		resp, err := tfapigatewayv2.FindRoutingRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.APIGatewayV2, create.ErrActionCheckingExistence, tfapigatewayv2.ResNameRoutingRule, rs.Primary.Attributes[names.AttrARN], err)
		}

		*routingrule = *resp

		return nil
	}
}

func testAccRoutingRuleConfig_basic(rName, certificate, key string, count, index int) string {
	return acctest.ConfigCompose(
		testAccDomainNameImportedCertsConfig(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  routing_mode = "ROUTING_RULE_THEN_API_MAPPING"
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "http://example.com"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id
}

resource "aws_apigatewayv2_routing_rule" "test" {
  domain_name = aws_apigatewayv2_domain_name.test.domain_name

  conditions {
    match_headers {
      any_of {
        header     = "X-Example-Header"
        value_glob = "example-*"
      }
    }
  }
  conditions {
	match_headers {
      any_of {
        header     = "X-Example-Header2"
        value_glob = "example-*"
      }
    }
  }
  conditions {
    match_base_paths {
      any_of = ["example-path"]
    }
  }
  actions {
    invoke_api {
      api_id          = aws_api_gateway_rest_api.test.id
      stage           = aws_api_gateway_stage.test.stage_name
      strip_base_path = true
    }
  }

  priority = 1
}
`, rName, index))
}
