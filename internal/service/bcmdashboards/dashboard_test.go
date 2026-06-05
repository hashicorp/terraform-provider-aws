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

func TestAccBCMDashboardsDashboard_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard bcmdashboards.GetDashboardOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bcmdashboards_dashboard.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDashboardsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDashboardExists(ctx, t, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "dashboard_type", string(awstypes.DashboardTypeCustom)),
					resource.TestCheckResourceAttr(resourceName, "widget.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "widget.0.title", "example"),
					resource.TestCheckResourceAttr(resourceName, "widget.0.configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "widget.0.configs.0.query_parameters.0.cost_and_usage.0.granularity", string(awstypes.GranularityMonthly)),
					resource.TestCheckResourceAttr(resourceName, "widget.0.configs.0.display_config.0.graph.0.visual_type", string(awstypes.VisualTypeBar)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
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

func TestAccBCMDashboardsDashboard_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard bcmdashboards.GetDashboardOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bcmdashboards_dashboard.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDashboardsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDashboardExists(ctx, t, resourceName, &dashboard),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbcmdashboards.ResourceDashboard, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBCMDashboardsDashboard_update(t *testing.T) {
	ctx := acctest.Context(t)
	var dashboard bcmdashboards.GetDashboardOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bcmdashboards_dashboard.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDashboardsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDashboardExists(ctx, t, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "widget.#", "1"),
				),
			},
			{
				Config: testAccDashboardConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDashboardExists(ctx, t, resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "widget.0.configs.0.query_parameters.0.cost_and_usage.0.group_by.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "widget.0.configs.0.query_parameters.0.cost_and_usage.0.filter.0.and.#", "2"),
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

func testAccCheckDashboardDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BCMDashboardsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bcmdashboards_dashboard" {
				continue
			}

			_, err := tfbcmdashboards.FindDashboardByARN(ctx, conn, rs.Primary.ID)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.BCMDashboards, create.ErrActionCheckingDestroyed, tfbcmdashboards.ResNameDashboard, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDashboardExists(ctx context.Context, t *testing.T, name string, dashboard *bcmdashboards.GetDashboardOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BCMDashboards, create.ErrActionCheckingExistence, tfbcmdashboards.ResNameDashboard, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BCMDashboards, create.ErrActionCheckingExistence, tfbcmdashboards.ResNameDashboard, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BCMDashboardsClient(ctx)
		resp, err := tfbcmdashboards.FindDashboardByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BCMDashboards, create.ErrActionCheckingExistence, tfbcmdashboards.ResNameDashboard, rs.Primary.ID, err)
		}

		*dashboard = *resp

		return nil
	}
}

func testAccDashboardConfig_basic(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccDashboardConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_bcmdashboards_dashboard" "test" {
  name        = %[1]q
  description = "Managed by Terraform"

  widget {
    title       = "monthly data transfer"
    description = "Managed by Terraform"
    height      = 4
    width       = 4

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
              value = "2025-07-31"
            }
          }

          group_by {
            key  = "SERVICE"
            type = "DIMENSION"
          }

          filter {
            and {
              tags {
                key    = "Environment"
                values = ["production"]
              }
            }
            and {
              dimensions {
                key    = "USAGE_TYPE"
                values = ["DataTransfer-In-Bytes"]
              }
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
`, rName)
}
