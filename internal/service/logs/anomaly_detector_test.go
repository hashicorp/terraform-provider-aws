// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsAnomalyDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_frequency", "TEN_MIN"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "log_group_arn_list.#"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccAnomalyDetectorImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccLogsAnomalyDetector_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_update(rName, "TEN_MIN", acctest.CtFalse, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_frequency", "TEN_MIN"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "log_group_arn_list.#"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "7"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccAnomalyDetectorImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrEnabled},
			},
			{
				Config: testAccAnomalyDetectorConfig_update(rName, "FIVE_MIN", acctest.CtTrue, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_frequency", "FIVE_MIN"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "8"),
					resource.TestCheckResourceAttrSet(resourceName, "log_group_arn_list.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccAnomalyDetectorImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccLogsAnomalyDetector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceAnomalyDetector, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAnomalyDetectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_anomaly_detector" {
				continue
			}

			_, err := tflogs.FindLogAnomalyDetectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Anomaly Detector still exists: %s", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckAnomalyDetectorExists(ctx context.Context, n string, v *cloudwatchlogs.GetLogAnomalyDetectorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		output, err := tflogs.FindLogAnomalyDetectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAnomalyDetectorImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccAnomalyDetectorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_log_anomaly_detector" "test" {
  detector_name           = %[1]q
  log_group_arn_list      = [aws_cloudwatch_log_group.test[0].arn]
  anomaly_visibility_time = 7
  evaluation_frequency    = "TEN_MIN"
  enabled                 = false
}
`, rName)
}

func testAccAnomalyDetectorConfig_update(rName string, ef string, enabled string, avt int64) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_log_anomaly_detector" "test" {
  detector_name           = %[1]q
  log_group_arn_list      = [aws_cloudwatch_log_group.test[0].arn]
  anomaly_visibility_time = %[4]d
  evaluation_frequency    = %[2]q
  enabled                 = %[3]q
}
`, rName, ef, enabled, avt)
}
