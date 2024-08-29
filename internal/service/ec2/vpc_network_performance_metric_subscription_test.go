// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkPerformanceMetricSubscription_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccNetworkPerformanceMetricSubscription_basic,
		acctest.CtDisappears: testAccNetworkPerformanceMetricSubscription_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccNetworkPerformanceMetricSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_network_performance_metric_subscription.test"
	src := acctest.AlternateRegion()
	dst := acctest.Region()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkPerformanceMetricSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkPerformanceMetricSubscription_basic(src, dst),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPerformanceMetricSubscriptionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDestination, dst),
					resource.TestCheckResourceAttr(resourceName, "metric", "aggregate-latency"),
					resource.TestCheckResourceAttr(resourceName, "period", "five-minutes"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSource, src),
					resource.TestCheckResourceAttr(resourceName, "statistic", "p50"),
				),
			},
		},
	})
}

func testAccNetworkPerformanceMetricSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_network_performance_metric_subscription.test"
	src := acctest.AlternateRegion()
	dst := acctest.Region()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkPerformanceMetricSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkPerformanceMetricSubscription_basic(src, dst),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkPerformanceMetricSubscriptionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkPerformanceMetricSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNetworkPerformanceMetricSubscriptionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindNetworkPerformanceMetricSubscriptionByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrSource], rs.Primary.Attributes[names.AttrDestination], rs.Primary.Attributes["metric"], rs.Primary.Attributes["statistic"])

		return err
	}
}

func testAccCheckNetworkPerformanceMetricSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_network_performance_metric_subscription" {
				continue
			}

			_, err := tfec2.FindNetworkPerformanceMetricSubscriptionByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrSource], rs.Primary.Attributes[names.AttrDestination], rs.Primary.Attributes["metric"], rs.Primary.Attributes["statistic"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 AWS Network Performance Metric Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCNetworkPerformanceMetricSubscription_basic(src, dst string) string {
	return fmt.Sprintf(`
resource "aws_vpc_network_performance_metric_subscription" "test" {
  source      = %[1]q
  destination = %[2]q
}
`, src, dst)
}
