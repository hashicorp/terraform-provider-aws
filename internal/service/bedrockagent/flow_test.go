// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
)

func TestAccBedrockAgentFlow_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "description", "basic"),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "customer_encryption_key_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBedrockAgentFlow_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceFlow, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockAgentFlow_withEncryptionKey(t *testing.T) {
	ctx := acctest.Context(t)

	var flow bedrockagent.GetFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_flow.test"
	foundationModel := "amazon.titan-text-express-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_withEncryptionKey(rName, foundationModel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`flow/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "customer_encryption_key_arn", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}
func testAccCheckFlowDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_flow" {
				continue
			}

			_, err := tfbedrockagent.FindFlowByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameFlow, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameFlow, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFlowExists(ctx context.Context, name string, flow *bedrockagent.GetFlowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameFlow, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameFlow, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		resp, err := tfbedrockagent.FindFlowByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameFlow, rs.Primary.ID, err)
		}

		*flow = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

	input := &bedrockagent.ListFlowsInput{}

	_, err := conn.ListFlows(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFlowConfig_base(model string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_tf"
  assume_role_policy = data.aws_iam_policy_document.test_trust.json
}

data "aws_iam_policy_document" "test_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role_policy" "test_permissions" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test_permissions.json
}

data "aws_iam_policy_document" "test_permissions" {
  statement {
    actions   = ["bedrock:GetFlow"]
    effect    = "Allow"
    resources = ["*"]
  }
  statement {
    actions   = ["bedrock:InvokeModel"]
    effect    = "Allow"
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[1]s",
    ]
  }
}`, model)
}

func testAccFlowConfig_basic(rName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_bedrockagent_flow" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn
  description        = "basic"
}
`, rName))
}

func testAccFlowConfig_withEncryptionKey(rName, model string) string {
	return acctest.ConfigCompose(testAccFlowConfig_base(model), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrockagent_flow" "test" {
  name                        = %[1]q
  execution_role_arn          = aws_iam_role.test.arn
  customer_encryption_key_arn = aws_kms_key.test.arn
}
`, rName))
}
