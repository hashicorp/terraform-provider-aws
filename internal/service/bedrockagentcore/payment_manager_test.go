// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPaymentManager_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomPaymentManagerName(t)
	resourceName := "aws_bedrockagentcore_payment_manager.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentManagers(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentManagerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentManagerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentManagerExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorizer_type"), knownvalue.StringExact("AWS_IAM")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("payment_manager_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("payment_manager_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`payment-manager/.+`))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "payment_manager_id"),
				ImportStateVerifyIdentifierAttribute: "payment_manager_id",
			},
		},
	})
}

func testAccPaymentManager_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomPaymentManagerName(t)
	resourceName := "aws_bedrockagentcore_payment_manager.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentManagers(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentManagerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentManagerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentManagerExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourcePaymentManager, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccPaymentManager_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomPaymentManagerName(t)
	resourceName := "aws_bedrockagentcore_payment_manager.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentManagers(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentManagerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentManagerConfig_description(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentManagerExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial description")),
				},
			},
			{
				Config: testAccPaymentManagerConfig_description(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentManagerExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated description")),
				},
			},
		},
	})
}

func testAccCheckPaymentManagerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_payment_manager" {
				continue
			}

			_, err := tfbedrockagentcore.FindPaymentManagerByID(ctx, conn, rs.Primary.Attributes["payment_manager_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Payment Manager %s still exists", rs.Primary.Attributes["payment_manager_id"])
		}

		return nil
	}
}

func testAccCheckPaymentManagerExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindPaymentManagerByID(ctx, conn, rs.Primary.Attributes["payment_manager_id"])
		return err
	}
}

func testAccPreCheckPaymentManagers(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)
	input := bedrockagentcorecontrol.ListPaymentManagersInput{}

	_, err := conn.ListPaymentManagers(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPaymentManagerConfig_base(rName string) string {
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
        Effect   = "Allow"
        Action   = ["bedrock-agentcore:CreateWorkloadIdentity"]
        Resource = "*"
      }
    ]
  })
}
`, rName)
}

func testAccPaymentManagerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPaymentManagerConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_payment_manager" "test" {
  name            = %[1]q
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.test.arn
}
`, rName))
}

func testAccPaymentManagerConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccPaymentManagerConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_payment_manager" "test" {
  name            = %[1]q
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.test.arn
  description     = %[2]q
}
`, rName, description))
}

func testAccRandomPaymentManagerName(t *testing.T) string {
	return strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "")
}
