// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestMemoryNamespaceKeysAutoFlex(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	model := fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfbedrockagentcore.NamespaceKeyModel{
		{
			Key: types.StringValue("environment"),
			Validation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.NamespaceKeyValidationModel{
				AllowedValues: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"development", "production"}),
				RegexPattern:  types.StringValue("^[a-z]+$"),
			}),
		},
		{
			Key:        types.StringValue("tenant"),
			Validation: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.NamespaceKeyValidationModel](ctx),
		},
	})
	apiObjects := []awstypes.NamespaceKeyEntry{
		{
			Key: aws.String("environment"),
			Validation: &awstypes.NamespaceKeyValidation{
				AllowedValues: []string{"development", "production"},
				RegexPattern:  aws.String("^[a-z]+$"),
			},
		},
		{
			Key: aws.String("tenant"),
		},
	}

	var expanded []awstypes.NamespaceKeyEntry
	if diags := fwflex.Expand(ctx, model, &expanded); diags.HasError() {
		t.Fatalf("unexpected Expand error: %v", diags)
	}
	if diff := cmp.Diff(apiObjects, expanded, cmpopts.IgnoreUnexported(awstypes.NamespaceKeyEntry{}, awstypes.NamespaceKeyValidation{}), cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
		t.Errorf("unexpected Expand diff (+got, -want): %s", diff)
	}

	flattened := fwtypes.NewSetNestedObjectValueOfNull[tfbedrockagentcore.NamespaceKeyModel](ctx)
	if diags := fwflex.Flatten(ctx, apiObjects, &flattened); diags.HasError() {
		t.Fatalf("unexpected Flatten error: %v", diags)
	}
	if !flattened.Equal(model) {
		t.Errorf("unexpected Flatten result: got %s, want %s", flattened, model)
	}
}

