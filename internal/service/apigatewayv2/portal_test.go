// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
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
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2Portal_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorization"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"cognito_config": knownvalue.ListExact([]knownvalue.Check{}),
							"none": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("endpoint_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"acm_managed": knownvalue.ListExact([]knownvalue.Check{}),
							"none": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("included_portal_product_arns"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_modified"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_published"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_published_description"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logo_uri"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_arn"), tfknownvalue.RegionalARNRegexp("apigateway", regexache.MustCompile(`/portals/[0-9A-Za-z]{10,30}`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrDescription: knownvalue.Null(),
							names.AttrDisplayName: knownvalue.StringExact(rName),
							"theme": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"custom_colors": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"accent_color":           knownvalue.StringExact("#FF9900"),
											"background_color":       knownvalue.StringExact("#FFFFFF"),
											"error_validation_color": knownvalue.StringExact("#D13212"),
											"header_color":           knownvalue.StringExact("#232F3E"),
											"navigation_color":       knownvalue.StringExact("#232F3E"),
											"text_color":             knownvalue.StringExact("#16191F"),
										}),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_default_domain_name"), knownvalue.StringRegexp(regexache.MustCompile(`^p-[0-9a-z]+\.apigw-portal\.`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_domain_hosted_zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_id"), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9A-Za-z]{10,30}$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_description"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_status"), tfknownvalue.StringExact(awstypes.PublishStatusDisabled)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rum_app_monitor_name"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_id"),
				ImportStateVerifyIdentifierAttribute: "portal_id",
			},
		},
	})
}

func TestAccAPIGatewayV2Portal_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfapigatewayv2.ResourcePortal, resourceName),
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

func TestAccAPIGatewayV2Portal_portalContent(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	portalID := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/portalContent/"),
				ConfigVariables: config.Variables{
					"display_name":       config.StringVariable(rName),
					"portal_description": config.StringVariable("original"),
					"accent_color":       config.StringVariable("#FF9900"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					portalID.AddStateValue(resourceName, tfjsonpath.New("portal_id")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey(names.AttrDescription), knownvalue.StringExact("original")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey(names.AttrDisplayName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey("theme").AtSliceIndex(0).AtMapKey("custom_colors").AtSliceIndex(0).AtMapKey("accent_color"), knownvalue.StringExact("#FF9900")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/portalContent/"),
				ConfigVariables: config.Variables{
					"display_name":       config.StringVariable(rNameUpdated),
					"portal_description": config.StringVariable("updated"),
					"accent_color":       config.StringVariable("#0073BB"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					portalID.AddStateValue(resourceName, tfjsonpath.New("portal_id")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey(names.AttrDescription), knownvalue.StringExact("updated")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey(names.AttrDisplayName), knownvalue.StringExact(rNameUpdated)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey("theme").AtSliceIndex(0).AtMapKey("custom_colors").AtSliceIndex(0).AtMapKey("accent_color"), knownvalue.StringExact("#0073BB")),
				},
			},
			// Removing the description clears it.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/portalContent/"),
				ConfigVariables: config.Variables{
					"display_name": config.StringVariable(rNameUpdated),
					"accent_color": config.StringVariable("#0073BB"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					portalID.AddStateValue(resourceName, tfjsonpath.New("portal_id")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccAPIGatewayV2Portal_publish(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	lastPublished := statecheck.CompareValue(compare.ValuesDiffer())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Create published.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/publish/"),
				ConfigVariables: config.Variables{
					"display_name":        config.StringVariable(rName),
					"publish":             config.BoolVariable(true),
					"publish_description": config.StringVariable("first"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					lastPublished.AddStateValue(resourceName, tfjsonpath.New("last_published")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_published"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_published_description"), knownvalue.StringExact("first")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_status"), tfknownvalue.StringExact(awstypes.PublishStatusPublished)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/publish/"),
				ConfigVariables: config.Variables{
					"display_name":        config.StringVariable(rName),
					"publish":             config.BoolVariable(true),
					"publish_description": config.StringVariable("first"),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_id"),
				ImportStateVerifyIdentifierAttribute: "portal_id",
				ImportStateVerifyIgnore:              []string{"publish_description"},
			},
			// A content change to a published portal publishes it again.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/publish/"),
				ConfigVariables: config.Variables{
					"display_name":        config.StringVariable(rNameUpdated),
					"publish":             config.BoolVariable(true),
					"publish_description": config.StringVariable("second"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					lastPublished.AddStateValue(resourceName, tfjsonpath.New("last_published")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_published_description"), knownvalue.StringExact("second")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_content").AtSliceIndex(0).AtMapKey(names.AttrDisplayName), knownvalue.StringExact(rNameUpdated)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_status"), tfknownvalue.StringExact(awstypes.PublishStatusPublished)),
				},
			},
			// Disable.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/publish/"),
				ConfigVariables: config.Variables{
					"display_name": config.StringVariable(rNameUpdated),
					"publish":      config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_status"), tfknownvalue.StringExact(awstypes.PublishStatusDisabled)),
				},
			},
			// Publish again, then destroy the published portal.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/publish/"),
				ConfigVariables: config.Variables{
					"display_name": config.StringVariable(rNameUpdated),
					"publish":      config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("publish_status"), tfknownvalue.StringExact(awstypes.PublishStatusPublished)),
				},
			},
		},
	})
}

func TestAccAPIGatewayV2Portal_rumAppMonitorName(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/rumAppMonitorName/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"rum_app_monitor": config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rum_app_monitor_name"), knownvalue.StringExact(rName)),
				},
			},
			// Removing the argument clears the RUM app monitor.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/rumAppMonitorName/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:   config.StringVariable(rName),
					"rum_app_monitor": config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
					testAccCheckPortalRUMAppMonitorNameEmpty(&v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rum_app_monitor_name"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccAPIGatewayV2Portal_logoURI(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/logoURI/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"logo":          config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logo_uri"), knownvalue.StringExact(fmt.Sprintf("s3://%s/logo.png", rName))),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/logoURI/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"logo":          config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logo_uri"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccAPIGatewayV2Portal_authorizationCognito(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/authorizationCognito/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorization"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"cognito_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"app_client_id":    knownvalue.NotNull(),
									"user_pool_arn":    tfknownvalue.RegionalARNRegexp("cognito-idp", regexache.MustCompile(`userpool/.+`)),
									"user_pool_domain": knownvalue.StringRegexp(regexache.MustCompile(`^` + rName + `\.auth\.`)),
								}),
							}),
							"none": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/authorizationCognito/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_id"),
				ImportStateVerifyIdentifierAttribute: "portal_id",
			},
		},
	})
}

