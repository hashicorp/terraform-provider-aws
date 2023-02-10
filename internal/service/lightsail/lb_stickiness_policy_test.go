package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
)

func TestAccLightsailLoadBalancerStickinessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "150"
	enabled := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, enabled, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration),
					resource.TestCheckResourceAttr(resourceName, "enabled", enabled),
					resource.TestCheckResourceAttr(resourceName, "lb_name", rName),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerStickinessPolicy_CookieDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration1 := "200"
	cookieDuration2 := "500"
	enabled := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, enabled, cookieDuration1),
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
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, enabled, cookieDuration2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration2),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerStickinessPolicy_Enabled(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "200"
	enabledTrue := "true"
	enabledFalse := "false"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, enabledTrue, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", enabledTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, enabledFalse, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", enabledFalse),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerStickinessPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "200"
	enabled := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, enabled, cookieDuration),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn()

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
