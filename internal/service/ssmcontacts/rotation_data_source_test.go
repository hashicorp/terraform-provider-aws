// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"fmt"
	"testing"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRotationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssmcontacts_rotation.test"
	resourceName := "aws_ssmcontacts_rotation.test"
	startTime := time.Now().UTC().AddDate(0, 0, 2).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testRotationDataSourceConfig_basic(rName, startTime),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.number_of_on_calls", dataSourceName, "recurrence.0.number_of_on_calls"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.recurrence_multiplier", dataSourceName, "recurrence.0.recurrence_multiplier"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.#", dataSourceName, "recurrence.0.weekly_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.0.day_of_week", dataSourceName, "recurrence.0.weekly_settings.0.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.hour_of_day", dataSourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.minute_of_hour", dataSourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.minute_of_hour"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.1.day_of_week", dataSourceName, "recurrence.0.weekly_settings.1.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.hour_of_day", dataSourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.minute_of_hour", dataSourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.minute_of_hour"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.2.day_of_week", dataSourceName, "recurrence.0.weekly_settings.2.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.2.hand_off_time.0.hour_of_day", dataSourceName, "recurrence.0.weekly_settings.2.hand_off_time.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.2.hand_off_time.0.minute_of_hour", dataSourceName, "recurrence.0.weekly_settings.2.hand_off_time.0.minute_of_hour"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.#", dataSourceName, "recurrence.0.shift_coverages.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStartTime, dataSourceName, names.AttrStartTime),
					resource.TestCheckResourceAttrPair(resourceName, "time_zone_id", dataSourceName, "time_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_ids.#", dataSourceName, "contact_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_ids.0", dataSourceName, "contact_ids.0"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsKey1, dataSourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsKey2, dataSourceName, acctest.CtTagsKey2),
				),
			},
		},
	})
}

func testAccRotationDataSource_dailySettings(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssmcontacts_rotation.test"
	resourceName := "aws_ssmcontacts_rotation.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testRotationDataSourceConfig_dailySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.daily_settings.#", dataSourceName, "recurrence.0.daily_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.daily_settings.0.hour_of_day", dataSourceName, "recurrence.0.daily_settings.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.daily_settings.0.minute_of_hour", dataSourceName, "recurrence.0.daily_settings.0.minute_of_hour"),
				),
			},
		},
	})
}

func testAccRotationDataSource_monthlySettings(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssmcontacts_rotation.test"
	resourceName := "aws_ssmcontacts_rotation.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testRotationDataSourceConfig_monthlySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.#", dataSourceName, "recurrence.0.monthly_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.0.day_of_month", dataSourceName, "recurrence.0.monthly_settings.0.day_of_month"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.hour_of_day", dataSourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.minute_of_hour", dataSourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.minute_of_hour"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.1.day_of_month", dataSourceName, "recurrence.0.monthly_settings.1.day_of_month"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.hour_of_day", dataSourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.minute_of_hour", dataSourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.minute_of_hour"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.2.day_of_month", dataSourceName, "recurrence.0.monthly_settings.2.day_of_month"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.2.hand_off_time.0.hour_of_day", dataSourceName, "recurrence.0.monthly_settings.2.hand_off_time.0.hour_of_day"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.2.hand_off_time.0.minute_of_hour", dataSourceName, "recurrence.0.monthly_settings.2.hand_off_time.0.minute_of_hour"),
				),
			},
		},
	})
}

func testRotationDataSourceConfig_base(alias string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}

resource "aws_ssmcontacts_contact" "test" {
  alias = "test-contact-one-for-%[2]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, acctest.Region(), alias)
}

func testRotationDataSourceConfig_basic(rName, startTime string) string {
	return acctest.ConfigCompose(
		testRotationDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
    aws_ssmcontacts_contact.test.arn,
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    weekly_settings {
      day_of_week = "MON"
      hand_off_time {
        hour_of_day    = 4
        minute_of_hour = 25
      }
    }
    weekly_settings {
      day_of_week = "WED"
      hand_off_time {
        hour_of_day    = 7
        minute_of_hour = 34
      }
    }
    weekly_settings {
      day_of_week = "FRI"
      hand_off_time {
        hour_of_day    = 15
        minute_of_hour = 57
      }
    }
    shift_coverages {
      map_block_key = "MON"
      coverage_times {
        start {
          hour_of_day    = 1
          minute_of_hour = 0
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 0
        }
      }
    }
    shift_coverages {
      map_block_key = "WED"
      coverage_times {
        start {
          hour_of_day    = 1
          minute_of_hour = 0
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 0
        }
      }
    }
    shift_coverages {
      map_block_key = "FRI"
      coverage_times {
        start {
          hour_of_day    = 1
          minute_of_hour = 0
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 0
        }
      }
    }
  }

  start_time = %[2]q

  time_zone_id = "Australia/Sydney"

  tags = {
    key1 = "tag1"
    key2 = "tag2"
  }

  depends_on = [aws_ssmincidents_replication_set.test]
}

data "aws_ssmcontacts_rotation" "test" {
  arn = aws_ssmcontacts_rotation.test.arn
}
`, rName, startTime))
}

func testRotationDataSourceConfig_dailySettings(rName string) string {
	return acctest.ConfigCompose(
		testRotationDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
    aws_ssmcontacts_contact.test.arn,
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 18
      minute_of_hour = 0
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}

data "aws_ssmcontacts_rotation" "test" {
  arn = aws_ssmcontacts_rotation.test.arn
}
`, rName))
}

func testRotationDataSourceConfig_monthlySettings(rName string) string {
	return acctest.ConfigCompose(
		testRotationDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
    aws_ssmcontacts_contact.test.arn,
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    monthly_settings {
      day_of_month = 20
      hand_off_time {
        hour_of_day    = 8
        minute_of_hour = 0
      }
    }
    monthly_settings {
      day_of_month = 13
      hand_off_time {
        hour_of_day    = 12
        minute_of_hour = 34
      }
    }
    monthly_settings {
      day_of_month = 1
      hand_off_time {
        hour_of_day    = 4
        minute_of_hour = 58
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}

data "aws_ssmcontacts_rotation" "test" {
  arn = aws_ssmcontacts_rotation.test.arn
}
`, rName))
}
