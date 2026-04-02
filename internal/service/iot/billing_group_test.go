// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTBillingGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", regexache.MustCompile(fmt.Sprintf("billinggroup/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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

func TestAccIoTBillingGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfiot.NewResourceBillingGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTBillingGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBillingGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBillingGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIoTBillingGroup_properties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_properties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBillingGroupConfig_propertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
			},
		},
	})
}

func TestAccIoTBillingGroup_migrateFromPluginSDK(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_billing_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.IoTServiceID),
		CheckDestroy: testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.74.0",
					},
				},
				Config: testAccBillingGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", regexache.MustCompile(fmt.Sprintf("billinggroup/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBillingGroupConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccIoTBillingGroup_migrateFromPluginSDK_properties(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_billing_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.IoTServiceID),
		CheckDestroy: testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.74.0",
					},
				},
				Config: testAccBillingGroupConfig_properties(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBillingGroupConfig_properties(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBillingGroupConfig_propertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccIoTBillingGroup_requiredTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccBillingGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
			// Creation with required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccBillingGroupConfig_tags1(rName, tagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
			// Updates which remove required tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccBillingGroupConfig_tags1(rName, nonRequiredTagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
		},
	})
}

func TestAccIoTBillingGroup_requiredTags_defaultTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccBillingGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
			// Creation with required tags in default_tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyComplianceAndDefaultTags1("error", tagKey, acctest.CtValue1),
					testAccBillingGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
			// Updates which remove required tags from default_tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyComplianceAndDefaultTags1("error", nonRequiredTagKey, acctest.CtValue1),
					testAccBillingGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
		},
	})
}

func TestAccIoTBillingGroup_requiredTags_warning(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("warning"),
					testAccBillingGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
			// Updates adding required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("warning"),
					testAccBillingGroupConfig_tags1(rName, tagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
			// Updates which remove required tags also succeed
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("warning"),
					testAccBillingGroupConfig_tags1(rName, nonRequiredTagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccIoTBillingGroup_requiredTags_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("disabled"),
					testAccBillingGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
			// Updates adding required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("disabled"),
					testAccBillingGroupConfig_tags1(rName, tagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
			// Updates which remove required tags also succeed
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("disabled"),
					testAccBillingGroupConfig_tags1(rName, nonRequiredTagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
				),
			},
		},
	})
}

// A smoke test to verify setting provider_meta does not trigger unexpected
// behavior for Plugin Framework based resources
func TestAccIoTBillingGroup_providerMeta(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigProviderMeta(),
					testAccBillingGroupConfig_basic(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, t, resourceName),
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

func testAccCheckBillingGroupExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		_, err := tfiot.FindBillingGroupByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckBillingGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_billing_group" {
				continue
			}

			_, err := tfiot.FindBillingGroupByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Billing Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBillingGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccBillingGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBillingGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccBillingGroupConfig_properties(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  properties {
    description = "test description 1"
  }
}
`, rName)
}

func testAccBillingGroupConfig_propertiesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  properties {
    description = "test description 2"
  }
}
`, rName)
}
