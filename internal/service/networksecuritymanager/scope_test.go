// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
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

const (
	scopeResourceTypeALB   = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"
	scopeResourceTypeAPIGW = "AWS::ApiGateway::Stage"
	scopeResourceTypeCFD   = "AWS::CloudFront::Distribution"
	scopeResourceTypeCLB   = "AWS::ElasticLoadBalancing::LoadBalancer"
	scopeResourceTypeEIP   = "AWS::EC2::EIP"
)

func testAccScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.account_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scope_configuration.0.resource_scope.*", map[string]string{
						names.AttrResourceType:                               scopeResourceTypeCFD,
						"include.#":                                          "1",
						"include.0.expression.#":                             "1",
						"include.0.expression.0.criteria.#":                  "1",
						"include.0.expression.0.criteria.0.tags.%":           "1",
						"include.0.expression.0.criteria.0.tags.nsm-acctest": rName,
						"include.0.expression.0.criteria.0.alb_config.#":     "0",
						"include.0.expression.0.and.#":                       "0",
						"include.0.expression.0.or.#":                        "0",
						"include.0.expression.0.not.#":                       "0",
						"exclude.#":                                          "0",
					}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`scope:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scope_id"), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`scope:[0-9a-z]+$`)),
					}),
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

func testAccScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceScope, resourceName),
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

// Every optional argument set on creation; the scope is created as a draft
// and selects every distribution in the account.
func testAccScope_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scope_configuration.0.resource_scope.*", map[string]string{
						names.AttrResourceType: scopeResourceTypeCFD,
						"include_all":          acctest.CtTrue,
						"include.#":            "0",
						"exclude.#":            "0",
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Every distribution in the account")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`scope:[0-9a-z]+$`))),
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

// Every regional resource type, alone and combined. The global type
// (CloudFront) is covered by the other tests; it cannot share a scope with a
// regional type.
func testAccScope_resourceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_includeAll(rName, scopeResourceTypeALB, scopeResourceTypeAPIGW, scopeResourceTypeCLB, scopeResourceTypeEIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.#", "4"),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeALB),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeAPIGW),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeCLB),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeEIP),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
				Config: testAccScopeConfig_includeAll(rName, scopeResourceTypeEIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.#", "1"),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeEIP),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccScopeConfig_includeAll(rName, scopeResourceTypeEIP, scopeResourceTypeAPIGW),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.#", "2"),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeAPIGW),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeEIP),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccScopeConfig_includeAll(rName, scopeResourceTypeALB, scopeResourceTypeCLB),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.#", "2"),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeALB),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeCLB),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
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

// include_all, include and exclude on one resource type, and every
// transition between them.
func testAccScope_selection(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	includeAll := "include_all = true"
	include := fmt.Sprintf(`
      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
            }
          }
        }
      }
`, rName)
	exclude := fmt.Sprintf(`
      exclude {
        expression {
          criteria {
            tags = {
              nsm-acctest = %[1]q
              nsm-exclude = "true"
            }
          }
        }
      }
