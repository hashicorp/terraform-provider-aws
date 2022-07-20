package elb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccELBProxyProtocolPolicy_basic(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyProtocolPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyProtocolPolicyConfig_basic(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_proxy_protocol_policy.smtp", "load_balancer", lbName),
					resource.TestCheckResourceAttr(
						"aws_proxy_protocol_policy.smtp", "instance_ports.#", "1"),
					resource.TestCheckTypeSetElemAttr("aws_proxy_protocol_policy.smtp", "instance_ports.*", "25"),
				),
			},
			{
				Config: testAccProxyProtocolPolicyConfig_update(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_proxy_protocol_policy.smtp", "load_balancer", lbName),
					resource.TestCheckResourceAttr("aws_proxy_protocol_policy.smtp", "instance_ports.#", "2"),
					resource.TestCheckTypeSetElemAttr("aws_proxy_protocol_policy.smtp", "instance_ports.*", "25"),
					resource.TestCheckTypeSetElemAttr("aws_proxy_protocol_policy.smtp", "instance_ports.*", "587"),
				),
			},
		},
	})
}

func testAccCheckProxyProtocolPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_placement_group" {
			continue
		}

		req := &elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{
				aws.String(rs.Primary.Attributes["load_balancer"])},
		}
		_, err := conn.DescribeLoadBalancers(req)
		if err != nil {
			// Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
				continue
			}
			return err
		}

		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccProxyProtocolPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 25
    instance_protocol = "tcp"
    lb_port           = 25
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 587
    instance_protocol = "tcp"
    lb_port           = 587
    lb_protocol       = "tcp"
  }
}

resource "aws_proxy_protocol_policy" "smtp" {
  load_balancer  = aws_elb.lb.name
  instance_ports = ["25"]
}
`, rName))
}

func testAccProxyProtocolPolicyConfig_update(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "lb" {
  name               = "%s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 25
    instance_protocol = "tcp"
    lb_port           = 25
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 587
    instance_protocol = "tcp"
    lb_port           = 587
    lb_protocol       = "tcp"
  }
}

resource "aws_proxy_protocol_policy" "smtp" {
  load_balancer  = aws_elb.lb.name
  instance_ports = ["25", "587"]
}
`, rName))
}
