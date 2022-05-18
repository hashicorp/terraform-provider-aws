package elb_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
)

func TestAccELBListenerPolicy_basic(t *testing.T) {
	rChar := sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlpha)
	lbName := rChar
	mcName := rChar
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckListenerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerPolicyConfig_basic0(lbName, mcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.magic-cookie-sticky"),
					testAccCheckListenerPolicyState(lbName, int64(80), mcName, true),
				),
			},
			{
				Config: testAccListenerPolicyConfig_basic1(lbName, mcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.magic-cookie-sticky"),
					testAccCheckListenerPolicyState(lbName, int64(80), mcName, true),
				),
			},
			{
				Config: testAccListenerPolicyConfig_basic2(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerPolicyState(lbName, int64(80), mcName, false),
				),
			},
		},
	})
}

func policyInListenerPolicies(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func testAccCheckListenerPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

	for _, rs := range s.RootModule().Resources {
		switch {
		case rs.Type == "aws_load_balancer_policy":
			loadBalancerName, policyName := tfelb.ListenerPoliciesParseID(rs.Primary.ID)
			out, err := conn.DescribeLoadBalancerPolicies(
				&elb.DescribeLoadBalancerPoliciesInput{
					LoadBalancerName: aws.String(loadBalancerName),
					PolicyNames:      []*string{aws.String(policyName)},
				})
			if err != nil {
				if ec2err, ok := err.(awserr.Error); ok && (ec2err.Code() == "PolicyNotFound" || ec2err.Code() == "LoadBalancerNotFound") {
					continue
				}
				return err
			}
			if len(out.PolicyDescriptions) > 0 {
				return fmt.Errorf("Policy still exists")
			}
		case rs.Type == "aws_load_listener_policy":
			loadBalancerName, _ := tfelb.ListenerPoliciesParseID(rs.Primary.ID)
			out, err := conn.DescribeLoadBalancers(
				&elb.DescribeLoadBalancersInput{
					LoadBalancerNames: []*string{aws.String(loadBalancerName)},
				})

			if tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			policyNames := []string{}
			for k := range rs.Primary.Attributes {
				if strings.HasPrefix(k, "policy_names.") && strings.HasSuffix(k, ".name") {
					value_key := fmt.Sprintf("%s.value", strings.TrimSuffix(k, ".name"))
					policyNames = append(policyNames, rs.Primary.Attributes[value_key])
				}
			}
			for _, policyName := range policyNames {
				for _, listener := range out.LoadBalancerDescriptions[0].ListenerDescriptions {
					policyStrings := []string{}
					for _, pol := range listener.PolicyNames {
						policyStrings = append(policyStrings, *pol)
					}
					if policyInListenerPolicies(policyName, policyStrings) {
						return fmt.Errorf("Policy still exists and is assigned")
					}
				}
			}
		default:
			continue
		}
	}
	return nil
}

func testAccCheckListenerPolicyState(loadBalancerName string, loadBalancerListenerPort int64, loadBalancerListenerPolicyName string, assigned bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

		loadBalancerDescription, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{aws.String(loadBalancerName)},
		})
		if err != nil {
			return err
		}

		for _, listener := range loadBalancerDescription.LoadBalancerDescriptions[0].ListenerDescriptions {
			if *listener.Listener.LoadBalancerPort != loadBalancerListenerPort {
				continue
			}
			policyStrings := []string{}
			for _, pol := range listener.PolicyNames {
				policyStrings = append(policyStrings, *pol)
			}
			if policyInListenerPolicies(loadBalancerListenerPolicyName, policyStrings) != assigned {
				if assigned {
					return fmt.Errorf("Policy no longer assigned %s not in %+v", loadBalancerListenerPolicyName, policyStrings)
				} else {
					return fmt.Errorf("Policy exists and is assigned")
				}
			}
		}

		return nil
	}
}

func testAccListenerPolicyConfig_basic0(lbName, mcName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test-lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "magic-cookie-sticky" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "%s"
  policy_type_name   = "AppCookieStickinessPolicyType"

  policy_attribute {
    name  = "CookieName"
    value = "magic_cookie"
  }
}

resource "aws_load_balancer_listener_policy" "test-lb-listener-policies-80" {
  load_balancer_name = aws_elb.test-lb.name
  load_balancer_port = 80

  policy_names = [
    aws_load_balancer_policy.magic-cookie-sticky.policy_name,
  ]
}
`, lbName, mcName))
}

func testAccListenerPolicyConfig_basic1(lbName, mcName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test-lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "magic-cookie-sticky" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "%s"
  policy_type_name   = "AppCookieStickinessPolicyType"

  policy_attribute {
    name  = "CookieName"
    value = "unicorn_cookie"
  }
}

resource "aws_load_balancer_listener_policy" "test-lb-listener-policies-80" {
  load_balancer_name = aws_elb.test-lb.name
  load_balancer_port = 80

  policy_names = [
    aws_load_balancer_policy.magic-cookie-sticky.policy_name,
  ]
}
`, lbName, mcName))
}

func testAccListenerPolicyConfig_basic2(lbName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test-lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`, lbName))
}
