// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_byARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// action
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.ListExact([]knownvalue.Check{}),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.NotNull(),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"stickiness": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrDuration: knownvalue.Null(),
								names.AttrEnabled:  knownvalue.Bool(false),
							}),
						}),
						"target_group": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:    knownvalue.NotNull(),
								names.AttrWeight: knownvalue.NotNull(),
							}),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("target_group").AtSliceIndex(0).AtMapKey(names.AttrARN),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("target_group_arn"),
						compare.ValuesSame(),
					),

					// condition
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("host_header", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"regex_values": knownvalue.Null(),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("example.com"),
								}),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_byListenerAndPriority(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// action
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.ListExact([]knownvalue.Check{}),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.NotNull(),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"stickiness": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrDuration: knownvalue.Null(),
								names.AttrEnabled:  knownvalue.Bool(false),
							}),
						}),
						"target_group": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:    knownvalue.NotNull(),
								names.AttrWeight: knownvalue.NotNull(),
							}),
						}),
					})),

					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("target_group").AtSliceIndex(0).AtMapKey(names.AttrARN),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("target_group_arn"),
						compare.ValuesSame(),
					),

					// condition
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("host_header", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"regex_values": knownvalue.Null(),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("example.com"),
								}),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionAuthenticateCognito(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.NotNull(),
							"authenticate_oidc":    knownvalue.ListExact([]knownvalue.Check{}),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.ListExact([]knownvalue.Check{}),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_cognito").AtSliceIndex(0), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_cognito").AtSliceIndex(0),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionAuthenticateOIDC(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.NotNull(),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.ListExact([]knownvalue.Check{}),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("authentication_request_extra_params"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("authentication_request_extra_params"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("authorization_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("authorization_endpoint"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrClientID),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrClientID),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrIssuer),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrIssuer),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("on_unauthenticated_request"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("on_unauthenticated_request"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrScope),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey(names.AttrScope),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("session_cookie_name"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("session_cookie_name"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("session_timeout"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("session_timeout"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("token_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("token_endpoint"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("user_info_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("authenticate_oidc").AtSliceIndex(0).AtMapKey("user_info_endpoint"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_actionAuthenticateJWTValidation(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionAuthenticateJWTValidation(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.NotNull(),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.ListExact([]knownvalue.Check{}),
							"jwt_validation":       knownvalue.NotNull(),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0).AtMapKey(names.AttrIssuer),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0).AtMapKey(names.AttrIssuer),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0).AtMapKey("jwks_endpoint"),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0).AtMapKey("jwks_endpoint"),
						compare.ValuesSame(),
					),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0).AtMapKey("additional_claim"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(
						dataSourceName,
						tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("jwt_validation").AtSliceIndex(0).AtMapKey("additional_claim"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrFormat: knownvalue.StringExact("string-array"),
								names.AttrName:   knownvalue.StringExact("claim_name1"),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact(acctest.CtValue1),
									knownvalue.StringExact(acctest.CtValue2),
								}),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrFormat: knownvalue.StringExact("single-string"),
								names.AttrName:   knownvalue.StringExact("claim_name2"),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact(acctest.CtValue1),
								}),
							}),
						}),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionFixedResponse(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.ListExact([]knownvalue.Check{}),
							"fixed_response":       knownvalue.NotNull(),
							"forward":              knownvalue.ListExact([]knownvalue.Check{}),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("fixed_response").AtSliceIndex(0), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("fixed_response").AtSliceIndex(0),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionForwardWeightedStickiness(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.ListExact([]knownvalue.Check{}),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.NotNull(),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
							"order":                knownvalue.NotNull(),
							"redirect":             knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"stickiness":   knownvalue.NotNull(),
						"target_group": knownvalue.NotNull(),
					})),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("stickiness").AtSliceIndex(0),
						resourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("stickiness").AtSliceIndex(0),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("forward").AtSliceIndex(0).AtMapKey("target_group"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"
	resourceName := "aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_actionRedirect(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "listener_arn", resourceName, "listener_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"authenticate_cognito": knownvalue.ListExact([]knownvalue.Check{}),
							"authenticate_oidc":    knownvalue.ListExact([]knownvalue.Check{}),
							"fixed_response":       knownvalue.ListExact([]knownvalue.Check{}),
							"forward":              knownvalue.ListExact([]knownvalue.Check{}),
							"jwt_validation":       knownvalue.ListExact([]knownvalue.Check{}),
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

					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("redirect").AtSliceIndex(0), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("redirect").AtSliceIndex(0),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionHostHeader(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("host_header", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"regex_values": knownvalue.Null(),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("example.com"),
									knownvalue.StringExact("www.example.com"),
								}),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionHostHeaderRegex(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionHostHeaderRegex(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("host_header", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"regex_values": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("^example\\.com$"),
									knownvalue.StringExact("^www[0-9]+\\.example\\.com$"),
								}),
								names.AttrValues: knownvalue.Null(),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionHTTPHeader(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("http_header", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"http_header_name": knownvalue.StringExact("X-Forwarded-For"),
								"regex_values":     knownvalue.Null(),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("192.168.1.*"),
									knownvalue.StringExact("10.0.0.*"),
								}),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionHTTPHeaderRegex(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionHTTPHeaderRegex(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("http_header", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"http_header_name": knownvalue.StringExact("User-Agent"),
								"regex_values": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("A.+"),
									knownvalue.StringExact("B.*C"),
								}),
								names.AttrValues: knownvalue.Null(),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionHTTPRequestMethod(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("http_request_method", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("GET"),
									knownvalue.StringExact("POST"),
								}),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionPathPattern(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("path_pattern", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"regex_values": knownvalue.Null(),
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("/public/*"),
									knownvalue.StringExact("/cgi-bin/*"),
								}),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_conditionPathPatternRegex(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionPathPatternRegex(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("path_pattern", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"regex_values": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("^\\/api\\/(.*)$"),
									knownvalue.StringExact("^\\/api2\\/(.*)$"),
								}),
								names.AttrValues: knownvalue.Null(),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionQueryString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("query_string", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_conditionSourceIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCondition), knownvalue.SetExact([]knownvalue.Check{
						expectKnownCondition("source_ip", knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrValues: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("192.168.0.0/16"),
									knownvalue.StringExact("dead:cafe::/64"),
								}),
							}),
						})),
					})),
				},
			},
		},
	})
}

