// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2ListenerRuleDataSource_byARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_byARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// action
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.Null(),
							"authenticate_oidc":    knownvalue.Null(),
							"fixed_response":       knownvalue.Null(),
							"forward":              knownvalue.NotNull(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.Null(),
							"target_group_arn":     knownvalue.NotNull(),
							names.AttrType:         knownvalue.NotNull(),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("target_group_arn"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("target_group_arn"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						compare.ValuesSame(),
					),

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"stickiness": knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrDuration: knownvalue.Null(),
							names.AttrEnabled:  knownvalue.Bool(false),
						}),
						"target_group": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:    knownvalue.NotNull(),
								names.AttrWeight: knownvalue.NotNull(),
							}),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtMapKey("target_group").AtSliceIndex(0).AtMapKey(names.AttrARN),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("target_group_arn"),
						compare.ValuesSame(),
					),

					// condition
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("host_header", knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("example.com"),
							}),
						})),
					})),

					// tags
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionHostHeader(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionHostHeader(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("host_header", knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("example.com"),
								knownvalue.StringExact("www.example.com"),
							}),
						})),
					})),
				},
			},
		},
	})
}

func expectKnownCondition(key string, check knownvalue.Check) knownvalue.Check {
	checks := map[string]knownvalue.Check{
		"host_header":         knownvalue.Null(),
		"http_header":         knownvalue.Null(),
		"http_request_method": knownvalue.Null(),
		"path_pattern":        knownvalue.Null(),
		"query_string":        knownvalue.Null(),
		"source_ip":           knownvalue.Null(),
	}
	checks[key] = check
	return knownvalue.ObjectExact(checks)
}

func testAccListenerRuleDataSourceConfig_byARN(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_baseWithHTTPListener(rName), `
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    host_header {
      values = ["example.com"]
    }
  }
}
`)
}

func testAccListenerRuleDataSourceConfig_conditionHostHeader(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_baseWithHTTPListener(rName), `
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    host_header {
      values = ["example.com", "www.example.com"]
    }
  }
}
`)
}
