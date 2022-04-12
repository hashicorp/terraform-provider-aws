package elbv2_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(elbv2.EndpointsID, testAccErrorCheckSkipELBv2)

}

func testAccErrorCheckSkipELBv2(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"ValidationError: Type must be one of: 'application, network'",
		"ValidationError: Action type 'authenticate-cognito' must be one of 'redirect,fixed-response,forward,authenticate-oidc'",
	)
}

func TestLBCloudwatchSuffixFromARN(t *testing.T) {
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
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/app/%s/.+", lbName))),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "defensive"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NLB_basic(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_networkLoadbalancer(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/net/%s/.+", lbName))),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_LoadBalancerType_gateway(t *testing.T) {
	var conf elbv2.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_LoadBalancerType_Gateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", elbv2.LoadBalancerTypeEnumGateway),
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

func TestAccELBV2LoadBalancer_ipv6SubnetMapping(t *testing.T) {
	var conf elbv2.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_IPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "subnet_mapping.*", map[string]*regexp.Regexp{
						"ipv6_address": regexp.MustCompile("[a-f0-6]+:[a-f0-6:]+"),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"idle_timeout",
				},
			},
		},
	})
}

func TestAccELBV2LoadBalancer_LoadBalancerTypeGateway_enableCrossZoneLoadBalancing(t *testing.T) {
	var conf elbv2.LoadBalancer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", elbv2.LoadBalancerTypeEnumGateway),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "true"),
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
				Config: testAccLoadBalancerConfig_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", elbv2.LoadBalancerTypeEnumGateway),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "false"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ALB_outpost(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-outpost-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_outpost(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/app/%s/.+", lbName))),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_outpost"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_mapping.0.outpost_id"),
					resource.TestCheckResourceAttrSet(resourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_networkLoadBalancerEIP(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_networkLoadBalancerEIP(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "2"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NLB_privateIPv4Address(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-pipv4a-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_networkLoadBalancerPrivateIPV4Address(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
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

func TestAccELBV2LoadBalancer_backwardsCompatibility(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_alb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerBackwardsCompatibilityConfig(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_generatedName(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_generatedName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_generatesNameForZeroValue(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_zeroValueName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_namePrefix(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_namePrefix(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestMatchResourceAttr(resourceName, "name",
						regexp.MustCompile("^tf-lb-")),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_tags(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_updatedTags(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Type", "Sample Type Tag"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_NetworkLoadBalancer_updateCrossZone(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-nlbcz-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_networkLoadbalancer(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					testAccCheckLoadBalancerAttribute(resourceName, "load_balancing.cross_zone.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "true"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_networkLoadbalancer(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &mid),
					testAccCheckLoadBalancerAttribute(resourceName, "load_balancing.cross_zone.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "false"),
					testAccChecklbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_networkLoadbalancer(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					testAccCheckLoadBalancerAttribute(resourceName, "load_balancing.cross_zone.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "true"),
					testAccChecklbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateHTTP2(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSalb-http2-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableHTTP2(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					testAccCheckLoadBalancerAttribute(resourceName, "routing.http2.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", "false"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableHTTP2(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &mid),
					testAccCheckLoadBalancerAttribute(resourceName, "routing.http2.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", "true"),
					testAccChecklbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableHTTP2(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					testAccCheckLoadBalancerAttribute(resourceName, "routing.http2.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", "false"),
					testAccChecklbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateDropInvalidHeaderFields(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSalb-headers-%s", sdkacctest.RandString(10))

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableDropInvalidHeaderFields(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("aws_lb.lb_test", &pre),
					testAccCheckLoadBalancerAttribute("aws_lb.lb_test", "routing.http.drop_invalid_header_fields.enabled", "false"),
					resource.TestCheckResourceAttr("aws_lb.lb_test", "drop_invalid_header_fields", "false"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDropInvalidHeaderFields(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("aws_lb.lb_test", &mid),
					testAccCheckLoadBalancerAttribute("aws_lb.lb_test", "routing.http.drop_invalid_header_fields.enabled", "true"),
					resource.TestCheckResourceAttr("aws_lb.lb_test", "drop_invalid_header_fields", "true"),
					testAccChecklbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDropInvalidHeaderFields(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("aws_lb.lb_test", &post),
					testAccCheckLoadBalancerAttribute("aws_lb.lb_test", "routing.http.drop_invalid_header_fields.enabled", "false"),
					resource.TestCheckResourceAttr("aws_lb.lb_test", "drop_invalid_header_fields", "false"),
					testAccChecklbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateDeletionProtection(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSalb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_enableDeletionProtection(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					testAccCheckLoadBalancerAttribute(resourceName, "deletion_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDeletionProtection(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &mid),
					testAccCheckLoadBalancerAttribute(resourceName, "deletion_protection.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "true"),
					testAccChecklbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccLoadBalancerConfig_enableDeletionProtection(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					testAccCheckLoadBalancerAttribute(resourceName, "deletion_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					testAccChecklbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ApplicationLoadBalancer_updateWafFailOpen(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSalb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBConfig_enableWafFailOpen(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "enable_waf_fail_open", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLBConfig_enableWafFailOpen(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &mid),
					resource.TestCheckResourceAttr(resourceName, "enable_waf_fail_open", "true"),
					testAccChecklbARNs(&pre, &mid),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLBConfig_enableWafFailOpen(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "enable_waf_fail_open", "false"),
					testAccChecklbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_updatedSecurityGroups(t *testing.T) {
	var pre, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_updateSecurityGroups(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					testAccChecklbARNs(&pre, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_updatedSubnets(t *testing.T) {
	var pre, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_updateSubnets(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "3"),
					testAccChecklbARNs(&pre, &post),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_updatedIPAddressType(t *testing.T) {
	var pre, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerWithIPAddressTypeConfig(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				),
			},
			{
				Config: testAccLoadBalancerWithIPAddressTypeUpdatedConfig(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
				),
			},
		},
	})
}

// TestAccAWSALB_noSecurityGroup regression tests the issue in #8264,
// where if an ALB is created without a security group, a default one
// is assigned.
func TestAccELBV2LoadBalancer_noSecurityGroup(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-nosg-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nosg(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_ALB_accessLogs(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", sdkacctest.RandString(6))
	lbName := fmt.Sprintf("testAccAWSlbaccesslog-%s", sdkacctest.RandString(4))
	resourceName := "aws_lb.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerALBAccessLogsConfig(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerALBAccessLogsConfig(false, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerALBAccessLogsConfig(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerALBAccessLogsNoBlocksConfig(lbName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
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

func TestAccELBV2LoadBalancer_ALBAccessLogs_prefix(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", sdkacctest.RandString(6))
	lbName := fmt.Sprintf("testAccAWSlbaccesslog-%s", sdkacctest.RandString(4))
	resourceName := "aws_lb.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerALBAccessLogsConfig(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerALBAccessLogsConfig(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerALBAccessLogsConfig(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
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

func TestAccELBV2LoadBalancer_NLB_accessLogs(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", sdkacctest.RandString(6))
	lbName := fmt.Sprintf("testAccAWSlbaccesslog-%s", sdkacctest.RandString(4))
	resourceName := "aws_lb.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerNLBAccessLogsConfig(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerNLBAccessLogsConfig(false, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerNLBAccessLogsConfig(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerNLBAccessLogsNoBlocksConfig(lbName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
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

func TestAccELBV2LoadBalancer_NLBAccessLogs_prefix(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", sdkacctest.RandString(6))
	lbName := fmt.Sprintf("testAccAWSlbaccesslog-%s", sdkacctest.RandString(4))
	resourceName := "aws_lb.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerNLBAccessLogsConfig(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerNLBAccessLogsConfig(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerNLBAccessLogsConfig(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckLoadBalancerAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
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

func TestAccELBV2LoadBalancer_NetworkLoadBalancerSubnet_change(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testAccAWSlb-basic-%s", sdkacctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_networkLoadbalancer_subnets(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "testAccLoadBalancerConfig_networkLoadbalancer_subnets"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancer_updateDesyncMitigationMode(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawsalb-desync-%s", sdkacctest.RandString(4))
	resourceName := "aws_lb.lb_test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elbv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_desyncMitigationMode(lbName, "strictest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &pre),
					testAccCheckLoadBalancerAttribute(resourceName, "routing.http.desync_mitigation_mode", "strictest"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "strictest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfig_desyncMitigationMode(lbName, "monitor"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &mid),
					testAccCheckLoadBalancerAttribute(resourceName, "routing.http.desync_mitigation_mode", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "monitor"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfig_desyncMitigationMode(lbName, "defensive"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists(resourceName, &post),
					testAccCheckLoadBalancerAttribute(resourceName, "routing.http.desync_mitigation_mode", "defensive"),
					resource.TestCheckResourceAttr(resourceName, "desync_mitigation_mode", "defensive"),
				),
			},
		},
	})
}

func testAccChecklbARNs(pre, post *elbv2.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(pre.LoadBalancerArn) != aws.StringValue(post.LoadBalancerArn) {
			return errors.New("LB has been recreated. ARNs are different")
		}

		return nil
	}
}

func testAccCheckLoadBalancerExists(n string, res *elbv2.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LB ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		lb, err := tfelbv2.FindLoadBalancerByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading LB (%s): %w", rs.Primary.ID, err)
		}

		if lb != nil {
			*res = *lb
			return nil
		}

		return fmt.Errorf("LB (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckLoadBalancerAttribute(n, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LB ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn
		attributesResp, err := conn.DescribeLoadBalancerAttributes(&elbv2.DescribeLoadBalancerAttributesInput{
			LoadBalancerArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Error retrieving LB Attributes: %s", err)
		}

		for _, attr := range attributesResp.Attributes {
			if aws.StringValue(attr.Key) == key {
				if aws.StringValue(attr.Value) == value {
					return nil
				}
				return fmt.Errorf("LB attribute %s expected: %q actual: %q", key, value, aws.StringValue(attr.Value))
			}
		}
		return fmt.Errorf("LB attribute %s does not exist on LB: %s", key, rs.Primary.ID)
	}
}

func testAccCheckLoadBalancerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb" && rs.Type != "aws_alb" {
			continue
		}

		lb, err := tfelbv2.FindLoadBalancerByARN(conn, rs.Primary.ID)

		if tfawserr.ErrCodeContains(err, elb.ErrCodeAccessPointNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("Unexpected error checking LB (%s) destroyed: %w", rs.Primary.ID, err)
		}

		if lb != nil && aws.StringValue(lb.LoadBalancerArn) == rs.Primary.ID {
			return fmt.Errorf("LB %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPreCheckElbv2GatewayLoadBalancer(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

	input := &elbv2.DescribeAccountLimitsInput{}

	output, err := conn.DescribeAccountLimits(input)

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
		if limit == nil {
			continue
		}

		if aws.StringValue(limit.Name) == "gateway-load-balancers" {
			return
		}
	}

	t.Skip("skipping acceptance testing: region does not support ELBv2 Gateway Load Balancers")
}

func testAccLoadBalancerWithIPAddressTypeUpdatedConfig(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test_1.id, aws_subnet.alb_test_2.id]

  ip_address_type = "dualstack"

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.lb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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

resource "aws_egress_only_internet_gateway" "igw" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_vpc" "alb_test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-lb-with-ip-address-type-updated"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_subnet" "alb_test_1" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-updated-1"
  }
}

resource "aws_subnet" "alb_test_2" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 2)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-updated-2"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, lbName))
}

func testAccLoadBalancerWithIPAddressTypeConfig(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test_1.id, aws_subnet.alb_test_2.id]

  ip_address_type = "ipv4"

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.lb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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

resource "aws_egress_only_internet_gateway" "igw" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_vpc" "alb_test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-lb-with-ip-address-type"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_subnet" "alb_test_1" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-1"
  }
}

resource "aws_subnet" "alb_test_2" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 2)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-2"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, lbName))
}

func testAccLoadBalancerConfig_basic(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccLoadBalancerConfig_outpost(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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
  vpc_id                       = aws_vpc.alb_test.id
}

resource "aws_route_table" "test" {
  vpc_id     = aws_vpc.alb_test.id
  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.test]
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.alb_test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_lb" "lb_test" {
  name                       = "%s"
  security_groups            = [aws_security_group.alb_test.id]
  customer_owned_ipv4_pool   = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  idle_timeout               = 30
  enable_deletion_protection = false
  subnets                    = [aws_subnet.alb_test.id]

  tags = {
    Name = "TestAccAWSALB_outpost"
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-outpost"
  }
}

resource "aws_subnet" "alb_test" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = "tf-acc-lb-outpost"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_outpost"
  }
}
`, lbName))
}

func testAccLoadBalancerConfig_enableHTTP2(lbName string, http2 bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_http2 = %t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, http2))
}

func testAccLoadBalancerConfig_enableDropInvalidHeaderFields(lbName string, dropInvalid bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  drop_invalid_header_fields = %t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, dropInvalid))
}

func testAccLoadBalancerConfig_enableDeletionProtection(lbName string, deletion_protection bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = %t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, deletion_protection))
}

func testAccLBConfig_enableWafFailOpen(lbName string, wafFailOpen bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_waf_fail_open = %[2]t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, wafFailOpen))
}

func testAccLoadBalancerConfig_networkLoadbalancer_subnets(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-network-load-balancer-subnets"
  }
}

resource "aws_lb" "lb_test" {
  name = "%s"

  subnets = [
    aws_subnet.alb_test_1.id,
    aws_subnet.alb_test_2.id,
    aws_subnet.alb_test_3.id,
  ]

  load_balancer_type               = "network"
  internal                         = true
  idle_timeout                     = 60
  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = false

  tags = {
    Name = "testAccLoadBalancerConfig_networkLoadbalancer_subnets"
  }
}

resource "aws_subnet" "alb_test_1" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-lb-network-load-balancer-subnets-1"
  }
}

resource "aws_subnet" "alb_test_2" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-lb-network-load-balancer-subnets-2"
  }
}

resource "aws_subnet" "alb_test_3" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "tf-acc-lb-network-load-balancer-subnets-3"
  }
}
`, lbName))
}

func testAccLoadBalancerConfig_networkLoadbalancer(lbName string, cz bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name               = "%s"
  internal           = true
  load_balancer_type = "network"

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = %t

  subnet_mapping {
    subnet_id = aws_subnet.alb_test.id
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "terraform-testacc-network-load-balancer"
  }
}

resource "aws_subnet" "alb_test" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-load-balancer"
  }
}
`, lbName, cz))
}

func testAccLoadBalancerConfig_LoadBalancerType_Gateway(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName))
}

func testAccLoadBalancerConfig_IPv6(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 16)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  name                       = %[1]q
  load_balancer_type         = "network"
  enable_deletion_protection = false

  subnet_mapping {
    subnet_id    = aws_subnet.test.id
    ipv6_address = cidrhost(cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 16), 5)
  }

  tags = {
    Name = "TestAccAWSALB_ipv6address"
  }

  depends_on = [aws_internet_gateway.gw]
}
`, rName))
}

func testAccLoadBalancerConfig_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(rName string, enableCrossZoneLoadBalancing bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  enable_cross_zone_load_balancing = %[2]t
  load_balancer_type               = "gateway"
  name                             = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName, enableCrossZoneLoadBalancing))
}

func testAccLoadBalancerConfig_networkLoadBalancerEIP(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-network-load-balancer-eip"
  }
}

resource "aws_subnet" "public" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.10.${count.index}.0/24"
  vpc_id            = aws_vpc.main.id

  tags = {
    Name = "tf-acc-lb-network-load-balancer-eip-${count.index}"
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.default.id
  }
}

resource "aws_route_table_association" "a" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_lb" "lb_test" {
  name               = "%s"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = aws_subnet.public[0].id
    allocation_id = aws_eip.lb[0].id
  }

  subnet_mapping {
    subnet_id     = aws_subnet.public[1].id
    allocation_id = aws_eip.lb[1].id
  }

  depends_on = [aws_internet_gateway.default]
}

resource "aws_eip" "lb" {
  count = "2"
}
`, lbName))
}

func testAccLoadBalancerConfig_networkLoadBalancerPrivateIPV4Address(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = "%s"
  internal                   = true
  load_balancer_type         = "network"
  enable_deletion_protection = false

  subnet_mapping {
    subnet_id            = aws_subnet.test.id
    private_ipv4_address = "10.10.0.15"
  }

  tags = {
    Name = "TestAccAWSALB_privateipv4address"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "TestAccAWSALB_privateipv4address"
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "TestAccAWSALB_privateipv4address"
  }
}
`, lbName))
}

func testAccLoadBalancerBackwardsCompatibilityConfig(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_alb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-bc-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccLoadBalancerConfig_updateSubnets(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-update-subnets"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 3
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-update-subnets-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccLoadBalancerConfig_generatedName() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_lb" "lb_test" {
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-generated-name"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.alb_test.id

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-generated-name-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`)
}

func testAccLoadBalancerConfig_zeroValueName() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_lb" "lb_test" {
  name            = ""
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

# See https://github.com/hashicorp/terraform-provider-aws/issues/2498
output "lb_name" {
  value = aws_lb.lb_test.name
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-zero-value-name"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.alb_test.id

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-zero-value-name-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`)
}

func testAccLoadBalancerConfig_namePrefix() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_lb" "lb_test" {
  name_prefix     = "tf-lb-"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-name-prefix"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-name-prefix-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`)
}

func testAccLoadBalancerConfig_updatedTags(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Environment = "Production"
    Type        = "Sample Type Tag"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-updated-tags"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-updated-tags-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccLoadBalancerALBAccessLogsBaseConfig(bucketName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_elb_service_account" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-access-logs"
  }
}

resource "aws_subnet" "alb_test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-lb-access-logs-${count.index}"
  }
}

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
`, bucketName))
}

func testAccLoadBalancerALBAccessLogsConfig(enabled bool, lbName, bucketName, bucketPrefix string) string {
	return acctest.ConfigCompose(testAccLoadBalancerALBAccessLogsBaseConfig(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.alb_test.*.id

  access_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, lbName, enabled, bucketPrefix))
}

func testAccLoadBalancerALBAccessLogsNoBlocksConfig(lbName, bucketName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerALBAccessLogsBaseConfig(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.alb_test.*.id
}
`, lbName))
}

