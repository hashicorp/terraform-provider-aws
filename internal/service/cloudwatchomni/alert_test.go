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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkAlertARN                = tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`alert/[0-9a-f-]+`))
	checkAlertARNAlternateRegion = tfknownvalue.RegionalARNAlternateRegionRegexp("cloudwatch", regexache.MustCompile(`alert/[0-9a-f-]+`))
	checkAlertID                 = knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-f-]+$`))
)

func testAccAlert_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Alert
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_alert.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlertDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Alert/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("alert_id"), checkAlertID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkAlertARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notification_rule"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notifications_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"telemetry_rule": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrCondition: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"comparator":         tfknownvalue.StringExact(awstypes.ComparatorGt),
											"critical_threshold": knownvalue.Float64Exact(0),
											"threshold_field":    knownvalue.Null(),
											"threshold_mode":     tfknownvalue.StringExact(awstypes.ThresholdModeCountOfResults),
											"warning_threshold":  knownvalue.Null(),
										}),
									}),
									"evaluation": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"interval_seconds": knownvalue.Int32Exact(60),
										}),
									}),
									"no_data": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"treat_as": tfknownvalue.StringExact(awstypes.AlertStateOk),
										}),
									}),
									"query": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											names.AttrExpression: knownvalue.StringExact("up"),
											"language":           tfknownvalue.StringExact(awstypes.QueryLanguagePromql),
										}),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("space_id"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Alert/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "space_id", "alert_id"),
				ImportStateVerifyIdentifierAttribute: "alert_id",
			},
		},
	})
}

func testAccAlert_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Alert
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_alert.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlertDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Alert/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceAlert, resourceName),
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

// testAccAlert_disappearsSpace verifies that deleting the parent space, which the service reports
// as AccessDeniedException rather than ResourceNotFoundException, removes the alert from state.
func testAccAlert_disappearsSpace(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Alert
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_alert.test"
	spaceResourceName := "aws_cloudwatchomni_space.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlertDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Alert/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceSpace, spaceResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(spaceResourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAlert_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Alert
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_alert.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlertDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Alert/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:            config.StringVariable(rName),
					"alert_name":               config.StringVariable(rName),
					"description":              config.StringVariable("description 1"),
					"critical_threshold":       config.IntegerVariable(10),
					"no_data_treat_as":         config.StringVariable(string(awstypes.AlertStateOk)),
					"notification_states":      config.ListVariable(config.ListVariable(config.StringVariable(string(awstypes.AlertStateCritical)))),
					"pending_duration_seconds": config.IntegerVariable(120),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notification_rule"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrTarget: knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrARN:  knownvalue.NotNull(),
									"metadata":     knownvalue.Null(),
									names.AttrType: tfknownvalue.StringExact(awstypes.NotificationTargetTypeSns),
								}),
							}),
							"trigger": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"state_values": knownvalue.SetExact([]knownvalue.Check{
										tfknownvalue.StringExact(awstypes.AlertStateCritical),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("telemetry_rule").AtSliceIndex(0).AtMapKey(names.AttrCondition), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"comparator":         tfknownvalue.StringExact(awstypes.ComparatorGt),
							"critical_threshold": knownvalue.Float64Exact(10),
							"threshold_field":    knownvalue.Null(),
							"threshold_mode":     tfknownvalue.StringExact(awstypes.ThresholdModeCountOfResults),
							"warning_threshold":  knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("telemetry_rule").AtSliceIndex(0).AtMapKey("evaluation").AtSliceIndex(0).AtMapKey("pending_duration_seconds"), knownvalue.Int32Exact(120)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("telemetry_rule").AtSliceIndex(0).AtMapKey("no_data"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"treat_as": tfknownvalue.StringExact(awstypes.AlertStateOk),
						}),
					})),
				},
			},
			{
				// Update every updatable argument in place.
				ConfigDirectory: config.StaticDirectory("testdata/Alert/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:             config.StringVariable(rName),
					"alert_name":                config.StringVariable(rName2),
					"description":               config.StringVariable("description 2"),
					"critical_threshold":        config.IntegerVariable(20),
					"warning_threshold":         config.IntegerVariable(5),
					"expression":                config.StringVariable("up == 0"),
					"interval_seconds":          config.IntegerVariable(120),
					"no_data_treat_as":          config.StringVariable(string(awstypes.AlertStateCritical)),
					"notification_states":       config.ListVariable(config.ListVariable(config.StringVariable(string(awstypes.AlertStateCritical)), config.StringVariable(string(awstypes.AlertStateWarning))), config.ListVariable(config.StringVariable(string(awstypes.AlertStateOk)))),
					"notifications_enabled":     config.BoolVariable(false),
					"pending_duration_seconds":  config.IntegerVariable(240),
					"recovery_duration_seconds": config.IntegerVariable(360),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notification_rule"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notifications_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("telemetry_rule").AtSliceIndex(0).AtMapKey(names.AttrCondition).AtSliceIndex(0).AtMapKey("warning_threshold"), knownvalue.Float64Exact(5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("telemetry_rule").AtSliceIndex(0).AtMapKey("evaluation"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"interval_seconds":          knownvalue.Int32Exact(120),
							"pending_duration_seconds":  knownvalue.Int32Exact(240),
							"recovery_duration_seconds": knownvalue.Int32Exact(360),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("telemetry_rule").AtSliceIndex(0).AtMapKey("query").AtSliceIndex(0).AtMapKey(names.AttrExpression), knownvalue.StringExact("up == 0")),
				},
			},
			{
				// Removing notification rules is an in-place update.
				ConfigDirectory: config.StaticDirectory("testdata/Alert/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"alert_name":         config.StringVariable(rName2),
					"description":        config.StringVariable("description 2"),
					"critical_threshold": config.IntegerVariable(20),
					"no_data_treat_as":   config.StringVariable(string(awstypes.AlertStateCritical)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notification_rule"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notifications_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				// Removing the description requires replacement.
				ConfigDirectory: config.StaticDirectory("testdata/Alert/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"alert_name":    config.StringVariable(rName2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
			{
				// Changing the access profile requires replacement.
				ConfigDirectory: config.StaticDirectory("testdata/Alert/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"alert_name":         config.StringVariable(rName2),
					"use_second_profile": config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("profile_id"), "aws_cloudwatchomni_access_profile.test2", tfjsonpath.New("profile_id"), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccCheckAlertDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_alert" {
				continue
			}

			_, err := tfcloudwatchomni.FindAlertByTwoPartKey(ctx, conn, rs.Primary.Attributes["space_id"], rs.Primary.Attributes["alert_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Alert %s still exists", rs.Primary.Attributes["alert_id"])
		}

		return nil
	}
}

func testAccCheckAlertExists(ctx context.Context, t *testing.T, n string, v *awstypes.Alert) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindAlertByTwoPartKey(ctx, conn, rs.Primary.Attributes["space_id"], rs.Primary.Attributes["alert_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
