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
	"github.com/hashicorp/terraform-provider-aws/names"

	tfdevopsagent "github.com/hashicorp/terraform-provider-aws/internal/service/devopsagent"
)

func TestAccDevOpsAgentAsset_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var asset devopsagent.GetAssetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_version"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "asset_id",
				ImportStateVerifyIgnore:              []string{"content_body", "content_path"},
			},
		},
	})
}

func TestAccDevOpsAgentAsset_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var asset devopsagent.GetAssetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourceAsset, resourceName),
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

func TestAccDevOpsAgentAsset_updateMetadata(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var asset devopsagent.GetAssetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devopsagent_asset.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetConfig_metadata(rName, `["GENERIC"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
				),
			},
			{
				Config: testAccAssetConfig_metadata(rName, `["INCIDENT_TRIAGE"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
				),
			},
		},
	})
}

func testAccCheckAssetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsagent_asset" {
				continue
			}

			_, err := tfdevopsagent.FindAssetByID(ctx, conn, rs.Primary.Attributes["agent_space_id"], rs.Primary.Attributes["asset_id"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNameAsset, rs.Primary.Attributes["asset_id"], err)
			}

			return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNameAsset, rs.Primary.Attributes["asset_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssetExists(ctx context.Context, t *testing.T, name string, asset *devopsagent.GetAssetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAsset, name, errors.New("not found"))
		}

		agentSpaceID := rs.Primary.Attributes["agent_space_id"]
		assetID := rs.Primary.Attributes["asset_id"]
		if agentSpaceID == "" || assetID == "" {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAsset, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		resp, err := tfdevopsagent.FindAssetByID(ctx, conn, agentSpaceID, assetID)
		if err != nil {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNameAsset, assetID, err)
		}

		*asset = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

	input := &devopsagent.ListAgentSpacesInput{}

	_, err := conn.ListAgentSpaces(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAssetConfig_basic(rName string) string {
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
`, rName)
}

func testAccAssetConfig_metadata(rName, agentTypes string) string {
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
    agent_types = %[2]s
  })
}
`, rName, agentTypes)
}
