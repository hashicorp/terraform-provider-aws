// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsAnomalyDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AnomalyDetector/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_frequency", "TEN_MIN"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "log_group_arn_list.#"),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AnomalyDetector/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					names.AttrEnabled,
				},
			},
		},
	})
}

func TestAccLogsAnomalyDetector_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AnomalyDetector/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:         config.StringVariable(rName),
					"evaluationFrequency":   config.StringVariable("TEN_MIN"),
					"enabled":               config.BoolVariable(false),
					"anomalyVisibilityTime": config.IntegerVariable(7),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_frequency", "TEN_MIN"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "log_group_arn_list.#"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "7"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AnomalyDetector/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:         config.StringVariable(rName),
					"evaluationFrequency":   config.StringVariable("TEN_MIN"),
					"enabled":               config.BoolVariable(false),
					"anomalyVisibilityTime": config.IntegerVariable(7),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					names.AttrEnabled,
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AnomalyDetector/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:         config.StringVariable(rName),
					"evaluationFrequency":   config.StringVariable("FIVE_MIN"),
					"enabled":               config.BoolVariable(true),
					"anomalyVisibilityTime": config.IntegerVariable(8),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "detector_name"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_frequency", "FIVE_MIN"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_visibility_time", "8"),
					resource.TestCheckResourceAttrSet(resourceName, "log_group_arn_list.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLogsAnomalyDetector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.GetLogAnomalyDetectorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_anomaly_detector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalyDetectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AnomalyDetector/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalyDetectorExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflogs.ResourceAnomalyDetector, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAnomalyDetectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_anomaly_detector" {
				continue
			}

			_, err := tflogs.FindLogAnomalyDetectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
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

func testAccCheckAnomalyDetectorExists(ctx context.Context, t *testing.T, n string, v *cloudwatchlogs.GetLogAnomalyDetectorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindLogAnomalyDetectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