func TestAccBedrockAgentCoreMemory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "bedrock-agentcore", "memory/{id}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+`-[a-zA-Z0-9]{10}$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceMemory, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_description(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_description(rName, "desc1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc1")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMemoryConfig_description(rName, "desc2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc2")),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_memoryExecutionRole(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_indexedKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Memory/indexed_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"keys":          acctest.ListOfStringsVariable("customer_id", "score"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("indexed_key"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("customer_id"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeString),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("score"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeNumber),
						}),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Memory/indexed_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"keys":          acctest.ListOfStringsVariable("customer_id", "score"),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Reordering indexed_key entries is a no-op: it is modeled as a set, so config order is irrelevant.
				ConfigDirectory: config.StaticDirectory("testdata/Memory/indexed_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"keys":          acctest.ListOfStringsVariable("score", "customer_id"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("indexed_key"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				// Adding an indexed key is applied in place via UpdateMemory (AddIndexedKeys), not a replacement.
				ConfigDirectory: config.StaticDirectory("testdata/Memory/indexed_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"keys":          acctest.ListOfStringsVariable("customer_id", "channel", "score"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("indexed_key"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("customer_id"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeString),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("channel"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeString),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("score"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeNumber),
						}),
					})),
				},
			},
			{
				// Removing an indexed key forces replacement: the API cannot remove previously indexed keys.
				ConfigDirectory: config.StaticDirectory("testdata/Memory/indexed_key/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"keys":          acctest.ListOfStringsVariable("channel", "score"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("indexed_key"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("channel"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeString),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:  knownvalue.StringExact("score"),
							names.AttrType: tfknownvalue.StringExact(awstypes.MetadataValueTypeNumber),
						}),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_namespaceKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_namespaceKeys(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_key"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("tenant"),
							"validation": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"allowed_values": knownvalue.Null(),
									"regex_pattern":  knownvalue.StringExact("^[a-z0-9-]+$"),
								}),
							}),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("environment"),
							"validation": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"allowed_values": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("development"),
										knownvalue.StringExact("production"),
									}),
									"regex_pattern": knownvalue.Null(),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Changing a key's validation and adding a key without validation is an in-place update.
				Config: testAccMemoryConfig_namespaceKeysUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_key"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("environment"),
							"validation": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"allowed_values": knownvalue.SetExact([]knownvalue.Check{
										knownvalue.StringExact("development"),
										knownvalue.StringExact("production"),
										knownvalue.StringExact("staging"),
									}),
									"regex_pattern": knownvalue.StringExact("^[a-z]+$"),
								}),
							}),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("region1"),
							"validation":  knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Removing every namespace_key block clears the keys in place.
				Config: testAccMemoryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_key"), knownvalue.SetSizeExact(0)),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_namespaceKeysValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccMemoryConfig_namespaceKey(rName, "tenantId", `regex_pattern = "^[a-z]+$"`),
				ExpectError: regexache.MustCompile(`Invalid Attribute Value Match`),
			},
			{
				Config:      testAccMemoryConfig_namespaceKey(rName, "tenant", `allowed_values = ["Production"]`),
				ExpectError: regexache.MustCompile(`Invalid Attribute Value Match`),
			},
			{
				Config:      testAccMemoryConfig_namespaceKey(rName, "tenant", ""),
				ExpectError: regexache.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccBedrockAgentCoreMemory_streamDeliveryResources(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.Memory
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_memory.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryConfig_streamDeliveryResources(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("stream_delivery_resources").AtSliceIndex(0).AtMapKey("resource").AtSliceIndex(0).AtMapKey("kinesis").AtSliceIndex(0).AtMapKey("data_stream_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("stream_delivery_resources").AtSliceIndex(0).AtMapKey("resource").AtSliceIndex(0).AtMapKey("kinesis").AtSliceIndex(0).AtMapKey("content_configuration").AtSliceIndex(0).AtMapKey(names.AttrType), knownvalue.StringExact("MEMORY_RECORDS")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMemoryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_memory" {
				continue
			}

			_, err := tfbedrockagentcore.FindMemoryByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Memory %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMemoryExists(ctx context.Context, t *testing.T, n string, v *awstypes.Memory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindMemoryByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckMemories(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListMemoriesInput{}

	_, err := conn.ListMemories(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMemoryConfig_baseIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}

resource "aws_kms_key" "test" {
  description             = "Test key for %[1]s"
  deletion_window_in_days = 7
}
`, rName)
}

func testAccMemoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                  = %[1]q
  event_expiry_duration = 7
}
`, rName)
}

func testAccMemoryConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                  = %[1]q
  event_expiry_duration = 7
  description           = %[2]q
}
`, rName, description)
}

func testAccMemoryConfig_memoryExecutionRole(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                      = %[1]q
  event_expiry_duration     = 7
  memory_execution_role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccMemoryConfig_streamDeliveryResources(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}

resource "aws_iam_role_policy" "test_kinesis" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "kinesis:PutRecord",
        "kinesis:PutRecords",
        "kinesis:DescribeStream",
        "kinesis:DescribeStreamSummary",
        "kinesis:ListShards",
      ]
      Resource = aws_kinesis_stream.test.arn
    }]
  })
}

resource "aws_bedrockagentcore_memory" "test" {
  name                      = %[1]q
  event_expiry_duration     = 7
  memory_execution_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test_kinesis]

  stream_delivery_resources {
    resource {
      kinesis {
        data_stream_arn = aws_kinesis_stream.test.arn

        content_configuration {
          type  = "MEMORY_RECORDS"
          level = "METADATA_ONLY"
        }
      }
    }
  }
}
`, rName))
}

func testAccMemoryConfig_namespaceKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                  = %[1]q
  event_expiry_duration = 7

  namespace_key {
    key = "tenant"

    validation {
      regex_pattern = "^[a-z0-9-]+$"
    }
  }

  namespace_key {
    key = "environment"

    validation {
      allowed_values = ["development", "production"]
    }
  }
}
`, rName)
}

func testAccMemoryConfig_namespaceKeysUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                  = %[1]q
  event_expiry_duration = 7

  namespace_key {
    key = "environment"

    validation {
      allowed_values = ["development", "staging", "production"]
      regex_pattern  = "^[a-z]+$"
    }
  }

  namespace_key {
    key = "region1"
  }
}
`, rName)
}

func testAccMemoryConfig_namespaceKey(rName, key, validation string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory" "test" {
  name                  = %[1]q
  event_expiry_duration = 7

  namespace_key {
    key = %[2]q

    validation {
      %[3]s
    }
  }
}
`, rName, key, validation)
}
