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
	"github.com/aws/aws-sdk-go-v2/service/m2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentActionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_agent_action_group.test"
	var v bedrockagent.GetAgentActionGroupOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBedrockAgentActionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBedrockAgentActionGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBedrockAgentActionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "action_group_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBedrockAgentActionGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_agent_action_group" {
				continue
			}

			_, err := conn.GetAgentActionGroup(ctx, &bedrockagent.GetAgentActionGroupInput{
				ActionGroupId: aws.String(rs.Primary.ID),
				AgentId:       aws.String(rs.Primary.Attributes["agent_id"]),
				AgentVersion:  aws.String(rs.Primary.Attributes["agent_version"]),
			})

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
func testAccCheckBedrockAgentActionGroupExists(ctx context.Context, n string, v *bedrockagent.GetAgentActionGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := conn.GetAgentActionGroup(ctx, &bedrockagent.GetAgentActionGroupInput{
			ActionGroupId: aws.String(rs.Primary.ID),
			AgentId:       aws.String(*v.AgentActionGroup.AgentId),
			AgentVersion:  aws.String(*v.AgentActionGroup.AgentVersion),
		})

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBedrockAgentActionGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_agent_action_group" "test" {
  action_group_name = %[1]q
  agent_id          = "TLJ2L6TKGM"
  agent_version     = "DRAFT"
  action_group_executor {
	lambda = "arn:aws:lambda:us-west-2:xxxxxxxxxxxx:function:sairam:1"
  }
  api_schema {
	s3 {
		s3_bucket_name = "tf-acc-test-3678246384762388142"
		s3_object_key  = "sai.yaml"
	}
  }
}
`, rName)
}
