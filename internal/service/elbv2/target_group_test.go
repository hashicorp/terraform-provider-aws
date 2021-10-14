package elbv2_test

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_lb_target_group", &resource.Sweeper{
		Name: "aws_lb_target_group",
		F:    sweepTargetGroups,
		Dependencies: []string{
			"aws_lb",
		},
	})
}

func sweepTargetGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ELBV2Conn

	err = conn.DescribeTargetGroupsPages(&elbv2.DescribeTargetGroupsInput{}, func(page *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
		if page == nil || len(page.TargetGroups) == 0 {
			log.Print("[DEBUG] No LB Target Groups to sweep")
			return false
		}

		for _, targetGroup := range page.TargetGroups {
			name := aws.StringValue(targetGroup.TargetGroupName)

			log.Printf("[INFO] Deleting LB Target Group: %s", name)
			_, err := conn.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
				TargetGroupArn: targetGroup.TargetGroupArn,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete LB Target Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping LB Target Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving LB Target Groups: %w", err)
	}
	return nil
}

func TestLBTargetGroupCloudwatchSuffixFromARN(t *testing.T) {
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

func TestALBTargetGroupCloudwatchSuffixFromARN(t *testing.T) {
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

func TestAccAWSLBTargetGroup_BackwardsCompatibility(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupBackwardsCompatibilityConfig(rName),
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

func TestAccAWSLBTargetGroup_ProtocolVersion(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_ProtocolVersion(rName),
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

func TestAccAWSLBTargetGroup_ProtocolVersion_GRPC_HealthCheck(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_GRPC_ProtocolVersion(rName),
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

func TestAccAWSLBTargetGroup_ProtocolVersion_HTTP_GRPC_Update(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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
				Config: testAccTargetGroupConfig_GRPC_ProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "GRPC"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_Protocol_Tcp_HealthCheck_Protocol(t *testing.T) {
	var targetGroup1, targetGroup2 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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
				Config: testAccTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, "/", 5),
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

func TestAccAWSLBTargetGroup_Protocol_Tls(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_Protocol_TLS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_TCP_HTTPHealthCheck(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, "/healthz", 2),
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
				Config: testAccTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, "/healthz2", 4),
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

func TestAccAWSLBTargetGroup_attrsOnCreate(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_basic(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_basicUdp(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_changeNameForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rNameBefore := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameAfter := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_changePortForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_changeProtocolForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_changeVpcForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_defaults_application(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALB_defaults(rName),
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

func TestAccAWSLBTargetGroup_defaults_network(t *testing.T) {
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNLB_defaults(rName, healthCheckInvalid1),
				ExpectError: regexp.MustCompile("health_check.path is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccNLB_defaults(rName, healthCheckInvalid2),
				ExpectError: regexp.MustCompile("health_check.matcher is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccNLB_defaults(rName, healthCheckInvalid3),
				ExpectError: regexp.MustCompile("health_check.timeout is not supported for target_groups with TCP protocol"),
			},
			{
				Config: testAccNLB_defaults(rName, healthCheckValid),
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

func TestAccAWSLBTargetGroup_enableHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_withoutHealthcheck(rName),
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

func TestAccAWSLBTargetGroup_generatedName(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_namePrefix(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_networkLB_TargetGroup(t *testing.T) {
	var targetGroup1, targetGroup2, targetGroup3 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_networkLB_TargetGroupWithProxy(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confBefore),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "false"),
				),
			},
			{
				Config: testAccTargetGroupConfig_typeTCP_withProxyProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &confAfter),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "true"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_preserveClientIPValid(t *testing.T) {
	var conf elbv2.TargetGroup
	resourceName := "aws_lb_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_protocolGeneve(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupProtocolGeneveConfig(rName),
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
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccAWSLBTargetGroup_protocolGeneveNotSticky(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupProtocolGeneveConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
				),
			},
			{
				Config: testAccTargetGroupProtocolGeneveHealthConfig(rName),
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

func TestAccAWSLBTargetGroup_stickinessDefaultALB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_stickinessDefaultNLB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_stickinessInvalidALB(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_stickinessInvalidNLB(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", true),
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

func TestAccAWSLBTargetGroup_stickinessValidALB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_stickinessValidNLB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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
				// this test should be invalid but allowed to avoid breaking changes
				Config: testAccTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
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

func TestAccAWSLBTargetGroup_tags(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccTargetGroupTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTargetGroupTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_updateAppSticknessEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_updateHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_updateSticknessEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
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

func TestAccAWSLBTargetGroup_withoutHealthcheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_withoutHealthcheck(rName),
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

func TestAccAWSALBTargetGroup_basic(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
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

func TestAccAWSALBTargetGroup_changeNameForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameAfter := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccATargetGroupConfig_basic(rNameAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameAfter),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_changePortForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
				),
			},
			{
				Config: testAccATargetGroupConfig_updatedPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "port", "442"),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_changeProtocolForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccATargetGroupConfig_updatedProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_changeVpcForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
				),
			},
			{
				Config: testAccATargetGroupConfig_updatedVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_generatedName(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_generatedName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_lambda(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_lambda(rName),
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
					"deregistration_delay",
					"proxy_protocol_v2",
					"slow_start",
					"load_balancing_algorithm_type",
				},
			},
		},
	})
}

func TestAccAWSALBTargetGroup_lambdaMultiValueHeadersEnabled(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_lambdaMultiValueHeadersEnabled(rName, true),
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
					"deregistration_delay",
					"proxy_protocol_v2",
					"slow_start",
					"load_balancing_algorithm_type",
				},
			},
			{
				Config: testAccATargetGroupConfig_lambdaMultiValueHeadersEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
			{
				Config: testAccATargetGroupConfig_lambdaMultiValueHeadersEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "lambda_multi_value_headers_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "lambda"),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_missingPortProtocolVpc(t *testing.T) {
	rName := fmt.Sprintf("test-target-group-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccATargetGroupConfig_missing_port(rName),
				ExpectError: regexp.MustCompile(`port should be set when target type is`),
			},
			{
				Config:      testAccATargetGroupConfig_missing_protocol(rName),
				ExpectError: regexp.MustCompile(`protocol should be set when target type is`),
			},
			{
				Config:      testAccATargetGroupConfig_missing_vpc(rName),
				ExpectError: regexp.MustCompile(`vpc_id should be set when target type is`),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_namePrefix(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_setAndUpdateSlowStart(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_updateSlowStart(rName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "30"),
				),
			},
			{
				Config: testAccATargetGroupConfig_updateSlowStart(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "60"),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroup_tags(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
			{
				Config: testAccATargetGroupConfig_updateTags(rName),
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

func TestAccAWSALBTargetGroup_updateHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_basic(rName),
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
				Config: testAccATargetGroupConfig_updateHealthCheck(rName),
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

func TestAccAWSALBTargetGroup_updateLoadBalancingAlgorithmType(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_loadBalancingAlgorithm(rName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "round_robin"),
				),
			},
			{
				Config: testAccATargetGroupConfig_loadBalancingAlgorithm(rName, true, "round_robin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckATargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "load_balancing_algorithm_type", "round_robin"),
				),
			},
			{
				Config: testAccATargetGroupConfig_loadBalancingAlgorithm(rName, true, "least_outstanding_requests"),
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

func TestAccAWSALBTargetGroup_updateSticknessEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckATargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccATargetGroupConfig_stickiness(rName, false, false),
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
				Config: testAccATargetGroupConfig_stickiness(rName, true, true),
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
				Config: testAccATargetGroupConfig_stickiness(rName, true, false),
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

func testAccALB_defaults(rName string) string {
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

func testAccTargetGroupBackwardsCompatibilityConfig(rName string) string {
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

func testAccTargetGroupProtocolGeneveConfig(rName string) string {
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

func testAccTargetGroupTags1Config(rName, tagKey1, tagValue1 string) string {
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

func testAccTargetGroupTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccTargetGroupConfig_GRPC_ProtocolVersion(rName string) string {
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

func testAccTargetGroupConfig_ProtocolVersion(rName string) string {
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

func testAccTargetGroupConfig_Protocol_TLS(rName string) string {
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

func testAccTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, path string, threshold int) string {
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

func testAccTargetGroupConfig_typeTCP_withProxyProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  proxy_protocol_v2    = "true"
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

func testAccTargetGroupConfig_withoutHealthcheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}
`, rName)
}

func testAccTargetGroupProtocolGeneveHealthConfig(rName string) string {
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

func testAccNLB_defaults(rName, healthCheckBlock string) string {
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

func testAccATargetGroupConfig_basic(rName string) string {
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

func testAccATargetGroupConfig_generatedName(rName string) string {
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

func testAccATargetGroupConfig_lambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}`, rName)
}

func testAccATargetGroupConfig_lambdaMultiValueHeadersEnabled(rName string, lambdaMultiValueHadersEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  lambda_multi_value_headers_enabled = %[1]t
  name                               = %[2]q
  target_type                        = "lambda"
}
`, lambdaMultiValueHadersEnabled, rName)
}

func testAccATargetGroupConfig_loadBalancingAlgorithm(rName string, nonDefault bool, algoType string) string {
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

func testAccATargetGroupConfig_missing_port(rName string) string {
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

func testAccATargetGroupConfig_missing_protocol(rName string) string {
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

func testAccATargetGroupConfig_missing_vpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
}
`, rName)
}

func testAccATargetGroupConfig_namePrefix(rName string) string {
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

func testAccATargetGroupConfig_stickiness(rName string, addStickinessBlock bool, enabled bool) string {
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

func testAccATargetGroupConfig_updateHealthCheck(rName string) string {
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

func testAccATargetGroupConfig_updateSlowStart(rName string, slowStartDuration int) string {
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

func testAccATargetGroupConfig_updateTags(rName string) string {
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

func testAccATargetGroupConfig_updatedPort(rName string) string {
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

func testAccATargetGroupConfig_updatedProtocol(rName string) string {
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

func testAccATargetGroupConfig_updatedVPC(rName string) string {
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
		if tfawserr.ErrMessageContains(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
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
