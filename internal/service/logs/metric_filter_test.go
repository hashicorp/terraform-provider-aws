// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsMetricFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mf types.MetricFilter
	resourceName := "aws_cloudwatch_log_metric_filter.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricFilterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricFilterExists(ctx, t, resourceName, &mf),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.default_value", ""),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.name", "metric1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.namespace", "ns1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.unit", "None"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pattern", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMetricFilterImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsMetricFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mf types.MetricFilter
	resourceName := "aws_cloudwatch_log_metric_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricFilterExists(ctx, t, resourceName, &mf),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceMetricFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsMetricFilter_Disappears_logGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var mf types.MetricFilter
	resourceName := "aws_cloudwatch_log_metric_filter.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricFilterExists(ctx, t, resourceName, &mf),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsMetricFilter_many(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_log_metric_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricFilterConfig_many(rName, 15),
				Check:  testAccCheckMetricFilterManyExists(ctx, t, resourceName, 15),
			},
		},
	})
}

func TestAccLogsMetricFilter_update(t *testing.T) {
	ctx := acctest.Context(t)
	var mf types.MetricFilter
	resourceName := "aws_cloudwatch_log_metric_filter.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricFilterConfig_allAttributes1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricFilterExists(ctx, t, resourceName, &mf),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.default_value", "2.5"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.name", "metric1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.namespace", "ns1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.unit", "Terabytes"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.value", "3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pattern", "[TEST]"),
					resource.TestCheckResourceAttr(resourceName, "apply_on_transformed_logs", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMetricFilterImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccMetricFilterConfig_allAttributes2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricFilterExists(ctx, t, resourceName, &mf),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.default_value", ""),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.d1", "$.d1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.d2", "$.d2"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.d3", "$.d3"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.name", "metric2"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.namespace", "ns2"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.unit", "Gigabits"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.value", "10"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pattern", `{ $.d1 = "OK" }`),
					resource.TestCheckResourceAttr(resourceName, "apply_on_transformed_logs", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccMetricFilterImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrLogGroupName] + ":" + rs.Primary.Attributes[names.AttrName], nil
	}
}

func testAccCheckMetricFilterExists(ctx context.Context, t *testing.T, n string, v *types.MetricFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindMetricFilterByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMetricFilterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_metric_filter" {
				continue
			}

			_, err := tflogs.FindMetricFilterByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName], rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Metric Filter still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMetricFilterManyExists(ctx context.Context, t *testing.T, basename string, n int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := range n {
			n := fmt.Sprintf("%s.%d", basename, i)
			var v types.MetricFilter

			err := testAccCheckMetricFilterExists(ctx, t, n, &v)(s)

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccMetricFilterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_metric_filter" "test" {
  name           = %[1]q
  pattern        = ""
  log_group_name = aws_cloudwatch_log_group.test.name

  metric_transformation {
    name      = "metric1"
    namespace = "ns1"
    value     = "1"
  }
}
`, rName)
}

func testAccMetricFilterConfig_many(rName string, n int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_metric_filter" "test" {
  count = %[2]d

  name           = "%[1]s-${count.index}"
  pattern        = "TEST"
  log_group_name = aws_cloudwatch_log_group.test.name

  metric_transformation {
    name      = "metric${count.index}"
    namespace = "ns1"
    value     = count.index
  }
}
`, rName, n)
}

func testAccMetricFilterConfig_allAttributes1(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_metric_filter" "test" {
  name           = %[1]q
  pattern        = "[TEST] "
  log_group_name = aws_cloudwatch_log_group.test.name

  metric_transformation {
    name          = "metric1"
    namespace     = "ns1"
    unit          = "Terabytes"
    value         = "3"
    default_value = "2.5"
  }
}
`, rName)
}

func testAccMetricFilterConfig_allAttributes2(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_metric_filter" "test" {
  name    = %[1]q
  pattern = <<EOS
    { $.d1 = "OK" }
EOS

  log_group_name            = aws_cloudwatch_log_group.test.name
  apply_on_transformed_logs = true

  metric_transformation {
    name      = "metric2"
    namespace = "ns2"
    unit      = "Gigabits"
    value     = "10"

    dimensions = {
      d1 = "$.d1"
      d2 = "$.d2"
      d3 = "$.d3"
    }
  }
}
`, rName)
}
