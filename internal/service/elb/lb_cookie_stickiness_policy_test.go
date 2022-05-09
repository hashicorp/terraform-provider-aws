package elb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
)

func TestAccELBCookieStickinessPolicy_basic(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", sdkacctest.RandString(5))
	resourceName := "aws_lb_cookie_stickiness_policy.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cookie_expiration_period", "300"),
					testAccCheckLBCookieStickinessPolicy(
						"aws_elb.lb",
						"aws_lb_cookie_stickiness_policy.foo",
					),
				),
			},
			{
				Config: testAccLBCookieStickinessPolicyConfigUpdate(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cookie_expiration_period", "0"),
					testAccCheckLBCookieStickinessPolicy(
						"aws_elb.lb",
						"aws_lb_cookie_stickiness_policy.foo",
					),
				),
			},
		},
	})
}

func testAccCheckLBCookieStickinessPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_cookie_stickiness_policy" {
			continue
		}

		lbName, _, policyName := tfelb.CookieStickinessPolicyParseID(rs.Primary.ID)
		out, err := conn.DescribeLoadBalancerPolicies(
			&elb.DescribeLoadBalancerPoliciesInput{
				LoadBalancerName: aws.String(lbName),
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
	}

	return nil
}

func testAccCheckLBCookieStickinessPolicy(elbResource string, policyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[elbResource]
		if !ok {
			return fmt.Errorf("Not found: %s", elbResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[policyResource]
		if !ok {
			return fmt.Errorf("Not found: %s", policyResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn
		elbName, _, policyName := tfelb.CookieStickinessPolicyParseID(policy.Primary.ID)
		_, err := conn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(elbName),
			PolicyNames:      []*string{aws.String(policyName)},
		})

		return err
	}
}

func TestAccELBCookieStickinessPolicy_disappears(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", sdkacctest.RandString(5))
	elbResourceName := "aws_elb.lb"
	resourceName := "aws_lb_cookie_stickiness_policy.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(elbResourceName, resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfelb.ResourceCookieStickinessPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBCookieStickinessPolicy_Disappears_elb(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", sdkacctest.RandString(5))
	elbResourceName := "aws_elb.lb"
	resourceName := "aws_lb_cookie_stickiness_policy.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(elbResourceName, resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfelb.ResourceLoadBalancer(), elbResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccLBCookieStickinessPolicyConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_lb_cookie_stickiness_policy" "foo" {
  name                     = "foo-policy"
  load_balancer            = aws_elb.lb.id
  lb_port                  = 80
  cookie_expiration_period = 300
}
`, rName))
}

// Sets the cookie_expiration_period to 0s.
func testAccLBCookieStickinessPolicyConfigUpdate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_lb_cookie_stickiness_policy" "foo" {
  name                     = "foo-policy"
  load_balancer            = aws_elb.lb.id
  lb_port                  = 80
  cookie_expiration_period = 0
}
`, rName))
}
