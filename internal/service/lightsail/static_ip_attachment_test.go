// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailStaticIPAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	instanceName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	keypairName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStaticIPAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPAttachmentExists(ctx, "aws_lightsail_static_ip_attachment.test"),
					resource.TestCheckResourceAttrSet("aws_lightsail_static_ip_attachment.test", names.AttrIPAddress),
				),
			},
		},
	})
}

func TestAccLightsailStaticIPAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	instanceName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	keypairName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	staticIpDestroy := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)
		_, err := conn.DetachStaticIp(ctx, &lightsail.DetachStaticIpInput{
			StaticIpName: aws.String(staticIpName),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Static IP in disappear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStaticIPAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPAttachmentExists(ctx, "aws_lightsail_static_ip_attachment.test"),
					staticIpDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStaticIPAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Static IP Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		resp, err := conn.GetStaticIp(ctx, &lightsail.GetStaticIpInput{
			StaticIpName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if resp == nil || resp.StaticIp == nil {
			return fmt.Errorf("Static IP (%s) not found", rs.Primary.ID)
		}

		if !*resp.StaticIp.IsAttached {
			return fmt.Errorf("Static IP (%s) not attached", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStaticIPAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_static_ip_attachment" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

			resp, err := conn.GetStaticIp(ctx, &lightsail.GetStaticIpInput{
				StaticIpName: aws.String(rs.Primary.ID),
			})

			if tflightsail.IsANotFoundError(err) {
				continue
			}

			if err == nil {
				if *resp.StaticIp.IsAttached {
					return fmt.Errorf("Lightsail Static IP %q is still attached (to %q)", rs.Primary.ID, *resp.StaticIp.AttachedTo)
				}
			}

			return err
		}

		return nil
	}
}

func testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_static_ip_attachment" "test" {
  static_ip_name = aws_lightsail_static_ip.test.name
  instance_name  = aws_lightsail_instance.test.name
}

resource "aws_lightsail_static_ip" "test" {
  name = "%s"
}

resource "aws_lightsail_instance" "test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "micro_1_0"
  key_pair_name     = aws_lightsail_key_pair.test.name
}

resource "aws_lightsail_key_pair" "test" {
  name = "%s"
}
`, staticIpName, instanceName, keypairName)
}