func TestAccAPIGatewayV2Portal_endpointConfigurationACMManaged(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalOutput
	resourceName := "aws_apigatewayv2_portal.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/endpointConfigurationACMManaged/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"root_domain":   config.StringVariable(rootDomain),
					"domain_name":   config.StringVariable(domainName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("endpoint_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"acm_managed": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrCertificateARN: tfknownvalue.RegionalARNRegexp("acm", regexache.MustCompile(`certificate/.+`)),
									names.AttrDomainName:     knownvalue.StringExact(domainName),
								}),
							}),
							"none": knownvalue.ListExact([]knownvalue.Check{}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_default_domain_name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_domain_hosted_zone_id"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Portal/endpointConfigurationACMManaged/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"root_domain":   config.StringVariable(rootDomain),
					"domain_name":   config.StringVariable(domainName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_id"),
				ImportStateVerifyIdentifierAttribute: "portal_id",
			},
		},
	})
}

func testAccCheckPortalDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_portal" {
				continue
			}

			_, err := tfapigatewayv2.FindPortalByID(ctx, conn, rs.Primary.Attributes["portal_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Portal %s still exists", rs.Primary.Attributes["portal_id"])
		}

		return nil
	}
}

func testAccCheckPortalExists(ctx context.Context, t *testing.T, n string, v *apigatewayv2.GetPortalOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindPortalByID(ctx, conn, rs.Primary.Attributes["portal_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPortalRUMAppMonitorNameEmpty(v *apigatewayv2.GetPortalOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got := v.RumAppMonitorName; got != nil && *got != "" {
			return fmt.Errorf("RumAppMonitorName = %q, want empty", *got)
		}

		return nil
	}
}
