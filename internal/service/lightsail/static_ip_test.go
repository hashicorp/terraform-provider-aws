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
)

func TestAccLightsailStaticIP_basic(t *testing.T) {
	ctx := acctest.Context(t)
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStaticIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPConfig_basic(staticIpName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPExists(ctx, "aws_lightsail_static_ip.test"),
				),
			},
		},
	})
}

func TestAccLightsailStaticIP_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	staticIpDestroy := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)
		_, err := conn.ReleaseStaticIp(ctx, &lightsail.ReleaseStaticIpInput{
			StaticIpName: aws.String(staticIpName),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Static IP in disapear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStaticIPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPConfig_basic(staticIpName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPExists(ctx, "aws_lightsail_static_ip.test"),
					staticIpDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStaticIPExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Static IP ID is set")
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

		return nil
	}
}

func testAccCheckStaticIPDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_static_ip" {
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
				if resp.StaticIp != nil {
					return fmt.Errorf("Lightsail Static IP %q still exists", rs.Primary.ID)
				}
			}

			return err
		}

		return nil
	}
}

func testAccStaticIPConfig_basic(staticIpName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_static_ip" "test" {
  name = "%s"
}
`, staticIpName)
}
