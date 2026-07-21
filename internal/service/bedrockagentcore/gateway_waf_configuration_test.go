// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreGatewayWAFConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_waf_configuration.test"
	gatewayResourceName := "aws_bedrockagentcore_gateway.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayWAFConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayWAFConfigurationConfig_basic(rName, string(awstypes.WafFailureModeFailOpen)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayWAFConfigurationExists(ctx, t, gatewayResourceName, &gateway, awstypes.WafFailureModeFailOpen),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("failure_mode"), knownvalue.StringExact(string(awstypes.WafFailureModeFailOpen))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("web_acl_arn"), webACLResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_identifier"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_identifier",
			},
			{
				Config: testAccGatewayWAFConfigurationConfig_basic(rName, string(awstypes.WafFailureModeFailClose)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayWAFConfigurationExists(ctx, t, gatewayResourceName, &gateway, awstypes.WafFailureModeFailClose),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("failure_mode"), knownvalue.StringExact(string(awstypes.WafFailureModeFailClose))),
				},
			},
		},
	})
}

// testAccCheckGatewayWAFConfigurationExists verifies the gateway exists and carries the
// expected WAF failure mode.
func testAccCheckGatewayWAFConfigurationExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetGatewayOutput, expected awstypes.WafFailureMode) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayByID(ctx, conn, rs.Primary.Attributes["gateway_id"])
		if err != nil {
			return err
		}

		if resp.WafConfiguration == nil {
			return fmt.Errorf("Bedrock Agent Core Gateway %s has no WAF configuration", rs.Primary.Attributes["gateway_id"])
		}
		if resp.WafConfiguration.FailureMode != expected {
			return fmt.Errorf("expected WAF failure mode %s, got %s", expected, resp.WafConfiguration.FailureMode)
		}

		*v = *resp

		return nil
	}
}

// testAccCheckGatewayWAFConfigurationDestroy confirms the WAF configuration was cleared
// from the gateway. The gateway itself is owned by the aws_bedrockagentcore_gateway
// resource, so its own CheckDestroy validates the gateway is gone.
func testAccCheckGatewayWAFConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway_waf_configuration" {
				continue
			}

			out, err := tfbedrockagentcore.FindGatewayByID(ctx, conn, rs.Primary.Attributes["gateway_identifier"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			if out.WafConfiguration != nil {
				return fmt.Errorf("Bedrock Agent Core Gateway %s still has WAF configuration", rs.Primary.Attributes["gateway_identifier"])
			}
		}

		return nil
	}
}

func testAccGatewayWAFConfigurationConfig_basic(rName, failureMode string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "AWS_IAM"
  protocol_type   = "MCP"
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_association" "test" {
  resource_arn = aws_bedrockagentcore_gateway.test.gateway_arn
  web_acl_arn  = aws_wafv2_web_acl.test.arn
}

resource "aws_bedrockagentcore_gateway_waf_configuration" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  failure_mode       = %[2]q

  depends_on = [aws_wafv2_web_acl_association.test]
}
`, rName, failureMode))
}