`, rName)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, includeAll),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeALB),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, include),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scope_configuration.0.resource_scope.*", map[string]string{
						names.AttrResourceType: scopeResourceTypeALB,
						"include.#":            "1",
						"include.0.expression.0.criteria.0.tags.nsm-acctest": rName,
						"exclude.#": "0",
					}),
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
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, exclude),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scope_configuration.0.resource_scope.*", map[string]string{
						names.AttrResourceType: scopeResourceTypeALB,
						"include.#":            "0",
						"exclude.#":            "1",
						"exclude.0.expression.0.criteria.0.tags.%":           "2",
						"exclude.0.expression.0.criteria.0.tags.nsm-exclude": acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
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
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, includeAll),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeALB),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, exclude),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
				),
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, include),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"),
				),
			},
			{
				// The same configuration again is not written back.
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeALB, include),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// Explicit ARNs of resources created by the test: an Application Load
// Balancer in a regional scope and a CloudFront distribution in a global one.
func testAccScope_explicitARNs(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	albScopeName := "aws_networksecuritymanager_scope.alb"
	cfdScopeName := "aws_networksecuritymanager_scope.cfd"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_explicitARNs(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, albScopeName, &v),
					testAccCheckScopeExists(ctx, t, cfdScopeName, &v),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(albScopeName, "scope_configuration.0.resource_scope.0.include.0.explicit_arns.*", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.include.0.explicit_arns.#", "1"),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.include.0.expression.#", "1"),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.include.0.expression.0.and.0.criteria.#", "2"),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.include.0.expression.0.and.0.criteria.0.tags.nsm-acctest", rName),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.include.0.expression.0.and.0.criteria.1.alb_config.0.scheme", "internal"),
					resource.TestCheckResourceAttr(cfdScopeName, "scope_configuration.0.resource_scope.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(cfdScopeName, "scope_configuration.0.resource_scope.0.include.0.explicit_arns.*", "aws_cloudfront_distribution.test", names.AttrARN),
					resource.TestCheckResourceAttr(cfdScopeName, "scope_configuration.0.resource_scope.0.include.0.explicit_arns.#", "1"),
					resource.TestCheckResourceAttr(cfdScopeName, "scope_configuration.0.resource_scope.0.include.0.expression.#", "0"),
				),
			},
			{
				ResourceName:                         albScopeName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(albScopeName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ResourceName:                         cfdScopeName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(cfdScopeName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				// ARNs moved to the exclude set, expression dropped.
				Config: testAccScopeConfig_explicitARNs(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, albScopeName, &v),
					testAccCheckScopeExists(ctx, t, cfdScopeName, &v),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.include.#", "0"),
					resource.TestCheckTypeSetElemAttrPair(albScopeName, "scope_configuration.0.resource_scope.0.exclude.0.explicit_arns.*", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(albScopeName, "scope_configuration.0.resource_scope.0.exclude.0.expression.#", "0"),
					resource.TestCheckResourceAttr(albScopeName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(cfdScopeName, "scope_configuration.0.resource_scope.0.include.#", "0"),
					resource.TestCheckTypeSetElemAttrPair(cfdScopeName, "scope_configuration.0.resource_scope.0.exclude.0.explicit_arns.*", "aws_cloudfront_distribution.test", names.AttrARN),
					resource.TestCheckResourceAttr(cfdScopeName, names.AttrVersion, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(albScopeName, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction(cfdScopeName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         albScopeName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(albScopeName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

// Every expression shape the API accepts: a single criteria, and, or and not
// operators, the operand limit, and tag maps of every size.
func testAccScope_expressions(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"
	prefix := "scope_configuration.0.resource_scope.0.include.0.expression.0."

	var twenty strings.Builder
	for i := range 20 {
		fmt.Fprintf(&twenty, `
              criteria {
                tags = {
                  nsm-acctest-%[1]d = %[2]q
                }
              }
`, i, rName)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_expression(rName, fmt.Sprintf(`
            criteria {
              tags = {
                nsm-acctest = %[1]q
                nsm-env     = "test"
                nsm-team    = "edge"
              }
            }
`, rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"criteria.0.tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, prefix+"criteria.0.tags.nsm-env", "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				Config: testAccScopeConfig_expression(rName, fmt.Sprintf(`
            and {
              criteria {
                tags = {
                  nsm-acctest = %[1]q
                }
              }
              criteria {
                tags = {
                  nsm-env = "test"
                }
              }
            }
`, rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.0.criteria.#", "2"),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.0.criteria.0.tags.nsm-acctest", rName),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.0.criteria.1.tags.nsm-env", "test"),
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
				Config: testAccScopeConfig_expression(rName, fmt.Sprintf(`
            or {
              criteria {
                tags = {
                  nsm-acctest = %[1]q
                }
              }
              criteria {
                tags = {
                  nsm-acctest-alt = %[1]q
                }
              }
            }
`, rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.#", "0"),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.0.criteria.#", "2"),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.0.criteria.1.tags.nsm-acctest-alt", rName),
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
			{
				Config: testAccScopeConfig_expression(rName, fmt.Sprintf(`
            not {
              criteria {
                tags = {
                  nsm-acctest = %[1]q
                }
              }
            }
`, rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.#", "0"),
					resource.TestCheckResourceAttr(resourceName, prefix+"not.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"not.0.criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"not.0.criteria.0.tags.nsm-acctest", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
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
				// The operand limit.
				Config: testAccScopeConfig_expression(rName, fmt.Sprintf(`
            and {
%[1]s
            }
`, twenty.String())),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.0.criteria.#", "20"),
					resource.TestCheckResourceAttr(resourceName, prefix+"and.0.criteria.19.tags.nsm-acctest-19", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
				),
			},
			{
				// A single-operand operator and an empty tag map are both valid.
				Config: testAccScopeConfig_expression(rName, `
            or {
              criteria {
                tags = {}
              }
            }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.0.criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.0.criteria.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"),
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
				// Back to a single criteria; a value-less tag matches on the key.
				Config: testAccScopeConfig_expression(rName, `
            criteria {
              tags = {
                nsm-acctest = ""
              }
            }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"or.#", "0"),
					resource.TestCheckResourceAttr(resourceName, prefix+"criteria.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"criteria.0.tags.nsm-acctest", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "7"),
				),
			},
		},
	})
}

// Every ip_address_type and scheme, alone, combined and empty.
func testAccScope_albConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"
	prefix := "scope_configuration.0.resource_scope.0.include.0.expression.0.criteria.0.alb_config.0."

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_albConfig(rName, `ip_address_type = "ipv4"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+names.AttrIPAddressType, "ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, prefix+"scheme"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
				Config: testAccScopeConfig_albConfig(rName, `ip_address_type = "dualstack"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+names.AttrIPAddressType, "dualstack"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
			},
			{
				Config: testAccScopeConfig_albConfig(rName, `ip_address_type = "dualstack-without-public-ipv4"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+names.AttrIPAddressType, "dualstack-without-public-ipv4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
				),
			},
			{
				Config: testAccScopeConfig_albConfig(rName, `scheme = "internet-facing"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, prefix+names.AttrIPAddressType),
					resource.TestCheckResourceAttr(resourceName, prefix+"scheme", "internet-facing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
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
				Config: testAccScopeConfig_albConfig(rName, `scheme = "internal"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"scheme", "internal"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
				),
			},
			{
				Config: testAccScopeConfig_albConfig(rName, `
                ip_address_type = "ipv4"
                scheme          = "internet-facing"
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, prefix+"scheme", "internet-facing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"),
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
				// An empty alb_config is accepted by the API and matches every load balancer.
				Config: testAccScopeConfig_albConfig(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.resource_scope.0.include.0.expression.0.criteria.0.alb_config.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, prefix+names.AttrIPAddressType),
					resource.TestCheckNoResourceAttr(resourceName, prefix+"scheme"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "7"),
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

func testAccScope_publish(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Created as a draft: no published version exists yet.
				Config: testAccScopeConfig_published(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`scope:[0-9a-z]+$`))),
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
				// Publishing the draft keeps its version.
				Config: testAccScopeConfig_published(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				// Unpublishing an active scope saves a draft on top of it.
				Config: testAccScopeConfig_published(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
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
				Config: testAccScopeConfig_published(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				// Destroyed with a pending draft.
				Config: testAccScopeConfig_published(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
		},
	})
}

func testAccScope_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				Config: testAccScopeConfig_description(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
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
				Config: testAccScopeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
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

// The update token is read back before every write, so a change made outside
// Terraform does not fail the next update, and a draft saved outside
// Terraform is published back.
func testAccScope_updateToken(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					testAccCheckScopeUpdateOutOfBand(ctx, t, &v, true, "out of band"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// The out-of-band description is reverted and the new
				// configuration applied in a single update.
				Config: testAccScopeConfig_descriptionIncludeAll(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeCFD),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
					testAccCheckScopeUpdateOutOfBand(ctx, t, &v, false, "draft outside terraform"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				// The pending draft is published back.
				Config: testAccScopeConfig_descriptionIncludeAll(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.EntityStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "has_published_version", acctest.CtFalse),
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

// A new name, a change between global and regional resource types, and an
// added account filter all replace the scope.
func testAccScope_replace(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_includeAll(rName, scopeResourceTypeCFD),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccScopeConfig_includeAll(rNameUpdated, scopeResourceTypeCFD),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				// Global to regional.
				Config: testAccScopeConfig_includeAll(rNameUpdated, scopeResourceTypeALB, scopeResourceTypeEIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeALB),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				// Regional to global.
				Config: testAccScopeConfig_includeAll(rNameUpdated, scopeResourceTypeCFD),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					testAccCheckScopeIncludesAll(resourceName, scopeResourceTypeCFD),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				// Adding an account filter is planned as a replacement. The
				// account of the test is a single-account administrator, for
				// which the API rejects any account filter, so the replacement
				// itself fails.
				Config: testAccScopeConfig_accountFilter(rNameUpdated, "include_all = true"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ExpectError: regexache.MustCompile(`accountFilter\s+is\s+not\s+supported\s+for\s+this\s+account`),
			},
		},
	})
}

func testAccScope_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccScopeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func testAccScope_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	var twentyOne strings.Builder
	for i := range 21 {
		fmt.Fprintf(&twentyOne, `
              criteria {
                tags = {
                  k%[1]d = "v"
                }
              }
`, i)
	}
	var arns []string
	for i := range 101 {
		arns = append(arns, fmt.Sprintf(`"arn:%s:cloudfront::123456789012:distribution/E%013d"`, acctest.Partition(), i))
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Plan-time validation.
			{
				Config:      testAccScopeConfig_includeAll("-"+rName, scopeResourceTypeCFD),
				ExpectError: regexache.MustCompile(`must\s+start\s+with\s+an\s+alphanumeric\s+character`),
			},
			{
				Config:      testAccScopeConfig_includeAll(strings.Repeat("a", 129), scopeResourceTypeCFD),
				ExpectError: regexache.MustCompile(`string\s+length\s+must\s+be\s+between\s+1\s+and\s+128`),
			},
			{
				Config:      testAccScopeConfig_description(rName, strings.Repeat("a", 257)),
				ExpectError: regexache.MustCompile(`string\s+length\s+must\s+be\s+at\s+most\s+256`),
			},
			{
				Config:      testAccScopeConfig_description(rName, "not # allowed"),
				ExpectError: regexache.MustCompile(`must\s+contain\s+only\s+alphanumeric\s+characters`),
			},
			{
				Config:      testAccScopeConfig_includeAll(rName, "AWS::EC2::VPC"),
				ExpectError: regexache.MustCompile(`does\s+not\s+match\s+any\s+valid\s+values`),
			},
			{
				// Two identical blocks collapse into one set element, so the
				// duplicate differs in its selection.
				Config: fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::EC2::EIP"
      include_all   = true
    }

    resource_scope {
      resource_type = "AWS::EC2::EIP"

      include {
        expression {
          criteria {
            tags = {}
          }
        }
      }
    }
  }
}
`, rName),
				ExpectError: regexache.MustCompile(`must\s+not\s+contain\s+the\s+same\s+resource_type\s+more\s+than\s+once`),
			},
			{
				Config:      testAccScopeConfig_includeAll(rName, scopeResourceTypeCFD, scopeResourceTypeALB),
				ExpectError: regexache.MustCompile(`must\s+not\s+combine\s+global\s+and\s+regional\s+resource\s+types`),
			},
			{
				Config:      testAccScopeConfig_selection(rName, scopeResourceTypeCFD, "include_all = false"),
				ExpectError: regexache.MustCompile(`include_all[\s\S]*must\s+be\s+true\s+when\s+set`),
			},
			{
				Config:      testAccScopeConfig_selection(rName, scopeResourceTypeCFD, ""),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"include_all",\s+"include"\s+or\s+"exclude"\s+must\s+be\s+set`),
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeCFD, `
      include_all = true
      include {}
`),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"include_all",\s+"include"\s+or\s+"exclude"\s+must\s+be\s+set`),
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeCFD, `
      include {
        explicit_arns = ["arn:${data.aws_partition.current.partition}:cloudfront::123456789012:distribution/E1"]
      }
      exclude {
        explicit_arns = ["arn:${data.aws_partition.current.partition}:cloudfront::123456789012:distribution/E2"]
      }
`),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"include_all",\s+"include"\s+or\s+"exclude"\s+must\s+be\s+set`),
			},
			{
				Config:      testAccScopeConfig_expression(rName, ""),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"and",\s+"criteria",\s+"not"\s+or\s+"or"\s+must\s+be\s+set`),
			},
			{
				Config: testAccScopeConfig_expression(rName, `
            criteria {
              tags = {}
            }
            not {
              criteria {
                tags = {}
              }
            }
`),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"and",\s+"criteria",\s+"not"\s+or\s+"or"\s+must\s+be\s+set`),
			},
			{
				Config: testAccScopeConfig_expression(rName, `
            criteria {}
`),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"alb_config"\s+or\s+"tags"\s+must\s+be\s+set`),
			},
			{
				Config: testAccScopeConfig_expression(rName, `
            criteria {
              tags = {}
              alb_config {}
            }
`),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"alb_config"\s+or\s+"tags"\s+must\s+be\s+set`),
			},
			{
				Config: testAccScopeConfig_expression(rName, `
            and {}
`),
				ExpectError: regexache.MustCompile(`must\s+have\s+a\s+configuration\s+value\s+as\s+the\s+provider\s+has\s+marked\s+it\s+as\s+required`),
			},
			{
				Config: testAccScopeConfig_expression(rName, fmt.Sprintf(`
            and {
%[1]s
            }
`, twentyOne.String())),
				ExpectError: regexache.MustCompile(`at\s+most\s+20\s+elements`),
			},
			{
				Config: testAccScopeConfig_expression(rName, `
            not {
              criteria {
                tags = {}
              }
              criteria {
                tags = {}
              }
            }
`),
				ExpectError: regexache.MustCompile(`at\s+most\s+1\s+elements`),
			},
			{
				Config:      testAccScopeConfig_albConfig(rName, `ip_address_type = "ipv6"`),
				ExpectError: regexache.MustCompile(`does\s+not\s+match\s+any\s+valid\s+values`),
			},
			{
				Config:      testAccScopeConfig_albConfig(rName, `scheme = "public"`),
				ExpectError: regexache.MustCompile(`does\s+not\s+match\s+any\s+valid\s+values`),
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeCFD, `
      include {
        explicit_arns = ["not-an-arn"]
      }
`),
				ExpectError: regexache.MustCompile(`value\s+must\s+be\s+a\s+valid\s+ARN`),
			},
			{
				Config: testAccScopeConfig_selection(rName, scopeResourceTypeCFD, fmt.Sprintf(`
      include {
        explicit_arns = [%[1]s]
      }
`, strings.Join(arns, ", "))),
				ExpectError: regexache.MustCompile(`at\s+most\s+100\s+elements`),
			},
			{
				Config: testAccScopeConfig_accountFilter(rName, `
      include {
        account_ids = ["12345678901"]
      }
`),
				ExpectError: regexache.MustCompile(`must\s+be\s+a\s+valid\s+AWS\s+account\s+ID`),
			},
			{
				Config: testAccScopeConfig_accountFilter(rName, `
      exclude {
        organizational_units = ["r-abcd"]
      }
`),
				ExpectError: regexache.MustCompile(`must\s+be\s+an\s+organizational\s+unit\s+id`),
			},
			{
				Config: testAccScopeConfig_accountFilter(rName, `
      include_all = true
      exclude {
        account_ids = ["123456789012"]
      }
`),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"include_all",\s+"include"\s+or\s+"exclude"\s+must\s+be\s+set`),
			},
			{
				Config:      testAccScopeConfig_accountFilter(rName, ""),
				ExpectError: regexache.MustCompile(`(?i)exactly\s+one\s+of\s+"include_all",\s+"include"\s+or\s+"exclude"\s+must\s+be\s+set`),
			},
			{
				Config: fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q
}
`, rName),
				ExpectError: regexache.MustCompile(`scope_configuration[\s\S]*must\s+have\s+a\s+configuration\s+value`),
			},
			{
				Config: fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {}
}
`, rName),
				ExpectError: regexache.MustCompile(`resource_scope[\s\S]*must\s+have\s+a\s+configuration\s+value`),
			},
			// API validation: a single-account administrator cannot set an
			// account filter.
			{
				Config:      testAccScopeConfig_accountFilter(rName, "include_all = true"),
				ExpectError: regexache.MustCompile(`accountFilter\s+is\s+not\s+supported\s+for\s+this\s+account`),
			},
		},
	})
}

func testAccCheckScopeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_scope" {
				continue
			}

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

func testAccCheckScopeExists(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetScopeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindScopeByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckScopeIncludesAll checks that the scope has an include_all
// resource scope for the resource type.
func testAccCheckScopeIncludesAll(n, resourceType string) resource.TestCheckFunc {
	return resource.TestCheckTypeSetElemNestedAttrs(n, "scope_configuration.0.resource_scope.*", map[string]string{
		names.AttrResourceType: resourceType,
		"include_all":          acctest.CtTrue,
		"include.#":            "0",
		"exclude.#":            "0",
	})
}

// testAccCheckScopeUpdateOutOfBand changes the scope's description outside
// Terraform, published or as a draft, invalidating the update token.
func testAccCheckScopeUpdateOutOfBand(ctx context.Context, t *testing.T, v *networksecuritymanager.GetScopeOutput, publish bool, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		input := networksecuritymanager.UpdateScopeInput{
			IsPublished:      aws.Bool(publish),
			ScopeDescription: aws.String(description),
			ScopeIdentifier:  v.ScopeArn,
			UpdateToken:      v.UpdateToken,
		}
		_, err := conn.UpdateScope(ctx, &input)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.ListScopesInput{
		MaxResults: aws.Int32(1),
	}
	_, err := conn.ListScopes(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccScopeConfig_basic(rName string) string {
	return fmt.Sprintf(`
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

func testAccScopeConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name         = %[1]q
  description  = "Every distribution in the account"
  is_published = false

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"
      include_all   = true
    }
  }

  tags = {
    key1 = "value1"
  }
}
`, rName)
}

func testAccScopeConfig_includeAll(rName string, resourceTypes ...string) string {
	var resourceScopes strings.Builder
	for _, resourceType := range resourceTypes {
		fmt.Fprintf(&resourceScopes, `
    resource_scope {
      resource_type = %[1]q
      include_all   = true
    }
`, resourceType)
	}

	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
%[2]s
  }
}
`, rName, resourceScopes.String())
}

func testAccScopeConfig_selection(rName, resourceType, selection string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    resource_scope {
      resource_type = %[2]q
%[3]s
    }
  }
}
`, rName, resourceType, selection)
}