func TestAccELBV2ListenerRuleDataSource_transform(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listenerRule awstypes.Rule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleDataSourceConfig_transform(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, dataSourceName, &listenerRule),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("transform"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrType: knownvalue.StringExact(string(awstypes.TransformTypeEnumHostHeaderRewrite)),
							"host_header_rewrite_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"rewrite": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"regex":   knownvalue.StringExact("^mywebsite-(.+).com$"),
											"replace": knownvalue.StringExact("internal.dev.$1.myweb.com"),
										}),
									}),
								}),
							}),
							"url_rewrite_config": knownvalue.ListExact([]knownvalue.Check{}),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrType:               knownvalue.StringExact(string(awstypes.TransformTypeEnumUrlRewrite)),
							"host_header_rewrite_config": knownvalue.ListExact([]knownvalue.Check{}),
							"url_rewrite_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"rewrite": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"regex":   knownvalue.StringExact("^/dp/([A-Za-z0-9]+)/?$"),
											"replace": knownvalue.StringExact("/product.php?id=$1"),
										}),
									}),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func expectKnownCondition(key string, check knownvalue.Check) knownvalue.Check {
	checks := map[string]knownvalue.Check{
		"host_header":         knownvalue.ListExact([]knownvalue.Check{}),
		"http_header":         knownvalue.ListExact([]knownvalue.Check{}),
		"http_request_method": knownvalue.ListExact([]knownvalue.Check{}),
		"path_pattern":        knownvalue.ListExact([]knownvalue.Check{}),
		"query_string":        knownvalue.ListExact([]knownvalue.Check{}),
		"source_ip":           knownvalue.ListExact([]knownvalue.Check{}),
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

func testAccListenerRuleDataSourceConfig_actionAuthenticateJWTValidation(rName, key, certificate string) string {
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
    type = "jwt-validation"

    jwt_validation {
      issuer        = "https://example.com"
      jwks_endpoint = "https://example.com/.well-known/jwks.json"
      additional_claim {
        format = "string-array"
        name   = "claim_name1"
        values = ["value1", "value2"]
      }
      additional_claim {
        format = "single-string"
        name   = "claim_name2"
        values = ["value1"]
      }
    }
  }

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }

  tags = {
    Name = %[1]q
  }
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

func testAccListenerRuleDataSourceConfig_conditionHostHeaderRegex(rName string) string {
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
      regex_values = ["^example\\.com$", "^www[0-9]+\\.example\\.com$"]
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

func testAccListenerRuleDataSourceConfig_conditionHTTPHeaderRegex(rName string) string {
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
      http_header_name = "User-Agent"
      regex_values     = ["A.+", "B.*C"]
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

func testAccListenerRuleDataSourceConfig_conditionPathPatternRegex(rName string) string {
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
      regex_values = ["^\\/api\\/(.*)$", "^\\/api2\\/(.*)$"]
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

func testAccListenerRuleDataSourceConfig_transform(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_baseWithHTTPListener(rName), `
data "aws_lb_listener_rule" "test" {
  arn = aws_lb_listener_rule.test.arn
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["*"]
    }
  }

  transform {
    type = "host-header-rewrite"
    host_header_rewrite_config {
      rewrite {
        regex   = "^mywebsite-(.+).com$"
        replace = "internal.dev.$1.myweb.com"
      }
    }
  }

  transform {
    type = "url-rewrite"
    url_rewrite_config {
      rewrite {
        regex   = "^/dp/([A-Za-z0-9]+)/?$"
        replace = "/product.php?id=$1"
      }
    }
  }
}
`)
}