func testAccLoadBalancerNLBAccessLogsBaseConfig(bucketName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_elb_service_account" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-access-logs"
  }
}

resource "aws_subnet" "alb_test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-lb-access-logs-${count.index}"
  }
}

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
`, bucketName))
}

func testAccLoadBalancerNLBAccessLogsConfig(enabled bool, lbName, bucketName, bucketPrefix string) string {
	return acctest.ConfigCompose(testAccLoadBalancerNLBAccessLogsBaseConfig(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.alb_test.*.id

  access_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, lbName, enabled, bucketPrefix))
}

func testAccLoadBalancerNLBAccessLogsNoBlocksConfig(lbName, bucketName string) string {
	return acctest.ConfigCompose(testAccLoadBalancerNLBAccessLogsBaseConfig(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.alb_test.*.id
}
`, lbName))
}

func testAccLoadBalancerConfig_nosg(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name     = "%s"
  internal = true
  subnets  = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-no-sg"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-no-sg-${count.index}"
  }
}
`, lbName))
}

func testAccLoadBalancerConfig_updateSecurityGroups(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id, aws_security_group.alb_test_2.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-update-security-groups"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-update-security-groups-${count.index}"
  }
}

resource "aws_security_group" "alb_test_2" {
  name        = "allow_all_alb_test_2"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic_2"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccAWSLBConfig_desyncMitigationMode(lbName string, mode string) string {
	return fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  desync_mitigation_mode = %q

  tags = {
    Name = "TestAccAWSALB_desync"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-desync"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-desync-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test_desync"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    Name = "TestAccAWSALB_desync"
  }
}
`, lbName, mode)
}
