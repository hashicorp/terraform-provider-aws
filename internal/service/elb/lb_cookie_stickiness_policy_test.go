// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBCookieStickinessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_cookie_stickiness_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLBCookieStickinessPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig_basic(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_expiration_period", "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccLBCookieStickinessPolicyConfig_basic(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_expiration_period", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccELBCookieStickinessPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_cookie_stickiness_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLBCookieStickinessPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig_basic(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelb.ResourceCookieStickinessPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBCookieStickinessPolicy_Disappears_elb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_cookie_stickiness_policy.test"
	elbResourceName := "aws_elb.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLBCookieStickinessPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLBCookieStickinessPolicyConfig_basic(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelb.ResourceLoadBalancer(), elbResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLBCookieStickinessPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_cookie_stickiness_policy" {
				continue
			}

			lbName, lbPort, policyName, err := tfelb.LBCookieStickinessPolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic LB Cookie Stickiness Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLBCookieStickinessPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		lbName, lbPort, policyName, err := tfelb.LBCookieStickinessPolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).ELBClient(ctx)

		_, err = tfelb.FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

		return err
	}
}

func testAccLBCookieStickinessPolicyConfig_basic(rName string, expirationPeriod int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_lb_cookie_stickiness_policy" "test" {
  name                     = %[1]q
  load_balancer            = aws_elb.test.id
  lb_port                  = 80
  cookie_expiration_period = %[2]d
}
`, rName, expirationPeriod))
}
