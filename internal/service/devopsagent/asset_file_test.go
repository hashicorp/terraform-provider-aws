// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assetFile devopsagent.GetAssetFileOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName, &assetFile),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrPair(resourceName, "asset_id", "aws_devopsagent_asset.test", "asset_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "README.md"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Hello\n\nThis is a test file."),
					resource.TestCheckResourceAttrSet(resourceName, "asset_version"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrPath,
				ImportStateIdFunc:                    testAccAssetFileImportStateIDFunc(resourceName),
				// content_body: GetAssetFile returns content but text encoding can differ slightly.
				ImportStateVerifyIgnore: []string{"content_body"},
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assetFile devopsagent.GetAssetFileOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName, &assetFile),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourceAssetFile, resourceName),
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

func TestAccDevOpsAgentAssetFile_updateContent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assetFile devopsagent.GetAssetFileOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset_file.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_content(rName, "# Original\\n\\nFirst version."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName, &assetFile),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrPair(resourceName, "asset_id", "aws_devopsagent_asset.test", "asset_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "docs/guide.md"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Original\n\nFirst version."),
					resource.TestCheckResourceAttrSet(resourceName, "asset_version"),
				),
			},
			{
				Config: testAccAssetFileConfig_content(rName, "# Updated\\n\\nSecond version."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName, &assetFile),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrPair(resourceName, "asset_id", "aws_devopsagent_asset.test", "asset_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "docs/guide.md"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Updated\n\nSecond version."),
					resource.TestCheckResourceAttrSet(resourceName, "asset_version"),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_multipleFiles(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assetFile1, assetFile2 devopsagent.GetAssetFileOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_devopsagent_asset_file.readme"
	resourceName2 := "aws_devopsagent_asset_file.config"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetFileConfig_multipleFiles(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetFileExists(ctx, t, resourceName1, &assetFile1),
					testAccCheckAssetFileExists(ctx, t, resourceName2, &assetFile2),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPath, "README.md"),
					resource.TestCheckResourceAttr(resourceName1, "content_body", "# My Skill"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPath, "config.yaml"),
					resource.TestCheckResourceAttr(resourceName2, "content_body", "version: 1"),
				),
			},
		},
	})
}

// Helper functions

func testAccCheckAssetFileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsagent_asset_file" {
				continue
			}

			_, err := tfdevopsagent.FindAssetFileByPath(ctx, conn, rs.Primary.Attributes["agent_space_id"], rs.Primary.Attributes["asset_id"], rs.Primary.Attributes[names.AttrPath])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNameAssetFile, rs.Primary.Attributes[names.AttrPath], err)
			}

			return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNameAssetFile, rs.Primary.Attributes[names.AttrPath], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssetFileExists(ctx context.Context, t *testing.T, name string, assetFile *devopsagent.GetAssetFileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAssetFile, name, errors.New("not found"))
		}

		agentSpaceID := rs.Primary.Attributes["agent_space_id"]
		assetID := rs.Primary.Attributes["asset_id"]
		path := rs.Primary.Attributes[names.AttrPath]
		if agentSpaceID == "" || assetID == "" || path == "" {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAssetFile, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		resp, err := tfdevopsagent.FindAssetFileByPath(ctx, conn, agentSpaceID, assetID, path)
		if err != nil {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAssetFile, path, err)
		}

		*assetFile = *resp

		return nil
	}
}

func testAccAssetFileImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return rs.Primary.Attributes["agent_space_id"] + "," + rs.Primary.Attributes["asset_id"] + "," + rs.Primary.Attributes[names.AttrPath], nil
	}
}

// Config functions

func testAccAssetFileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# Test Skill\n\nThis is a test skill."

  metadata = jsonencode({
    name        = %[1]q
    description = "A test skill"
    agent_types = ["GENERIC"]
  })
}

resource "aws_devopsagent_asset_file" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "README.md"
  content_body   = "# Hello\n\nThis is a test file."
}
`, rName)
}

func testAccAssetFileConfig_content(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# Test Skill\n\nThis is a test skill."

  metadata = jsonencode({
    name        = %[1]q
    description = "A test skill"
    agent_types = ["GENERIC"]
  })
}

resource "aws_devopsagent_asset_file" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "docs/guide.md"
  content_body   = "%[2]s"
}
`, rName, content)
}

func testAccAssetFileConfig_multipleFiles(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# Test Skill\n\nThis is a test skill."

  metadata = jsonencode({
    name        = %[1]q
    description = "A test skill"
    agent_types = ["GENERIC"]
  })
}

resource "aws_devopsagent_asset_file" "readme" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "README.md"
  content_body   = "# My Skill"
}

resource "aws_devopsagent_asset_file" "config" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "config.yaml"
  content_body   = "version: 1"

  depends_on = [aws_devopsagent_asset_file.readme]
}
`, rName)
}
