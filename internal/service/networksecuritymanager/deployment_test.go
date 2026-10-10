// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	shieldtypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworksecuritymanager "github.com/hashicorp/terraform-provider-aws/internal/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "scope_arn", "aws_networksecuritymanager_scope.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.enable_cross_account_visibility", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.firewall_type", "WAF"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.in_scope_resource_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.in_scope_resource_types.0", "AWS::CloudFront::Distribution"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.policy_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "coverage.0.policy_arns.0", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "warnings.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`deployment:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deployment_id"), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
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

func testAccDeployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceDeployment, resourceName),
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

// Every optional argument at once, on a deployment that is never published.
func testAccDeployment_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.shield", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.enable_cross_account_visibility", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "coverage.*", map[string]string{
						"firewall_type":             "WAF",
						"in_scope_resource_types.#": "1",
						"policy_arns.#":             "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "coverage.*", map[string]string{
						"firewall_type":             "SHIELD_ADVANCED",
						"in_scope_resource_types.#": "1",
						"policy_arns.#":             "1",
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(names.AttrDescription)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
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

// Every transition of the policy set: add a second WAF policy, drop the
// first, mix WAF and Shield Advanced, Shield Advanced alone.
func testAccDeployment_policies(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.test.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.test.arn, aws_networksecuritymanager_policy.second.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.second", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.firewall_type", "WAF"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.policy_arns.#", "2"),
				),
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
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.second.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.second", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
				),
			},
			{
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.shield.arn, aws_networksecuritymanager_policy.test.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.shield", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "coverage.*", map[string]string{
						"firewall_type": "WAF",
						"policy_arns.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "coverage.*", map[string]string{
						"firewall_type": "SHIELD_ADVANCED",
						"policy_arns.#": "1",
					}),
				),
			},
			{
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.shield.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.shield", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.firewall_type", "SHIELD_ADVANCED"),
				),
			},
			{
				// The same set in the other order is not a change.
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.test.arn, aws_networksecuritymanager_policy.shield.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"),
				),
			},
			{
				Config: testAccDeploymentConfig_policies(rName, "aws_networksecuritymanager_policy.shield.arn, aws_networksecuritymanager_policy.test.arn"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
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

// The scope is swapped in place; the coverage follows the new scope's class.
func testAccDeployment_scope(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_scope(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "scope_arn", "aws_networksecuritymanager_scope.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.in_scope_resource_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.in_scope_resource_types.0", "AWS::CloudFront::Distribution"),
				),
			},
			{
				Config: testAccDeploymentConfig_scope(rName, "regional"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "scope_arn", "aws_networksecuritymanager_scope.regional", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.in_scope_resource_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "coverage.0.in_scope_resource_types.*", "AWS::ElasticLoadBalancingV2::LoadBalancer::application"),
					resource.TestCheckTypeSetElemAttr(resourceName, "coverage.0.in_scope_resource_types.*", "AWS::ApiGateway::Stage"),
				),
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

func testAccDeployment_deploymentConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_crossAccountVisibility(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.enable_cross_account_visibility", acctest.CtTrue),
				),
			},
			{
				Config: testAccDeploymentConfig_crossAccountVisibility(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.enable_cross_account_visibility", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccDeploymentConfig_crossAccountVisibility(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.enable_cross_account_visibility", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
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

// The draft lifecycle: created unpublished, published, a draft saved on top
// of the published version, published again, and destroyed with a pending
// draft.
func testAccDeployment_publish(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "is_published", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusDraft)),
					resource.TestCheckResourceAttr(resourceName, "has_published_version", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-security-manager", regexache.MustCompile(`deployment:[0-9a-z]+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccDeploymentConfig_isPublished(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "is_published", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccDeploymentConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "is_published", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusDraft)),
					resource.TestCheckResourceAttr(resourceName, "has_published_version", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccDeploymentConfig_isPublished(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "is_published", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
			},
			{
				Config: testAccDeploymentConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusDraft)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
				),
			},
		},
	})
}

