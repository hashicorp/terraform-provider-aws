// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightRefreshSchedule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var schedule awstypes.RefreshSchedule
	resourceName := "aws_quicksight_refresh_schedule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_basic(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.time_of_the_day", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.timezone", "Europe/London"),
					// acctest.CheckResourceAttrRFC3339(resourceName, "schedule.0.start_after_date_time"),
					resource.TestMatchResourceAttr(resourceName, "schedule.0.start_after_date_time", regexache.MustCompile(`^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_basic(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceRefreshSchedule, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expectNoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_WeeklyRefresh(rId, rName, sId, awstypes.DayOfWeekMonday),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "WEEKLY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_month"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_week", string(awstypes.DayOfWeekMonday)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRefreshScheduleConfig_WeeklyRefresh(rId, rName, sId, awstypes.DayOfWeekWednesday),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "WEEKLY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_month"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_week", string(awstypes.DayOfWeekWednesday)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expectNoChange := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_MonthlyRefresh(rId, rName, sId, "15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "MONTHLY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_month", "15"),
					resource.TestCheckNoResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_week"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRefreshScheduleConfig_MonthlyRefresh(rId, rName, sId, "21"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight",
						fmt.Sprintf("dataset/%s/refresh-schedule/%s", rId, sId)),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "MONTHLY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_month", "21"),
					resource.TestCheckNoResourceAttr(resourceName, "schedule.0.schedule_frequency.0.refresh_on_day.0.day_of_week"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
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

func TestAccQuickSightRefreshSchedule_startAfterDateTime(t *testing.T) {
	ctx := acctest.Context(t)
	var schedule awstypes.RefreshSchedule
	resourceName := "aws_quicksight_refresh_schedule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// We expect the value to change between steps 1 and 3
	expectChange := statecheck.CompareValue(compare.ValuesDiffer())
	// ... and not the change between 3 and 5
	expectNoChange := statecheck.CompareValue(compare.ValuesSame())

	now := time.Now()
	startTime1 := now.AddDate(1, 0, 0).Format(tfquicksight.StartAfterDateTimeLayout)
	startTime2 := now.AddDate(1, 1, 0).Format(tfquicksight.StartAfterDateTimeLayout)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRefreshScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRefreshScheduleConfig_startAfterDateTime(rId, rName, sId, startTime1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.time_of_the_day", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.timezone", "Europe/London"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.start_after_date_time", startTime1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRefreshScheduleConfig_startAfterDateTime(rId, rName, sId, startTime2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.time_of_the_day", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.timezone", "Europe/London"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.start_after_date_time", startTime2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
					expectNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRefreshScheduleConfig_startAfterDateTime_Removed(rId, rName, sId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRefreshScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.refresh_type", "FULL_REFRESH"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.interval", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.time_of_the_day", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_frequency.0.timezone", "Europe/London"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.start_after_date_time", startTime2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrSchedule).AtSliceIndex(0).AtMapKey("start_after_date_time")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRefreshScheduleExists(ctx context.Context, t *testing.T, n string, v *awstypes.RefreshSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, output, err := tfquicksight.FindRefreshScheduleByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_set_id"], rs.Primary.Attributes["schedule_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRefreshScheduleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_refresh_schedule" {
				continue
			}

			_, _, err := tfquicksight.FindRefreshScheduleByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_set_id"], rs.Primary.Attributes["schedule_id"])

			if retry.NotFound(err) {
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

func testAccRefreshScheduleConfig_WeeklyRefresh(rId, rName, sId string, dayOfWeek awstypes.DayOfWeek) string {
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
        day_of_week = %[2]q
      }
    }
  }
}
`, sId, dayOfWeek))
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

func testAccRefreshScheduleConfig_MonthlyRefresh(rId, rName, sId, dayOfMonth string) string {
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
        day_of_month = %[2]q
      }
    }
  }
}
`, sId, dayOfMonth))
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

func testAccRefreshScheduleConfig_startAfterDateTime(rId, rName, sId, startAfter string) string {
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
    start_after_date_time = %[2]q
  }
}
`, sId, startAfter))
}

func testAccRefreshScheduleConfig_startAfterDateTime_Removed(rId, rName, sId string) string {
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
