// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointConnectionNotification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_endpoint_connection_notification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointConnectionNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConnectionNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointConnectionNotificationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "connection_events.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "Topic"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "Enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConnectionNotificationConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointConnectionNotificationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "connection_events.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "Topic"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "Enabled"),
				),
			},
		},
	})
}

func testAccCheckVPCEndpointConnectionNotificationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_endpoint_connection_notification" {
				continue
			}

			_, err := tfec2.FindVPCEndpointConnectionNotificationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC Endpoint Connection Notification %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCEndpointConnectionNotificationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Endpoint Connection Notification ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindVPCEndpointConnectionNotificationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccVPCEndpointConnectionNotificationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_lb" "nlb_test" {
  name = %[1]q

  subnets = aws_subnet.test[*].id

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false
}

data "aws_caller_identity" "current" {}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    aws_lb.nlb_test.id,
  ]

  allowed_principals = [
    data.aws_caller_identity.current.arn
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "vpce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "SNS:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:*:*:%[1]s"
    }
  ]
}
POLICY
}
`, rName))
}

func testAccVPCEndpointConnectionNotificationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConnectionNotificationConfig_base(rName), `
resource "aws_vpc_endpoint_connection_notification" "test" {
  vpc_endpoint_service_id     = aws_vpc_endpoint_service.test.id
  connection_notification_arn = aws_sns_topic.test.arn
  connection_events           = ["Accept", "Reject"]
}
`)
}

func testAccVPCEndpointConnectionNotificationConfig_modified(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConnectionNotificationConfig_base(rName), `
resource "aws_vpc_endpoint_connection_notification" "test" {
  vpc_endpoint_service_id     = aws_vpc_endpoint_service.test.id
  connection_notification_arn = aws_sns_topic.test.arn
  connection_events           = ["Accept"]
}
`)
}
