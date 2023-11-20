// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestLBTargetGroupCloudWatchSuffixFromARN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		arn    *string
		suffix string
	}{
		{
			name:   "valid suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup/my-targets/73e2d6bc24d8a067`), //lintignore:AWSAT003,AWSAT005
			suffix: `targetgroup/my-targets/73e2d6bc24d8a067`,
		},
		{
			name:   "no suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup`), //lintignore:AWSAT003,AWSAT005
			suffix: ``,
		},
		{
			name:   "nil ARN",
			arn:    nil,
			suffix: ``,
		},
	}

	for _, tc := range cases {
		actual := tfelbv2.TargetGroupSuffixFromARN(tc.arn)
		if actual != tc.suffix {
			t.Fatalf("bad suffix: %q\nExpected: %s\n     Got: %s", tc.name, tc.suffix, actual)
		}
	}
}

func TestALBTargetGroupCloudWatchSuffixFromARN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		arn    *string
		suffix string
	}{
		{
			name:   "valid suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup/my-targets/73e2d6bc24d8a067`), //lintignore:AWSAT003,AWSAT005
			suffix: `targetgroup/my-targets/73e2d6bc24d8a067`,
		},
		{
			name:   "no suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup`), //lintignore:AWSAT003,AWSAT005
			suffix: ``,
		},
		{
			name:   "nil ARN",
			arn:    nil,
			suffix: ``,
		},
	}

	for _, tc := range cases {
		actual := tfelbv2.TargetGroupSuffixFromARN(tc.arn)
		if actual != tc.suffix {
			t.Fatalf("bad suffix: %q\nExpected: %s\n     Got: %s", tc.name, tc.suffix, actual)
		}
	}
}

func TestAccELBV2TargetGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckNoResourceAttr(resourceName, "preserve_client_ip"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceTargetGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2TargetGroup_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameGeneratedWithPrefix(resourceName, "name", "tf-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_namePrefix(rName, "tf-px-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-px-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-px-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_duplicateName(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				),
			},
			{
				Config:      testAccTargetGroupConfig_duplicateName(rName, 200),
				ExpectError: regexache.MustCompile(`already exists`),
			},
		},
	})
}

