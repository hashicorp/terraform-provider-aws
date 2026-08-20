// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

import (
	"context"
	"encoding/json"
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
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Test Skill\n\nThis is a test skill."),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "asset_id",
				ImportStateIdFunc:                    testAccAssetImportStateIDFunc(resourceName),
				// content_body, content_path: GetAsset API does not return file content (retrieved via separate GetAssetContent/GetAssetFile APIs).
				// metadata: Not read back from API during refresh; treated as write-only to avoid perpetual diffs from server-added defaults (skill_type, status).
				ImportStateVerifyIgnore: []string{"content_body", "content_path", "metadata"},
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
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Test Skill\n\nThis is a test skill."),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
			{
				Config: testAccAssetConfig_metadata(rName, `["INCIDENT_TRIAGE"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Test Skill\n\nThis is a test skill."),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_updateContentBody(t *testing.T) {
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
				Config: testAccAssetConfig_contentBody(rName, "# Original Content\\n\\nFirst version."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Original Content\n\nFirst version."),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
			{
				Config: testAccAssetConfig_contentBody(rName, "# Updated Content\\n\\nSecond version."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "content_body", "# Updated Content\n\nSecond version."),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_updateFilename(t *testing.T) {
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
				Config: testAccAssetConfig_filename(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill.md"),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill from file") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
			{
				Config: testAccAssetConfig_filenameUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill_updated.md"),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill from file") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_updateZipFile(t *testing.T) {
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
				Config: testAccAssetConfig_zipFile(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill_bundle.zip"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
			{
				Config: testAccAssetConfig_zipFileUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill_bundle_updated.zip"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_attachment(t *testing.T) {
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
				Config: testAccAssetConfig_attachment(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "attachment"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/test_attachment.png"),
					resource.TestCheckResourceAttr(resourceName, "content_path", "test_attachment.png"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, "filename", "test_attachment.png") {
							return fmt.Errorf("metadata missing expected filename")
						}
						if !containsJSON(value, "extension", "png") {
							return fmt.Errorf("metadata missing expected extension")
						}
						if !containsJSON(value, names.AttrSize, nil) {
							return fmt.Errorf("metadata missing size")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_filenameTextToZip(t *testing.T) {
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
				Config: testAccAssetConfig_filename(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill.md"),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill from file") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
			{
				Config: testAccAssetConfig_filenameTextToZip(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill_bundle.zip"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_filename(t *testing.T) {
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
				Config: testAccAssetConfig_filename(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill.md"),
					resource.TestCheckResourceAttr(resourceName, "content_path", "SKILL.md"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, acctest.CtName, rName) {
							return fmt.Errorf("metadata missing expected name %q", rName)
						}
						if !containsJSON(value, names.AttrDescription, "A test skill from file") {
							return fmt.Errorf("metadata missing expected description")
						}
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDevOpsAgentAsset_zipFile(t *testing.T) {
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
				Config: testAccAssetConfig_zipFile(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAssetExists(ctx, t, resourceName, &asset),
					resource.TestCheckResourceAttrPair(resourceName, "agent_space_id", "aws_devopsagent_agent_space.test", "agent_space_id"),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttr(resourceName, "asset_type", "skill"),
					resource.TestCheckResourceAttr(resourceName, "asset_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "filename", "test-fixtures/skill_bundle.zip"),
					resource.TestCheckResourceAttrWith(resourceName, "metadata", func(value string) error {
						if !containsJSON(value, "agent_types", nil) {
							return fmt.Errorf("metadata missing agent_types")
						}
						return nil
					}),
				),
			},
		},
	})
}

// Helper functions

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

	input := devopsagent.ListAgentSpacesInput{}

	_, err := conn.ListAgentSpaces(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAssetImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return rs.Primary.Attributes["agent_space_id"] + "," + rs.Primary.Attributes["asset_id"], nil
	}
}

// containsJSON checks if a JSON string contains a key, optionally with a specific string value.
// Pass nil for expectedValue to only check key existence.
func containsJSON(jsonStr, key string, expectedValue interface{}) bool {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return false
	}
	val, exists := m[key]
	if !exists {
		return false
	}
	if expectedValue == nil {
		return true
	}
	if strVal, ok := expectedValue.(string); ok {
		actual, ok := val.(string)
		return ok && actual == strVal
	}
	return true
}

// Config functions

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

func testAccAssetConfig_contentBody(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "%[2]s"

  metadata = jsonencode({
    name        = %[1]q
    description = "A test skill"
    agent_types = ["GENERIC"]
  })
}
`, rName, content)
}

func testAccAssetConfig_filename(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  filename       = "test-fixtures/skill.md"
  content_path   = "SKILL.md"

  metadata = jsonencode({
    name        = %[1]q
    description = "A test skill from file"
    agent_types = ["GENERIC"]
  })
}
`, rName)
}

func testAccAssetConfig_filenameUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  filename       = "test-fixtures/skill_updated.md"
  content_path   = "SKILL.md"

  metadata = jsonencode({
    name        = %[1]q
    description = "A test skill from file"
    agent_types = ["GENERIC"]
  })
}
`, rName)
}

func testAccAssetConfig_zipFile(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  filename       = "test-fixtures/skill_bundle.zip"

  metadata = jsonencode({
    agent_types = ["GENERIC"]
  })
}
`, rName)
}

func testAccAssetConfig_zipFileUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  filename       = "test-fixtures/skill_bundle_updated.zip"

  metadata = jsonencode({
    agent_types = ["GENERIC"]
  })
}
`, rName)
}

func testAccAssetConfig_filenameTextToZip(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  filename       = "test-fixtures/skill_bundle.zip"

  metadata = jsonencode({
    agent_types = ["GENERIC"]
  })
}
`, rName)
}

func testAccAssetConfig_attachment(rName string) string {
	return fmt.Sprintf(`
resource "aws_devopsagent_agent_space" "test" {
  name = %[1]q
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "attachment"
  filename       = "test-fixtures/test_attachment.png"
  content_path   = "test_attachment.png"

  metadata = jsonencode({
    filename  = "test_attachment.png"
    extension = "png"
    size      = 67
  })
}
`, rName)
}
