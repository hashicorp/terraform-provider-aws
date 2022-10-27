package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	var enabled bool
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "150"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(resourceName, enabled),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "lb_name", rName),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerStickinessPolicy_CookieDuration(t *testing.T) {
	var enabled bool
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration1 := "200"
	cookieDuration2 := "500"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, cookieDuration1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(resourceName, enabled),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, cookieDuration2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(resourceName, enabled),
					resource.TestCheckResourceAttr(resourceName, "cookie_duration", cookieDuration2),
				),
			},
		},
	})
}

func TestAccLightsailLoadBalancerStickinessPolicy_disappears(t *testing.T) {
	var enabled bool
	resourceName := "aws_lightsail_lb_stickiness_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cookieDuration := "200"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerStickinessPolicyConfig_basic(rName, cookieDuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerStickinessPolicyExists(resourceName, enabled),
					acctest.CheckResourceDisappears(acctest.Provider, tflightsail.ResourceLoadBalancerStickinessPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoadBalancerStickinessPolicyExists(n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailLoadBalancerStickinessPolicy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		out, err := tflightsail.FindLoadBalancerStickinessPolicyById(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("Load Balancer Stickiness Policy %q does not exist", rs.Primary.ID)
		}

		boolValue, err := strconv.ParseBool(*out["SessionStickinessEnabled"])
		if err != nil {
			return fmt.Errorf("Load Balancer Stickiness Policy %q does not exist. Error parsing enabled bool", rs.Primary.ID)
		}

		enabled = boolValue

		return nil
	}
}

func testAccLoadBalancerStickinessPolicyConfig_basic(rName string, cookieDuration string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
resource "aws_lightsail_lb_stickiness_policy" "test" {
  lb_name = aws_lightsail_lb.test.name
  cookie_duration = %[2]s
}
`, rName, cookieDuration)
}
