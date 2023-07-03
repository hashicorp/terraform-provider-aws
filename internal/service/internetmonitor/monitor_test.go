// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package internetmonitor_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfinternetmonitor "github.com/hashicorp/terraform-provider-aws/internal/service/internetmonitor"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInternetMonitorMonitor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_internetmonitor_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InternetMonitorEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "internetmonitor", regexp.MustCompile(`monitor/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "internet_measurements_log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "max_city_networks_to_monitor", "0"),
					resource.TestCheckResourceAttr(resourceName, "monitor_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "traffic_percentage_to_monitor", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorConfig_status(rName, "INACTIVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "INACTIVE"),
				),
			},
		},
	})
}

func TestAccInternetMonitorMonitor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_internetmonitor_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InternetMonitorEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinternetmonitor.ResourceMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccInternetMonitorMonitor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_internetmonitor_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InternetMonitorEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMonitorConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccInternetMonitorMonitor_healthEventsConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_internetmonitor_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InternetMonitorEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_healthEventsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.0.availability_score_threshold", "50"),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.0.performance_score_threshold", "95"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorConfig_healthEventsConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.0.availability_score_threshold", "75"),
					resource.TestCheckResourceAttr(resourceName, "health_events_config.0.performance_score_threshold", "85"),
				),
			},
		},
	})
}

func TestAccInternetMonitorMonitor_log(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_internetmonitor_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InternetMonitorEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorConfig_log(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "internet_measurements_log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "internet_measurements_log_delivery.0.s3_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "internet_measurements_log_delivery.0.s3_config.0.bucket_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMonitorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InternetMonitorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_internetmonitor_monitor" {
				continue
			}

			_, err := tfinternetmonitor.FindMonitorByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Internet Monitor Monitor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMonitorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).InternetMonitorClient(ctx)

		_, err := tfinternetmonitor.FindMonitorByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_internetmonitor_monitor" "test" {
  monitor_name                  = %[1]q
  traffic_percentage_to_monitor = 1
}
`, rName)
}

func testAccMonitorConfig_status(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_internetmonitor_monitor" "test" {
  monitor_name                  = %[1]q
  traffic_percentage_to_monitor = 1
  status                        = %[2]q
}
`, rName, status)
}

func testAccMonitorConfig_healthEventsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_internetmonitor_monitor" "test" {
  monitor_name                 = %[1]q
  max_city_networks_to_monitor = 2

  health_events_config {
    availability_score_threshold = 50
  }
}
`, rName)
}

func testAccMonitorConfig_healthEventsConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_internetmonitor_monitor" "test" {
  monitor_name                 = %[1]q
  max_city_networks_to_monitor = 2

  health_events_config {
    availability_score_threshold = 75
    performance_score_threshold  = 85
  }
}
`, rName)
}

func testAccMonitorConfig_log(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["s3:PutObject"]
    effect    = "Allow"
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }

  statement {
    actions   = ["s3:GetBucketAcl"]
    effect    = "Allow"
    resources = [aws_s3_bucket.test.arn]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_internetmonitor_monitor" "test" {
  monitor_name                  = %[1]q
  traffic_percentage_to_monitor = 1

  internet_measurements_log_delivery {
    s3_config {
      bucket_name = aws_s3_bucket_policy.test.bucket
    }
  }
}
`, rName)
}

func testAccMonitorConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_internetmonitor_monitor" "test" {
  monitor_name                  = %[1]q
  traffic_percentage_to_monitor = 1

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccMonitorConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_internetmonitor_monitor" "test" {
  monitor_name                  = %[1]q
  traffic_percentage_to_monitor = 1

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
