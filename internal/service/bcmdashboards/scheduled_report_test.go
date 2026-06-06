// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bcmdashboards"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbcmdashboards "github.com/hashicorp/terraform-provider-aws/internal/service/bcmdashboards"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBCMDashboardsScheduledReport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var scheduledReport bcmdashboards.GetScheduledReportOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bcmdashboards_scheduled_report.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDashboardsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledReportDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledReportConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledReportExists(ctx, t, resourceName, &scheduledReport),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "dashboard_arn", "aws_bcmdashboards_dashboard.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_report_execution_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.state", string(awstypes.ScheduleStateEnabled)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
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

func TestAccBCMDashboardsScheduledReport_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var scheduledReport bcmdashboards.GetScheduledReportOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bcmdashboards_scheduled_report.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDashboardsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledReportDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledReportConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledReportExists(ctx, t, resourceName, &scheduledReport),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbcmdashboards.ResourceScheduledReport, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScheduledReportDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BCMDashboardsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bcmdashboards_scheduled_report" {
				continue
			}

			_, err := tfbcmdashboards.FindScheduledReportByARN(ctx, conn, rs.Primary.ID)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.BCMDashboards, create.ErrActionCheckingDestroyed, tfbcmdashboards.ResNameScheduledReport, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckScheduledReportExists(ctx context.Context, t *testing.T, name string, scheduledReport *bcmdashboards.GetScheduledReportOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BCMDashboards, create.ErrActionCheckingExistence, tfbcmdashboards.ResNameScheduledReport, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BCMDashboards, create.ErrActionCheckingExistence, tfbcmdashboards.ResNameScheduledReport, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BCMDashboardsClient(ctx)
		resp, err := tfbcmdashboards.FindScheduledReportByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BCMDashboards, create.ErrActionCheckingExistence, tfbcmdashboards.ResNameScheduledReport, rs.Primary.ID, err)
		}

		*scheduledReport = *resp

		return nil
	}
}

func testAccScheduledReportConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "bcm-dashboards.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["bcm-dashboards:GetDashboard"]
        Resource = "arn:${data.aws_partition.current.partition}:bcm-dashboards::*:dashboard/*"
      },
      {
        Effect = "Allow"
        Action = [
          "ce:GetCostAndUsage",
          "ce:GetDimensionValues",
          "ce:GetTags",
          "ce:GetCostCategories",
          "ce:GetSavingsPlansCoverage",
          "ce:GetReservationUtilization",
          "ce:GetReservationCoverage",
          "ce:GetSavingsPlansUtilization",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_bcmdashboards_dashboard" "test" {
  name = %[1]q

  widget {
    title = "example"

    configs {
      query_parameters {
        cost_and_usage {
          granularity = "MONTHLY"
          metrics     = ["UnblendedCost"]

          time_range {
            start_time {
              type  = "ABSOLUTE"
              value = "2025-01-01"
            }
            end_time {
              type  = "ABSOLUTE"
              value = "2025-03-31"
            }
          }
        }
      }

      display_config {
        graph {
          metric      = "UnblendedCost"
          visual_type = "BAR"
        }
      }
    }
  }
}

resource "aws_bcmdashboards_scheduled_report" "test" {
  name                                = %[1]q
  dashboard_arn                       = aws_bcmdashboards_dashboard.test.arn
  scheduled_report_execution_role_arn = aws_iam_role.test.arn

  schedule_config {
    schedule_expression           = "cron(0 9 1 * ? *)"
    schedule_expression_time_zone = "UTC"
    state                         = "ENABLED"
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}
