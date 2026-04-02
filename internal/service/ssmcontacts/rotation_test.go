// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssmcontacts "github.com/hashicorp/terraform-provider-aws/internal/service/ssmcontacts"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRotation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	timeZoneId := "Australia/Sydney"
	recurrenceMultiplier := 1

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, recurrenceMultiplier, timeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", timeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.number_of_on_calls", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.hour_of_day", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile(`rotation/.+$`)),
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
				Check:  testAccCheckRotationDestroy(ctx, t),
			},
		},
	})
}

// testAccSSMContactsRotation_Identity_regionOverride cannot be generated, because the test requires `aws_ssmincidents_replication_set`, which doesn't support region override
func testAccSSMContactsRotation_Identity_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	timeZoneId := "Australia/Sydney"
	recurrenceMultiplier := 1

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_regionOverride(rName, recurrenceMultiplier, timeZoneId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},

			// Import command with appended "@<region>"
			{
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.CrossRegionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},

			// Import command without appended "@<region>"
			{
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Import block with Import ID and appended "@<region>"
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithID,
				ImportStateIdFunc: acctest.CrossRegionImportStateIdFunc(resourceName),
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
				},
			},

			// Import block with Import ID and no appended "@<region>"
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
				},
			},

			// Import block with Resource Identity
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
				},
			},
		},
	})
}

func testAccRotation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"
	timeZoneId := "Australia/Sydney"
	recurrenceMultiplier := 1

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, recurrenceMultiplier, timeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssmcontacts.ResourceRotation, resourceName),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	iniTimeZoneId := "Australia/Sydney"
	updTimeZoneId := "America/Los_Angeles"
	iniRecurrenceMultiplier := 1
	updRecurrenceMultiplier := 2

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, iniRecurrenceMultiplier, iniTimeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", iniTimeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", "1"),
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
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", updTimeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", "2"),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	iniStartTime := time.Now().UTC().AddDate(0, 0, 2).Format(time.RFC3339)
	updStartTime := time.Now().UTC().AddDate(20, 2, 10).Format(time.RFC3339)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_startTime(rName, iniStartTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
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
					testAccCheckRotationExists(ctx, t, resourceName),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_twoContacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", "2"),
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
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", "3"),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_recurrenceDailySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.hour_of_day", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0.minute_of_hour", "0"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceOneMonthlySetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.day_of_month", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.hour_of_day", "8"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.minute_of_hour", "0"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleMonthlySetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.day_of_month", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.hour_of_day", "8"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.day_of_month", "13"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.hour_of_day", "12"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.hand_off_time.0.minute_of_hour", "34"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceOneWeeklySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.day_of_week", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.hour_of_day", "10"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.minute_of_hour", "30"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleWeeklySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.day_of_week", "WED"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.hour_of_day", "4"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time.0.minute_of_hour", "25"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.day_of_week", "FRI"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.hour_of_day", "15"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.hand_off_time.0.minute_of_hour", "57"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceOneShiftCoverages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.map_block_key", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.hour_of_day", "8"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.hour_of_day", "17"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.minute_of_hour", "0"),
				),
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleShiftCoverages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.map_block_key", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.hour_of_day", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.hour_of_day", "23"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.map_block_key", "WED"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start.0.hour_of_day", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end.0.hour_of_day", "23"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.map_block_key", "FRI"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start.0.hour_of_day", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end.0.hour_of_day", "23"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end.0.minute_of_hour", "0"),
				),
			},
		},
	})
}

func testAccCheckRotationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_rotation" {
				continue
			}

			_, err := tfssmcontacts.FindRotationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckRotationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameRotation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameRotation, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SSMContactsClient(ctx)
		_, err := tfssmcontacts.FindRotationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameRotation, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SSMContactsClient(ctx)

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

func testAccRotationConfig_replicationSetBase_regionOverride() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  provider = awsalternate

  region {
    name = %[1]q
  }
}
`, acctest.AlternateRegion())
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

func testAccRotationConfig_base_regionOverride(alias string, contactCount int) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_replicationSetBase_regionOverride(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test" {
  region = %[3]q

  count = %[2]d
  alias = "%[1]s-${count.index}"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias, contactCount, acctest.AlternateRegion()))
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

func testAccRotationConfig_regionOverride(rName string, recurrenceMultiplier int, timeZoneId string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccRotationConfig_base_regionOverride(rName, 1),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  region = %[4]q

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
}`, rName, recurrenceMultiplier, timeZoneId, acctest.AlternateRegion()))
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
