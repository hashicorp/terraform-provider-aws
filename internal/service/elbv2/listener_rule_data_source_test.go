// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
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
							names.AttrType:         knownvalue.NotNull(),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
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

func TestAccELBV2ListenerRuleDataSource_byListenerAndPriority(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_byListenerAndPriority(rName),
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
							names.AttrType:         knownvalue.NotNull(),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
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

func TestAccELBV2ListenerRuleDataSource_actionAuthenticateCognito(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionAuthenticateCognito(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.NotNull(),
							"authenticate_oidc":    knownvalue.Null(),
							"fixed_response":       knownvalue.Null(),
							"forward":              knownvalue.Null(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.Null(),
							names.AttrType:         knownvalue.NotNull(),
						}),
						knownvalue.NotNull(),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						compare.ValuesSame(),
					),

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_cognito"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_cognito"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_cognito").AtSliceIndex(0),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_actionAuthenticateOIDC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionAuthenticateOIDC(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.Null(),
							"authenticate_oidc":    knownvalue.NotNull(),
							"fixed_response":       knownvalue.Null(),
							"forward":              knownvalue.Null(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.Null(),
							names.AttrType:         knownvalue.NotNull(),
						}),
						knownvalue.NotNull(),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						compare.ValuesSame(),
					),

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("authentication_request_extra_params"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("authentication_request_extra_params"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("authorization_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("authorization_endpoint"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey(names.AttrClientID),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrClientID),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey(names.AttrIssuer),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrIssuer),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("on_unauthenticated_request"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("on_unauthenticated_request"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey(names.AttrScope),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrScope),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("session_cookie_name"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("session_cookie_name"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("session_timeout"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("session_timeout"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("token_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("token_endpoint"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtMapKey("user_info_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("user_info_endpoint"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_actionFixedResponse(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_actionFixedResponse(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.Null(),
							"authenticate_oidc":    knownvalue.Null(),
							"fixed_response":       knownvalue.NotNull(),
							"forward":              knownvalue.Null(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.Null(),
							names.AttrType:         knownvalue.NotNull(),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						compare.ValuesSame(),
					),

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("fixed_response"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("fixed_response"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("fixed_response").AtSliceIndex(0),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_actionForwardWeightedStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionForwardWeightedStickiness(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.Null(),
							"authenticate_oidc":    knownvalue.Null(),
							"fixed_response":       knownvalue.Null(),
							"forward":              knownvalue.NotNull(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.Null(),
							names.AttrType:         knownvalue.NotNull(),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						compare.ValuesSame(),
					),

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"stickiness":   knownvalue.NotNull(),
						"target_group": knownvalue.NotNull(),
					})),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtMapKey("stickiness"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("stickiness").AtSliceIndex(0),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtMapKey("target_group"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("target_group"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_actionRedirect(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_actionRedirect(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.Null(),
							"authenticate_oidc":    knownvalue.Null(),
							"fixed_response":       knownvalue.Null(),
							"forward":              knownvalue.Null(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.NotNull(),
							names.AttrType:         knownvalue.NotNull(),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("order"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey(names.AttrType),
						compare.ValuesSame(),
					),

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("redirect"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("redirect"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("redirect").AtSliceIndex(0),
						compare.ValuesSame(),
					),
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

func TestAccELBV2ListenerRuleDataSource_conditionHTTPHeader(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_conditionHTTPHeader(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("http_header", knownvalue.ObjectExact(map[string]knownvalue.Check{
							"http_header_name": knownvalue.StringExact("X-Forwarded-For"),
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("192.168.1.*"),
								knownvalue.StringExact("10.0.0.*"),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionHTTPRequestMethod(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_conditionHTTPRequestMethod(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("http_request_method", knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("GET"),
								knownvalue.StringExact("POST"),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionPathPattern(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_conditionPathPattern(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("path_pattern", knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("/public/*"),
								knownvalue.StringExact("/cgi-bin/*"),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionQueryString(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_conditionQueryString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("query_string", knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrKey:   knownvalue.StringExact("one"),
									names.AttrValue: knownvalue.StringExact("un"),
								}),
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrKey:   knownvalue.StringExact("two"),
									names.AttrValue: knownvalue.StringExact("deux"),
								}),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionSourceIP(t *testing.T) {
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
				Config: testAccListenerRuleDataSourceConfig_conditionSourceIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("source_ip", knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("192.168.0.0/16"),
								knownvalue.StringExact("dead:cafe::/64"),
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

func testAccListenerRuleDataSourceConfig_byListenerAndPriority(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_baseWithHTTPListener(rName), `
data "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener_rule.test.listener_arn
  priority     = aws_lb_listener_rule.test.priority
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

func testAccListenerRuleDataSourceConfig_actionAuthenticateCognito(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerRuleConfig_baseWithHTTPSListener(rName, key, certificate),
		fmt.Sprintf(`
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type = "authenticate-cognito"

    authenticate_cognito {
      user_pool_arn       = aws_cognito_user_pool.test.arn
      user_pool_client_id = aws_cognito_user_pool_client.test.id
      user_pool_domain    = aws_cognito_user_pool_domain.test.domain

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }
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

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                                 = %[1]q
  user_pool_id                         = aws_cognito_user_pool.test.id
  generate_secret                      = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_scopes                 = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]
  callback_urls                        = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri                 = "https://www.example.com/redirect"
  logout_urls                          = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName))
}

func testAccListenerRuleDataSourceConfig_actionAuthenticateOIDC(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerRuleConfig_baseWithHTTPSListener(rName, key, certificate),
		fmt.Sprintf(`
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }
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

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                                 = %[1]q
  user_pool_id                         = aws_cognito_user_pool.test.id
  generate_secret                      = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_scopes                 = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]
  callback_urls                        = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri                 = "https://www.example.com/redirect"
  logout_urls                          = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName))
}

func testAccListenerRuleDataSourceConfig_actionFixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_baseWithHTTPListener(rName), `
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Hello"
      status_code  = "200"
    }
  }

  condition {
    host_header {
      values = ["example.com"]
    }
  }
}
`)
}

func testAccListenerRuleDataSourceConfig_actionForwardWeightedStickiness(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test1.arn
        weight = 1
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 1
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }

  condition {
    host_header {
      values = ["example.com"]
    }
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test1.arn
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}

resource "aws_lb_target_group" "test1" {
  name     = "%[1]s-1"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}

resource "aws_lb_target_group" "test2" {
  name     = "%[1]s-2"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName))
}

func testAccListenerRuleDataSourceConfig_actionRedirect(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_baseWithHTTPListener(rName), `
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100

  action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      query       = "param1=value1"
      status_code = "HTTP_301"
    }
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

func testAccListenerRuleDataSourceConfig_conditionHTTPHeader(rName string) string {
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
    http_header {
      http_header_name = "X-Forwarded-For"
      values           = ["192.168.1.*", "10.0.0.*"]
    }
  }
}
`)
}

func testAccListenerRuleDataSourceConfig_conditionHTTPRequestMethod(rName string) string {
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
    http_request_method {
      values = ["GET", "POST"]
    }
  }
}
`)
}

func testAccListenerRuleDataSourceConfig_conditionPathPattern(rName string) string {
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
    path_pattern {
      values = ["/public/*", "/cgi-bin/*"]
    }
  }
}
`)
}

func testAccListenerRuleDataSourceConfig_conditionQueryString(rName string) string {
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
    query_string {
      key   = "one"
      value = "un"
    }
    query_string {
      key   = "two"
      value = "deux"
    }
  }
}
`)
}

func testAccListenerRuleDataSourceConfig_conditionSourceIP(rName string) string {
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
    source_ip {
      values = [
        "192.168.0.0/16",
        "dead:cafe::/64",
      ]
    }
  }
}
`)
}
