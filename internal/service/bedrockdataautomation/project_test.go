// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockdataautomation "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockdataautomation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkProjectARN                knownvalue.Check = tfknownvalue.RegionalARNRegexp("bedrock", regexache.MustCompile(`data-automation-project/.+`))
	checkProjectARNAlternateRegion knownvalue.Check = tfknownvalue.RegionalARNAlternateRegionRegexp("bedrock", regexache.MustCompile(`data-automation-project/.+`))
)

func TestAccBedrockDataAutomationProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkProjectARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_output_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_modified_time"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("project_stage"), tfknownvalue.StringExact(awstypes.DataAutomationProjectStageLive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.DataAutomationProjectStatusCompleted)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.DataAutomationProjectTypeAsync)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccBedrockDataAutomationProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockdataautomation.ResourceProject, resourceName),
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

func TestAccBedrockDataAutomationProject_standardOutputConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"
	documentPath := tfjsonpath.New("standard_output_configuration").AtSliceIndex(0).AtMapKey("document").AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/standard_output_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:          config.StringVariable(rName),
					"granularity_types":      config.ListVariable(config.StringVariable("PAGE")),
					"generative_field_state": config.StringVariable("DISABLED"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("standard_output_configuration").AtSliceIndex(0).AtMapKey("audio"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, documentPath.AtMapKey("extraction").AtSliceIndex(0).AtMapKey("granularity").AtSliceIndex(0).AtMapKey("types"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("PAGE"),
					})),
					statecheck.ExpectKnownValue(resourceName, documentPath.AtMapKey("generative_field").AtSliceIndex(0).AtMapKey(names.AttrState), knownvalue.StringExact("DISABLED")),
					statecheck.ExpectKnownValue(resourceName, documentPath.AtMapKey("output_format"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("standard_output_configuration").AtSliceIndex(0).AtMapKey("video"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/standard_output_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:          config.StringVariable(rName),
					"granularity_types":      config.ListVariable(config.StringVariable("PAGE")),
					"generative_field_state": config.StringVariable("DISABLED"),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"override_configuration", "standard_output_configuration"},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/standard_output_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:          config.StringVariable(rName),
					"granularity_types":      config.ListVariable(config.StringVariable("PAGE"), config.StringVariable("WORD")),
					"generative_field_state": config.StringVariable("ENABLED"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, documentPath.AtMapKey("extraction").AtSliceIndex(0).AtMapKey("granularity").AtSliceIndex(0).AtMapKey("types"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("PAGE"),
						knownvalue.StringExact("WORD"),
					})),
					statecheck.ExpectKnownValue(resourceName, documentPath.AtMapKey("generative_field").AtSliceIndex(0).AtMapKey(names.AttrState), knownvalue.StringExact("ENABLED")),
				},
			},
		},
	})
}

func TestAccBedrockDataAutomationProject_overrideConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"
	overridePath := tfjsonpath.New("override_configuration").AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/override_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"splitter_state": config.StringVariable("ENABLED"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, overridePath.AtMapKey("audio"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, overridePath.AtMapKey("document").AtSliceIndex(0).AtMapKey("modality_processing"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, overridePath.AtMapKey("document").AtSliceIndex(0).AtMapKey("splitter").AtSliceIndex(0).AtMapKey(names.AttrState), knownvalue.StringExact("ENABLED")),
					statecheck.ExpectKnownValue(resourceName, overridePath.AtMapKey("modality_routing").AtSliceIndex(0).AtMapKey("jpeg"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, overridePath.AtMapKey("modality_routing").AtSliceIndex(0).AtMapKey("png"), knownvalue.StringExact("DOCUMENT")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/override_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"splitter_state": config.StringVariable("DISABLED"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, overridePath.AtMapKey("document").AtSliceIndex(0).AtMapKey("splitter").AtSliceIndex(0).AtMapKey(names.AttrState), knownvalue.StringExact("DISABLED")),
				},
			},
		},
	})
}

func TestAccBedrockDataAutomationProject_customOutputConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"
	blueprintPath := tfjsonpath.New("custom_output_configuration").AtSliceIndex(0).AtMapKey("blueprint")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/custom_output_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"description":     config.StringVariable("first"),
					"blueprint_index": config.IntegerVariable(0),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, blueprintPath, knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, blueprintPath.AtSliceIndex(0).AtMapKey("blueprint_arn"), "aws_bedrockdataautomation_blueprint.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("first")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/custom_output_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"description":     config.StringVariable("first"),
					"blueprint_index": config.IntegerVariable(0),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/custom_output_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"description":     config.StringVariable("second"),
					"blueprint_index": config.IntegerVariable(1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, blueprintPath.AtSliceIndex(0).AtMapKey("blueprint_arn"), "aws_bedrockdataautomation_blueprint.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("second")),
				},
			},
		},
	})
}

func TestAccBedrockDataAutomationProject_dataAutomationLibraryConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"
	libraryPath := tfjsonpath.New("data_automation_library_configuration").AtSliceIndex(0).AtMapKey("library")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/data_automation_library_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"library_index": config.IntegerVariable(0),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, libraryPath, knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, libraryPath.AtSliceIndex(0).AtMapKey("library_arn"), "aws_bedrockdataautomation_library.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/data_automation_library_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"library_index": config.IntegerVariable(0),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/data_automation_library_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"library_index": config.IntegerVariable(1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, libraryPath.AtSliceIndex(0).AtMapKey("library_arn"), "aws_bedrockdataautomation_library.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccBedrockDataAutomationProject_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationProject
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_project.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProjectExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0).AtMapKey(names.AttrKMSKeyID), "aws_kms_key.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0).AtMapKey("kms_encryption_context"), knownvalue.MapExact(map[string]knownvalue.Check{
						"Purpose": knownvalue.StringExact("test"),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Project/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckProjectExists(ctx context.Context, t *testing.T, n string, v *awstypes.DataAutomationProject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).BedrockDataAutomationClient(ctx)

		output, err := tfbedrockdataautomation.FindProjectByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckProjectDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockdataautomation_project" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).BedrockDataAutomationClient(ctx)
			_, err := tfbedrockdataautomation.FindProjectByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Data Automation Project %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}
