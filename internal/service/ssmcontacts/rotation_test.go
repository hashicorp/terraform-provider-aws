// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmcontacts "github.com/hashicorp/terraform-provider-aws/internal/service/ssmcontacts"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRotation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	timeZoneId := "Australia/Sydney"
	recurrenceMultiplier := 1

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
				Config: testAccRotationConfig_basic(rName, recurrenceMultiplier, timeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", timeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.number_of_on_calls", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.hour_of_day", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile(`rotation\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// We need to explicitly test destroying this resource instead of just using CheckDestroy,
				// because CheckDestroy will run after the replication set has been destroyed and destroying
				// the replication set will destroy all other resources.
				Config: testAccRotationConfig_replicationSetBase(),
				Check:  testAccCheckRotationDestroy(ctx),
			},
		},
	})
}

func testAccRotation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"
	timeZoneId := "Australia/Sydney"
	recurrenceMultiplier := 1

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
				Config: testAccRotationConfig_basic(rName, recurrenceMultiplier, timeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssmcontacts.ResourceRotation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRotation_updateRequiredFields(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	iniTimeZoneId := "Australia/Sydney"
	updTimeZoneId := "America/Los_Angeles"
	iniRecurrenceMultiplier := 1
	updRecurrenceMultiplier := 2

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
				Config: testAccRotationConfig_basic(rName, iniRecurrenceMultiplier, iniTimeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", iniTimeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_basic(rName, updRecurrenceMultiplier, updTimeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", updTimeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", acctest.Ct2),
				),
			},
		},
	})
}

func testAccRotation_startTime(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	iniStartTime := time.Now().UTC().AddDate(0, 0, 2).Format(time.RFC3339)
	updStartTime := time.Now().UTC().AddDate(20, 2, 10).Format(time.RFC3339)

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
				Config: testAccRotationConfig_startTime(rName, iniStartTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, iniStartTime),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_startTime(rName, updStartTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, updStartTime),
				),
			},
		},
	})
}

func testAccRotation_contactIds(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccRotationConfig_twoContacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_threeContacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", acctest.Ct3),
				),
			},
		},
	})
}

func testAccRotation_recurrence(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccRotationConfig_recurrenceDailySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.hour_of_day", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.minute_of_hour", acctest.Ct0),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceOneMonthlySetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.day_of_month", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.hour_of_day", "8"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.minute_of_hour", acctest.Ct0),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleMonthlySetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.day_of_month", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.hour_of_day", "8"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.day_of_month", "13"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.hour_of_day", "12"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.minute_of_hour", "34"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceOneWeeklySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.day_of_week", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.hour_of_day", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.minute_of_hour", "30"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleWeeklySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.day_of_week", "WED"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.hour_of_day", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.minute_of_hour", "25"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.day_of_week", "FRI"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.hour_of_day", "15"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.minute_of_hour", "57"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceOneShiftCoverages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.map_block_key", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.hour_of_day", "8"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.hour_of_day", "17"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.minute_of_hour", acctest.Ct0),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleShiftCoverages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.map_block_key", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.hour_of_day", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.hour_of_day", "23"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.map_block_key", "WED"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start.0.hour_of_day", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end.0.hour_of_day", "23"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.map_block_key", "FRI"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start.0.hour_of_day", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start.0.minute_of_hour", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end.0.hour_of_day", "23"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end.0.minute_of_hour", acctest.Ct0),
				),
			},
		},
	})
}

func testAccRotation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	tagKey1 := sdkacctest.RandString(26)
	tagVal1 := sdkacctest.RandString(26)
	tagVal1Updated := sdkacctest.RandString(26)
	tagKey2 := sdkacctest.RandString(26)
	tagVal2 := sdkacctest.RandString(26)

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
				Config: testAccRotationConfig_oneTag(rName, tagKey1, tagVal1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1),
				),
			},
			{
				Config: testAccRotationConfig_multipleTags(rName, tagKey1, tagVal1, tagKey2, tagVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey2, tagVal2),
				),
			},
			{
				Config: testAccRotationConfig_oneTag(rName, tagKey1, tagVal1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1Updated),
				),
			},
		},
	})
}

func testAccCheckRotationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_rotation" {
				continue
			}

			_, err := tfssmcontacts.FindRotationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				if strings.Contains(err.Error(), "Invalid value provided - Account not found for the request") {
					continue
				}

				return err
			}

			return create.Error(names.SSMContacts, create.ErrActionCheckingDestroyed, tfssmcontacts.ResNameRotation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRotationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameRotation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameRotation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)
		_, err := tfssmcontacts.FindRotationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameRotation, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

	input := &ssmcontacts.ListRotationsInput{}
	_, err := conn.ListRotations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		// ListRotations returns a 400 ValidationException if not onboarded with a replication set, and isn't an awsErr
		if strings.Contains(err.Error(), "Invalid value provided - Account not found for the request") {
			return
		}

		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRotationConfig_replicationSetBase() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}

func testAccRotationConfig_base(alias string, contactCount int) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_replicationSetBase(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test" {
  count = %[2]d
  alias = "%[1]s-${count.index}"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias, contactCount))
}

func testAccRotationConfig_basic(rName string, recurrenceMultiplier int, timeZoneId string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = %[2]d
    daily_settings {
      hour_of_day    = 1
      minute_of_hour = 00
    }
  }

  time_zone_id = %[3]q

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, recurrenceMultiplier, timeZoneId))
}

func testAccRotationConfig_startTime(rName, startTime string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 1
      minute_of_hour = 00
    }
  }

  start_time = %[2]q

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, startTime))
}

func testAccRotationConfig_twoContacts(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 2),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 1
      minute_of_hour = 00
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_threeContacts(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 3),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 1
      minute_of_hour = 00
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceDailySettings(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 1
      minute_of_hour = 00
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceOneMonthlySetting(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    monthly_settings {
      day_of_month = 20
      hand_off_time {
        hour_of_day    = 8
        minute_of_hour = 00
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceMultipleMonthlySetting(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    monthly_settings {
      day_of_month = 20
      hand_off_time {
        hour_of_day    = 8
        minute_of_hour = 00
      }
    }
    monthly_settings {
      day_of_month = 13
      hand_off_time {
        hour_of_day    = 12
        minute_of_hour = 34
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceOneWeeklySettings(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    weekly_settings {
      day_of_week = "MON"
      hand_off_time {
        hour_of_day    = 10
        minute_of_hour = 30
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceMultipleWeeklySettings(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn
  name        = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    weekly_settings {
      day_of_week = "WED"
      hand_off_time {
        hour_of_day    = 04
        minute_of_hour = 25
      }
    }

    weekly_settings {
      day_of_week = "FRI"
      hand_off_time {
        hour_of_day    = 15
        minute_of_hour = 57
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceOneShiftCoverages(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 9
      minute_of_hour = 00
    }
    shift_coverages {
      map_block_key = "MON"
      coverage_times {
        start {
          hour_of_day    = 08
          minute_of_hour = 00
        }
        end {
          hour_of_day    = 17
          minute_of_hour = 00
        }
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceMultipleShiftCoverages(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 9
      minute_of_hour = 00
    }

    shift_coverages {
      map_block_key = "MON"
      coverage_times {
        start {
          hour_of_day    = 01
          minute_of_hour = 00
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 00
        }
      }
    }
    shift_coverages {
      map_block_key = "WED"
      coverage_times {
        start {
          hour_of_day    = 01
          minute_of_hour = 00
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 00
        }
      }
    }
    shift_coverages {
      map_block_key = "FRI"
      coverage_times {
        start {
          hour_of_day    = 01
          minute_of_hour = 00
        }
        end {
          hour_of_day    = 23
          minute_of_hour = 00
        }
      }
    }
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_oneTag(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 18
      minute_of_hour = 00
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, tagKey, tagValue))
}

func testAccRotationConfig_multipleTags(rName, tagKey1, tagVal1, tagKey2, tagVal2 string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = aws_ssmcontacts_contact.test[*].arn

  name = %[1]q

  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 18
      minute_of_hour = 00
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, tagKey1, tagVal1, tagKey2, tagVal2))
}
