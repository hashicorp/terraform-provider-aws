// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func TestAccBedrockAgentAgentKnowledgeBaseAssociation_basic(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentknowledgebaseassociation types.AgentKnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_knowledge_base_association.test"
	agentModel := "anthropic.claude-v2"
	foundationModel := "amazon.titan-embed-text-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckAgentKnowledgeBaseAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentKnowledgeBaseAssociationConfig_basic(rName, agentModel, foundationModel, "test desc", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentKnowledgeBaseAssociationExists(ctx, resourceName, &agentknowledgebaseassociation),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test desc"),
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

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func TestAccBedrockAgentAgentKnowledgeBaseAssociation_update(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentknowledgebaseassociation types.AgentKnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_knowledge_base_association.test"
	agentModel := "anthropic.claude-v2"
	foundationModel := "amazon.titan-embed-text-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckAgentKnowledgeBaseAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentKnowledgeBaseAssociationConfig_basic(rName, agentModel, foundationModel, "test desc", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentKnowledgeBaseAssociationExists(ctx, resourceName, &agentknowledgebaseassociation),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test desc"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentKnowledgeBaseAssociationConfig_basic(rName, agentModel, foundationModel, "test desc2", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentKnowledgeBaseAssociationExists(ctx, resourceName, &agentknowledgebaseassociation),
					resource.TestCheckResourceAttr(resourceName, "knowledge_base_state", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test desc2"),
				),
			},
		},
	})
}

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func TestAccBedrockAgentAgentKnowledgeBaseAssociation_disappears(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var agentknowledgebaseassociation types.AgentKnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_agent_knowledge_base_association.test"
	agentModel := "anthropic.claude-v2"
	foundationModel := "amazon.titan-embed-text-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckAgentKnowledgeBaseAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentKnowledgeBaseAssociationConfig_basic(rName, agentModel, foundationModel, "test desc", "ENABLED"),
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

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Knowledge Base Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAgentKnowledgeBaseAssociationExists(ctx context.Context, n string, v *types.AgentKnowledgeBase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindAgentKnowledgeBaseAssociationByThreePartID(ctx, conn, rs.Primary.Attributes["agent_id"], rs.Primary.Attributes["agent_version"], rs.Primary.Attributes["knowledge_base_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAgentKnowledgeBaseAssociationConfig_basic(rName, agentModel, embeddingModel, description, state string) string {
	return acctest.ConfigCompose(
		testAccAgentConfig_basic(rName, agentModel, description),
		testAccKnowledgeBaseConfig_basicRDS(rName, embeddingModel),
		fmt.Sprintf(`
resource "aws_bedrockagent_agent_knowledge_base_association" "test" {
  agent_id             = aws_bedrockagent_agent.test.id
  description          = %[1]q
  knowledge_base_id    = aws_bedrockagent_knowledge_base.test.id
  knowledge_base_state = %[2]q
}
`, description, state),
	)
}
