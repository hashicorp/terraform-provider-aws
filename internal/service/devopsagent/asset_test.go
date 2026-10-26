// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdevopsagent "github.com/hashicorp/terraform-provider-aws/internal/service/devopsagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDevOpsAgentAsset_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetConfig_basic(rName, "GENERIC", "Follow the operational runbook."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "agents_md"),
					resource.TestCheckResourceAttr(resourceName, "content.0.file.0.path", "AGENTS.md"),
					resource.TestCheckResourceAttr(resourceName, "content.0.file.0.body.0.text", "Follow the operational runbook."),
					acctest.CheckResourceAttrJMES(resourceName, "metadata", "agent_type", "GENERIC"),
					acctest.CheckResourceAttrJMES(resourceName, "metadata_all", "agent_type", "GENERIC"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAssetImportStateID(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "asset_id",
				ImportStateVerifyIgnore:              []string{names.AttrContent, "metadata"},
			},
		},
	})
}

func TestAccDevOpsAgentAsset_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetConfig_basic(rName, "GENERIC", "Follow the operational runbook."),
				Check:  testAccCheckAssetExists(ctx, t, resourceName),
			},
			{
				Config: testAccAssetConfig_basic(rName, "CHAT", "Follow the operational runbook."),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName),
					acctest.CheckResourceAttrJMES(resourceName, "metadata_all", "agent_type", "CHAT"),
				),
			},
			{
				Config: testAccAssetConfig_basic(rName, "CHAT", "Follow the updated operational runbook."),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "content.0.file.0.body.0.text", "Follow the updated operational runbook."),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAssetImportStateID(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "asset_id",
				ImportStateVerifyIgnore:              []string{names.AttrContent, "metadata"},
			},
		},
	})
}

func TestAccDevOpsAgentAsset_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgent),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetConfig_basic(rName, "GENERIC", "Follow the operational runbook."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourceAsset, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate)},
				},
			},
		},
	})
}

func testAccCheckAssetExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, "Asset", name, errors.New("not found"))
		}
		agentSpaceID, assetID := rs.Primary.Attributes["agent_space_id"], rs.Primary.Attributes["asset_id"]
		if agentSpaceID == "" || assetID == "" {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, "Asset", name, errors.New("identity not set"))
		}
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)
		_, err := tfdevopsagent.FindAssetByTwoPartKey(ctx, conn, agentSpaceID, assetID)
		if err != nil {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, "Asset", assetID, err)
		}
		return nil
	}
}

func testAccCheckAssetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsagent_asset" {
				continue
			}
			assetID := rs.Primary.Attributes["asset_id"]
			_, err := tfdevopsagent.FindAssetByTwoPartKey(ctx, conn, rs.Primary.Attributes["agent_space_id"], assetID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, "Asset", assetID, err)
			}
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, "Asset", assetID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccAssetImportStateID(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "agent_space_id", "asset_id")
}

func testAccAssetConfig_basic(rName, agentType, text string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "agents_md"
  metadata       = jsonencode({ agent_type = %[2]q })

  content {
    file {
      path = "AGENTS.md"
      body {
        text = %[3]q
      }
    }
  }
}
`, rName, agentType, text)
}
