// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccLoadBalancerStickinessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "150"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, acctest.CtTrue, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "lb_name", rName),
				),
			},
		},
	})
}

func testAccLoadBalancerStickinessPolicy_cookieDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration1 := "200"
	cookieDuration2 := "500"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, acctest.CtTrue, cookieDuration1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, acctest.CtTrue, cookieDuration2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration2),
				),
			},
		},
	})
}

func testAccLoadBalancerStickinessPolicy_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "200"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, acctest.CtTrue, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, acctest.CtFalse, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccLoadBalancerStickinessPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "200"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, acctest.CtTrue, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceLoadBalancerStickinessPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoadBalancerStickinessPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailLoadBalancerStickinessPolicy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		out, err := tflightsail.FindLoadBalancerStickinessPolicyById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Load Balancer Stickiness Policy %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLoadBalancerStickinessPolicyConfig_basic(rName string, enabled string, cookieDuration string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
resource "aws_lightsail_lb_stickiness_policy" "test" {
  enabled         = %[2]s
  cookie_duration = %[3]s
  lb_name         = aws_lightsail_lb.test.name
}
`, rName, enabled, cookieDuration)
}
