// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreCheckScraperLoggingConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

	var input amp.ListScrapersInput
	_, err := conn.ListScrapers(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccAMPScraperLoggingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v amp.DescribeScraperLoggingConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_scraper_logging_configuration.test"
	scraperResourceName := "aws_prometheus_scraper.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScraperLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScraperLoggingConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScraperLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScraperLoggingConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "scraper_id", scraperResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "logging_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_destination.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_destination.0.cloudwatch_logs.0.log_group_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "scraper_id"),
				ImportStateVerifyIdentifierAttribute: "scraper_id",
			},
		},
	})
}

func TestAccAMPScraperLoggingConfiguration_scraperComponents(t *testing.T) {
	ctx := acctest.Context(t)
	var v amp.DescribeScraperLoggingConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_scraper_logging_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScraperLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScraperLoggingConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScraperLoggingConfigurationConfig_scraperComponents(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScraperLoggingConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scraper_components.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scraper_components.0.type", "COLLECTOR"),
					resource.TestCheckResourceAttr(resourceName, "scraper_components.1.type", "EXPORTER"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "scraper_id"),
				ImportStateVerifyIdentifierAttribute: "scraper_id",
				ImportStateVerifyIgnore:              []string{"scraper_components"},
			},
		},
	})
}

func TestAccAMPScraperLoggingConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v amp.DescribeScraperLoggingConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_scraper_logging_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckScraperLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScraperLoggingConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScraperLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScraperLoggingConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfamp.ResourceScraperLoggingConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScraperLoggingConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_scraper_logging_configuration" {
				continue
			}

			_, err := tfamp.FindScraperLoggingConfigurationByID(ctx, conn, rs.Primary.Attributes["scraper_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Prometheus Scraper Logging Configuration %s still exists", rs.Primary.Attributes["scraper_id"])
		}

		return nil
	}
}

func testAccCheckScraperLoggingConfigurationExists(ctx context.Context, t *testing.T, n string, v *amp.DescribeScraperLoggingConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		output, err := tfamp.FindScraperLoggingConfigurationByID(ctx, conn, rs.Primary.Attributes["scraper_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccScraperLoggingConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccScraperConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/prometheus/scraper-logs/%[1]s"
}

resource "aws_prometheus_scraper_logging_configuration" "test" {
  scraper_id = aws_prometheus_scraper.test.id

  logging_destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
    }
  }
}
`, rName))
}

func testAccScraperLoggingConfigurationConfig_scraperComponents(rName string) string {
	return acctest.ConfigCompose(
		testAccScraperConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/prometheus/scraper-logs/%[1]s"
}

resource "aws_prometheus_scraper_logging_configuration" "test" {
  scraper_id = aws_prometheus_scraper.test.id

  scraper_components {
    type = "COLLECTOR"
  }

  scraper_components {
    type = "EXPORTER"
  }

  logging_destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
    }
  }
}
`, rName))
}
