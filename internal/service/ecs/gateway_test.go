// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
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
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestGatewayStatusError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		gateway  *awstypes.Gateway
		want     string
	}{
		{
			testName: "no reason and no failures",
			gateway:  &awstypes.Gateway{},
		},
		{
			testName: "reason only",
			gateway: &awstypes.Gateway{
				StatusReason: aws.String("update failed"),
			},
			want: "update failed",
		},
		{
			testName: "reason and failures",
			gateway: &awstypes.Gateway{
				StatusReason: aws.String("some load balancers were not updated"),
				Failures: []awstypes.GatewayResourceFailure{
					{
						Arn: aws.String("arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/a/1"), // lintignore:AWSAT003,AWSAT005
						Errors: []awstypes.GatewayOperationError{
							{
								Code:      aws.String("AccessDenied"),
								Message:   aws.String("not authorized"),
								Operation: aws.String("elasticloadbalancing:ModifyLoadBalancerAttributes"),
							},
							{
								Code:      aws.String("Throttling"),
								Message:   aws.String("rate exceeded"),
								Operation: aws.String("elasticloadbalancing:SetWebAcl"),
							},
						},
					},
				},
			},
			want: "some load balancers were not updated\n" +
				"arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/a/1: elasticloadbalancing:ModifyLoadBalancerAttributes AccessDenied: not authorized\n" + // lintignore:AWSAT003,AWSAT005
				"arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/a/1: elasticloadbalancing:SetWebAcl Throttling: rate exceeded", // lintignore:AWSAT003,AWSAT005
		},
		{
			testName: "failures only",
			gateway: &awstypes.Gateway{
				Failures: []awstypes.GatewayResourceFailure{
					{
						Arn: aws.String("arn:aws:wafv2:us-west-2:123456789012:regional/webacl/a/1"), // lintignore:AWSAT003,AWSAT005
						Errors: []awstypes.GatewayOperationError{
							{
								Code:      aws.String("WAFNonexistentItemException"),
								Message:   aws.String("web ACL not found"),
								Operation: aws.String("wafv2:AssociateWebACL"),
							},
						},
					},
				},
			},
			want: "arn:aws:wafv2:us-west-2:123456789012:regional/webacl/a/1: wafv2:AssociateWebACL WAFNonexistentItemException: web ACL not found", // lintignore:AWSAT003,AWSAT005
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := tfecs.GatewayStatusError(testCase.gateway)

			if testCase.want == "" {
				if err != nil {
					t.Errorf("got %q, want no error", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("got no error, want %q", testCase.want)
			}
			if got := err.Error(); got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestAccECSGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecs", regexache.MustCompile(`gateway/`+rName+`/[0-9a-f-]+$`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_type"), tfknownvalue.StringExact(awstypes.AccessTypePublic)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDeletionProtection), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPropagateTags), tfknownvalue.StringExact(awstypes.PropagateGatewayTagsGateway)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("access_logs"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("http_desync"), tfknownvalue.StringExact(awstypes.GatewayAlbDesyncMitigationModeDefensive)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("max_alb_count"), knownvalue.Int32Exact(10)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("scheme"), tfknownvalue.StringExact(awstypes.GatewayElbSchemeInternetFacing)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("test_traffic_rules"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("web_acl_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath("inbound_security_group"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.GatewayElbIpAddressTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath(names.AttrSubnets), knownvalue.SetSizeExact(2)),
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
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccECSGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceGateway, resourceName),
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

func TestAccECSGateway_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_deletionProtectionOmitted(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDeletionProtection), knownvalue.Bool(true)),
				},
			},
			{
				Config:      testAccGatewayConfig_deletionProtectionOmitted(rName),
				Destroy:     true,
				ExpectError: regexache.MustCompile(`Gateway\s+has\s+deletion\s+protection\s+enabled`),
			},
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDeletionProtection), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccECSGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
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
			{
				Config: testAccGatewayConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccGatewayConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAccECSGateway_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_update(rName, "monitor", 2, "NONE", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("infrastructure_role_arn"), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPropagateTags), tfknownvalue.StringExact(awstypes.PropagateGatewayTagsNone)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("http_desync"), tfknownvalue.StringExact(awstypes.GatewayAlbDesyncMitigationModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("max_alb_count"), knownvalue.Int32Exact(2)),
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
				Config: testAccGatewayConfig_update(rName, "strictest", 3, "GATEWAY", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("infrastructure_role_arn"), "aws_iam_role.test2", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPropagateTags), tfknownvalue.StringExact(awstypes.PropagateGatewayTagsGateway)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("http_desync"), tfknownvalue.StringExact(awstypes.GatewayAlbDesyncMitigationModeStrictest)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("max_alb_count"), knownvalue.Int32Exact(3)),
				},
			},
		},
	})
}

