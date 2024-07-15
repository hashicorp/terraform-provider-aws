// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.ELBV2ServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"ValidationError: Action type 'authenticate-cognito' must be one",
		"ValidationError: Protocol 'GENEVE' must be one of",
		"ValidationError: Type must be one of: 'application, network'",
	)
}

func TestLBCloudWatchSuffixFromARN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		arn    *string
		suffix string
	}{
		{
			name:   "valid suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:loadbalancer/app/my-alb/abc123`), //lintignore:AWSAT003,AWSAT005
			suffix: `app/my-alb/abc123`,
		},
		{
			name:   "no suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:loadbalancer`), //lintignore:AWSAT003,AWSAT005
			suffix: ``,
		},
		{
			name:   "nil ARN",
			arn:    nil,
			suffix: ``,
		},
	}

	for _, tc := range cases {
		actual := tfelbv2.SuffixFromARN(tc.arn)
		if actual != tc.suffix {
			t.Fatalf("bad suffix: %q\nExpected: %s\n     Got: %s", tc.name, tc.suffix, actual)
		}
	}
}

func TestAccELBV2LoadBalancer_ALB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "defensive"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_tls_version_and_cipher_suite_headers", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_xff_client_port", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "internal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "xff_header_processing_mode", "append"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NLB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/net/%s/.+", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "connection_logs.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttr(resourceName, "dns_record_client_routing_policy", "any_availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "internal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_LoadBalancerType_gateway(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_typeGateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", string(awstypes.LoadBalancerTypeEnumGateway)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"enable_waf_fail_open",
					"idle_timeout",
				},
			},
		},
	})
}

func TestAccELBV2LoadBalancer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceLoadBalancer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2LoadBalancer_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nameGenerated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrName, "tf-lb-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-lb-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"enable_waf_fail_open",
					"idle_timeout",
				},
			},
		},
	})
}

func TestAccELBV2LoadBalancer_nameGeneratedForZeroValue(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_zeroValueName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, names.AttrName, "tf-lb-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-lb-"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_namePrefix(rName, "tf-px-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-px-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-px-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"enable_waf_fail_open",
					"idle_timeout",
				},
			},
		},
	})
}

func TestAccELBV2LoadBalancer_duplicateName(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config:      testAccLoadBalancerConfig_duplicateName(rName),
				ExpectError: regexache.MustCompile(`already exists`),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ipv6SubnetMapping(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	importStateVerifyIgnore := []string{
		"drop_invalid_header_fields",
		"enable_http2",
		"idle_timeout",
	}
	// GovCloud doesn't support dns_record_client_routing_policy.
	if acctest.Partition() == names.USGovCloudPartitionID {
		importStateVerifyIgnore = append(importStateVerifyIgnore, "dns_record_client_routing_policy")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_ipv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "subnet_mapping.*", map[string]*regexp.Regexp{
						"ipv6_address": regexache.MustCompile("[0-6a-f]+:[0-6a-f:]+"),
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
		},
	})
}

