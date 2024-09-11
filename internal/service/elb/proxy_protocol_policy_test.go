// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBProxyProtocolPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	lbName := fmt.Sprintf("tf-test-lb-%s", sdkacctest.RandString(5))
	resourceName := "aws_proxy_protocol_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyProtocolPolicyConfig_basic(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "load_balancer", lbName),
					resource.TestCheckResourceAttr(resourceName, "instance_ports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_ports.*", "25"),
				),
			},
			{
				Config: testAccProxyProtocolPolicyConfig_update(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "load_balancer", lbName),
					resource.TestCheckResourceAttr(resourceName, "instance_ports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_ports.*", "25"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_ports.*", "587"),
				),
			},
		},
	})
}

func testAccProxyProtocolPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = %[1]q
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

resource "aws_proxy_protocol_policy" "test" {
  load_balancer  = aws_elb.test.name
  instance_ports = ["25"]
}
`, rName))
}

func testAccProxyProtocolPolicyConfig_update(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = %[1]q
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

resource "aws_proxy_protocol_policy" "test" {
  load_balancer  = aws_elb.test.name
  instance_ports = ["25", "587"]
}
`, rName))
}
