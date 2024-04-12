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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_agent_alias.test"
	var v bedrockagent.GetAgentAliasOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentAliasConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentAliasExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_alias_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"agent_id"},
			},
		},
	})
}

func testAccCheckAgentAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_agent_alias" {
				continue
			}

			// _, err := conn.GetAgentAlias(ctx, &bedrockagent.GetAgentAliasInput{
			// 	AgentAliasId: aws.String(rs.Primary.ID),
			// 	AgentId:      aws.String(rs.Primary.Attributes["agent_id"]),
			// })
			_, err := findAgentAliasByID(ctx, conn, rs.Primary.ID)

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

func testAccCheckAgentAliasExists(ctx context.Context, n string, v *bedrockagent.GetAgentAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		// output, err := conn.GetAgentAlias(ctx, &bedrockagent.GetAgentAliasInput{
		// 	AgentAliasId: aws.String(rs.Primary.ID),
		// 	AgentId:      aws.String(rs.Primary.Attributes["agent_id"]),
		// })
		output, err := findAgentAliasByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func findAgentAliasByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetAgentAliasOutput, error) {
	parts, err := intflex.ExpandResourceId(id, 2, false)
	if err != nil {
		return nil, err
	}
	input := &bedrockagent.GetAgentAliasInput{
		AgentAliasId: aws.String(parts[0]),
		AgentId:      aws.String(parts[1]),
	}

	output, err := conn.GetAgentAlias(ctx, input)

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

func testAccAgentAliasConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAgentConfig_basic(rName, "anthropic.claude-v2", "basic claude"), fmt.Sprintf(`
resource "aws_bedrock_agent_alias" "test" {
  agent_alias_name = %[1]q
  agent_id         = aws_bedrockagent_agent.test.agent_id
}
`, rName))
}