func testAccDeployment_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
				),
			},
			{
				Config: testAccDeploymentConfig_description(rName, "description 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
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
				Config: testAccDeploymentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
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

// An update token read on the last refresh is stale after any change outside
// Terraform; the resource reads the current token right before updating.
func testAccDeployment_updateToken(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					testAccDeploymentUpdateOutOfBand(ctx, t, &v, true, "changed outside Terraform")
				},
				Config: testAccDeploymentConfig_descriptionAndCrossAccountVisibility(rName, "description 1", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.enable_cross_account_visibility", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
				),
			},
			{
				// A draft saved outside Terraform is published back.
				PreConfig: func() {
					testAccDeploymentUpdateOutOfBand(ctx, t, &v, false, "draft outside Terraform")
				},
				Config: testAccDeploymentConfig_descriptionAndCrossAccountVisibility(rName, "description 1", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
					resource.TestCheckResourceAttr(resourceName, "is_published", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
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

func testAccDeployment_replace(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_name(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccDeploymentConfig_name(rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v2),
					testAccCheckDeploymentRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccDeployment_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccDeploymentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v2),
					testAccCheckDeploymentRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// A scope or policy referenced by a published deployment cannot be deleted;
// replacing one needs create_before_destroy so that the deployment is moved
// first.
func testAccDeployment_referenceReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_referenceNames(rName, rName, rName, 1, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "scope_arn", "aws_networksecuritymanager_scope.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
				),
			},
			{
				// The scope is renamed with the default lifecycle.
				Config:      testAccDeploymentConfig_referenceNames(rName, rNameUpdated, rName, 1, false, false),
				ExpectError: regexache.MustCompile(`ConflictException: The resource is currently in use`),
			},
			{
				Config: testAccDeploymentConfig_referenceNames(rName, rNameUpdated, rName, 1, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "scope_arn", "aws_networksecuritymanager_scope.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_scope.test", plancheck.ResourceActionCreateBeforeDestroy),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// The policy is renamed with the default lifecycle.
				Config:      testAccDeploymentConfig_referenceNames(rName, rNameUpdated, rNameUpdated, 1, true, false),
				ExpectError: regexache.MustCompile(`ConflictException: The resource is currently in use`),
			},
			{
				// With create_before_destroy the old policy still holds its
				// priority while the new one is created, so the replacement
				// needs a new priority too.
				Config:      testAccDeploymentConfig_referenceNames(rName, rNameUpdated, rNameUpdated, 1, true, true),
				ExpectError: regexache.MustCompile(`ConflictException: Priority 1 is already used`),
			},
			{
				Config: testAccDeploymentConfig_referenceNames(rName, rNameUpdated, rNameUpdated, 2, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_policy.test", plancheck.ResourceActionCreateBeforeDestroy),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

// A deployment that exists only as a draft does not protect its scope or
// policy: both can be replaced with the default lifecycle, and the draft
// follows the new ARNs.
func testAccDeployment_draftReferences(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_draftReferenceNames(rName, rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusDraft)),
				),
			},
			{
				Config: testAccDeploymentConfig_draftReferenceNames(rName, rNameUpdated, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "scope_arn", "aws_networksecuritymanager_scope.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusDraft)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_scope.test", plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccDeploymentConfig_draftReferenceNames(rName, rNameUpdated, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", "aws_networksecuritymanager_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusDraft)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_policy.test", plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccDeployment_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	fakeARN := func(kind string, n int) string {
		return fmt.Sprintf("arn:%s:network-security-manager:%s:123456789012:%s:tfacc%021d", acctest.Partition(), acctest.Region(), kind, n)
	}
	policy := "aws_networksecuritymanager_policy.test.arn"
	scope := "aws_networksecuritymanager_scope.test.arn"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccDeploymentConfig_literal(rName, "-"+rName, "["+policy+"]", scope, true),
				ExpectError: regexache.MustCompile(`must start with an alphanumeric character`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, strings.Repeat("a", 129), "["+policy+"]", scope, true),
				ExpectError: regexache.MustCompile(`string length must be between 1 and 128`),
			},
			{
				Config:      testAccDeploymentConfig_description(rName, strings.Repeat("a", 257)),
				ExpectError: regexache.MustCompile(`string length must be at most 256`),
			},
			{
				Config:      testAccDeploymentConfig_description(rName, "not # allowed"),
				ExpectError: regexache.MustCompile(`must contain only alphanumeric characters`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, "[]", scope, true),
				ExpectError: regexache.MustCompile(`set must contain at least 1 elements and at most 2\s+elements`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, fmt.Sprintf("[%q, %q, %q]", fakeARN(names.AttrPolicy, 1), fakeARN(names.AttrPolicy, 2), fakeARN(names.AttrPolicy, 3)), scope, true),
				ExpectError: regexache.MustCompile(`set must contain at least 1 elements and at most 2\s+elements`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, `["not-an-arn"]`, scope, true),
				ExpectError: regexache.MustCompile(`value must be a valid ARN`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, "["+policy+"]", `"not-an-arn"`, true),
				ExpectError: regexache.MustCompile(`cannot be parsed as an ARN`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, "["+policy+"]", scope, false),
				ExpectError: regexache.MustCompile(`Block deployment_configuration must have a configuration value`),
			},
			// The API rejects the following.
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, fmt.Sprintf("[%q]", fakeARN(names.AttrPolicy, 1)), scope, true),
				ExpectError: regexache.MustCompile(`AccessDeniedException|UnauthorizedException`),
			},
			{
				Config:      testAccDeploymentConfig_missingReference(rName, names.AttrPolicy),
				ExpectError: regexache.MustCompile(`(?s)Cannot reference policy.*because it doesn't exist`),
			},
			{
				Config:      testAccDeploymentConfig_missingReference(rName, names.AttrScope),
				ExpectError: regexache.MustCompile(`(?s)Cannot reference scope.*because it doesn't exist`),
			},
			{
				Config:      testAccDeploymentConfig_draftReference(rName, names.AttrPolicy),
				ExpectError: regexache.MustCompile(`(?s)Cannot reference policy.*because it doesn't exist`),
			},
			{
				Config:      testAccDeploymentConfig_draftReference(rName, names.AttrScope),
				ExpectError: regexache.MustCompile(`(?s)Cannot reference scope.*because it doesn't exist`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, `["${aws_networksecuritymanager_policy.test.arn}:DRAFT"]`, scope, true),
				ExpectError: regexache.MustCompile(`DRAFT-qualified ARNs not allowed`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, "["+policy+"]", `"${aws_networksecuritymanager_scope.test.arn}:DRAFT"`, true),
				ExpectError: regexache.MustCompile(`DRAFT-qualified ARNs not allowed`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, "["+scope+"]", scope, true),
				ExpectError: regexache.MustCompile(`Expected policy ARN but found scope ARN`),
			},
			{
				Config:      testAccDeploymentConfig_literal(rName, rName, "["+policy+"]", policy, true),
				ExpectError: regexache.MustCompile(`Expected scope ARN but found policy ARN`),
			},
		},
	})
}

func testAccCheckDeploymentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_deployment" {
				continue
			}

			_, err := tfnetworksecuritymanager.FindDeploymentByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Deployment %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckDeploymentExists(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindDeploymentByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDeploymentRecreated(before, after *networksecuritymanager.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.DeploymentId), aws.ToString(after.DeploymentId); before == after {
			return fmt.Errorf("Network Security Manager Deployment (%s) not recreated", before)
		}

		return nil
	}
}

func testAccDeploymentUpdateOutOfBand(ctx context.Context, t *testing.T, v *networksecuritymanager.GetDeploymentOutput, publish bool, description string) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.UpdateDeploymentInput{
		DeploymentDescription: aws.String(description),
		DeploymentIdentifier:  v.DeploymentArn,
		IsPublished:           aws.Bool(publish),
		UpdateToken:           v.UpdateToken,
	}
	_, err := conn.UpdateDeployment(ctx, &input)
	if err != nil {
		t.Fatalf("updating Network Security Manager Deployment (%s) outside Terraform: %s", aws.ToString(v.DeploymentArn), err)
	}
}

// A WAF policy that does not remediate, on a scope that selects only
// CloudFront distributions carrying a tag no real resource has: the safe base
// for tests that need a published deployment without touching any resource.
func testAccDeploymentConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.test.arn
  }
}

resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccDeploymentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName))
}

// A second WAF policy and a Shield Advanced policy, neither remediating,
// with distinct priorities.
func testAccDeploymentConfig_morePolicies(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "second" {
  name          = "%[1]s-second"
  firewall_type = "WAF"
  priority      = 2

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.test.arn
  }
}

resource "aws_networksecuritymanager_policy" "shield" {
  name          = "%[1]s-shield"
  firewall_type = "SHIELD_ADVANCED"
  priority      = 3

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false
  }
}
`, rName)
}

func testAccDeploymentConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), testAccDeploymentConfig_morePolicies(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name         = %[1]q
  description  = %[4]q
  is_published = false
  policy_arns  = [aws_networksecuritymanager_policy.test.arn, aws_networksecuritymanager_policy.shield.arn]
  scope_arn    = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = true
  }

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, acctest.CtKey1, acctest.CtValue1, names.AttrDescription))
}

func testAccDeploymentConfig_policies(rName, policyARNs string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), testAccDeploymentConfig_morePolicies(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [%[2]s]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, policyARNs))
}

// A regional scope selecting load balancers and API stages by the same
// test-only tag, next to the CloudFront scope of the base.
func testAccDeploymentConfig_scope(rName, scope string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "regional" {
  name = "%[1]s-regional"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
    }

    resource_scope {
      resource_type = "AWS::ApiGateway::Stage"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
    }
  }
}

resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.%[2]s.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, scope))
}

func testAccDeploymentConfig_crossAccountVisibility(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = %[2]t
  }
}
`, rName, enabled))
}

func testAccDeploymentConfig_isPublished(rName string, isPublished bool) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name         = %[1]q
  is_published = %[2]t
  policy_arns  = [aws_networksecuritymanager_policy.test.arn]
  scope_arn    = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, isPublished))
}

func testAccDeploymentConfig_description(rName, description string) string {
	return testAccDeploymentConfig_descriptionAndCrossAccountVisibility(rName, description, false)
}

func testAccDeploymentConfig_descriptionAndCrossAccountVisibility(rName, description string, enabled bool) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  description = %[2]q
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = %[3]t
  }
}
`, rName, description, enabled))
}

func testAccDeploymentConfig_name(rName, deploymentName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[2]q
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, deploymentName))
}

// A published deployment whose scope and policy each have a name and a
// lifecycle under test.
func testAccDeploymentConfig_referenceNames(rName, scopeName, policyName string, policyPriority int32, scopeCreateBeforeDestroy, policyCreateBeforeDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[3]q
  firewall_type = "WAF"
  priority      = %[6]d

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.test.arn
  }

  lifecycle {
    create_before_destroy = %[5]t
  }
}

resource "aws_networksecuritymanager_scope" "test" {
  name = %[2]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
    }
  }

  lifecycle {
    create_before_destroy = %[4]t
  }
}

resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, scopeName, policyName, scopeCreateBeforeDestroy, policyCreateBeforeDestroy, policyPriority)
}

// The same, never published.
func testAccDeploymentConfig_draftReferenceNames(rName, scopeName, policyName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[3]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.test.arn
  }
}

resource "aws_networksecuritymanager_scope" "test" {
  name = %[2]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
    }
  }
}

resource "aws_networksecuritymanager_deployment" "test" {
  name         = %[1]q
  is_published = false
  policy_arns  = [aws_networksecuritymanager_policy.test.arn]
  scope_arn    = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, scopeName, policyName)
}

func testAccDeploymentConfig_literal(rName, deploymentName, policyARNs, scopeARN string, withConfiguration bool) string {
	configuration := ""
	if withConfiguration {
		configuration = `
  deployment_configuration {
    enable_cross_account_visibility = false
  }
`
	}

	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = %[2]s
  scope_arn   = %[3]s
%[4]s
}
`, deploymentName, policyARNs, scopeARN, configuration))
}

