// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseRoutingRuleARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		input               string
		expectErr           bool
		expectDomainName    string
		expectRoutingRuleID string
	}{
		{
			name:      "empty ARN",
			expectErr: true,
		},
		{
			name:      "unparsable ARN",
			input:     "test",
			expectErr: true,
		},
		{
			name:      "invalid ARN resource parts 1",
			input:     "arn:aws:apigateway:us-west-2:123456789012:/apis/test", //lintignore:AWSAT003,AWSAT005
			expectErr: true,
		},
		{
			name:      "invalid ARN resource parts 2",
			input:     "arn:aws:apigateway:us-west-2:123456789012:/domainnames/example.com", //lintignore:AWSAT003,AWSAT005
			expectErr: true,
		},
		{
			name:      "invalid ARN resource parts 3",
			input:     "arn:aws:apigateway:us-west-2:123456789012:/domainnames/example.com/test", //lintignore:AWSAT003,AWSAT005
			expectErr: true,
		},
		{
			name:                "valid ARN",
			input:               "arn:aws:apigateway:us-west-2:123456789012:/domainnames/example.com/routingrules/test", //lintignore:AWSAT003,AWSAT005
			expectDomainName:    "example.com",
			expectRoutingRuleID: "test",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotDomainName, gotRoutingRuleID, err := tfapigatewayv2.ParseRoutingRuleARN(testCase.input)
			if got, want := (err != nil), testCase.expectErr; got != want {
				t.Errorf("ParseRoutingRuleARN(%q) error = %v, want error presence = %t", testCase.input, err, want)
			}
			if err != nil {
				return
			}
			if got, want := gotDomainName, testCase.expectDomainName; got != want {
				t.Errorf("ParseRoutingRuleARN(%q) DomainName = %v, want = %v", testCase.input, got, want)
			}
			if got, want := gotRoutingRuleID, testCase.expectRoutingRuleID; got != want {
				t.Errorf("ParseRoutingRuleARN(%q) RoutingRuleID = %v, want = %v", testCase.input, got, want)
			}
		})
	}
}

func TestAccAPIGatewayV2RoutingRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
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
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "routing_rule_arn", "apigateway", regexache.MustCompile(`/domainnames/.+/routingrules/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "routing_rule_arn"),
				ImportStateVerifyIdentifierAttribute: "routing_rule_arn",
			},
		},
	})
}

func TestAccAPIGatewayV2RoutingRule_v2Domain(t *testing.T) {
	ctx := acctest.Context(t)
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
				Config: testAccRoutingRuleConfig_basicV2(rName, certificate, key, 1, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingRuleExists(ctx, resourceName, &routingrule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "routing_rule_arn", "apigateway", regexache.MustCompile(`/domainnames/.+/routingrules/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "routing_rule_arn"),
				ImportStateVerifyIdentifierAttribute: "routing_rule_arn",
			},
		},
	})
}

func TestAccAPIGatewayV2RoutingRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
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
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfapigatewayv2.ResourceRoutingRule, resourceName),
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

			_, err := tfapigatewayv2.FindRoutingRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["routing_rule_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Routing Rule %s still exists", rs.Primary.Attributes["routing_rule_id"])
		}

		return nil
	}
}

func testAccCheckRoutingRuleExists(ctx context.Context, n string, v *apigatewayv2.GetRoutingRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindRoutingRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["routing_rule_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRoutingRuleConfig_basicV2(rName, certificate, key string, count, index int) string {
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

  condition {
    match_headers {
      any_of {
        header     = "X-Example-Header"
        value_glob = "example-*"
      }
    }
  }

  condition {
    match_headers {
      any_of {
        header     = "X-Example-Header2"
        value_glob = "example-*"
      }
    }
  }

  condition {
    match_base_paths {
      any_of = ["example-path"]
    }
  }

  action {
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

func testAccRoutingRuleConfig_basic(rName, certificate, key string, count, index int) string {
	return acctest.ConfigCompose(
		testAccDomainNameImportedCertsConfig(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  regional_certificate_arn = aws_acm_certificate.test[%[2]d].arn
  endpoint_configuration {
    types = ["REGIONAL"]
  }

  security_policy = "TLS_1_2"

  routing_mode = "ROUTING_RULE_THEN_BASE_PATH_MAPPING"
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
  domain_name = aws_api_gateway_domain_name.test.domain_name

  condition {
    match_headers {
      any_of {
        header     = "X-Example-Header"
        value_glob = "example-*"
      }
    }
  }

  condition {
    match_headers {
      any_of {
        header     = "X-Example-Header2"
        value_glob = "example-*"
      }
    }
  }

  condition {
    match_base_paths {
      any_of = ["example-path"]
    }
  }

  action {
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
