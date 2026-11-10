// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
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
	tfnetworksecuritymanager "github.com/hashicorp/terraform-provider-aws/internal/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var checkScopeARN knownvalue.Check = tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`scope:[0-9a-z]+$`))

func TestAccNetworkSecurityManagerScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkScopeARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scope_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"account_filter": knownvalue.ListSizeExact(0),
							"resource_scope": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"exclude":              knownvalue.ListSizeExact(0),
									"include":              knownvalue.ListSizeExact(0),
									"include_all":          knownvalue.Bool(true),
									names.AttrResourceType: tfknownvalue.StringExact(awstypes.ScopeResourceTypeAlb),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/basic/"),
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

func TestAccNetworkSecurityManagerScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceScope, resourceName),
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

// Updating description and scope_configuration publishes a new version in place.
func TestAccNetworkSecurityManagerScope_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("initial"),
					"tag_value":     config.StringVariable("initial"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("updated"),
					"tag_value":     config.StringVariable("updated"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
				},
			},
		},
	})
}

func TestAccNetworkSecurityManagerScope_descriptionRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("initial"),
					"tag_value":     config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

// The and, or and not operators, alb_config and an empty tag map, across include and exclude.
func TestAccNetworkSecurityManagerScope_expressions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/expressions/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scope_configuration").AtSliceIndex(0).AtMapKey("resource_scope"), knownvalue.SetSizeExact(3)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/expressions/"),
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

// A resource scope switches in place between include_all, include and exclude, and
// a resource set selects resources by explicit ARN, alone or with an expression.
func TestAccNetworkSecurityManagerScope_resourceSet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"
	checkEIPARN := tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`elastic-ip/eipalloc-[0-9a-f]+$`))

	step := func(selection string, action plancheck.ResourceActionType, resourceScope map[string]knownvalue.Check) resource.TestStep {
		resourceScope[names.AttrResourceType] = tfknownvalue.StringExact(awstypes.ScopeResourceTypeEip)

		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Scope/resource_set/"),
			ConfigVariables: config.Variables{
				acctest.CtRName: config.StringVariable(rName),
				"selection":     config.StringVariable(selection),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckScopeExists(ctx, t, resourceName),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scope_configuration").AtSliceIndex(0).AtMapKey("resource_scope"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.ObjectExact(resourceScope),
				})),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			step("include_all", plancheck.ResourceActionCreate, map[string]knownvalue.Check{
				"exclude":     knownvalue.ListSizeExact(0),
				"include":     knownvalue.ListSizeExact(0),
				"include_all": knownvalue.Bool(true),
			}),
			step("include", plancheck.ResourceActionUpdate, map[string]knownvalue.Check{
				"exclude": knownvalue.ListSizeExact(0),
				"include": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"explicit_arns":      knownvalue.SetExact([]knownvalue.Check{checkEIPARN}),
						names.AttrExpression: knownvalue.ListSizeExact(0),
					}),
				}),
				"include_all": knownvalue.Null(),
			}),
			step("exclude", plancheck.ResourceActionUpdate, map[string]knownvalue.Check{
				"exclude": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"explicit_arns": knownvalue.SetExact([]knownvalue.Check{checkEIPARN}),
						names.AttrExpression: knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"and": knownvalue.ListSizeExact(0),
								"criteria": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"alb_config": knownvalue.ListSizeExact(0),
										names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
											"Name": knownvalue.StringExact(rName),
										}),
									}),
								}),
								"not": knownvalue.ListSizeExact(0),
								"or":  knownvalue.ListSizeExact(0),
							}),
						}),
					}),
				}),
				"include":     knownvalue.ListSizeExact(0),
				"include_all": knownvalue.Null(),
			}),
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/resource_set/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"selection":     config.StringVariable("exclude"),
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

// Each union in scope_configuration requires exactly one of its members, include_all can only be true,
// and each resource_scope needs a different resource_type.
func TestAccNetworkSecurityManagerScope_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	step := func(dir, expectError string) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Scope/validation/" + dir + "/"),
			ConfigVariables: config.Variables{
				acctest.CtRName: config.StringVariable(rName),
			},
			PlanOnly:    true,
			ExpectError: regexache.MustCompile(expectError),
		}
	}
	const invalidCombination = `Invalid Attribute Combination`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			step("account_filter_include_all_and_include", invalidCombination),
			step("resource_scope_none", invalidCombination),
			step("resource_scope_include_all_and_include", invalidCombination),
			step("resource_scope_duplicate_type", `Duplicate resource_type`),
			step("include_all_false", `Invalid Attribute Value Match`),
			step("expression_criteria_and_or", invalidCombination),
			step("criteria_alb_config_and_tags", invalidCombination),
		},
	})
}

// is_published drives the DRAFT/ACTIVE lifecycle. The scope is created as a draft,
// published in place, and finally given a pending draft on top of the published
// version, so that destroy removes both.
func TestAccNetworkSecurityManagerScope_isPublished(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	step := func(isPublished bool, tagValue string, action plancheck.ResourceActionType) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Scope/is_published/"),
			ConfigVariables: config.Variables{
				acctest.CtRName: config.StringVariable(rName),
				"is_published":  config.BoolVariable(isPublished),
				"tag_value":     config.StringVariable(tagValue),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckScopeExists(ctx, t, resourceName),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkScopeARN),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(isPublished)),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Created as a draft.
			step(false, "initial", plancheck.ResourceActionCreate),
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/is_published/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"is_published":  config.BoolVariable(false),
					"tag_value":     config.StringVariable("initial"),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// DRAFT -> ACTIVE: the draft itself is published.
			step(true, "initial", plancheck.ResourceActionUpdate),
			// ACTIVE -> DRAFT: a draft is saved on top of the published version.
			step(false, "updated", plancheck.ResourceActionUpdate),
		},
	})
}

// A scope cannot change between global and regional resource types. The scope is
// managed in us-east-1, where the global CloudFront resource type is selectable,
// whatever the default Region.
func TestAccNetworkSecurityManagerScope_regionalToGlobal(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/regional/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(endpoints.UsEast1RegionID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Scope/global/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(endpoints.UsEast1RegionID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scope_configuration").AtSliceIndex(0).AtMapKey("resource_scope"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"include_all":          knownvalue.Bool(true),
							names.AttrResourceType: tfknownvalue.StringExact(awstypes.ScopeResourceTypeCfd),
						}),
					})),
				},
			},
		},
	})
}

func testAccCheckScopeExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		_, err := tfnetworksecuritymanager.FindScopeByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccCheckScopeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_scope" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
			_, err := tfnetworksecuritymanager.FindScopeByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Scope %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccPreCheckScope(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	var input networksecuritymanager.ListScopesInput
	_, err := conn.ListScopes(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
