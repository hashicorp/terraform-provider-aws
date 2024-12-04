// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsAnomalyDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var loganomalydetector cloudwatchlogs.GetLogAnomalyDetectorOutput
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
				Config: testAccLogAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, resourceName, &loganomalydetector),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var loganomalydetector cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.LogsServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLogAnomalyDetectorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, resourceName, &loganomalydetector),
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

			return fmt.Errorf("CloudwatchLogs Anomaly Detector %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckAnomalyDetectorExists(ctx context.Context, name string, loganomalydetector *cloudwatchlogs.GetLogAnomalyDetectorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameAnomalyDetector, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameAnomalyDetector, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		resp, err := tflogs.FindLogAnomalyDetectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameAnomalyDetector, rs.Primary.Attributes[names.AttrARN], err)
		}

		*loganomalydetector = *resp

		return nil
	}
}

func testAccAnomalyDetectorImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccLogAnomalyDetectorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_log_anomaly_detector" "test" {
  detector_name           = %[1]q
  log_group_arn_list      = [aws_cloudwatch_log_group.test[0].arn]
  anomaly_visibility_time = 7
  evaluation_frequency    = "TEN_MIN"
  enabled                 = "false"
}
`, rName)
}