func testAccScopeConfig_expression(rName, expression string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
%[2]s
        }
      }
    }
  }
}
`, rName, expression)
}

func testAccScopeConfig_albConfig(rName, albConfig string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          criteria {
            alb_config {
%[2]s
            }
          }
        }
      }
    }
  }
}
`, rName, albConfig)
}

func testAccScopeConfig_explicitARNs(rName string, include bool) string {
	selection := "exclude"
	expression := ""
	if include {
		selection = "include"
		expression = fmt.Sprintf(`
        expression {
          and {
            criteria {
              tags = {
                nsm-acctest = %[1]q
              }
            }
            criteria {
              alb_config {
                scheme = "internal"
              }
            }
          }
        }
`, rName)
	}

	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  tags = {
    nsm-acctest = %[1]q
  }
}

resource "aws_cloudfront_distribution" "test" {
  enabled             = false
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = {
    nsm-acctest = %[1]q
  }
}

resource "aws_networksecuritymanager_scope" "alb" {
  name = "%[1]s-alb"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      %[2]s {
        explicit_arns = [aws_lb.test.arn]
%[3]s
      }
    }
  }
}

resource "aws_networksecuritymanager_scope" "cfd" {
  name = "%[1]s-cfd"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      %[2]s {
        explicit_arns = [aws_cloudfront_distribution.test.arn]
      }
    }
  }
}
`, rName, selection, expression))
}

func testAccScopeConfig_published(rName string, published bool) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name         = %[1]q
  is_published = %[2]t

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
`, rName, published)
}

func testAccScopeConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name        = %[1]q
  description = %[2]q

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
`, rName, description)
}

func testAccScopeConfig_descriptionIncludeAll(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name        = %[1]q
  description = %[2]q

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"
      include_all   = true
    }
  }
}
`, rName, description)
}

func testAccScopeConfig_accountFilter(rName, accountFilter string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    account_filter {
%[2]s
    }

    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"
      include_all   = true
    }
  }
}
`, rName, accountFilter)
}

// Every account_filter member and every transition between them. Needs the
// account to be an onboarded Network Security Manager administrator of an
// AWS Organizations organization; skipped otherwise.
func testAccScope_accountFilter(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetScopeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_scope.test"
	prefix := "scope_configuration.0.account_filter.0."

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckAdminAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_accountFilterOrganization(rName, "include_all = true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.account_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"include_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, prefix+"include.#", "0"),
					resource.TestCheckResourceAttr(resourceName, prefix+"exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
				Config: testAccScopeConfig_accountFilterOrganization(rName, `
      include {
        account_ids = [data.aws_caller_identity.current.account_id]
      }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, prefix+"include_all"),
					resource.TestCheckResourceAttr(resourceName, prefix+"include.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"include.0.account_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, prefix+"include.0.account_ids.*", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckNoResourceAttr(resourceName, prefix+"include.0.organizational_units"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccScopeConfig_accountFilterOrganization(rName, `
      include {
        account_ids          = [data.aws_caller_identity.current.account_id]
        organizational_units = [aws_organizations_organizational_unit.test.id]
      }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"include.0.account_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, prefix+"include.0.organizational_units.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, prefix+"include.0.organizational_units.*", "aws_organizations_organizational_unit.test", names.AttrID),
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
			{
				Config: testAccScopeConfig_accountFilterOrganization(rName, `
      exclude {
        organizational_units = [aws_organizations_organizational_unit.test.id]
      }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"include.#", "0"),
					resource.TestCheckResourceAttr(resourceName, prefix+"exclude.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, prefix+"exclude.0.account_ids"),
					resource.TestCheckResourceAttr(resourceName, prefix+"exclude.0.organizational_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
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
				Config: testAccScopeConfig_accountFilterOrganization(rName, `
      exclude {
        account_ids = [data.aws_caller_identity.current.account_id]
      }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"exclude.0.account_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
				),
			},
			{
				Config: testAccScopeConfig_accountFilterOrganization(rName, "include_all = true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, prefix+"include_all", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, prefix+"exclude.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Removing the account filter replaces the scope.
				Config: acctest.ConfigCompose(testAccScopeConfig_organizationBase(rName), testAccScopeConfig_includeAll(rName, scopeResourceTypeCFD)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope_configuration.0.account_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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

// testAccPreCheckAdminAccount skips the test unless the account is an onboarded
// Network Security Manager administrator.
func testAccPreCheckAdminAccount(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.ListAdminAccountsInput{}
	output, err := conn.ListAdminAccounts(ctx, &input)

	if err != nil || output == nil || len(output.AdminAccounts) == 0 {
		t.Skipf("skipping acceptance testing: the account is not an onboarded Network Security Manager administrator: %v", err)
	}
}

func testAccScopeConfig_organizationBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}
`, rName)
}

func testAccScopeConfig_accountFilterOrganization(rName, accountFilter string) string {
	return acctest.ConfigCompose(testAccScopeConfig_organizationBase(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_scope" "test" {
  name = %[1]q

  scope_configuration {
    account_filter {
%[2]s
    }

    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"
      include_all   = true
    }
  }
}
`, rName, accountFilter))
}
