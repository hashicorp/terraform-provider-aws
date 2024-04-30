// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentAgentKnowledgeBaseAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentknowledgebaseassociation types.AgentKnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_knowledge_base_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentKnowledgeBaseAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentKnowledgeBaseAssociationConfig_basic(rName, "anthropic.claude-v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentKnowledgeBaseAssociationExists(ctx, resourceName, &agentknowledgebaseassociation),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "bedrockagent", regexache.MustCompile(`agentknowledgebaseassociation:+.`)),
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

func TestAccBedrockAgentAgentKnowledgeBaseAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentknowledgebaseassociation types.AgentKnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_knowledge_base_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentKnowledgeBaseAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentKnowledgeBaseAssociationConfig_basic(rName, "anthropic.claude-v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentKnowledgeBaseAssociationExists(ctx, resourceName, &agentknowledgebaseassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceAgentKnowledgeBaseAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAgentKnowledgeBaseAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_agent_knowledge_base_association" {
				continue
			}

			_, err := tfbedrockagent.FindAgentKnowledgeBaseAssociationByThreePartID(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["knowledge_base_id"])

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameAgentKnowledgeBaseAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgent, create.ErrActionCheckingDestroyed, tfbedrockagent.ResNameAgentKnowledgeBaseAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAgentKnowledgeBaseAssociationExists(ctx context.Context, name string, agentknowledgebaseassociation *types.AgentKnowledgeBase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentKnowledgeBaseAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentKnowledgeBaseAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)
		resp, err := tfbedrockagent.FindAgentKnowledgeBaseAssociationByThreePartID(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["knowledge_base_id"])

		if err != nil {
			return create.Error(names.BedrockAgent, create.ErrActionCheckingExistence, tfbedrockagent.ResNameAgentKnowledgeBaseAssociation, rs.Primary.ID, err)
		}

		*agentknowledgebaseassociation = *resp

		return nil
	}
}

/*
	func testAccCheckAgentKnowledgeBaseAssociationNotRecreated(before, after *bedrockagent.GetAgentKnowledgeBaseOutput) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			if before, after := aws.ToString(before.I), aws.ToString(after.AgentKnowledgeBaseAssociationId); before != after {
				return create.Error(names.BedrockAgent, create.ErrActionCheckingNotRecreated, tfbedrockagent.ResNameAgentKnowledgeBaseAssociation, aws.ToString(before.AgentKnowledgeBaseAssociationId), errors.New("recreated"))
			}

			return nil
		}
	}
*/

func testAccAgentKnowledgeBaseAssociationConfig_basic(rName, model string) string {
	return acctest.ConfigCompose(
		testAccAgentConfig_basic(rName, model, "basic claude"),
		testAccKnowledgeBaseConfig_basicOpenSearch(rName, model),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_knowledge_base_association" "test" {
    
}
`, rName),
	)
}
