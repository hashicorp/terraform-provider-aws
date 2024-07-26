// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2Listener_Application_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Application_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr("aws_lb.test", "load_balancer_type", "application"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckNoResourceAttr(resourceName, "alpn_policy"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Network_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Network_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr("aws_lb.test", "load_balancer_type", "network"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckNoResourceAttr(resourceName, "alpn_policy"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TCP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Gateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Gateway_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr("aws_lb.test", "load_balancer_type", "gateway"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckNoResourceAttr(resourceName, "alpn_policy"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", acctest.Ct0), // A Gateway Listener can only have one action, so the API never returns a value
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Application_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceListener(), resourceName),
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

func TestAccELBV2Listener_updateForwardBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardBasic(rName, "test1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_forwardBasic(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_forwardWeighted(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardWeighted(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerConfig_changeForwardWeightedStickiness(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerConfig_changeForwardWeightedToBasic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_forwardTargetARNAndBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccListenerConfig_forwardTargetARNAndBlock(rName),
				ExpectError: regexache.MustCompile(regexp.QuoteMeta(`Only one of "default_action[0].target_group_arn" or "default_action[0].forward" can be specified`)),
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_TargetGroupARNToForwardBlock_NoChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_TargetGroupARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_ForwardBlock_AddStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockStickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_ForwardBlock_RemoveStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockStickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_TargetGroupARNToForwardBlock_WeightAndStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_TargetGroupARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockWeightAndStickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_ForwardBlockToTargetGroupARN_NoChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_TargetGroupARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_ForwardBlockToTargetGroupARN_WeightAndStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockWeightAndStickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_TargetGroupARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_ForwardBlock_AddAction(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockAddAction(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_ForwardBlock_RemoveAction(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockAddAction(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_ActionForward_IgnoreFields(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_actionForward_ForwardBlockMultiTargetWithIgnore(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "440"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccELBV2Listener_Protocol_upd(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basicUdp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "UDP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "514"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

// TestAccELBV2Listener_backwardsCompatibility confirms that the resource type `aws_alb_listener` works
func TestAccELBV2Listener_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_alb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_alb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_alb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Protocol_https(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_https(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationOff),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.trust_store_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_mutualAuthentication(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_mutualAuthentication(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationVerify),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_authentication.0.trust_store_arn", "aws_lb_trust_store.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_mutualAuthenticationPassthrough(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_mutualAuthenticationPassthrough(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationPassthrough),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.trust_store_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_LoadBalancerARN_gatewayLoadBalancer(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbResourceName := "aws_lb.test"
	resourceName := "aws_lb_listener.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_arnGateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", lbResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2Listener_Protocol_tls(t *testing.T) {
	ctx := acctest.Context(t)
	var listener1 awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_protocolTLS(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener1),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TLS"),
					resource.TestCheckResourceAttr(resourceName, "alpn_policy", tfelbv2.AlpnPolicyHTTP2Preferred),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_redirect(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_redirect(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.host", "#{host}"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.path", "/#{path}"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.query", "#{query}"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.status_code", "HTTP_301"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccELBV2Listener_fixedResponse(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_fixedResponse(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "fixed-response"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.message_body", "Fixed response content"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.status_code", "200"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccELBV2Listener_cognito(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_cognito(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-cognito"),
					resource.TestCheckResourceAttrSet(resourceName, "default_action.0.authenticate_cognito.0.user_pool_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_action.0.authenticate_cognito.0.user_pool_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_action.0.authenticate_cognito.0.user_pool_domain"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.0.authentication_request_extra_params.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.1.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_oidc(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_oidc(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.authorization_endpoint", "https://example.com/authorization_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.client_id", "s6BhdRkqt3"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.client_secret", "7Fjfp0ZBr1KtDRbnfVdmIw"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.token_endpoint", "https://example.com/token_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.user_info_endpoint", "https://example.com/user_info_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.authentication_request_extra_params.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_DefaultAction_defaultOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var listener awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_defaultAction_defaultOrder(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_DefaultAction_specifyOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var listener awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_defaultAction_specifyOrder(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", acctest.Ct4),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.forward",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6171
func TestAccELBV2Listener_DefaultAction_actionDisappears(t *testing.T) {
	ctx := acctest.Context(t)
	var listener awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_defaultAction_defaultOrder(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", acctest.Ct2),
					testAccCheckListenerDefaultActionOrderDisappears(ctx, &listener, 1),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						// TODO: change `default_action[0]`
						// TODO: add `default_action[1]`
					},
				},
			},
		},
	})
}

func TestAccELBV2Listener_EmptyDefaultAction(t *testing.T) {
	t.Parallel()

	testcases := map[awstypes.ActionTypeEnum]struct {
		actionType    awstypes.ActionTypeEnum
		expectedError *regexp.Regexp
	}{
		awstypes.ActionTypeEnumForward: {
			actionType: awstypes.ActionTypeEnumForward,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Either %q or %q must be specified when %q is %q.",
				"default_action[0].target_group_arn", "default_action[0].forward",
				"default_action[0].type",
				awstypes.ActionTypeEnumForward,
			))),
		},

		awstypes.ActionTypeEnumAuthenticateOidc: {
			actionType: awstypes.ActionTypeEnumAuthenticateOidc,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].authenticate_oidc",
				"default_action[0].type",
				awstypes.ActionTypeEnumAuthenticateOidc,
			))),
		},

		awstypes.ActionTypeEnumAuthenticateCognito: {
			actionType: awstypes.ActionTypeEnumAuthenticateCognito,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].authenticate_cognito",
				"default_action[0].type",
				awstypes.ActionTypeEnumAuthenticateCognito,
			))),
		},

		awstypes.ActionTypeEnumRedirect: {
			actionType: awstypes.ActionTypeEnumRedirect,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].redirect",
				"default_action[0].type",
				awstypes.ActionTypeEnumRedirect,
			))),
		},

		awstypes.ActionTypeEnumFixedResponse: {
			actionType: awstypes.ActionTypeEnumFixedResponse,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].fixed_response",
				"default_action[0].type",
				awstypes.ActionTypeEnumFixedResponse,
			))),
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // uses t.Setenv
		testcase := testcase

		t.Run(string(name), func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckListenerDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config:      testAccListenerConfig_EmptyDefaultAction(rName, testcase.actionType),
						ExpectError: testcase.expectedError,
					},
				},
			})
		})
	}
}

