// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfresiliencehub "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubResiliencyPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "tier", "NotApplicable"),
					resource.TestCheckResourceAttr(resourceName, "data_location_constraint", "AnyLocation"),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", "1h0m0s"),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", "1h0m0s"),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", "1h0m0s"),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", "1h0m0s"),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rpo"),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rto"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", "1h0m0s"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", "1h0m0s"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_dataLocationConstraint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_dataLocationConstraint(rName, awstypes.DataLocationConstraintSameContinent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "data_location_constraint", string(awstypes.DataLocationConstraintSameContinent)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_dataLocationConstraint(rName, awstypes.DataLocationConstraintSameCountry),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, "data_location_constraint", string(awstypes.DataLocationConstraintSameCountry)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	const (
		initial = "initial"
		updated = "updated"
	)

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_description(rName, initial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, initial),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_description(rName, updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updated),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_policy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	const (
		initialDuration = "1h0m0s"
		updatedDuration = "24h0m0s"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_policy(rName, initialDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", initialDuration),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rpo"),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rto"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", initialDuration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_policy(rName, updatedDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", updatedDuration),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rpo"),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rto"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", updatedDuration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_policyWithRegion(rName, updatedDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", updatedDuration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_policyWithRegion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	const (
		initialDuration = "1h0m0s"
		updatedDuration = "24h0m0s"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_policyWithRegion(rName, initialDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rto", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", initialDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", initialDuration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_policyWithRegion(rName, updatedDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", updatedDuration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_policy(rName, updatedDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto", updatedDuration),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rpo"),
					resource.TestCheckNoResourceAttr(resourceName, "policy.region.rto"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo", updatedDuration),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto", updatedDuration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_tier(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	expectNoARNChange := statecheck.CompareValue(compare.ValuesSame())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_tier(rName, awstypes.ResiliencyPolicyTierMissionCritical),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "tier", string(awstypes.ResiliencyPolicyTierMissionCritical)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccResiliencyPolicyConfig_tier(rName, awstypes.ResiliencyPolicyTierCoreServices),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					resource.TestCheckResourceAttr(resourceName, "tier", string(awstypes.ResiliencyPolicyTierCoreServices)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoARNChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfresiliencehub.ResourceResiliencyPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResiliencyPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehub_resiliency_policy" {
				continue
			}

			input := &resiliencehub.DescribeResiliencyPolicyInput{
				PolicyArn: aws.String(rs.Primary.Attributes[names.AttrARN]),
			}
			_, err := conn.DescribeResiliencyPolicy(ctx, input)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ResilienceHub, create.ErrActionCheckingDestroyed, tfresiliencehub.ResNameResiliencyPolicy, rs.Primary.ID, err)
			}

			return create.Error(names.ResilienceHub, create.ErrActionCheckingDestroyed, tfresiliencehub.ResNameResiliencyPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResiliencyPolicyExists(ctx context.Context, name string, policy *resiliencehub.DescribeResiliencyPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingExistence, tfresiliencehub.ResNameResiliencyPolicy, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingExistence, tfresiliencehub.ResNameResiliencyPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)
		resp, err := conn.DescribeResiliencyPolicy(ctx, &resiliencehub.DescribeResiliencyPolicyInput{
			PolicyArn: aws.String(rs.Primary.Attributes[names.AttrARN]),
		})

		if err != nil {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingExistence, tfresiliencehub.ResNameResiliencyPolicy, rs.Primary.ID, err)
		}

		*policy = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)

	input := &resiliencehub.ListResiliencyPoliciesInput{}
	_, err := conn.ListResiliencyPolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccResiliencyPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  tier = "NotApplicable"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}
`, rName)
}

func testAccResiliencyPolicyConfig_dataLocationConstraint(rName string, dataLocationConstraint awstypes.DataLocationConstraint) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  tier = "NotApplicable"

  data_location_constraint = %[2]q

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}
`, rName, dataLocationConstraint)
}

func testAccResiliencyPolicyConfig_description(rName, resPolicyDescValue string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  description = %[2]q

  tier = "NotApplicable"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}
`, rName, resPolicyDescValue)
}

func testAccResiliencyPolicyConfig_tier(rName string, tier awstypes.ResiliencyPolicyTier) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  tier = %[2]q

  data_location_constraint = "AnyLocation"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }
}
`, rName, tier)
}

func testAccResiliencyPolicyConfig_policy(rName, duration string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  tier = "NotApplicable"

  policy {
    az {
      rpo = %[2]q
      rto = %[2]q
    }
    hardware {
      rpo = %[2]q
      rto = %[2]q
    }
    software {
      rpo = %[2]q
      rto = %[2]q
    }
  }
}
`, rName, duration)
}

func testAccResiliencyPolicyConfig_policyWithRegion(rName, duration string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  name = %[1]q

  tier = "NotApplicable"

  policy {
    az {
      rpo = %[2]q
      rto = %[2]q
    }
    hardware {
      rpo = %[2]q
      rto = %[2]q
    }
    region {
      rpo = %[2]q
      rto = %[2]q
    }
    software {
      rpo = %[2]q
      rto = %[2]q
    }
  }
}
`, rName, duration)
}