// A reference to a same-account policy or scope that does not exist.
func testAccDeploymentConfig_missingReference(rName, kind string) string {
	policyARN := "aws_networksecuritymanager_policy.test.arn"
	scopeARN := "aws_networksecuritymanager_scope.test.arn"
	switch kind {
	case names.AttrPolicy:
		policyARN = `"${replace(aws_networksecuritymanager_policy.test.arn, "/[0-9a-z]+$/", "tfacc000000000000000001")}"`
	case names.AttrScope:
		scopeARN = `"${replace(aws_networksecuritymanager_scope.test.arn, "/[0-9a-z]+$/", "tfacc000000000000000001")}"`
	}

	return testAccDeploymentConfig_literal(rName, rName, "["+policyARN+"]", scopeARN, true)
}

// A reference to a policy or scope that exists only as a draft.
func testAccDeploymentConfig_draftReference(rName, kind string) string {
	policyARN := "aws_networksecuritymanager_policy.test.arn"
	scopeARN := "aws_networksecuritymanager_scope.test.arn"
	switch kind {
	case names.AttrPolicy:
		policyARN = "aws_networksecuritymanager_policy.draft.arn"
	case names.AttrScope:
		scopeARN = "aws_networksecuritymanager_scope.draft.arn"
	}

	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "draft" {
  name          = "%[1]s-draft"
  firewall_type = "SHIELD_ADVANCED"
  is_published  = false
  priority      = 2

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false
  }
}

resource "aws_networksecuritymanager_scope" "draft" {
  name         = "%[1]s-draft"
  is_published = false

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
    }
  }
}

resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [%[2]s]
  scope_arn   = %[3]s

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, policyARN, scopeARN))
}

// A published deployment of remediating policies on a real load balancer:
// the service creates and associates a web ACL and a Shield Advanced
// protection, and removes both again when the deployment is deleted, since
// the policies clean up after themselves.
//
// The load balancer is not created by the test: the service only acts on
// resources in its own inventory, and a freshly created load balancer took
// well over half an hour to appear there. Set
// NSM_DEPLOYMENT_TEST_LOAD_BALANCER_ARN to the ARN of an existing Application
// Load Balancer, in the test region, that the test may protect; the test is
// skipped otherwise. The deployment creates and later removes a web ACL and a
// Shield Advanced protection on that load balancer.
func testAccDeployment_remediation(t *testing.T) {
	ctx := acctest.Context(t)
	loadBalancerARN := acctest.SkipIfEnvVarNotSet(t, "NSM_DEPLOYMENT_TEST_LOAD_BALANCER_ARN")
	var v networksecuritymanager.GetDeploymentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_deployment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_remediation(rName, loadBalancerARN, "aws_networksecuritymanager_policy.waf.arn", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "coverage.0.in_scope_resource_types.0", "AWS::ElasticLoadBalancingV2::LoadBalancer::application"),
					testAccCheckLoadBalancerWebACL(ctx, t, loadBalancerARN, rName, true),
					testAccCheckLoadBalancerShieldProtection(ctx, t, loadBalancerARN, false),
				),
			},
			{
				Config: testAccDeploymentConfig_remediation(rName, loadBalancerARN, "aws_networksecuritymanager_policy.waf.arn, aws_networksecuritymanager_policy.shield.arn", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "coverage.#", "2"),
					testAccCheckLoadBalancerWebACL(ctx, t, loadBalancerARN, rName, true),
					testAccCheckLoadBalancerShieldProtection(ctx, t, loadBalancerARN, true),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				// The deployment is removed; the protections it created go with it.
				Config: testAccDeploymentConfig_remediation(rName, loadBalancerARN, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerWebACL(ctx, t, loadBalancerARN, rName, false),
					testAccCheckLoadBalancerShieldProtection(ctx, t, loadBalancerARN, false),
				),
			},
		},
	})
}