// https://github.com/hashicorp/terraform-provider-aws/issues/35668.
func TestAccELBV2Listener_redirectWithTargetGroupARN(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.34.0",
					},
				},
				Config: testAccListenerConfig_redirectWithTargetGroupARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.36.0",
					},
				},
				Config:             testAccListenerConfig_redirectWithTargetGroupARN(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckListenerDefaultActionOrderDisappears(ctx context.Context, listener *awstypes.Listener, actionOrderToDelete int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var newDefaultActions []awstypes.Action

		for i, action := range listener.DefaultActions {
			if int(aws.ToInt32(action.Order)) == actionOrderToDelete {
				newDefaultActions = slices.Delete(listener.DefaultActions, i, i+1)
				break
			}
		}

		if len(newDefaultActions) == 0 {
			return fmt.Errorf("Unable to find default action order %d from default actions: %#v", actionOrderToDelete, listener.DefaultActions)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		input := &elasticloadbalancingv2.ModifyListenerInput{
			DefaultActions: newDefaultActions,
			ListenerArn:    listener.ListenerArn,
		}

		_, err := conn.ModifyListener(ctx, input)

		return err
	}
}

func testAccCheckListenerExists(ctx context.Context, n string, v *awstypes.Listener) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		output, err := tfelbv2.FindListenerByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckListenerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_listener" && rs.Type != "aws_alb_listener" {
				continue
			}

			_, err := tfelbv2.FindListenerByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Listener %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccListenerConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Application_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "application"
  internal           = true
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Network_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "TCP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "network"
  internal           = true
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Gateway_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "gateway"
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_forwardBasic(rName, targetName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "440"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.%[2]s.arn
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
`, rName, targetName))
}

func testAccListenerConfig_forwardWeighted(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
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
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test1" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = %[2]q
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

  tags = {
    Name = %[2]q
  }
}
`, rName, rName2))
}

