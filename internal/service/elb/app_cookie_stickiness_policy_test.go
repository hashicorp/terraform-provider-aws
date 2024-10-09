// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBAppCookieStickinessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_app_cookie_stickiness_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppCookieStickinessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppCookieStickinessPolicyConfig_basic(rName, "bourbon"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppCookieStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_name", "bourbon"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppCookieStickinessPolicyConfig_basic(rName, "custard-cream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppCookieStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_name", "custard-cream"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccELBAppCookieStickinessPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_app_cookie_stickiness_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppCookieStickinessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppCookieStickinessPolicyConfig_basic(rName, "bourbon"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppCookieStickinessPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourceAppCookieStickinessPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBAppCookieStickinessPolicy_Disappears_elb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_app_cookie_stickiness_policy.test"
	elbResourceName := "aws_elb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppCookieStickinessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppCookieStickinessPolicyConfig_basic(rName, "bourbon"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppCookieStickinessPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourceLoadBalancer(), elbResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppCookieStickinessPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_app_cookie_stickiness_policy" {
				continue
			}

			lbName, lbPort, policyName, err := tfelb.AppCookieStickinessPolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic App Cookie Stickiness Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppCookieStickinessPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		lbName, lbPort, policyName, err := tfelb.AppCookieStickinessPolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		_, err = tfelb.FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

		return err
	}
}

func testAccAppCookieStickinessPolicyConfig_basic(rName, cookieName string) string {
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

resource "aws_app_cookie_stickiness_policy" "test" {
  name          = %[1]q
  load_balancer = aws_elb.test.id
  lb_port       = 80
  cookie_name   = %[2]q
}
`, rName, cookieName))
}
