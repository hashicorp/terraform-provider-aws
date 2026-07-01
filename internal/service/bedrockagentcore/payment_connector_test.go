// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
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

func testAccPaymentConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomPaymentConnectorName(t)
	resourceName := "aws_bedrockagentcore_payment_connector.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentConnectors(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentConnectorExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("StripePrivy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("payment_connector_id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccPaymentConnectorImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "payment_connector_id",
				// The credential provider configuration holds write-only secrets that
				// the API does not return, so the block cannot be reconstructed on import.
				ImportStateVerifyIgnore: []string{
					"credential_provider_configuration",
				},
			},
		},
	})
}

func testAccPaymentConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomPaymentConnectorName(t)
	resourceName := "aws_bedrockagentcore_payment_connector.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentConnectors(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentConnectorExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourcePaymentConnector, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPaymentConnectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_payment_connector" {
				continue
			}

			_, err := tfbedrockagentcore.FindPaymentConnectorByTwoPartKey(ctx, conn, rs.Primary.Attributes["payment_manager_id"], rs.Primary.Attributes["payment_connector_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Payment Connector %s still exists", rs.Primary.Attributes["payment_connector_id"])
		}

		return nil
	}
}

func testAccCheckPaymentConnectorExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindPaymentConnectorByTwoPartKey(ctx, conn, rs.Primary.Attributes["payment_manager_id"], rs.Primary.Attributes["payment_connector_id"])
		return err
	}
}

func testAccPreCheckPaymentConnectors(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	// Payment connectors are scoped to a payment manager; use the managers list to
	// verify service availability without an existing manager.
	input := bedrockagentcorecontrol.ListPaymentManagersInput{}

	_, err := conn.ListPaymentManagers(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPaymentConnectorConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock-agentcore:CreateWorkloadIdentity",
          "bedrock-agentcore:TagResource",
          "bedrock-agentcore:UntagResource",
          "bedrock-agentcore:ListTagsForResource",
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_bedrockagentcore_payment_manager" "test" {
  name            = %[2]q
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.test.arn
}

resource "aws_bedrockagentcore_payment_credential_provider" "test" {
  name                       = %[2]q
  credential_provider_vendor = "StripePrivy"

  provider_configuration {
    stripe_privy_configuration {
      app_id                    = "app_test_id"
      app_secret                = "sk_test_secret"
      authorization_id          = "auth_test_id"
      authorization_private_key = "MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgNlzLhFfPe/14eR6GlFuWOzYTHgfXgyKs1yHwtpFISo6hRANCAAQPrRtegKcGCGBALTzewz0OnIpa9AeOe5BpcT0OS+Ej7odZ7fsTN8YgZzq5kBAY3u2UcZNHn6YJC70Z4bgpiuKI"
    }
  }
}

resource "aws_bedrockagentcore_payment_connector" "test" {
  name               = %[1]q
  payment_manager_id = aws_bedrockagentcore_payment_manager.test.payment_manager_id
  type               = "StripePrivy"

  credential_provider_configuration {
    stripe_privy {
      credential_provider_arn = aws_bedrockagentcore_payment_credential_provider.test.credential_provider_arn
    }
  }
}
`, rName, strings.ReplaceAll(strings.ReplaceAll(rName, "-", ""), "_", ""))
}

func testAccPaymentConnectorImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "payment_manager_id", "payment_connector_id")
}

func testAccRandomPaymentConnectorName(t *testing.T) string {
	return strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
}