func testAccListenerConfig_changeForwardWeightedStickiness(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
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
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test1" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = %[2]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName, rName2))
}

func testAccListenerConfig_forwardTargetARNAndBlock(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "440"

  default_action {
    type = "forward"

    target_group_arn = aws_lb_target_group.test.arn

    forward {
      target_group {
        arn    = aws_lb_target_group.test.arn
        weight = 1
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_actionForward_TargetGroupARN(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
  `, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_actionForward_ForwardBlockBasic(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "440"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn = aws_lb_target_group.test.arn
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_actionForward_ForwardBlockStickiness(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "440"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn = aws_lb_target_group.test.arn
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_actionForward_ForwardBlockWeightAndStickiness(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test.arn
        weight = 2
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_actionForward_ForwardBlockAddAction(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
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

  default_action {
    type = "forward"

    forward {
      target_group {
        arn = aws_lb_target_group.test.arn
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_actionForward_ForwardBlockMultiTargetWithIgnore(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "440"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test.arn
        weight = 100
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 0
      }
    }
  }

  lifecycle {
    ignore_changes = [
      default_action[0].forward,
      default_action[0].target_group_arn,
    ]
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_changeForwardWeightedToBasic(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test1" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = %[2]q
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

  tags = {
    Name = %[2]q
  }
}
`, rName, rName2))
}

func testAccListenerConfig_basicUdp(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "UDP"
  port              = "514"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 514
  protocol = "UDP"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 514
    protocol = "TCP"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

// testAccListenerConfig_backwardsCompatibility should be the equivalent of `testAccListenerConfig_basic`
// but using the legacy `aws_alb*` resource types.
func testAccListenerConfig_backwardsCompatibility(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_alb_listener" "test" {
  load_balancer_arn = aws_alb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_alb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_alb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_alb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_https(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = false
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_mutualAuthentication(rName string, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		testAccTrustStoreConfig_baseS3BucketCA(rName),
		fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name                             = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  mutual_authentication {
    mode            = "verify"
    trust_store_arn = aws_lb_trust_store.test.arn
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_mutualAuthenticationPassthrough(rName string, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  mutual_authentication {
    mode = "passthrough"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_arnGateway(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 80
    protocol = "HTTP"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_protocolTLS(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  certificate_arn   = aws_acm_certificate.test.arn
  load_balancer_arn = aws_lb.test.arn
  port              = "443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  alpn_policy       = "HTTP2Preferred"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TLS"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_redirect(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_fixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code  = "200"
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_cognito(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = %[1]q
  internal                   = false
  security_groups            = [aws_security_group.test.id]
  subnets                    = aws_subnet.test[*].id
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
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

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_oidc(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = %[1]q
  internal                   = false
  security_groups            = [aws_security_group.test.id]
  subnets                    = aws_subnet.test[*].id
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
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

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_defaultAction_defaultOrder(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
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

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb" "test" {
  internal        = true
  name            = %[1]q
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_defaultAction_specifyOrder(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    order = 2
    type  = "authenticate-oidc"

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

  default_action {
    order            = 4
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb" "test" {
  internal        = true
  name            = %[1]q
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_EmptyDefaultAction(rName string, action awstypes.ActionTypeEnum) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = %[2]q
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName, action))
}

func testAccListenerConfig_redirectWithTargetGroupARN(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
