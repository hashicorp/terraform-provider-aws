// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightRefreshSchedule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var schedule awstypes.RefreshSchedule
	resourceName := "aws_quicksight_refresh_schedule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_basic(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.time_of_the_day", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.timezone", "Europe/London"),
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

func TestAccQuickSightRefreshSchedule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var schedule awstypes.RefreshSchedule
	resourceName := "aws_quicksight_refresh_schedule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_basic(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, resourceName, &schedule),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceRefreshSchedule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightRefreshSchedule_weeklyRefresh(t *testing.T) {
	ctx := acctest.Context(t)
	var schedule awstypes.RefreshSchedule
	resourceName := "aws_quicksight_refresh_schedule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_WeeklyRefresh(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "WEEKLY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_week", "MONDAY"),
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

func TestAccQuickSightRefreshSchedule_invalidWeeklyRefresh(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_WeeklyRefresh_NoRefreshOnDay(rId, rName, sId),
				ExpectError: fwdiag.ExpectAttributeRequiredWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day[0].day_of_week",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalWeekly),
				),
			},
			{
				Config: testAccRefreshScheduleConfig_WeeklyRefresh_NoDayOfWeek(rId, rName, sId),
				ExpectError: fwdiag.ExpectAttributeRequiredWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day[0].day_of_week",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalWeekly),
				),
			},
		},
	})
}

func TestAccQuickSightRefreshSchedule_monthlyRefresh(t *testing.T) {
	ctx := acctest.Context(t)
	var schedule awstypes.RefreshSchedule
	resourceName := "aws_quicksight_refresh_schedule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_MonthlyRefresh(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "MONTHLY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_month", acctest.Ct1),
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

func TestAccQuickSightRefreshSchedule_invalidMonthlyRefresh(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_MonthlyRefresh_NoRefreshOnDay(rId, rName, sId),
				ExpectError: fwdiag.ExpectAttributeRequiredWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day[0].day_of_month",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalMonthly),
				),
			},
			{
				Config: testAccRefreshScheduleConfig_MonthlyRefresh_NoDayOfMonth(rId, rName, sId),
				ExpectError: fwdiag.ExpectAttributeRequiredWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day[0].day_of_month",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalMonthly),
				),
			},
		},
	})
}

func TestAccQuickSightRefreshSchedule_invalidRefreshInterval(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_InvalidRefreshInterval(rId, rName, sId, string(awstypes.RefreshIntervalDaily)),
				ExpectError: fwdiag.ExpectAttributeConflictsWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalDaily),
				),
			},
			{
				Config: testAccRefreshScheduleConfig_InvalidRefreshInterval(rId, rName, sId, string(awstypes.RefreshIntervalHourly)),
				ExpectError: fwdiag.ExpectAttributeConflictsWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalHourly),
				),
			},
			{
				Config: testAccRefreshScheduleConfig_InvalidRefreshInterval(rId, rName, sId, string(awstypes.RefreshIntervalMinute30)),
				ExpectError: fwdiag.ExpectAttributeConflictsWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalMinute30),
				),
			},
			{
				Config: testAccRefreshScheduleConfig_InvalidRefreshInterval(rId, rName, sId, string(awstypes.RefreshIntervalMinute15)),
				ExpectError: fwdiag.ExpectAttributeConflictsWhenError(
					"schedule[0].schedule_frequency[0].refresh_on_day",
					"schedule[0].schedule_frequency[0].interval",
					string(awstypes.RefreshIntervalMinute15),
				),
			},
		},
	})
}

func testAccCheckRefreshScheduleExists(ctx context.Context, n string, v *awstypes.RefreshSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		_, output, err := tfquicksight.FindRefreshScheduleByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_set_id"], rs.Primary.Attributes["schedule_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRefreshScheduleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_refresh_schedule" {
				continue
			}

			_, _, err := tfquicksight.FindRefreshScheduleByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_set_id"], rs.Primary.Attributes["schedule_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Refresh Schedule (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRefreshScheduleConfig_base(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  type = "S3"
}

resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q
  import_mode = "SPICE"

  physical_table_map {
    physical_table_map_id = %[1]q
    s3_source {
      data_source_arn = aws_quicksight_data_source.test.arn
      input_columns {
        name = "Column1"
        type = "STRING"
      }
      upload_settings {
        format = "JSON"
      }
    }
  }
}
`, rId, rName))
}

func testAccRefreshScheduleConfig_basic(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval        = "DAILY"
      time_of_the_day = "12:00"
      timezone        = "Europe/London"
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_WeeklyRefresh(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = "WEEKLY"
      refresh_on_day {
        day_of_week = "MONDAY"
      }
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_WeeklyRefresh_NoRefreshOnDay(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = "WEEKLY"
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_WeeklyRefresh_NoDayOfWeek(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = "WEEKLY"
      refresh_on_day {
      }
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_MonthlyRefresh(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = "MONTHLY"
      refresh_on_day {
        day_of_month = "1"
      }
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_MonthlyRefresh_NoRefreshOnDay(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = "MONTHLY"
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_MonthlyRefresh_NoDayOfMonth(rId, rName, sId string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = "MONTHLY"
      refresh_on_day {
      }
    }
  }
}
`, sId))
}

func testAccRefreshScheduleConfig_InvalidRefreshInterval(rId, rName, sId, interval string) string {
	return acctest.ConfigCompose(
		testAccRefreshScheduleConfig_base(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_refresh_schedule" "test" {
  data_set_id = aws_quicksight_data_set.test.data_set_id
  schedule_id = %[1]q
  schedule {
    refresh_type = "FULL_REFRESH"
    schedule_frequency {
      interval = %[2]q
      refresh_on_day {
      }
    }
  }
}
`, sId, interval))
}
