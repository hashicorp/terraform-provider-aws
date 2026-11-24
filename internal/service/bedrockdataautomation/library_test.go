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
	checkLibraryARN                knownvalue.Check = tfknownvalue.RegionalARNRegexp("bedrock", regexache.MustCompile(`data-automation-library/.+`))
	checkLibraryARNAlternateRegion knownvalue.Check = tfknownvalue.RegionalARNAlternateRegionRegexp("bedrock", regexache.MustCompile(`data-automation-library/.+`))
)

func TestAccBedrockDataAutomationLibrary_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationLibrary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_library.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLibraryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Library/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLibraryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkLibraryARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.DataAutomationLibraryStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Library/basic/"),
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

func TestAccBedrockDataAutomationLibrary_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationLibrary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_library.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLibraryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Library/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLibraryExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockdataautomation.ResourceLibrary, resourceName),
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

func TestAccBedrockDataAutomationLibrary_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationLibrary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_library.test"
	step := func(description config.Variable, action plancheck.ResourceActionType, check knownvalue.Check) resource.TestStep {
		variables := config.Variables{
			acctest.CtRName: config.StringVariable(rName),
		}
		if description != nil {
			variables[names.AttrDescription] = description
		}

		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Library/description/"),
			ConfigVariables: variables,
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckLibraryExists(ctx, t, resourceName, &v),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), check),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLibraryDestroy(ctx, t),
		Steps: []resource.TestStep{
			step(config.StringVariable("first"), plancheck.ResourceActionCreate, knownvalue.StringExact("first")),
			step(config.StringVariable("second"), plancheck.ResourceActionUpdate, knownvalue.StringExact("second")),
			step(nil, plancheck.ResourceActionDestroyBeforeCreate, knownvalue.Null()),
		},
	})
}

func TestAccBedrockDataAutomationLibrary_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DataAutomationLibrary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockdataautomation_library.test"
	kmsKeyIDPath := tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0).AtMapKey(names.AttrKMSKeyID)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockDataAutomationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLibraryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Library/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"kms_key_index": config.IntegerVariable(0),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLibraryExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, kmsKeyIDPath, "aws_kms_key.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0).AtMapKey("kms_encryption_context"), knownvalue.MapExact(map[string]knownvalue.Check{
						"Purpose": knownvalue.StringExact("test"),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Library/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"kms_key_index": config.IntegerVariable(0),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Library/encryption_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"kms_key_index": config.IntegerVariable(1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLibraryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, kmsKeyIDPath, "aws_kms_key.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccCheckLibraryExists(ctx context.Context, t *testing.T, n string, v *awstypes.DataAutomationLibrary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).BedrockDataAutomationClient(ctx)

		output, err := tfbedrockdataautomation.FindLibraryByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLibraryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockdataautomation_library" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).BedrockDataAutomationClient(ctx)
			_, err := tfbedrockdataautomation.FindLibraryByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Data Automation Library %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}