// testAccCheckLoadBalancerWebACL waits for the load balancer to have (or no
// longer have) a web ACL created by the service. Remediation runs after the
// deployment is published and has been observed to take from under a minute
// to well over ten, so the check retries for a long time.
func testAccCheckLoadBalancerWebACL(ctx context.Context, t *testing.T, loadBalancerARN, rName string, want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)
		input := wafv2.GetWebACLForResourceInput{
			ResourceArn: aws.String(loadBalancerARN),
		}

		_, err := tfresource.RetryUntilEqual(ctx, 30*time.Minute, want, func(ctx context.Context) (bool, error) {
			output, err := conn.GetWebACLForResource(ctx, &input)
			if err != nil {
				return false, err
			}

			return output.WebACL != nil && strings.HasPrefix(aws.ToString(output.WebACL.Name), "NSMManagedWebACL-"), nil
		})
		if err != nil {
			return fmt.Errorf("waiting for Load Balancer (%s) web ACL managed by Network Security Manager = %t: %w", loadBalancerARN, want, err)
		}

		if want {
			return nil
		}

		// The service disassociates the web ACL first and deletes it a little
		// later; wait for the deletion too, so that nothing is left behind.
		_, err = tfresource.RetryUntilEqual(ctx, 30*time.Minute, false, func(ctx context.Context) (bool, error) {
			input := wafv2.ListWebACLsInput{
				Scope: wafv2types.ScopeRegional,
			}
			for {
				output, err := conn.ListWebACLs(ctx, &input)
				if err != nil {
					return false, err
				}

				for _, v := range output.WebACLs {
					if strings.HasPrefix(aws.ToString(v.Name), "NSMManagedWebACL-"+rName) {
						return true, nil
					}
				}

				if aws.ToString(output.NextMarker) == "" {
					return false, nil
				}
				input.NextMarker = output.NextMarker
			}
		})
		if err != nil {
			return fmt.Errorf("waiting for the web ACL managed by Network Security Manager for %s to be deleted: %w", rName, err)
		}

		return nil
	}
}

// testAccCheckLoadBalancerShieldProtection waits for the load balancer to
// have (or no longer have) a Shield Advanced protection created by the
// service.
func testAccCheckLoadBalancerShieldProtection(ctx context.Context, t *testing.T, loadBalancerARN string, want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ShieldClient(ctx)
		input := shield.DescribeProtectionInput{
			ResourceArn: aws.String(loadBalancerARN),
		}

		_, err := tfresource.RetryUntilEqual(ctx, 30*time.Minute, want, func(ctx context.Context) (bool, error) {
			output, err := conn.DescribeProtection(ctx, &input)
			if errs.IsA[*shieldtypes.ResourceNotFoundException](err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}

			return output.Protection != nil && strings.HasPrefix(aws.ToString(output.Protection.Name), "NSMManagedShieldProtection-"), nil
		})
		if err != nil {
			return fmt.Errorf("waiting for Load Balancer (%s) Shield Advanced protection managed by Network Security Manager = %t: %w", loadBalancerARN, want, err)
		}

		return nil
	}
}

// A scope pinned to an existing load balancer, and remediating WAF and Shield
// Advanced policies that clean up after themselves. The WAF policy carries
// both configuration rules a web ACL needs: without a VisibilityConfig rule
// the service cannot build the web ACL and never remediates.
func testAccDeploymentConfig_remediation(rName, loadBalancerARN, policyARNs string, withDeployment bool) string {
	deployment := ""
	if withDeployment {
		deployment = fmt.Sprintf(`
resource "aws_networksecuritymanager_deployment" "test" {
  name        = %[1]q
  policy_arns = [%[2]s]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
`, rName, policyARNs)
	}

	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "default_action" {
  name          = "%[1]s-default-action"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_rule" "visibility" {
  name          = "%[1]s-visibility"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    VisibilityConfig = {
      SampledRequestsEnabled   = true
      CloudWatchMetricsEnabled = true
      MetricName               = "tf-acc-test"
    }
  })
}

resource "aws_networksecuritymanager_policy" "waf" {
  name          = "%[1]s-waf"
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.default_action.arn
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.visibility.arn
  }
}

resource "aws_networksecuritymanager_policy" "shield" {
  name          = "%[1]s-shield"
  firewall_type = "SHIELD_ADVANCED"
  priority      = 2

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true
  }
}

resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        explicit_arns = [%[2]q]
      }
    }
  }
}
%[3]s
`, rName, loadBalancerARN, deployment)
}