func TestAccELBV2TargetGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccTargetGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTargetGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ProtocolVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolVersion(rName, "HTTP2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP2"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_protocolVersion(rName, "HTTP1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					testAccCheckTargetGroupRecreated(&after, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ProtocolVersion_grpcHealthCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_grpcProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/Test.Check/healthcheck"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "0-99"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ProtocolVersion_grpcUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				),
			},
			{
				Config: testAccTargetGroupConfig_grpcProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "GRPC"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_ipAddressType(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "target_type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_tls(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolTLS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_HealthCheck_tcpHTTPS(t *testing.T) {
	ctx := acctest.Context(t)
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCPHTTPHealthCheck(rName, "/healthz", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &confBefore),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/healthz"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPHTTPHealthCheck(rName, "/healthz2", 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &confAfter),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/healthz2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_attrsOnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "0"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
				),
			},
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_udp(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basicUdp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "514"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "UDP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "514"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_name(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rNameBefore := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameAfter := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rNameBefore, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rNameBefore),
				),
			},
			{
				Config: testAccTargetGroupConfig_basic(rNameAfter, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					testAccCheckTargetGroupRecreated(&after, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rNameAfter),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_port(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
				),
			},
			{
				Config: testAccTargetGroupConfig_updatedPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					testAccCheckTargetGroupRecreated(&after, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "442"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_protocol(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccTargetGroupConfig_updatedProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					testAccCheckTargetGroupRecreated(&after, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
				),
			},
			{
				Config: testAccTargetGroupConfig_updatedVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupRecreated(&after, &before),
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Defaults_application(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albDefaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Defaults_network(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"
	healthCheckValid := `
interval = 10
port     = 8081
protocol = "TCP"
timeout  = 4
    `

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_nlbDefaults(rName, healthCheckValid),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_HealthCheck_enable(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_noHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_enableHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_NetworkLB_tcpHealthCheckUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1, targetGroup2 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "false"),
					resource.TestCheckResourceAttr(resourceName, "connection_termination", "false"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPHealthCheckUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup2),
					testAccCheckTargetGroupNotRecreated(&targetGroup1, &targetGroup2),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "20"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "15"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_networkLB_TargetGroupWithConnectionTermination(t *testing.T) {
	ctx := acctest.Context(t)
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &confBefore),
					resource.TestCheckResourceAttr(resourceName, "connection_termination", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPConnectionTermination(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &confAfter),
					resource.TestCheckResourceAttr(resourceName, "connection_termination", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_NetworkLB_targetGroupWithProxy(t *testing.T) {
	ctx := acctest.Context(t)
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &confBefore),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPProxyProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &confAfter),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_preserveClientIPValid(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	resourceName := "aws_lb_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_preserveClientIP(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "true"),
				),
			},
			{
				Config: testAccTargetGroupConfig_preserveClientIP(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "false"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Geneve_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolGeneve(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_Geneve_notSticky(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolGeneve(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
				),
			},
			{
				Config: testAccTargetGroupConfig_protocolGeneveHealth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Geneve_Sticky(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolGeneveSticky(rName, "source_ip_dest_ip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip_dest_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_protocolGeneveSticky(rName, "source_ip_dest_ip_proto"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip_dest_ip_proto"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Geneve_targetFailover(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolGeneveTargetFailover(rName, "rebalance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
					resource.TestCheckResourceAttr(resourceName, "target_failover.0.on_deregistration", "rebalance"),
					resource.TestCheckResourceAttr(resourceName, "target_failover.0.on_unhealthy", "rebalance"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
			{
				Config: testAccTargetGroupConfig_protocolGeneveTargetFailover(rName, "no_rebalance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
					resource.TestCheckResourceAttr(resourceName, "target_failover.0.on_deregistration", "no_rebalance"),
					resource.TestCheckResourceAttr(resourceName, "target_failover.0.on_unhealthy", "no_rebalance"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_defaultALB(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "HTTP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_defaultNLB(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "TCP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "UDP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "TCP_UDP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_invalidALB(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "HTTP", "source_ip", true),
				ExpectError: regexache.MustCompile("Stickiness type 'source_ip' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "HTTPS", "source_ip", true),
				ExpectError: regexache.MustCompile("Stickiness type 'source_ip' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TLS", "lb_cookie", true),
				ExpectError: regexache.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:             testAccTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "lb_cookie", false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_invalidNLB(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", true),
				ExpectError: regexache.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", false),
				ExpectError: regexache.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "UDP", "lb_cookie", true),
				ExpectError: regexache.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "lb_cookie", true),
				ExpectError: regexache.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_validALB(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "HTTP", "lb_cookie", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "86400"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "HTTPS", "lb_cookie", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "86400"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_validNLB(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "source_ip", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "UDP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_updateAppEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_appStickiness(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccTargetGroupConfig_appStickiness(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "app_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_name", "Cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccTargetGroupConfig_appStickiness(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "app_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_name", "Cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_HealthCheck_update(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
			{
				Config: testAccTargetGroupConfig_updateHealthCheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_updateEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickiness(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickiness(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickiness(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_HealthCheck_without(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_noHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "instance"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameAfter := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_albBasic(rNameAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameAfter),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changePortForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdatedPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "port", "442"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changeProtocolForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdatedProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changeVPCForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdatedVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_generatedName(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albGeneratedName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"deregistration_delay",
					"proxy_protocol_v2",
					"slow_start",
					"load_balancing_algorithm_type",
				},
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_lambdaMultiValueHeadersEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLambdaMultiValueHeadersEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"connection_termination",
					"deregistration_delay",
					"proxy_protocol_v2",
					"slow_start",
					"load_balancing_algorithm_type",
				},
			},
			{
				Config: testAccTargetGroupConfig_albLambdaMultiValueHeadersEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLambdaMultiValueHeadersEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_missing(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		config     func(string) string
		errMessage string
	}{
		"Port": {
			config:     testAccTargetGroupConfig_albMissingPort,
			errMessage: `Attribute "port" must be specified when "target_type" is "instance".`,
		},
		"Protocol": {
			config:     testAccTargetGroupConfig_albMissingProtocol,
			errMessage: `Attribute "protocol" must be specified when "target_type" is "instance".`,
		},
		"VPC": {
			config:     testAccTargetGroupConfig_albMissingVPC,
			errMessage: `Attribute "vpc_id" must be specified when "target_type" is "instance".`,
		},
	}

	for name, tc := range testcases { //nolint:paralleltest // false positive
		tc := tc

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config:      tc.config(rName),
						ExpectError: regexache.MustCompile(tc.errMessage),
					},
				},
			})
		})
	}
}

func TestAccELBV2TargetGroup_ALBAlias_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name", regexache.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_setAndUpdateSlowStart(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albUpdateSlowStart(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "30"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdateSlowStart(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "60"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdateTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Type", "ALB Target Group"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_updateHealthCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdateHealthCheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_updateLoadBalancingAlgorithmType(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "round_robin"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName, true, "round_robin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "round_robin"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName, true, "least_outstanding_requests"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "least_outstanding_requests"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_updateLoadBalancingCrossZoneEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLoadBalancingCrossZoneEnabled(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_cross_zone_enabled", "use_load_balancer_configuration"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLoadBalancingCrossZoneEnabled(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_cross_zone_enabled", "true"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLoadBalancingCrossZoneEnabled(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_cross_zone_enabled", "false"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_updateStickinessEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albStickiness(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albStickiness(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albStickiness(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_targetHealthStateUnhealthyConnectionTermination(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_targetHealthStateConnectionTermination(rName, "TCP", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.0.enable_unhealthy_connection_termination", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_targetHealthStateConnectionTermination(rName, "TCP", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.0.enable_unhealthy_connection_termination", "true"),
				),
			},
			{
				Config: testAccTargetGroupConfig_targetHealthStateConnectionTermination(rName, "TLS", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.0.enable_unhealthy_connection_termination", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_targetHealthStateConnectionTermination(rName, "TLS", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_health_state.0.enable_unhealthy_connection_termination", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Instance_HealthCheck_defaults(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]map[string]struct {
		invalidHealthCheckProtocol bool
		expectedMatcher            string
		expectedPath               string
		expectedTimeout            string
	}{
		elbv2.ProtocolEnumHttp: {
			elbv2.ProtocolEnumHttp: {
				expectedMatcher: "200",
				expectedPath:    "/",
				expectedTimeout: "5",
			},
			elbv2.ProtocolEnumHttps: {
				expectedMatcher: "200",
				expectedPath:    "/",
				expectedTimeout: "5",
			},
			elbv2.ProtocolEnumTcp: {
				invalidHealthCheckProtocol: true,
			},
		},
		elbv2.ProtocolEnumHttps: {
			elbv2.ProtocolEnumHttp: {
				expectedMatcher: "200",
				expectedPath:    "/",
				expectedTimeout: "5",
			},
			elbv2.ProtocolEnumHttps: {
				expectedMatcher: "200",
				expectedPath:    "/",
				expectedTimeout: "5",
			},
			elbv2.ProtocolEnumTcp: {
				invalidHealthCheckProtocol: true,
			},
		},
		elbv2.ProtocolEnumTcp: {
			elbv2.ProtocolEnumHttp: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "6",
			},
			elbv2.ProtocolEnumHttps: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "10",
			},
			elbv2.ProtocolEnumTcp: {
				expectedMatcher: "",
				expectedPath:    "",
				expectedTimeout: "10",
			},
		},
		elbv2.ProtocolEnumTls: {
			elbv2.ProtocolEnumHttp: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "6",
			},
			elbv2.ProtocolEnumHttps: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "10",
			},
			elbv2.ProtocolEnumTcp: {
				expectedMatcher: "",
				expectedPath:    "",
				expectedTimeout: "10",
			},
		},
		elbv2.ProtocolEnumUdp: {
			elbv2.ProtocolEnumHttp: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "6",
			},
			elbv2.ProtocolEnumHttps: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "10",
			},
			elbv2.ProtocolEnumTcp: {
				expectedMatcher: "",
				expectedPath:    "",
				expectedTimeout: "10",
			},
		},
		elbv2.ProtocolEnumTcpUdp: {
			elbv2.ProtocolEnumHttp: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "6",
			},
			elbv2.ProtocolEnumHttps: {
				expectedMatcher: "200-399",
				expectedPath:    "/",
				expectedTimeout: "10",
			},
			elbv2.ProtocolEnumTcp: {
				expectedMatcher: "",
				expectedPath:    "",
				expectedTimeout: "10",
			},
		},
	}

	for _, protocol := range elbv2.ProtocolEnum_Values() {
		if protocol == elbv2.ProtocolEnumGeneve {
			continue
		}
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			protocolCase := testcases[protocol]
			if protocolCase == nil {
				t.Fatalf("missing case for target protocol %q", protocol)
			}

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := protocolCase[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)
					var targetGroup elbv2.TargetGroup

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealthCheck_basic(protocol, healthCheckProtocol),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else {
						step.Check = resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
							resource.TestCheckResourceAttr(resourceName, "protocol", protocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", tc.expectedMatcher),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.path", tc.expectedPath),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", healthCheckProtocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", tc.expectedTimeout),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
						)
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheck_matcher(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]map[string]struct {
		invalidHealthCheckProtocol bool
		invalidConfig              bool
		matcher                    string
	}{
		elbv2.ProtocolEnumHttp: {
			elbv2.ProtocolEnumHttp: {
				matcher: "200",
			},
			elbv2.ProtocolEnumHttps: {
				matcher: "200",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "200",
			},
		},
		elbv2.ProtocolEnumHttps: {
			elbv2.ProtocolEnumHttp: {
				matcher: "200",
			},
			elbv2.ProtocolEnumHttps: {
				matcher: "200",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "200",
			},
		},
		elbv2.ProtocolEnumTcp: {
			elbv2.ProtocolEnumHttp: {
				matcher: "200",
			},
			elbv2.ProtocolEnumHttps: {
				matcher: "200",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "200",
			},
		},
		elbv2.ProtocolEnumTls: {
			elbv2.ProtocolEnumHttp: {
				matcher: "200",
			},
			elbv2.ProtocolEnumHttps: {
				matcher: "200",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "200",
			},
		},
		elbv2.ProtocolEnumUdp: {
			elbv2.ProtocolEnumHttp: {
				matcher: "200",
			},
			elbv2.ProtocolEnumHttps: {
				matcher: "200",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "200",
			},
		},
		elbv2.ProtocolEnumTcpUdp: {
			elbv2.ProtocolEnumHttp: {
				matcher: "200",
			},
			elbv2.ProtocolEnumHttps: {
				matcher: "200",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "200",
			},
		},
	}

	for _, protocol := range elbv2.ProtocolEnum_Values() {
		if protocol == elbv2.ProtocolEnumGeneve {
			continue
		}
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			protocolCase := testcases[protocol]
			if protocolCase == nil {
				t.Fatalf("missing case for target protocol %q", protocol)
			}

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := protocolCase[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)
					var targetGroup elbv2.TargetGroup

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealthCheck_matcher(protocol, healthCheckProtocol, tc.matcher),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else if tc.invalidConfig {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.matcher" cannot be specified when "health_check.protocol" is "%s".`, healthCheckProtocol))
					} else {
						step.Check = resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "protocol", protocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", tc.matcher),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", healthCheckProtocol),
						)
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheck_path(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]map[string]struct {
		invalidHealthCheckProtocol bool
		invalidConfig              bool
		path                       string
	}{
		elbv2.ProtocolEnumHttp: {
			elbv2.ProtocolEnumHttp: {
				path: "/path",
			},
			elbv2.ProtocolEnumHttps: {
				path: "/path",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				path:          "/path",
			},
		},
		elbv2.ProtocolEnumHttps: {
			elbv2.ProtocolEnumHttp: {
				path: "/path",
			},
			elbv2.ProtocolEnumHttps: {
				path: "/path",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				path:          "/path",
			},
		},
		elbv2.ProtocolEnumTcp: {
			elbv2.ProtocolEnumHttp: {
				path: "/path",
			},
			elbv2.ProtocolEnumHttps: {
				path: "/path",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				path:          "/path",
			},
		},
		elbv2.ProtocolEnumTls: {
			elbv2.ProtocolEnumHttp: {
				path: "/path",
			},
			elbv2.ProtocolEnumHttps: {
				path: "/path",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				path:          "/path",
			},
		},
		elbv2.ProtocolEnumUdp: {
			elbv2.ProtocolEnumHttp: {
				path: "/path",
			},
			elbv2.ProtocolEnumHttps: {
				path: "/path",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				path:          "/path",
			},
		},
		elbv2.ProtocolEnumTcpUdp: {
			elbv2.ProtocolEnumHttp: {
				path: "/path",
			},
			elbv2.ProtocolEnumHttps: {
				path: "/path",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				path:          "/path",
			},
		},
	}

	for _, protocol := range elbv2.ProtocolEnum_Values() {
		if protocol == elbv2.ProtocolEnumGeneve {
			continue
		}
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			protocolCase := testcases[protocol]
			if protocolCase == nil {
				t.Fatalf("missing case for target protocol %q", protocol)
			}

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := protocolCase[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)
					var targetGroup elbv2.TargetGroup

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealthCheck_path(protocol, healthCheckProtocol, tc.path),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else if tc.invalidConfig {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.path" cannot be specified when "health_check.protocol" is "%s".`, healthCheckProtocol))
					} else {
						step.Check = resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "protocol", protocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.path", tc.path),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", healthCheckProtocol),
						)
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheck_matcherOutOfRange(t *testing.T) {
	t.Parallel()

	testcases := map[string]map[string]struct {
		invalidHealthCheckProtocol bool
		invalidConfig              bool
		matcher                    string
		validRange                 string
	}{
		elbv2.ProtocolEnumHttp: {
			elbv2.ProtocolEnumHttp: {
				matcher:    "500",
				validRange: "200-499",
			},
			elbv2.ProtocolEnumHttps: {
				matcher:    "500",
				validRange: "200-499",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "500",
			},
		},
		elbv2.ProtocolEnumHttps: {
			elbv2.ProtocolEnumHttp: {
				matcher:    "500",
				validRange: "200-499",
			},
			elbv2.ProtocolEnumHttps: {
				matcher:    "500",
				validRange: "200-499",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "500",
			},
		},
		elbv2.ProtocolEnumTcp: {
			elbv2.ProtocolEnumHttp: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumHttps: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "600",
			},
		},
		elbv2.ProtocolEnumTls: {
			elbv2.ProtocolEnumHttp: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumHttps: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "600",
			},
		},
		elbv2.ProtocolEnumUdp: {
			elbv2.ProtocolEnumHttp: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumHttps: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "600",
			},
		},
		elbv2.ProtocolEnumTcpUdp: {
			elbv2.ProtocolEnumHttp: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumHttps: {
				matcher:    "600",
				validRange: "200-599",
			},
			elbv2.ProtocolEnumTcp: {
				invalidConfig: true,
				matcher:       "600",
			},
		},
	}

	for _, protocol := range elbv2.ProtocolEnum_Values() {
		if protocol == elbv2.ProtocolEnumGeneve {
			continue
		}
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			protocolCase := testcases[protocol]
			if protocolCase == nil {
				t.Fatalf("missing case for target protocol %q", protocol)
			}

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := protocolCase[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealthCheck_matcher(protocol, healthCheckProtocol, tc.matcher),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else if tc.invalidConfig {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.matcher" cannot be specified when "health_check.protocol" is "%s".`, healthCheckProtocol))
					} else {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`ValidationError: Health check matcher HTTP code '%s' must be within '%s' inclusive`, tc.matcher, tc.validRange))
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheckGeneve_defaults(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]struct {
		expectedMatcher string
		expectedPath    string
		expectedTimeout string
	}{
		elbv2.ProtocolEnumHttp: {
			expectedMatcher: "200-399",
			expectedPath:    "/",
			expectedTimeout: "5",
		},
		elbv2.ProtocolEnumHttps: {
			expectedMatcher: "200-399",
			expectedPath:    "/",
			expectedTimeout: "5",
		},
		elbv2.ProtocolEnumTcp: {
			expectedMatcher: "",
			expectedPath:    "",
			expectedTimeout: "5",
		},
	}

	for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() { //nolint:paralleltest // false positive
		healthCheckProtocol := healthCheckProtocol

		t.Run(healthCheckProtocol, func(t *testing.T) {
			tc, ok := testcases[healthCheckProtocol]
			if !ok {
				t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
			}

			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccTargetGroupConfig_Instance_HealthCheckGeneve_basic(healthCheckProtocol),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
							resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", tc.expectedMatcher),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.path", tc.expectedPath),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"), // Should be 80
							resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", healthCheckProtocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", tc.expectedTimeout),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
						),
					},
				},
			})
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheckGRPC_defaults(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]struct {
		invalidHealthCheckProtocol bool
		expectedMatcher            string
		expectedPath               string
		expectedTimeout            string
	}{
		elbv2.ProtocolEnumHttp: {
			expectedMatcher: "12",
			expectedPath:    "/AWS.ALB/healthcheck",
			expectedTimeout: "5",
		},
		elbv2.ProtocolEnumHttps: {
			expectedMatcher: "12",
			expectedPath:    "/AWS.ALB/healthcheck",
			expectedTimeout: "5",
		},
		elbv2.ProtocolEnumTcp: {
			invalidHealthCheckProtocol: true,
		},
	}

	for _, protocol := range []string{elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps} {
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := testcases[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)
					var targetGroup elbv2.TargetGroup

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealhCheckGRPC_basic(protocol, healthCheckProtocol),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else {
						step.Check = resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "protocol", protocol),
							resource.TestCheckResourceAttr(resourceName, "protocol_version", "GRPC"),
							resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", tc.expectedMatcher),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.path", tc.expectedPath),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", healthCheckProtocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", tc.expectedTimeout),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
						)
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheckGRPC_path(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]struct {
		invalidHealthCheckProtocol bool
		invalidConfig              bool
		path                       string
	}{
		elbv2.ProtocolEnumHttp: {
			path: "/path",
		},
		elbv2.ProtocolEnumHttps: {
			path: "/path",
		},
		elbv2.ProtocolEnumTcp: {
			invalidConfig: true,
			path:          "/path",
		},
	}

	for _, protocol := range []string{elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps} {
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := testcases[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)
					var targetGroup elbv2.TargetGroup

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealhCheckGRPC_path(protocol, healthCheckProtocol, tc.path),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else if tc.invalidConfig {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.path" cannot be specified when "health_check.protocol" is "%s".`, healthCheckProtocol))
					} else {
						step.Check = resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "protocol", protocol),
							resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.path", tc.path),
							resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", healthCheckProtocol),
						)
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_HealthCheckGRPC_matcherOutOfRange(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		invalidHealthCheckProtocol bool
		matcher                    string
	}{
		elbv2.ProtocolEnumHttp: {
			matcher: "101",
		},
		elbv2.ProtocolEnumHttps: {
			matcher: "101",
		},
		elbv2.ProtocolEnumTcp: {
			invalidHealthCheckProtocol: true,
		},
	}

	for _, protocol := range []string{elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps} {
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			t.Parallel()

			for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() {
				healthCheckProtocol := healthCheckProtocol

				t.Run(healthCheckProtocol, func(t *testing.T) {
					tc, ok := testcases[healthCheckProtocol]
					if !ok {
						t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
					}

					ctx := acctest.Context(t)

					step := resource.TestStep{
						Config: testAccTargetGroupConfig_Instance_HealhCheckGRPC_matcher(protocol, healthCheckProtocol, tc.matcher),
					}
					if tc.invalidHealthCheckProtocol {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value "%s" when "protocol" is "%s".`, healthCheckProtocol, protocol))
					} else {
						step.ExpectError = regexache.MustCompile(fmt.Sprintf(`ValidationError: Health check matcher GRPC code '%s' must be within '0-99' inclusive`, tc.matcher))
					}
					resource.ParallelTest(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
						Steps: []resource.TestStep{
							step,
						},
					})
				})
			}
		})
	}
}

func TestAccELBV2TargetGroup_Instance_protocolVersion(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]struct {
		validConfig bool
	}{
		elbv2.ProtocolEnumHttp: {
			validConfig: true,
		},
		elbv2.ProtocolEnumHttps: {
			validConfig: true,
		},
		elbv2.ProtocolEnumTcp: {
			validConfig: false,
		},
		elbv2.ProtocolEnumTls: {
			validConfig: false,
		},
		elbv2.ProtocolEnumUdp: {
			validConfig: false,
		},
		elbv2.ProtocolEnumTcpUdp: {
			validConfig: false,
		},
	}

	for _, protocol := range elbv2.ProtocolEnum_Values() { //nolint:paralleltest // false positive
		if protocol == elbv2.ProtocolEnumGeneve {
			continue
		}
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			protocolCase, ok := testcases[protocol]
			if !ok {
				t.Fatalf("missing case for target protocol %q", protocol)
			}

			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			step := resource.TestStep{
				Config: testAccTargetGroupConfig_Instance_protocolVersion(protocol, "HTTP1"),
			}
			if protocolCase.validConfig {
				step.Check = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				)
			} else {
				step.Check = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", ""), // Should be Null
				)
			}

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
				Steps: []resource.TestStep{
					step,
				},
			})
		})
	}
}

func TestAccELBV2TargetGroup_Instance_protocolVersion_MigrateV0(t *testing.T) {
	t.Parallel()

	const resourceName = "aws_lb_target_group.test"

	testcases := map[string]struct {
		validConfig bool
	}{
		elbv2.ProtocolEnumHttp: {
			validConfig: true,
		},
		elbv2.ProtocolEnumHttps: {
			validConfig: true,
		},
		elbv2.ProtocolEnumTcp: {
			validConfig: false,
		},
		elbv2.ProtocolEnumTls: {
			validConfig: false,
		},
		elbv2.ProtocolEnumUdp: {
			validConfig: false,
		},
		elbv2.ProtocolEnumTcpUdp: {
			validConfig: false,
		},
	}

	for _, protocol := range elbv2.ProtocolEnum_Values() { //nolint:paralleltest // false positive
		if protocol == elbv2.ProtocolEnumGeneve {
			continue
		}
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			protocolCase, ok := testcases[protocol]
			if !ok {
				t.Fatalf("missing case for target protocol %q", protocol)
			}

			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			var (
				preCheck  resource.TestCheckFunc
				postCheck resource.TestCheckFunc
			)
			if protocolCase.validConfig {
				preCheck = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				)
				postCheck = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				)
			} else {
				preCheck = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
					resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
				)
				postCheck = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumInstance),
					resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
				)
			}

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
				CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
				Steps: testAccMigrateTest{
					PreviousVersion: "5.25.0",
					Config:          testAccTargetGroupConfig_Instance_protocolVersion(protocol, "HTTP1"),
					PreCheck:        preCheck,
					PostCheck:       postCheck,
				}.Steps(),
			})
		})
	}
}

func TestAccELBV2TargetGroup_Lambda_defaults(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_Lambda_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "port"),
					resource.TestCheckNoResourceAttr(resourceName, "protocol"),
					resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
					resource.TestCheckNoResourceAttr(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Lambda_defaults_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
		Steps: testAccMigrateTest{
			PreviousVersion: "5.25.0",
			Config:          testAccTargetGroupConfig_Lambda_basic(),
			PreCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				resource.TestCheckNoResourceAttr(resourceName, "port"),
				resource.TestCheckNoResourceAttr(resourceName, "protocol"),
				resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
				resource.TestCheckNoResourceAttr(resourceName, "vpc_id"),
				resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
			),
			PostCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				resource.TestCheckNoResourceAttr(resourceName, "port"),
				resource.TestCheckNoResourceAttr(resourceName, "protocol"),
				resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
				resource.TestCheckNoResourceAttr(resourceName, "vpc_id"),
				resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
			),
		}.Steps(),
	})
}

func TestAccELBV2TargetGroup_Lambda_vpc(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_Lambda_vpc(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", ""), // Should be Null
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Lambda_vpc_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
		Steps: testAccMigrateTest{
			PreviousVersion: "5.25.0",
			Config:          testAccTargetGroupConfig_Lambda_vpc(),
			PreCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
			),
			PostCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckResourceAttr(resourceName, "vpc_id", ""), // Should be Null
			),
		}.Steps(),
	})
}

func TestAccELBV2TargetGroup_Lambda_protocol(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	t.Parallel()

	for _, protocol := range elbv2.ProtocolEnum_Values() { //nolint:paralleltest // false positive
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccTargetGroupConfig_Lambda_protocol(protocol),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
							resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
							resource.TestCheckResourceAttr(resourceName, "protocol", ""), // Should be Null
						),
					},
				},
			})
		})
	}
}

func TestAccELBV2TargetGroup_Lambda_protocol_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	t.Parallel()

	for _, protocol := range elbv2.ProtocolEnum_Values() { //nolint:paralleltest // false positive
		protocol := protocol

		t.Run(protocol, func(t *testing.T) {
			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
				CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
				Steps: testAccMigrateTest{
					PreviousVersion: "5.25.0",
					Config:          testAccTargetGroupConfig_Lambda_protocol(protocol),
					PreCheck: resource.ComposeAggregateTestCheckFunc(
						testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
						resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
						resource.TestCheckResourceAttr(resourceName, "protocol", protocol),
					),
					PostCheck: resource.ComposeAggregateTestCheckFunc(
						testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
						resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
						resource.TestCheckResourceAttr(resourceName, "protocol", ""), // Should be Null
					),
				}.Steps(),
			})
		})
	}
}

func TestAccELBV2TargetGroup_Lambda_protocolVersion(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_Lambda_protocolVersion("HTTP1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", ""),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Lambda_protocolVersion_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
		Steps: testAccMigrateTest{
			PreviousVersion: "5.25.0",
			Config:          testAccTargetGroupConfig_Lambda_protocolVersion("GRPC"),
			PreCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
			),
			PostCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckNoResourceAttr(resourceName, "protocol_version"),
			),
		}.Steps(),
	})
}

func TestAccELBV2TargetGroup_Lambda_port(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_Lambda_port("443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
					resource.TestCheckResourceAttr(resourceName, "port", "0"), // Should be Null
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Lambda_port_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
		Steps: testAccMigrateTest{
			PreviousVersion: "5.25.0",
			Config:          testAccTargetGroupConfig_Lambda_port("443"),
			PreCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckResourceAttr(resourceName, "port", "443"),
			),
			PostCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "target_type", elbv2.TargetTypeEnumLambda),
				resource.TestCheckResourceAttr(resourceName, "port", "0"), // Should be Null
			),
		}.Steps(),
	})
}

func TestAccELBV2TargetGroup_Lambda_HealthCheck_basic(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_Lambda_HealthCheck_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "40"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "35"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Lambda_HealthCheck_basic_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	ctx := acctest.Context(t)
	var targetGroup elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
		Steps: testAccMigrateTest{
			PreviousVersion: "5.25.0",
			Config:          testAccTargetGroupConfig_Lambda_HealthCheck_basic(),
			PreCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "40"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.port", ""),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", ""),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "35"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
			),
			PostCheck: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
				resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "40"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.port", ""),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", ""),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "35"),
				resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
			),
		}.Steps(),
	})
}

func TestAccELBV2TargetGroup_Lambda_HealthCheck_protocol(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	t.Parallel()

	testcases := map[string]struct {
		invalidHealthCheckProtocol bool
		warning                    bool
	}{
		elbv2.ProtocolEnumHttp: {
			warning: true,
		},
		elbv2.ProtocolEnumHttps: {
			warning: true,
		},
		elbv2.ProtocolEnumTcp: {
			invalidHealthCheckProtocol: true,
		},
	}

	for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() { //nolint:paralleltest // false positive
		healthCheckProtocol := healthCheckProtocol

		t.Run(healthCheckProtocol, func(t *testing.T) {
			tc, ok := testcases[healthCheckProtocol]
			if !ok {
				t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
			}

			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			step := resource.TestStep{
				Config: testAccTargetGroupConfig_Lambda_HealthCheck_protocol(healthCheckProtocol),
			}
			if tc.invalidHealthCheckProtocol {
				step.ExpectError = regexache.MustCompile(fmt.Sprintf(`Attribute "health_check.protocol" cannot have value %q when "target_type" is "lambda"`, healthCheckProtocol))
			} else if tc.warning {
				step.Check = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", ""),
				)
			} else {
				t.Fatal("invalid test case, one of invalidHealthCheckProtocol or warning must be set")
			}
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
				Steps: []resource.TestStep{
					step,
				},
			})
		})
	}
}

func TestAccELBV2TargetGroup_Lambda_HealthCheck_protocol_MigrateV0(t *testing.T) {
	const resourceName = "aws_lb_target_group.test"

	t.Parallel()

	testcases := map[string]struct {
		invalidHealthCheckProtocol bool
		warning                    bool
	}{
		elbv2.ProtocolEnumHttp: {
			warning: true,
		},
		elbv2.ProtocolEnumHttps: {
			warning: true,
		},
		elbv2.ProtocolEnumTcp: {
			invalidHealthCheckProtocol: true,
		},
	}

	for _, healthCheckProtocol := range tfelbv2.HealthCheckProtocolEnumValues() { //nolint:paralleltest // false positive
		healthCheckProtocol := healthCheckProtocol

		t.Run(healthCheckProtocol, func(t *testing.T) {
			tc, ok := testcases[healthCheckProtocol]
			if !ok {
				t.Fatalf("missing case for health check protocol %q", healthCheckProtocol)
			}

			ctx := acctest.Context(t)
			var targetGroup elbv2.TargetGroup

			config := testAccTargetGroupConfig_Lambda_HealthCheck_protocol(healthCheckProtocol)

			step := resource.TestStep{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.25.0",
					},
				},
			}
			if tc.invalidHealthCheckProtocol {
				// Lambda health checks don't take a protocol, but are effectively an HTTP check.
				// So, they return a `matcher` on read. When Terraform validates the diff, the (incorrectly stored) protocol
				// `TCP` is checked against the `matcher`, and returns an error.
				step.ExpectError = regexache.MustCompile(`health_check.matcher is not supported for target_groups with TCP protocol`)
			} else if tc.warning {
				step.Check = resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", ""),
				)
			} else {
				t.Fatal("invalid test case, one of invalidHealthCheckProtocol or warning must be set")
			}

			steps := []resource.TestStep{step}
			if tc.warning {
				steps = append(steps, resource.TestStep{
					ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
					Config:                   config,
					PlanOnly:                 true,
				})
			}

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
				CheckDestroy: testAccCheckTargetGroupDestroy(ctx),
				Steps:        steps,
			})
		})
	}
}

func TestAccELBV2TargetGroup_Lambda_HealthCheck_matcherOutOfRange(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_Lambda_HealthCheck_matcher("999"),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`ValidationError: Health check matcher HTTP code '%s' must be within '200-499' inclusive`, "999")),
			},
		},
	})
}

func testAccCheckTargetGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_target_group" && rs.Type != "aws_alb_target_group" {
				continue
			}

			_, err := tfelbv2.FindTargetGroupByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Target Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTargetGroupExists(ctx context.Context, n string, v *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ELBv2 Target Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn(ctx)

		output, err := tfelbv2.FindTargetGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTargetGroupNotRecreated(i, j *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TargetGroupArn) != aws.StringValue(j.TargetGroupArn) {
			return errors.New("ELBv2 Target Group was recreated")
		}

		return nil
	}
}

func testAccCheckTargetGroupRecreated(i, j *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TargetGroupArn) == aws.StringValue(j.TargetGroupArn) {
			return errors.New("ELBv2 Target Group was not recreated")
		}

		return nil
	}
}

type testAccMigrateTest struct {
	// PreviousVersion is a version of the provider previous to the changes to be migrated
	PreviousVersion string

	// Config is the configuration to be deployed with the previous version and checked with the updated version
	Config string

	// PreCheck is a check function to validate the values prior to migration
	PreCheck resource.TestCheckFunc

	PostCheck resource.TestCheckFunc
}

func (t testAccMigrateTest) Steps() []resource.TestStep {
	return []resource.TestStep{
		{
			ExternalProviders: map[string]resource.ExternalProvider{
				"aws": {
					Source:            "hashicorp/aws",
					VersionConstraint: t.PreviousVersion,
				},
			},
			Config: t.Config,
			Check:  t.PreCheck,
		},
		{
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Config:                   t.Config,
			PlanOnly:                 true,
		},
		{
			ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
			Config:                   t.Config,
			Check:                    t.PostCheck,
		},
	}
}

func testAccTargetGroupConfig_basic(rName string, deregDelay int) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = %[2]d
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, deregDelay)
}

func testAccTargetGroupConfig_nameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_namePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name_prefix = %[2]q
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, namePrefix)
}

func testAccTargetGroupConfig_duplicateName(rName string, deregDelay int) string {
	return acctest.ConfigCompose(testAccTargetGroupConfig_basic(rName, deregDelay), fmt.Sprintf(`
resource "aws_lb_target_group" "test2" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTargetGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    %[2]q = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTargetGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccTargetGroupConfig_albDefaults(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  # HTTP Only

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "HTTP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

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
`, rName)
}

func testAccTargetGroupConfig_backwardsCompatibility(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_protocolGeneve(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

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
    port     = 80
    protocol = "HTTP"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_protocolGeneveSticky(rName, stickinessType string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id
  stickiness {
    enabled = true
    type    = %[2]q
  }
  tags = {
    Name = %[1]q
  }
}
`, rName, stickinessType)
}

func testAccTargetGroupConfig_protocolGeneveTargetFailover(rName, failoverType string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id
  target_failover {
    on_deregistration = %[2]q
    on_unhealthy      = %[2]q
  }
  tags = {
    Name = %[1]q
  }
}
`, rName, failoverType)
}

func testAccTargetGroupConfig_grpcProtocolVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name             = %[1]q
  port             = 80
  protocol         = "HTTP"
  protocol_version = "GRPC"
  vpc_id           = aws_vpc.test2.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/Test.Check/healthcheck"
    interval            = 60
    port                = 8080
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "0-99"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_protocolVersion(rName, protocolVersion string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name             = %[1]q
  port             = 443
  protocol         = "HTTPS"
  protocol_version = %[2]q
  vpc_id           = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, protocolVersion)
}

func testAccTargetGroupConfig_ipAddressType(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TLS"
  vpc_id   = aws_vpc.test.id

  target_type     = "ip"
  ip_address_type = "ipv6"

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
`, rName)
}

func testAccTargetGroupConfig_protocolTLS(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

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
`, rName)
}

func testAccTargetGroupConfig_appStickiness(rName string, addAppStickinessBlock bool, enabled bool) string {
	var appSstickinessBlock string

	if addAppStickinessBlock {
		appSstickinessBlock = fmt.Sprintf(`
stickiness {
  enabled         = "%[1]t"
  type            = "app_cookie"
  cookie_name     = "Cookie"
  cookie_duration = 10000
}
`, enabled)
	}

	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  %[2]s

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, appSstickinessBlock)
}

func testAccTargetGroupConfig_basicUdp(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 514
  protocol = "UDP"
  vpc_id   = aws_vpc.test.id

  health_check {
    protocol = "TCP"
    port     = 514
  }

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
`, rName)
}

func testAccTargetGroupConfig_enableHealthcheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"

  health_check {
    path     = "/health"
    interval = 60
  }
}
`, rName)
}

func testAccTargetGroupConfig_preserveClientIP(rName string, preserveClientIP bool) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  preserve_client_ip = %[2]t

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName, preserveClientIP)
}

func testAccTargetGroupConfig_stickiness(rName string, addStickinessBlock bool, enabled bool) string {
	var stickinessBlock string

	if addStickinessBlock {
		stickinessBlock = fmt.Sprintf(`
stickiness {
  enabled         = "%[1]t"
  type            = "lb_cookie"
  cookie_duration = 10000
}
`, enabled)
	}

	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  %[2]s

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, stickinessBlock)
}

func testAccTargetGroupConfig_stickinessDefault(rName, protocol string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 25
  protocol    = %[2]q
  vpc_id      = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, protocol)
}

func testAccTargetGroupConfig_stickinessValidity(rName, protocol, stickyType string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 25
  protocol    = %[1]q
  vpc_id      = aws_vpc.test.id

  stickiness {
    type    = %[2]q
    enabled = %[3]t
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[4]q
  }
}
`, protocol, stickyType, enabled, rName)
}

func testAccTargetGroupConfig_targetHealthStateConnectionTermination(rName, protocol string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 25
  protocol = %[2]q
  vpc_id   = aws_vpc.test.id

  target_health_state {
    enable_unhealthy_connection_termination = %[3]t
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, protocol, enabled)
}

func testAccTargetGroupConfig_typeTCP(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_typeTCPHealthCheckUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  health_check {
    interval            = 20
    port                = "8081"
    protocol            = "TCP"
    timeout             = 15
    healthy_threshold   = 4
    unhealthy_threshold = 5
  }

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
`, rName)
}

func testAccTargetGroupConfig_typeTCPHTTPHealthCheck(rName, path string, threshold int) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = "%[1]s"
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  health_check {
    healthy_threshold   = %[2]d
    unhealthy_threshold = %[2]d
    timeout             = "10"
    port                = "443"
    path                = "%[3]s"
    protocol            = "HTTPS"
    interval            = 30
    matcher             = "200-399"
  }

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
`, rName, threshold, path)
}

func testAccTargetGroupConfig_typeTCPProxyProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  proxy_protocol_v2    = true
  deregistration_delay = 200

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_typeTCPConnectionTermination(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  connection_termination = true
  deregistration_delay   = 200

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_updateHealthCheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_updatedPort(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 442
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_updatedProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTP"
  vpc_id   = aws_vpc.test2.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_updatedVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_noHealthcheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}
`, rName)
}

func testAccTargetGroupConfig_protocolGeneveHealth(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

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
    path                = "/health"
    interval            = 60
    port                = 80
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }
}
`, rName)
}

func testAccTargetGroupConfig_nlbDefaults(rName, healthCheckBlock string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  health_check {
    %[2]s
  }

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
`, rName, healthCheckBlock)
}

func testAccTargetGroupConfig_albBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    TestName = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccTargetGroupConfig_albGeneratedName(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_albLambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}`, rName)
}

func testAccTargetGroupConfig_albLambdaMultiValueHeadersEnabled(rName string, lambdaMultiValueHadersEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  lambda_multi_value_headers_enabled = %[1]t
  name                               = %[2]q
  target_type                        = "lambda"
}
`, lambdaMultiValueHadersEnabled, rName)
}

func testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName string, nonDefault bool, algoType string) string {
	var algoTypeParam string

	if nonDefault {
		algoTypeParam = fmt.Sprintf(`load_balancing_algorithm_type = "%s"`, algoType)
	}

	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  %[2]s
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName, algoTypeParam)
}

func testAccTargetGroupConfig_albLoadBalancingCrossZoneEnabled(rName string, nonDefault bool, enabled bool) string {
	var crossZoneParam string

	if nonDefault {
		crossZoneParam = fmt.Sprintf(`load_balancing_cross_zone_enabled = "%v"`, enabled)
	}

	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  %[2]s
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName, crossZoneParam)
}

func testAccTargetGroupConfig_albMissingPort(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}`, rName)
}

func testAccTargetGroupConfig_albMissingProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name   = %[1]q
  port   = 443
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}`, rName)
}

func testAccTargetGroupConfig_albMissingVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
}
`, rName)
}

func testAccTargetGroupConfig_albNamePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name_prefix = "tf-"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTargetGroupConfig_albStickiness(rName string, addStickinessBlock bool, enabled bool) string {
	var stickinessBlock string

	if addStickinessBlock {
		stickinessBlock = fmt.Sprintf(`
	  stickiness {
	    enabled         = "%t"
	    type            = "lb_cookie"
	    cookie_duration = 10000
	  }`, enabled)
	}

	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  %[2]s

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName, stickinessBlock)
}

