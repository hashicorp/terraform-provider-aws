package elbv2_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
)

func TestLBTargetGroupCloudWatchSuffixFromARN(t *testing.T) {
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

func TestAccELBV2TargetGroup_backwardsCompatibility(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
		},
	})
}

func TestAccELBV2TargetGroup_ProtocolVersion_grpcHealthCheck(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_grpcProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/Test.Check/healthcheck"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "0-99"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ProtocolVersion_grpcUpdate(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				),
			},
			{
				Config: testAccTargetGroupConfig_grpcProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "GRPC"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_HealthCheck_tcp(t *testing.T) {
	var targetGroup1, targetGroup2 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCPIntervalUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPHTTPHealthCheck(rName, "/", 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup2),
					testAccCheckTargetGroupRecreated(&targetGroup1, &targetGroup2),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_tls(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolTLS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_HealthCheck_tcpHTTPS(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCPHTTPHealthCheck(rName, "/healthz", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confBefore),
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
					testAccCheckTargetGroupExists(resourceName, &confAfter),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "0"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
				),
			},
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_basic(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_udp(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basicUdp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var before, after elbv2.TargetGroup
	rNameBefore := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameAfter := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rNameBefore, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rNameBefore),
				),
			},
			{
				Config: testAccTargetGroupConfig_basic(rNameAfter, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameAfter),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_port(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
				),
			},
			{
				Config: testAccTargetGroupConfig_updatedPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "port", "442"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_protocol(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccTargetGroupConfig_updatedProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ForceNew_vpc(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &before),
				),
			},
			{
				Config: testAccTargetGroupConfig_updatedVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &after),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Defaults_application(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albDefaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"
	healthCheckInvalid1 := `
path     = "/health"
interval = 10
port     = 8081
protocol = "TCP"
    `
	healthCheckInvalid2 := `
interval = 10
port     = 8081
protocol = "TCP"
matcher  = "200"
    `
	healthCheckInvalid3 := `
interval = 10
port     = 8081
protocol = "TCP"
timeout  = 4
    `
	healthCheckValid := `
interval = 10
port     = 8081
protocol = "TCP"
    `

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_nlbDefaults(rName, healthCheckInvalid1),
				ExpectError: regexp.MustCompile("health_check.path is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccTargetGroupConfig_nlbDefaults(rName, healthCheckInvalid2),
				ExpectError: regexp.MustCompile("health_check.matcher is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccTargetGroupConfig_nlbDefaults(rName, healthCheckInvalid3),
				ExpectError: regexp.MustCompile("health_check.timeout is not supported for target_groups with TCP protocol"),
			},
			{
				Config: testAccTargetGroupConfig_nlbDefaults(rName, healthCheckValid),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_noHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_enableHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Name_generated(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_generatedName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Name_prefix(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_NetworkLB_targetGroup(t *testing.T) {
	var targetGroup1, targetGroup2, targetGroup3 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
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
				Config:      testAccTargetGroupConfig_typeTCPInvalidThreshold(rName),
				ExpectError: regexp.MustCompile(`health_check\.healthy_threshold [0-9]+ and health_check\.unhealthy_threshold [0-9]+ must be the same for target_groups with TCP protocol`),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPThresholdUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup2),
					testAccCheckTargetGroupNotRecreated(&targetGroup1, &targetGroup2),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPIntervalUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup3),
					testAccCheckTargetGroupRecreated(&targetGroup2, &targetGroup3),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_networkLB_TargetGroupWithConnectionTermination(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confBefore),
					resource.TestCheckResourceAttr(resourceName, "connection_termination", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPConnectionTermination(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confAfter),
					resource.TestCheckResourceAttr(resourceName, "connection_termination", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_NetworkLB_targetGroupWithProxy(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confBefore),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCPProxyProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confAfter),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "true"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_preserveClientIPValid(t *testing.T) {
	var conf elbv2.TargetGroup
	resourceName := "aws_lb_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_preserveClientIP(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "true"),
				),
			},
			{
				Config: testAccTargetGroupConfig_preserveClientIP(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "false"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Geneve_basic(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolGeneve(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckGatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_protocolGeneve(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
				),
			},
			{
				Config: testAccTargetGroupConfig_protocolGeneveHealth(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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

func TestAccELBV2TargetGroup_Stickiness_defaultALB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "HTTP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_defaultNLB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "TCP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "UDP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessDefault(rName, "TCP_UDP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_invalidALB(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "HTTP", "source_ip", true),
				ExpectError: regexp.MustCompile("Stickiness type 'source_ip' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "HTTPS", "source_ip", true),
				ExpectError: regexp.MustCompile("Stickiness type 'source_ip' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TLS", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", false),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "UDP", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_validALB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "HTTP", "lb_cookie", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "86400"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "HTTPS", "lb_cookie", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "source_ip", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "UDP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_tags(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccTargetGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTargetGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_Stickiness_updateAppEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_appStickiness(targetGroupName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", targetGroupName),
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
				Config: testAccTargetGroupConfig_appStickiness(targetGroupName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", targetGroupName),
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
				Config: testAccTargetGroupConfig_appStickiness(targetGroupName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", targetGroupName),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_stickiness(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
					testAccCheckTargetGroupExists(resourceName, &conf),
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
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_noHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
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
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameAfter := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_albBasic(rNameAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameAfter),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changePortForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdatedPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "port", "442"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changeProtocolForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdatedProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_changeVPCForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdatedVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_generatedName(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albGeneratedName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_lambda(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &targetGroup1),
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
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLambdaMultiValueHeadersEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &targetGroup1),
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
					testAccCheckATargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLambdaMultiValueHeadersEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_missingPortProtocolVPC(t *testing.T) {
	rName := fmt.Sprintf("test-target-group-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_albMissingPort(rName),
				ExpectError: regexp.MustCompile(`port should be set when target type is`),
			},
			{
				Config:      testAccTargetGroupConfig_albMissingProtocol(rName),
				ExpectError: regexp.MustCompile(`protocol should be set when target type is`),
			},
			{
				Config:      testAccTargetGroupConfig_albMissingVPC(rName),
				ExpectError: regexp.MustCompile(`vpc_id should be set when target type is`),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_namePrefix(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_setAndUpdateSlowStart(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albUpdateSlowStart(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "30"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdateSlowStart(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "60"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_tags(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
			{
				Config: testAccTargetGroupConfig_albUpdateTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Type", "ALB Target Group"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_updateHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
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
					testAccCheckATargetGroupExists(resourceName, &conf),
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
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "round_robin"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName, true, "round_robin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "round_robin"),
				),
			},
			{
				Config: testAccTargetGroupConfig_albLoadBalancingAlgorithm(rName, true, "least_outstanding_requests"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "least_outstanding_requests"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroup_ALBAlias_updateStickinessEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_albStickiness(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
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
					testAccCheckATargetGroupExists(resourceName, &conf),
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
					testAccCheckATargetGroupExists(resourceName, &conf),
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
}
`, rName)
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

func testAccTargetGroupConfig_protocolVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name             = %[1]q
  port             = 443
  protocol         = "HTTPS"
  protocol_version = "HTTP2"
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

func testAccTargetGroupConfig_appStickiness(targetGroupName string, addAppStickinessBlock bool, enabled bool) string {
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
    Name = "terraform-testacc-lb-target-group-stickiness"
  }
}
`, targetGroupName, appSstickinessBlock)
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

func testAccTargetGroupConfig_generatedName(rName string) string {
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

func testAccTargetGroupConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
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
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 25
  protocol    = %[2]q
  vpc_id      = aws_vpc.test.id
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

func testAccTargetGroupConfig_typeTCPIntervalUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  health_check {
    interval            = 30
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 5
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

func testAccTargetGroupConfig_typeTCPInvalidThreshold(rName string) string {
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
    unhealthy_threshold = 4
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

func testAccTargetGroupConfig_typeTCPThresholdUpdated(rName string) string {
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
    healthy_threshold   = 5
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

func testAccCheckTargetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_target_group" && rs.Type != "aws_alb_target_group" {
			continue
		}

		targetGroup, err := tfelbv2.FindTargetGroupByARN(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("unexpected error checking ALB (%s) destroyed: %w", rs.Primary.ID, err)
		}

		if targetGroup == nil {
			continue
		}

		return fmt.Errorf("Target Group %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckTargetGroupExists(n string, res *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Target Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		targetGroup, err := tfelbv2.FindTargetGroupByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading ELBv2 Target Group (%s): %w", rs.Primary.ID, err)
		}

		if targetGroup == nil {
			return fmt.Errorf("Target Group (%s) not found", rs.Primary.ID)
		}

		*res = *targetGroup
		return nil
	}
}

func testAccCheckTargetGroupNotRecreated(i, j *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TargetGroupArn) != aws.StringValue(j.TargetGroupArn) {
			return fmt.Errorf("ELBv2 Target Group (%s) unexpectedly recreated (%s)", aws.StringValue(i.TargetGroupArn), aws.StringValue(j.TargetGroupArn))
		}

		return nil
	}
}

func testAccCheckTargetGroupRecreated(i, j *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TargetGroupArn) == aws.StringValue(j.TargetGroupArn) {
			return fmt.Errorf("ELBv2 Target Group (%s) not recreated", aws.StringValue(i.TargetGroupArn))
		}

		return nil
	}
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

func testAccCheckATargetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_alb_target_group" {
			continue
		}

		describe, err := conn.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.TargetGroups) != 0 &&
				*describe.TargetGroups[0].TargetGroupArn == rs.Primary.ID {
				return fmt.Errorf("Target Group %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
			return nil
		} else {
			return fmt.Errorf("Unexpected error checking ALB destroyed: %s", err)
		}
	}

	return nil
}

func testAccCheckATargetGroupExists(n string, res *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Target Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		describe, err := conn.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.TargetGroups) != 1 ||
			*describe.TargetGroups[0].TargetGroupArn != rs.Primary.ID {
			return errors.New("Target Group not found")
		}

		*res = *describe.TargetGroups[0]
		return nil
	}
}