func TestAccECSGateway_Configuration_full(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"
	fullStateChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_type"), tfknownvalue.StringExact(awstypes.AccessTypePrivate)),
		statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("access_logs"), knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				names.AttrDestination: knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"s3": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrBucket: knownvalue.StringExact(rName),
								names.AttrPrefix: knownvalue.StringExact("gateway"),
							}),
						}),
					}),
				}),
			}),
		})),
		statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("http_desync"), tfknownvalue.StringExact(awstypes.GatewayAlbDesyncMitigationModeStrictest)),
		statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("max_alb_count"), knownvalue.Int32Exact(4)),
		statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("scheme"), tfknownvalue.StringExact(awstypes.GatewayElbSchemeInternal)),
		statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("test_traffic_rules"), knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				names.AttrHeader: knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrName:   knownvalue.StringExact("x-canary"),
						names.AttrValues: knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("true")}),
					}),
				}),
			}),
		})),
		statecheck.CompareValuePairs(resourceName, testAccGatewayALBPath("web_acl_arn"), "aws_wafv2_web_acl.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
		statecheck.CompareValuePairs(resourceName, testAccGatewayNetworkPath("inbound_security_group"), "aws_security_group.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
		statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.GatewayElbIpAddressTypeIpv4)),
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_configurationFull(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: fullStateChecks,
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGatewayConfig_configurationRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("access_logs"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("scheme"), tfknownvalue.StringExact(awstypes.GatewayElbSchemeInternal)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("test_traffic_rules"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("web_acl_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath("inbound_security_group"), knownvalue.Null()),
				},
			},
			{
				Config: testAccGatewayConfig_configurationFull(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: fullStateChecks,
			},
		},
	})
}

func TestAccECSGateway_Configuration_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_ipAddressType(rName, string(awstypes.GatewayElbIpAddressTypeDualstack)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.GatewayElbIpAddressTypeDualstack)),
				},
			},
			{
				Config: testAccGatewayConfig_ipAddressType(rName, string(awstypes.GatewayElbIpAddressTypeIpv4)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, testAccGatewayNetworkPath(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.GatewayElbIpAddressTypeIpv4)),
				},
			},
		},
	})
}

func TestAccECSGateway_Configuration_scheme(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_type"), tfknownvalue.StringExact(awstypes.AccessTypePublic)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("scheme"), tfknownvalue.StringExact(awstypes.GatewayElbSchemeInternetFacing)),
				},
			},
			{
				Config: testAccGatewayConfig_scheme(rName, string(awstypes.GatewayElbSchemeInternal)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_type"), tfknownvalue.StringExact(awstypes.AccessTypePrivate)),
					statecheck.ExpectKnownValue(resourceName, testAccGatewayALBPath("scheme"), tfknownvalue.StringExact(awstypes.GatewayElbSchemeInternal)),
				},
			},
		},
	})
}

func testAccGatewayALBPath(key string) tfjsonpath.Path {
	return tfjsonpath.New(names.AttrConfiguration).AtSliceIndex(0).AtMapKey("host_shared_alb").AtSliceIndex(0).AtMapKey(key)
}

func testAccGatewayNetworkPath(key string) tfjsonpath.Path {
	return testAccGatewayALBPath(names.AttrNetworkConfiguration).AtSliceIndex(0).AtMapKey(key)
}

func testAccPreCheckGateways(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

	input := ecs.ListGatewaysInput{}
	_, err := conn.ListGateways(ctx, &input)

	if acctest.PreCheckSkipError(err) || errs.IsA[*awstypes.UnsupportedFeatureException](err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckGatewayDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_gateway" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]
			_, err := tfecs.FindGatewayByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Gateway %s still exists", arn)
		}

		return nil
	}
}

func testAccCheckGatewayExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		_, err := tfecs.FindGatewayByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccGatewayConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), testAccGatewayConfig_baseInfrastructureRole(rName))
}

func testAccGatewayConfig_baseInfrastructureRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRoleForGatewayHostSharedALB"
}
`, rName)
}

func testAccGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGatewayConfig_deletionProtectionOmitted(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGatewayConfig_update(rName, httpDesync string, maxALBCount int, propagateTags string, secondRole bool) string {
	config := testAccGatewayConfig_base(rName)
	role := "test"
	if secondRole {
		role = "test2"
		config = acctest.ConfigCompose(config, fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  assume_role_policy = data.aws_iam_policy_document.test_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test2.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRoleForGatewayHostSharedALB"
}
`, rName))
	}

	return acctest.ConfigCompose(config, fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.%[2]s.arn
  deletion_protection     = false
  propagate_tags          = %[5]q

  configuration {
    host_shared_alb {
      http_desync   = %[3]q
      max_alb_count = %[4]d

      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.%[2]s]
}
`, rName, role, httpDesync, maxALBCount, propagateTags))
}

func testAccGatewayConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccGatewayConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccGatewayConfig_scheme(rName, scheme string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      scheme = %[2]q

      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, scheme))
}

func testAccGatewayConfig_configurationBase(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
data "aws_elb_service_account" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test_bucket_policy" {
  statement {
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      type        = "AWS"
      identifiers = [data.aws_elb_service_account.current.arn]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test_bucket_policy.json
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName))
}

func testAccGatewayConfig_configurationFull(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_configurationBase(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      http_desync   = "strictest"
      max_alb_count = 4
      scheme        = "internal"
      web_acl_arn   = aws_wafv2_web_acl.test.arn

      access_logs {
        destination {
          s3 {
            bucket = aws_s3_bucket.test.bucket
            prefix = "gateway"
          }
        }
      }

      network_configuration {
        inbound_security_group = aws_security_group.test.id
        ip_address_type        = "ipv4"
        subnets                = aws_subnet.test[*].id
      }

      test_traffic_rules {
        header {
          name   = "x-canary"
          values = ["true"]
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_s3_bucket_policy.test]
}
`, rName))
}

func testAccGatewayConfig_configurationRemoved(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_configurationBase(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      http_desync   = "strictest"
      max_alb_count = 4
      scheme        = "internal"

      network_configuration {
        ip_address_type = "ipv4"
        subnets         = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_s3_bucket_policy.test]
}
`, rName))
}

func testAccGatewayConfig_ipAddressType(rName, ipAddressType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 2), testAccGatewayConfig_baseInfrastructureRole(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        ip_address_type = %[2]q
        subnets         = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, ipAddressType))
}