func testAccTargetGroupConfig_albUpdateHealthCheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccTargetGroupConfig_albUpdateSlowStart(rName string, slowStartDuration int) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = %[2]d

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    TestName = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName, slowStartDuration)
}

func testAccTargetGroupConfig_albUpdateTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    Environment = "Production"
    Type        = "ALB Target Group"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccTargetGroupConfig_albUpdatedPort(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 442
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    TestName = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccTargetGroupConfig_albUpdatedProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTP"
  vpc_id   = aws_vpc.test2.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    TestName = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccTargetGroupConfig_albUpdatedVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

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
    TestName = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccTargetGroupConfig_Instance_HealthCheck_basic(protocol, healthCheckProtocol string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port     = 443
  protocol = %[1]q
  vpc_id   = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, healthCheckProtocol)
}

func testAccTargetGroupConfig_Instance_HealthCheck_matcher(protocol, healthCheckProtocol, matcher string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port     = 443
  protocol = %[1]q
  vpc_id   = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[2]q
    matcher  = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, healthCheckProtocol, matcher)
}

func testAccTargetGroupConfig_Instance_HealthCheck_path(protocol, healthCheckProtocol, matcher string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port     = 443
  protocol = %[1]q
  vpc_id   = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[2]q
    path     = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, healthCheckProtocol, matcher)
}

