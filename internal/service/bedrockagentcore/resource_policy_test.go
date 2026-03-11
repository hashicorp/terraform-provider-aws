// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreResourcePolicy_runtime_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcepolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_runtime(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccBedrockAgentCoreResourcePolicy_endpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcepolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_endpoint(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccBedrockAgentCoreResourcePolicy_gateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcepolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_gateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcepolicy),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_resource_policy" {
				continue
			}

			// Call GetResourcePolicy directly to verify destruction
			input := bedrockagentcorecontrol.GetResourcePolicyInput{
				ResourceArn: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetResourcePolicy(ctx, &input)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameResourcePolicy, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameResourcePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, name string, resourcepolicy *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameResourcePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameResourcePolicy, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		input := bedrockagentcorecontrol.GetResourcePolicyInput{
			ResourceArn: aws.String(rs.Primary.ID),
		}

		out, err := conn.GetResourcePolicy(ctx, &input)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameResourcePolicy, rs.Primary.ID, err)
		}

		if out.Policy != nil {
			*resourcepolicy = aws.ToString(out.Policy)
		} else {
			*resourcepolicy = ""
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	// Use ListAgentRuntimes as a lightweight service availability check
	input := bedrockagentcorecontrol.ListAgentRuntimesInput{}

	_, err := conn.ListAgentRuntimes(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
func testAccResourcePolicyConfig_runtime(rName string) string {
	policy := `{"Version":"2012-10-17","Statement":[]}`

	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_bedrockagentcore_agent_runtime" "example" {
  name = "%s-runtime"
}

resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_agent_runtime.example.agent_runtime_arn
  policy       = %q
}
`, rName, policy)
}

func testAccResourcePolicyConfig_endpoint(rName string) string {
	policy := `{"Version":"2012-10-17","Statement":[]}`

	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_bedrockagentcore_agent_runtime" "example" {
  name = "%s-runtime"
}

resource "aws_bedrockagentcore_agent_runtime_endpoint" "example" {
  agent_runtime_id = aws_bedrockagentcore_agent_runtime.example.agent_runtime_id
  name             = "%s-endpoint"
}

resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_agent_runtime_endpoint.example.agent_runtime_endpoint_arn
  policy       = %q
}
`, rName, rName, policy)
}

func testAccResourcePolicyConfig_gateway(rName string) string {
	policy := `{"Version":"2012-10-17","Statement":[]}`

	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_bedrockagentcore_gateway" "example" {
  name = "%s-gateway"
}

resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_gateway.example.gateway_arn
  policy       = %q
}
`, rName, policy)
}