func TestAccELBV2LoadBalancer_LoadBalancerTypeGateway_enableCrossZoneLoadBalancing(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_typeGatewayEnableCrossZoneBalancing(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", string(awstypes.LoadBalancerTypeEnumGateway)),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"enable_waf_fail_open",
					"idle_timeout",
				},
			},
			{
				Config: testAccLoadBalancerConfig_typeGatewayEnableCrossZoneBalancing(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", string(awstypes.LoadBalancerTypeEnumGateway)),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ALB_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_outpost(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile(fmt.Sprintf("loadbalancer/app/%s/.+", rName))),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_mapping.0.outpost_id"),
					resource.TestCheckResourceAttrSet(resourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_networkLoadBalancerEIP(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	resourceName := "aws_lb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbEIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "internal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NLB_privateIPv4Address(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	var importStateVerifyIgnore []string
	// GovCloud doesn't support dns_record_client_routing_policy.
	if acctest.Partition() == names.USGovCloudPartitionID {
		importStateVerifyIgnore = append(importStateVerifyIgnore, "dns_record_client_routing_policy")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbPrivateIPV4Address(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "internal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
		},
	})
}

func TestAccELBV2LoadBalancer_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "internal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_updateCrossZone(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbCrossZone(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "load_balancing.cross_zone.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", acctest.CtTrue),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbCrossZone(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "load_balancing.cross_zone.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", acctest.CtFalse),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbCrossZone(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "load_balancing.cross_zone.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", acctest.CtTrue),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateHTTP2(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableHTTP2(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http2.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", acctest.CtFalse),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableHTTP2(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http2.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", acctest.CtTrue),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableHTTP2(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http2.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", acctest.CtFalse),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_clientKeepAlive(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_clientKeepAlive(rName, 3600),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, tfelbv2.LoadBalancerAttributeClientKeepAliveSeconds, "3600"),
					resource.TestCheckResourceAttr(resourceName, "client_keep_alive", "3600"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_clientKeepAlive(rName, 7200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, tfelbv2.LoadBalancerAttributeClientKeepAliveSeconds, "7200"),
					resource.TestCheckResourceAttr(resourceName, "client_keep_alive", "7200"),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_clientKeepAlive(rName, 14400),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, tfelbv2.LoadBalancerAttributeClientKeepAliveSeconds, "14400"),
					resource.TestCheckResourceAttr(resourceName, "client_keep_alive", "14400"),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateDropInvalidHeaderFields(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableDropInvalidHeaderFields(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.drop_invalid_header_fields.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "drop_invalid_header_fields", acctest.CtFalse),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDropInvalidHeaderFields(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.drop_invalid_header_fields.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "drop_invalid_header_fields", acctest.CtTrue),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDropInvalidHeaderFields(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.drop_invalid_header_fields.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "drop_invalid_header_fields", acctest.CtFalse),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updatePreserveHostHeader(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enablePreserveHostHeader(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.preserve_host_header.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "preserve_host_header", acctest.CtFalse),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enablePreserveHostHeader(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.preserve_host_header.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "preserve_host_header", acctest.CtTrue),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enablePreserveHostHeader(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.preserve_host_header.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "preserve_host_header", acctest.CtFalse),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateDeletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableDeletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "deletion_protection.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDeletionProtection(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "deletion_protection.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtTrue),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDeletionProtection(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "deletion_protection.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateWAFFailOpen(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableWAFFailOpen(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "enable_waf_fail_open", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_enableWAFFailOpen(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					resource.TestCheckResourceAttr(resourceName, "enable_waf_fail_open", acctest.CtTrue),
					testAccCheckLoadBalancerNotRecreated(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableWAFFailOpen(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "enable_waf_fail_open", acctest.CtFalse),
					testAccCheckLoadBalancerNotRecreated(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_updateIPAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_ipAddressType(rName, "ipv4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_ipAddressType(rName, "dualstack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_ipAddressType(rName, "dualstack-without-public-ipv4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack-without-public-ipv4"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updatedSecurityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albUpdateSecurityGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", ""),
					testAccCheckLoadBalancerNotRecreated(&pre, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_addSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_subnetCount(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
			{
				Config: testAccLoadBalancerConfig_subnetCount(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_deleteSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_subnetCount(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
			{
				Config: testAccLoadBalancerConfig_subnetCount(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_addSubnetMapping(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_subnetMappingCount(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
			{
				Config: testAccLoadBalancerConfig_subnetMappingCount(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_deleteSubnetMapping(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_subnetMappingCount(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
			{
				Config: testAccLoadBalancerConfig_subnetMappingCount(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
		},
	})
}

// TestAccELBV2LoadBalancer_noSecurityGroup regression tests the issue in #8264,
// where if an ALB is created without a security group, a default one
// is assigned.
func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_noSecurityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_albNoSecurityGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "internal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_accessLogs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_albAccessLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_albAccessLogs(false, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtFalse),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albAccessLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albAccessLogsNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtFalse),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_accessLogsPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_albAccessLogs(true, rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_albAccessLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albAccessLogs(true, rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_connectionLogs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_albConnectionLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_albConnectionLogs(false, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtFalse),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albConnectionLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albConnectionLogsNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtFalse),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", ""),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_connectionLogsPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_albConnectionLogs(true, rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_albConnectionLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_albConnectionLogs(true, rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "connection_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "connection_logs.0.prefix", "prefix1"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_accessLogs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	var importStateVerifyIgnore []string
	// GovCloud doesn't support dns_record_client_routing_policy.
	if acctest.Partition() == names.USGovCloudPartitionID {
		importStateVerifyIgnore = append(importStateVerifyIgnore, "dns_record_client_routing_policy")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogs(false, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtFalse),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogsNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtFalse),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_accessLogsPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	var importStateVerifyIgnore []string
	// GovCloud doesn't support dns_record_client_routing_policy.
	if acctest.Partition() == names.USGovCloudPartitionID {
		importStateVerifyIgnore = append(importStateVerifyIgnore, "dns_record_client_routing_policy")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogs(true, rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogs(true, rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbAccessLogs(true, rName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.bucket", rName),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.enabled", acctest.CtTrue),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_updateDNSRecordClientRoutingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud doesn't support dns_record_client_routing_policy.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbDNSRecordClientRoutingPolicyAffinity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dns_record_client_routing_policy", "availability_zone_affinity"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_nlbDNSRecordClientRoutingPolicyPartialAffinity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dns_record_client_routing_policy", "partial_availability_zone_affinity"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbDNSRecordClientRoutingPolicyAnyAvailabilityZone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dns_record_client_routing_policy", "any_availability_zone"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_updateSecurityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var lb1, lb2, lb3, lb4 awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroups(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb1),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroups(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb2),
					testAccCheckLoadBalancerRecreated(&lb2, &lb1),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroups(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb3),
					testAccCheckLoadBalancerNotRecreated(&lb3, &lb2),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", ""),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroups(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb4),
					testAccCheckLoadBalancerRecreated(&lb4, &lb3),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", ""),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_enforcePrivateLink(t *testing.T) {
	ctx := acctest.Context(t)
	var lb awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	var importStateVerifyIgnore []string
	// GovCloud doesn't support dns_record_client_routing_policy.
	if acctest.Partition() == names.USGovCloudPartitionID {
		importStateVerifyIgnore = append(importStateVerifyIgnore, "dns_record_client_routing_policy")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroupsEnforcePrivateLink(rName, 1, "off"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", "off")),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroups(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", "off"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroupsEnforcePrivateLink(rName, 1, "on"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", "on"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroupsEnforcePrivateLink(rName, 1, "off"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", "off"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSecurityGroups(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", "off"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_addSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbSubnetCount(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSubnetCount(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerNotRecreated(&post, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_deleteSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbSubnetCount(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSubnetCount(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerRecreated(&post, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_addSubnetMapping(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbSubnetMappingCount(rName, false, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSubnetMappingCount(rName, false, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_deleteSubnetMapping(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// GovCloud Regions don't always have 3 AZs.
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlbSubnetMappingCount(rName, false, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct3),
				),
			},
			{
				Config: testAccLoadBalancerConfig_nlbSubnetMappingCount(rName, false, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_updateDesyncMitigationMode(t *testing.T) {
	ctx := acctest.Context(t)
	var pre, mid, post awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_desyncMitigationMode(rName, "strictest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &pre),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.desync_mitigation_mode", "strictest"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "strictest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_desyncMitigationMode(rName, "monitor"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &mid),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.desync_mitigation_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "monitor"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_desyncMitigationMode(rName, "defensive"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &post),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.desync_mitigation_mode", "defensive"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "defensive"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ALB_updateTLSVersionAndCipherSuite(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_tlscipherSuiteEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.x_amzn_tls_version_and_cipher_suite.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_tls_version_and_cipher_suite_headers", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_tlscipherSuiteEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.x_amzn_tls_version_and_cipher_suite.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_tls_version_and_cipher_suite_headers", acctest.CtTrue),
				),
			},
			{
				Config: testAccLoadBalancerConfig_tlscipherSuiteEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.x_amzn_tls_version_and_cipher_suite.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_tls_version_and_cipher_suite_headers", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ALB_updateXffHeaderProcessingMode(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_xffHeaderProcessingMode(rName, "append"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.xff_header_processing.mode", "append"),
					resource.TestCheckResourceAttr(resourceName, "xff_header_processing_mode", "append"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_xffHeaderProcessingMode(rName, "preserve"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.xff_header_processing.mode", "preserve"),
					resource.TestCheckResourceAttr(resourceName, "xff_header_processing_mode", "preserve"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_xffHeaderProcessingMode(rName, "remove"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.xff_header_processing.mode", "remove"),
					resource.TestCheckResourceAttr(resourceName, "xff_header_processing_mode", "remove"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ALB_updateXffClientPort(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_xffClientPort(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.xff_client_port.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_xff_client_port", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerConfig_xffClientPort(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.xff_client_port.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_xff_client_port", acctest.CtTrue),
				),
			},
			{
				Config: testAccLoadBalancerConfig_xffClientPort(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(ctx, resourceName, &conf),
					testAccCheckLoadBalancerAttribute(ctx, resourceName, "routing.http.xff_client_port.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_xff_client_port", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckLoadBalancerNotRecreated(i, j *awstypes.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.LoadBalancerArn) != aws.ToString(j.LoadBalancerArn) {
			return errors.New("ELBv2 Load Balancer was recreated")
		}

		return nil
	}
}

func testAccCheckLoadBalancerRecreated(i, j *awstypes.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.LoadBalancerArn) == aws.ToString(j.LoadBalancerArn) {
			return errors.New("ELBv2 Load Balancer was not recreated")
		}

		return nil
	}
}

func testAccCheckLoadBalancerExists(ctx context.Context, n string, v *awstypes.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		output, err := tfelbv2.FindLoadBalancerByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLoadBalancerAttribute(ctx context.Context, n, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		attributes, err := tfelbv2.FindLoadBalancerAttributesByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		for _, v := range attributes {
			if aws.ToString(v.Key) == key {
				got := aws.ToString(v.Value)
				if got == value {
					return nil
				}

				return fmt.Errorf("ELBv2 Load Balancer (%s) attribute (%s) = %v, want %v", rs.Primary.ID, key, got, value)
			}
		}

		return fmt.Errorf("ELBv2 Load Balancer (%s) attribute (%s) not found", rs.Primary.ID, key)
	}
}

func testAccCheckLoadBalancerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb" && rs.Type != "aws_alb" {
				continue
			}

			_, err := tfelbv2.FindLoadBalancerByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Load Balancer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckGatewayLoadBalancer(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

	input := &elasticloadbalancingv2.DescribeAccountLimitsInput{}

	output, err := conn.DescribeAccountLimits(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected ELBv2 Gateway Load Balancer PreCheck error: %s", err)
	}

	if output == nil {
		t.Fatal("unexpected ELBv2 Gateway Load Balancer PreCheck error: empty response")
	}

	for _, limit := range output.Limits {
		if aws.ToString(limit.Name) == "gateway-load-balancers" {
			return
		}
	}

	t.Skip("skipping acceptance testing: region does not support ELBv2 Gateway Load Balancers")
}

func testAccLoadBalancerConfig_baseInternal(rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, subnetCount), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_basic(rName string) string {
	return testAccLoadBalancerConfig_subnetCount(rName, 2)
}

func testAccLoadBalancerConfig_subnetCount(rName string, subnetCount int) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, subnetCount), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}
`, rName))
}

func testAccLoadBalancerConfig_subnetMappingCount(rName string, subnetCount int) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, subnetCount), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]

  idle_timeout               = 30
  enable_deletion_protection = false

  dynamic "subnet_mapping" {
    for_each = aws_subnet.test[*]
    content {
      subnet_id = subnet_mapping.value.id
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
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

func testAccLoadBalancerConfig_zeroValueName(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = ""
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

# See https://github.com/hashicorp/terraform-provider-aws/issues/2498
output "lb_name" {
  value = aws_lb.test.name
}
`, rName))
}

func testAccLoadBalancerConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name_prefix     = %[2]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
`, rName, namePrefix))
}

func testAccLoadBalancerConfig_duplicateName(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_basic(rName), fmt.Sprintf(`
resource "aws_lb" "test2" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}
`, rName))
}

func testAccLoadBalancerConfig_ipAddressType(rName, ipAddressType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  ip_address_type = %[2]q

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_egress_only_internet_gateway.test, aws_internet_gateway.test]
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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
`, rName, ipAddressType))
}

func testAccLoadBalancerConfig_outpost(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id     = aws_vpc.test.id
  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_lb" "test" {
  name                       = %[1]q
  security_groups            = [aws_security_group.test.id]
  customer_owned_ipv4_pool   = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  idle_timeout               = 30
  enable_deletion_protection = false
  subnets                    = [aws_subnet.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}

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

func testAccLoadBalancerConfig_enableHTTP2(rName string, http2 bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_http2 = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, http2))
}

func testAccLoadBalancerConfig_clientKeepAlive(rName string, value int64) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_http2      = true
  client_keep_alive = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, value))
}

func testAccLoadBalancerConfig_enableDropInvalidHeaderFields(rName string, dropInvalid bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  drop_invalid_header_fields = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, dropInvalid))
}

func testAccLoadBalancerConfig_enablePreserveHostHeader(rName string, enablePreserveHostHeader bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  preserve_host_header = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, enablePreserveHostHeader))
}

func testAccLoadBalancerConfig_enableDeletionProtection(rName string, deletionProtection bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, deletionProtection))
}

func testAccLoadBalancerConfig_enableWAFFailOpen(rName string, wafFailOpen bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_waf_fail_open = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, wafFailOpen))
}

func testAccLoadBalancerConfig_nlbBasic(rName string) string {
	return testAccLoadBalancerConfig_nlbSubnetMappingCount(rName, false, 1)
}

func testAccLoadBalancerConfig_nlbCrossZone(rName string, cz bool) string {
	return testAccLoadBalancerConfig_nlbSubnetMappingCount(rName, cz, 1)
}

func testAccLoadBalancerConfig_nlbSubnetMappingCount(rName string, cz bool, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, subnetCount), fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = %[2]t

  dynamic "subnet_mapping" {
    for_each = aws_subnet.test[*]
    content {
      subnet_id = subnet_mapping.value.id
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, cz))
}

func testAccLoadBalancerConfig_typeGateway(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccLoadBalancerConfig_ipv6(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name                       = %[1]q
  load_balancer_type         = "network"
  enable_deletion_protection = false

  subnet_mapping {
    subnet_id    = aws_subnet.test[0].id
    ipv6_address = cidrhost(cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 0), 5)
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName))
}

func testAccLoadBalancerConfig_typeGatewayEnableCrossZoneBalancing(rName string, enableCrossZoneLoadBalancing bool) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  enable_cross_zone_load_balancing = %[2]t
  load_balancer_type               = "gateway"
  name                             = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName, enableCrossZoneLoadBalancing))
}

func testAccLoadBalancerConfig_nlbEIP(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = aws_subnet.test[0].id
    allocation_id = aws_eip.test[0].id
  }

  subnet_mapping {
    subnet_id     = aws_subnet.test[1].id
    allocation_id = aws_eip.test[1].id
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_eip" "test" {
  count = 2

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_nlbPrivateIPV4Address(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = %[1]q
  internal                   = true
  load_balancer_type         = "network"
  enable_deletion_protection = false

  subnet_mapping {
    subnet_id            = aws_subnet.test.id
    private_ipv4_address = "10.10.0.15"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_nlbDNSRecordClientRoutingPolicyAffinity(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                             = %[1]q
  dns_record_client_routing_policy = "availability_zone_affinity"
  internal                         = true
  load_balancer_type               = "network"

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_nlbDNSRecordClientRoutingPolicyAnyAvailabilityZone(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                             = %[1]q
  dns_record_client_routing_policy = "any_availability_zone"
  internal                         = true
  load_balancer_type               = "network"

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_nlbDNSRecordClientRoutingPolicyPartialAffinity(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                             = %[1]q
  dns_record_client_routing_policy = "partial_availability_zone_affinity"
  internal                         = true
  load_balancer_type               = "network"

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_backwardsCompatibility(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
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
`, rName))
}

func testAccLoadBalancerConfig_albUpdateSecurityGroups(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 3), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id, aws_security_group.test2.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccLoadBalancerConfig_albNoSecurityGroups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}
`, rName))
}

func testAccLoadBalancerConfig_baseALBAccessLogs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_elb_service_account" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["s3:PutObject"]
    effect    = "Allow"
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      type        = "AWS"
      identifiers = [data.aws_elb_service_account.current.arn]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, rName))
}

func testAccLoadBalancerConfig_albAccessLogs(enabled bool, rName, bucketPrefix string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseALBAccessLogs(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id

  access_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, rName, enabled, bucketPrefix))
}

func testAccLoadBalancerConfig_albAccessLogsNoBlocks(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseALBAccessLogs(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}
`, rName))
}

func testAccLoadBalancerConfig_albConnectionLogs(enabled bool, rName, bucketPrefix string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseALBAccessLogs(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id

  connection_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, rName, enabled, bucketPrefix))
}

func testAccLoadBalancerConfig_albConnectionLogsNoBlocks(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseALBAccessLogs(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}
`, rName))
}

