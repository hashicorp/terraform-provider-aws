// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_agent.test"
	var v bedrockagent.GetAgentOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBedrockAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBedrockAgentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBedrockAgentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"agent_version"},
			},
		},
	})
}

func testAccCheckBedrockAgentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_agent" {
				continue
			}

			_, err := findBedrockAgentByID(ctx, conn, rs.Primary.ID)

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, "Bedrock Agent", rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, "Bedrock Agent", rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

// func testAccCheckBedrockAgentDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_bedrock_agent" {
// 				continue
// 			}

// 			_, err := conn.GetAgent(ctx, &bedrockagent.GetAgentInput{
// 				AgentId: aws.String(rs.Primary.ID),
// 			})

// 			if errs.IsA[*types.ResourceNotFoundException](err) {
// 				return nil
// 			}
// 			if err != nil {
// 				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, "Bedrock Agent", rs.Primary.ID, err)
// 			}

// 			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, "Bedrock Agent", rs.Primary.ID, errors.New("not destroyed"))
// 		}

// 		return nil
// 	}
// }

func testAccCheckBedrockAgentExists(ctx context.Context, n string, v *bedrockagent.GetAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := findBedrockAgentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// func testAccCheckBedrockAgentExists(ctx context.Context, n string, v *bedrockagent.GetAgentOutput) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[n]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", n)
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

// 		output, err := conn.GetAgent(ctx, &bedrockagent.GetAgentInput{
// 			AgentId: aws.String(rs.Primary.ID),
// 		})

// 		if err != nil {
// 			return err
// 		}

// 		*v = *output

// 		return nil
// 	}
// }

func findBedrockAgentByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetAgentOutput, error) {
	input := &bedrockagent.GetAgentInput{
		AgentId: aws.String(id),
	}

	output, err := conn.GetAgent(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func testAccBedrockAgentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_agent" "test" {
  agent_name              = %[1]q
  agent_resource_role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/service-role/AmazonBedrockExecutionRoleForAgents_ADK5I5QWVLW"
  idle_ttl                = 500
  foundation_model        = "anthropic.claude-v2"
  
}
`, rName)
}
