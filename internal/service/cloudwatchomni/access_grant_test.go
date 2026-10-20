// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
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
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkAccessGrantARN                = tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`access-grant/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`))
	checkAccessGrantARNAlternateRegion = tfknownvalue.RegionalARNAlternateRegionRegexp("cloudwatch", regexache.MustCompile(`access-grant/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`))
	checkAccessGrantID                 = knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`))
)

func testAccAccessGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAccessGrantARN),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("domain_id"), "aws_cloudwatchomni_domain.test", tfjsonpath.New("domain_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("grant_id"), checkAccessGrantID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("grant_type"), tfknownvalue.StringExact(awstypes.AccessGrantTypeCustomerManaged)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("permission"), tfknownvalue.StringExact(awstypes.AccessGrantPermissionRead)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrincipal), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"principal_attribute": knownvalue.ListSizeExact(0),
						"principal_id":        knownvalue.NotNull(),
						"principal_type":      tfknownvalue.StringExact(awstypes.AccessGrantPrincipalTypeIamRole),
					})})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scoped_action"), knownvalue.ListSizeExact(0)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("space_id"), "aws_cloudwatchomni_space.test", tfjsonpath.New("space_id"), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "grant_id"),
				ImportStateVerifyIdentifierAttribute: "grant_id",
			},
		},
	})
}

func testAccAccessGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceAccessGrant, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccessGrant_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name1, name2 := rName, acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(name1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(name2),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name2)),
				},
			},
		},
	})
}

func testAccAccessGrant_permission(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/permission/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"permission":    config.StringVariable(string(awstypes.AccessGrantPermissionRead)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("permission"), tfknownvalue.StringExact(awstypes.AccessGrantPermissionRead)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/permission/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"permission":    config.StringVariable(string(awstypes.AccessGrantPermissionReadWriteDelete)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("permission"), tfknownvalue.StringExact(awstypes.AccessGrantPermissionReadWriteDelete)),
				},
			},
		},
	})
}

func testAccAccessGrant_principal(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/principal/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"principal_index": config.IntegerVariable(0),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_iam_role.principal[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/principal/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"principal_index": config.IntegerVariable(1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_iam_role.principal[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccAccessGrant_principalAccessProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/principal_access_profile/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAccessGrantARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("grant_id"), checkAccessGrantID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("permission"), tfknownvalue.StringExact(awstypes.AccessGrantPermissionRead)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrincipal), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"principal_attribute": knownvalue.ListSizeExact(0),
						"principal_id":        knownvalue.NotNull(),
						"principal_type":      tfknownvalue.StringExact(awstypes.AccessGrantPrincipalTypeAccessProfile),
					})})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrPrincipal).AtSliceIndex(0).AtMapKey("principal_id"), "aws_cloudwatchomni_access_profile.test", tfjsonpath.New("profile_id"), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/principal_access_profile/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "grant_id"),
				ImportStateVerifyIdentifierAttribute: "grant_id",
			},
		},
	})
}

func testAccAccessGrant_scopedAction(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AccessGrant
	resourceName := "aws_cloudwatchomni_access_grant.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/scoped_action/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("permission"), tfknownvalue.StringExact(awstypes.AccessGrantPermissionCustom)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scoped_action"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrActions: knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("cloudwatch:GetOmniDashboard"),
							knownvalue.StringExact("cloudwatch:ListOmniDashboards"),
						}),
						// "context_conditions": knownvalue.Null(),
						"resource": knownvalue.ListSizeExact(0),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/scoped_action/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "grant_id"),
				ImportStateVerifyIdentifierAttribute: "grant_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/scoped_action_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scoped_action"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrActions: knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("cloudwatch:GetRecords"),
						}),
						// "context_conditions": knownvalue.MapExact(map[string]knownvalue.Check{
						// 	"channelNames": knownvalue.ListExact([]knownvalue.Check{
						// 		knownvalue.StringExact("channel-a"),
						// 		knownvalue.StringExact("channel-b"),
						// 	}),
						// }),
						"resource": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"resource_arns":        knownvalue.Null(),
							names.AttrResourceType: knownvalue.StringExact("DataSet"),
							"row_scope_group": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"row_scope": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrField: knownvalue.StringExact("serviceName"),
											"operator":      tfknownvalue.StringExact(awstypes.RowScopeOperatorIn),
											names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
												knownvalue.StringExact("service-a"),
												knownvalue.StringExact("service-b"),
											}),
										}),
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrField:  knownvalue.StringExact("accountId"),
											"operator":       tfknownvalue.StringExact(awstypes.RowScopeOperatorIn),
											names.AttrValues: knownvalue.SetSizeExact(1),
										}),
									}),
								}),
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"row_scope": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrField: knownvalue.StringExact("serviceName"),
											"operator":      tfknownvalue.StringExact(awstypes.RowScopeOperatorIn),
											names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
												knownvalue.StringExact("service-c"),
											}),
										}),
									}),
								}),
							}),
							"signal_types": knownvalue.ListExact([]knownvalue.Check{
								tfknownvalue.StringExact(awstypes.SignalTypeLogs),
							}),
							names.AttrTags: knownvalue.MapExact(map[string]knownvalue.Check{
								"env": knownvalue.StringExact("test"),
							}),
						})}),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/scoped_action_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "grant_id"),
				ImportStateVerifyIdentifierAttribute: "grant_id",
			},
		},
	})
}

func testAccAccessGrant_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/validation_principal_attribute_type/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`principal_attribute must not be configured`),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/validation_principal_missing/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`At least one of.*principal_attribute.*principal_id`),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/validation_custom_without_scoped_action/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`scoped_action must be configured`),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/validation_row_scope_group_without_signal_types/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`These attributes must be configured together`),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessGrant/validation_action_name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

func testAccCheckAccessGrantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_access_grant" {
				continue
			}

			_, err := tfcloudwatchomni.FindAccessGrantByID(ctx, conn, rs.Primary.Attributes["grant_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Access Grant %s still exists", rs.Primary.Attributes["grant_id"])
		}

		return nil
	}
}

func testAccCheckAccessGrantExists(ctx context.Context, t *testing.T, n string, v *awstypes.AccessGrant) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindAccessGrantByID(ctx, conn, rs.Primary.Attributes["grant_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