func testAccLoadBalancerConfig_baseNLBAccessLogs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["s3:PutObject"]
    effect    = "Allow"
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }

  statement {
    actions   = ["s3:GetBucketAcl"]
    effect    = "Allow"
    resources = [aws_s3_bucket.test.arn]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, rName))
}

func testAccLoadBalancerConfig_nlbAccessLogs(enabled bool, rName, bucketPrefix string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseNLBAccessLogs(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id

  access_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, rName, enabled, bucketPrefix))
}

func testAccLoadBalancerConfig_nlbAccessLogsNoBlocks(rName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseNLBAccessLogs(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id
}
`, rName))
}

func testAccLoadBalancerConfig_nlbSecurityGroups(rName string, n int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 3

  name   = "%[1]s-${count.index}"
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id
  security_groups    = slice(aws_security_group.test[*].id, 0, %[2]d)
}
`, rName, n))
}

func testAccLoadBalancerConfig_nlbSecurityGroupsEnforcePrivateLink(rName string, n int, enforcePrivateLink string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 3

  name   = "%[1]s-${count.index}"
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id
  security_groups    = slice(aws_security_group.test[*].id, 0, %[2]d)

  enforce_security_group_inbound_rules_on_private_link_traffic = %[3]q
}
`, rName, n, enforcePrivateLink))
}

func testAccLoadBalancerConfig_nlbSubnetCount(rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, subnetCount), fmt.Sprintf(`
resource "aws_lb" "test" {
  name = %[1]q

  subnets = aws_subnet.test[*].id

  load_balancer_type               = "network"
  internal                         = true
  idle_timeout                     = 60
  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = false
}
`, rName))
}

func testAccLoadBalancerConfig_desyncMitigationMode(rName, mode string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  desync_mitigation_mode = %[2]q
}
`, rName, mode))
}

func testAccLoadBalancerConfig_tlscipherSuiteEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_tls_version_and_cipher_suite_headers = %[2]t
}
`, rName, enabled))
}

func testAccLoadBalancerConfig_xffHeaderProcessingMode(rName, mode string) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  xff_header_processing_mode = %[2]q
}
`, rName, mode))
}

func testAccLoadBalancerConfig_xffClientPort(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccLoadBalancerConfig_baseInternal(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_xff_client_port = %[2]t
}
`, rName, enabled))
}
