// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreMemory_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var m1, m2, m3 bedrockagentcorecontrol.GetMemoryOutput
	rName := "tf_acc_test_" + sdkacctest.RandString(10)
	resourceName := "aws_bedrockagentcore_memory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccMemoryConfig(rName, "test description", 30, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, resourceName, &m1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "event_expiry_duration", "30"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "memory_execution_role_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "encryption_key_arn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`memory/.+$`)),
				),
			},
			{
				Config: testAccMemoryConfig(rName, "updated test description", 10, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, resourceName, &m2),
					testAccCheckMemoryNotRecreated(&m1, &m2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated test description"),
					resource.TestCheckResourceAttr(resourceName, "event_expiry_duration", "10"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_execution_role_arn"),
				),
			},
			{
				Config: testAccMemoryConfig(rName, "updated test description", 10, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, resourceName, &m3),
					testAccCheckMemoryRecreated(&m2, &m3),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_key_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_token"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var memory bedrockagentcorecontrol.GetMemoryOutput
	rName := "tf_acc_test_" + sdkacctest.RandString(10)
	resourceName := "aws_bedrockagentcore_memory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig(rName, "test description", 30, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, resourceName, &memory),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceMemory, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckMemoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_memory" {
				continue
			}

			_, err := tfbedrockagentcore.FindMemoryByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameMemory, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameMemory, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMemoryExists(ctx context.Context, name string, memory *bedrockagentcorecontrol.GetMemoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameMemory, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameMemory, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindMemoryByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameMemory, rs.Primary.ID, err)
		}

		*memory = *resp

		return nil
	}
}

func testAccCheckMemoryRecreated(before, after *bedrockagentcorecontrol.GetMemoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.ToString(before.Memory.Id), aws.ToString(after.Memory.Id); beforeID == afterID {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingRecreated, tfbedrockagentcore.ResNameMemory, beforeID, errors.New("not recreated"))
		}
		return nil
	}
}

func testAccCheckMemoryNotRecreated(before, after *bedrockagentcorecontrol.GetMemoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.ToString(before.Memory.Id), aws.ToString(after.Memory.Id); beforeID != afterID {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameMemory, beforeID, errors.New("recreated"))
		}

		return nil
	}
}

func testAccMemoryConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume" {
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
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonBedrockAgentCoreMemoryBedrockModelInferenceExecutionRolePolicy"
}

resource "aws_kms_key" "test" {
  description             = "Test key for %[1]s"
  deletion_window_in_days = 7
}

`, rName)
}

func testAccMemoryConfig(rName, description string, expiry int, withRole, withCmk bool) string {
	role, cmk := "aws_iam_role.test.arn", "aws_kms_key.test.arn"
	if !withRole {
		role = "null"
	}
	if !withCmk {
		cmk = "null"
	}

	return acctest.ConfigCompose(testAccMemoryConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                      = %[1]q
  description               = %[2]q
  event_expiry_duration     = %[3]d
  memory_execution_role_arn = %[4]s
  encryption_key_arn        = %[5]s
}
`, rName, description, expiry, role, cmk))
}
