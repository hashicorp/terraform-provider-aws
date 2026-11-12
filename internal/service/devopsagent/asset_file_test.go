// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

import (
	"context"
	"encoding/base64"
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

func TestAccDevOpsAgentAssetFile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_text(rName, "Check the runbook first."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrPair(resourceName, "asset_id", "aws_devopsagent_asset.test", "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content.0.text", "Check the runbook first."),
					resource.TestCheckNoResourceAttr(resourceName, "content.0.bytes"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckNoResourceAttr(resourceName, "metadata"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "references/runbook.md"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAssetFileImportStateID(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrPath,
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_text(rName, "Check the runbook first."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourceAssetFile, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate)},
				},
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_Disappears_asset(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_text(rName, "Check the runbook first."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourceAsset, "aws_devopsagent_asset.test"),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate)},
				},
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_text(rName, "Check the runbook first."),
				Check:  testAccCheckAssetFileExists(ctx, t, resourceName),
			},
			{
				Config: testAccAssetFileConfig_text(rName, "Check the updated runbook first."),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "content.0.text", "Check the updated runbook first."),
				),
			},
			{
				Config: testAccAssetFileConfig_bytes(rName, "Check the updated runbook first."),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "content.0.bytes", base64.StdEncoding.EncodeToString([]byte("Check the updated runbook first."))),
					resource.TestCheckNoResourceAttr(resourceName, "content.0.text"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAssetFileImportStateID(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrPath,
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_metadata(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_metadata(rName, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					acctest.CheckResourceAttrJMES(resourceName, "metadata", acctest.CtKey1, acctest.CtValue1),
					acctest.CheckResourceAttrJMES(resourceName, "metadata_actual", acctest.CtKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccAssetFileConfig_metadata(rName, acctest.CtValue2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName),
					acctest.CheckResourceAttrJMES(resourceName, "metadata", acctest.CtKey1, acctest.CtValue2),
					acctest.CheckResourceAttrJMES(resourceName, "metadata_actual", acctest.CtKey1, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAssetFileExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAssetFile, name, errors.New("not found"))
		}
		agentSpaceID, assetID, path := rs.Primary.Attributes["agent_space_id"], rs.Primary.Attributes["asset_id"], rs.Primary.Attributes[names.AttrPath]
		if agentSpaceID == "" || assetID == "" || path == "" {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAssetFile, name, errors.New("identity not set"))
		}
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)
		_, err := tfdevopsagent.FindAssetFileByThreePartKey(ctx, conn, agentSpaceID, assetID, path)
		if err != nil {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAssetFile, path, err)
		}
		return nil
	}
}

func testAccCheckAssetFileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsagent_asset_file" {
				continue
			}
			path := rs.Primary.Attributes[names.AttrPath]
			_, err := tfdevopsagent.FindAssetFileByThreePartKey(ctx, conn, rs.Primary.Attributes["agent_space_id"], rs.Primary.Attributes["asset_id"], path)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNameAssetFile, path, err)
			}
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNameAssetFile, path, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccAssetFileImportStateID(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "agent_space_id", "asset_id", names.AttrPath)
}

func testAccAssetFileConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  metadata = jsonencode({
    name        = %[1]q
    description = "Terraform acceptance test skill."
    agent_types = ["GENERIC"]
  })

  content {
    file {
      path = "SKILL.md"
      body {
        text = "Follow the operational runbook."
      }
    }
  }
}
`, rName)
}

func testAccAssetFileConfig_text(rName, text string) string {
	return acctest.ConfigCompose(testAccAssetFileConfig_base(rName), fmt.Sprintf(`
resource "aws_devopsagent_asset_file" "test" {
  agent_space_id = aws_devopsagent_asset.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "references/runbook.md"

  content {
    text = %[1]q
  }
}
`, text))
}

func testAccAssetFileConfig_bytes(rName, text string) string {
	return acctest.ConfigCompose(testAccAssetFileConfig_base(rName), fmt.Sprintf(`
resource "aws_devopsagent_asset_file" "test" {
  agent_space_id = aws_devopsagent_asset.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "references/runbook.md"

  content {
    bytes = base64encode(%[1]q)
  }
}
`, text))
}

func testAccAssetFileConfig_metadata(rName, value string) string {
	return acctest.ConfigCompose(testAccAssetFileConfig_base(rName), fmt.Sprintf(`
resource "aws_devopsagent_asset_file" "test" {
  agent_space_id = aws_devopsagent_asset.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "references/runbook.md"
  metadata       = jsonencode({ key1 = %[1]q })

  content {
    text = "Check the runbook first."
  }
}
`, value))
}