func testAccTargetGroupConfig_Instance_HealthCheckGeneve_basic(healthCheckProtocol string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, healthCheckProtocol)
}

func testAccTargetGroupConfig_Instance_HealhCheckGRPC_basic(protocol, healthCheckProtocol string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port             = 443
  protocol         = %[1]q
  protocol_version = "GRPC"
  vpc_id           = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, healthCheckProtocol)
}

func testAccTargetGroupConfig_Instance_HealhCheckGRPC_path(protocol, healthCheckProtocol, path string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port             = 443
  protocol         = %[1]q
  protocol_version = "GRPC"
  vpc_id           = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[2]q
    path     = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, healthCheckProtocol, path)
}

func testAccTargetGroupConfig_Instance_HealhCheckGRPC_matcher(protocol, healthCheckProtocol, matcher string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port             = 443
  protocol         = %[1]q
  protocol_version = "GRPC"
  vpc_id           = aws_vpc.test.id

  target_type = "instance"

  health_check {
    protocol = %[2]q
    matcher  = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, healthCheckProtocol, matcher)
}

func testAccTargetGroupConfig_Instance_protocolVersion(protocol, protocolVersion string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  target_type = "instance"

  port             = 443
  protocol         = %[1]q
  protocol_version = %[2]q
  vpc_id           = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`, protocol, protocolVersion)
}

func testAccTargetGroupConfig_Lambda_basic() string {
	return `
resource "aws_lb_target_group" "test" {
  target_type = "lambda"
}
`
}

func testAccTargetGroupConfig_Lambda_vpc() string {
	return `
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
`
}

func testAccTargetGroupConfig_Lambda_protocol(protocol string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  protocol = %[1]q
}
`, protocol)
}

func testAccTargetGroupConfig_Lambda_protocolVersion(protocolVersion string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  protocol_version = %[1]q
}
`, protocolVersion)
}

func testAccTargetGroupConfig_Lambda_port(port string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  port = %[1]q
}
`, port)
}

func testAccTargetGroupConfig_Lambda_HealthCheck_basic() string {
	return `
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  health_check {
    timeout  = 35 # The Terraform default (30) is too short for Lambda.
    interval = 40 # Must be > timeout
  }
}
`
}

func testAccTargetGroupConfig_Lambda_HealthCheck_protocol(healthCheckProtocol string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  health_check {
    protocol = %[1]q
    timeout  = 35 # The Terraform default (30) is too short for Lambda.
    interval = 40 # Must be > timeout
  }
}
`, healthCheckProtocol)
}

func testAccTargetGroupConfig_Lambda_HealthCheck_matcher(matcher string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  target_type = "lambda"

  health_check {
    matcher  = %[1]q
    timeout  = 35 # The Terraform default (30) is too short for Lambda.
    interval = 40 # Must be > timeout
  }
}
`, matcher)
}
