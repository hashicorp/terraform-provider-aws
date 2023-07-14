package ssmcontacts_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	tfssmcontacts "github.com/hashicorp/terraform-provider-aws/internal/service/ssmcontacts"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMContactsRotation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"
	contactResourceName := "aws_ssmcontacts_contact.test_contact_one"

	timeZoneId := "Australia/Sydney"
	numberOfOncalls := 1
	recurrenceMultiplier := 1

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, recurrenceMultiplier, timeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, contactResourceName),
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", timeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.number_of_on_calls", strconv.Itoa(numberOfOncalls)),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", strconv.Itoa(recurrenceMultiplier)),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0", "01:00"),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "contact_ids.0", "ssm-contacts", "contact/test-contact-one-for-"+rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssm-contacts", regexp.MustCompile(`rotation\/+.`)),
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
				Config: testAccRotationConfig_none(),
				Check:  testAccCheckRotationDestroy(ctx),
			},
		},
	})
}

func TestAccSSMContactsRotation_disappears(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, recurrenceMultiplier, timeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssmcontacts.ResourceRotation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMContactsRotation_updateRequiredFields(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, iniRecurrenceMultiplier, iniTimeZoneId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "time_zone_id", iniTimeZoneId),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", strconv.Itoa(iniRecurrenceMultiplier)),
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
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.recurrence_multiplier", strconv.Itoa(updRecurrenceMultiplier)),
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

func TestAccSSMContactsRotation_startTime(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_startTime(rName, iniStartTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_time", iniStartTime),
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
					resource.TestCheckResourceAttr(resourceName, "start_time", updStartTime),
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

func TestAccSSMContactsRotation_contactIds(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_rotation.test"
	firstContactResourceName := "aws_ssmcontacts_contact.test_contact_one"
	secondContactResourceName := "aws_ssmcontacts_contact.test_contact_two"
	thirdContactResourceName := "aws_ssmcontacts_contact.test_contact_three"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_twoContacts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					testAccCheckContactExists(ctx, firstContactResourceName),
					testAccCheckContactExists(ctx, secondContactResourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", "2"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "contact_ids.0", "ssm-contacts", "contact/test-contact-one-for-"+rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "contact_ids.1", "ssm-contacts", "contact/test-contact-two-for-"+rName),
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
					testAccCheckContactExists(ctx, firstContactResourceName),
					testAccCheckContactExists(ctx, secondContactResourceName),
					testAccCheckContactExists(ctx, thirdContactResourceName),
					resource.TestCheckResourceAttr(resourceName, "contact_ids.#", "3"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "contact_ids.0", "ssm-contacts", "contact/test-contact-one-for-"+rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "contact_ids.1", "ssm-contacts", "contact/test-contact-two-for-"+rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "contact_ids.2", "ssm-contacts", "contact/test-contact-three-for-"+rName),
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

func TestAccSSMContactsRotation_recurrence(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_recurrenceDailySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.daily_settings.0", "18:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_recurrenceOneMonthlySetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.day_of_month", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time", "08:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleMonthlySetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.day_of_month", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.0.hand_off_time", "08:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.day_of_month", "13"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.monthly_settings.1.hand_off_time", "12:34"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_recurrenceOneWeeklySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.day_of_week", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time", "10:30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleWeeklySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.day_of_week", "WED"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.0.hand_off_time", "04:25"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.day_of_week", "FRI"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.weekly_settings.1.hand_off_time", "15:57"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_recurrenceOneShiftCoverages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.day_of_week", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start_time", "08:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end_time", "17:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_recurrenceMultipleShiftCoverages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.day_of_week", "FRI"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start_time", "01:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end_time", "23:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.day_of_week", "MON"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start_time", "01:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end_time", "23:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.day_of_week", "WED"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start_time", "01:00"),
					resource.TestCheckResourceAttr(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end_time", "23:00"),
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

func TestAccSSMContactsRotation_tags(t *testing.T) {
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
	tagVal2Updated := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRotationConfig_basic(rName, 1, "Australia/Sydney"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_oneTag(rName, tagKey1, tagVal1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_multipleTags(rName, tagKey1, tagVal1, tagKey2, tagVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey2, tagVal2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_multipleTags(rName, tagKey1, tagVal1Updated, tagKey2, tagVal2Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey2, tagVal2Updated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_oneTag(rName, tagKey1, tagVal1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tagKey1, tagVal1Updated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRotationConfig_basic(rName, 1, "Australia/Sydney"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRotationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckRotationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_rotation" {
				continue
			}

			input := &ssmcontacts.GetRotationInput{
				RotationId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetRotation(ctx, input)
			if err != nil {
				if strings.Contains(err.Error(), "Invalid value provided - Account not found for the request") {
					continue
				}

				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
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

func testAccRotationConfig_none() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}

func testAccRotationConfig_base(alias string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_none(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test_contact_one" {
  alias = "test-contact-one-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}

func testAccRotationConfig_secondContact(alias string) string {
	return fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test_contact_two" {
  alias = "test-contact-two-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias)
}

func testAccRotationConfig_thirdContact(alias string) string {
	return fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test_contact_three" {
  alias = "test-contact-three-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias)
}

func testAccRotationConfig_basic(rName string, recurrenceMultiplier int, timeZoneId string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = %[2]d
	daily_settings = [
		"01:00"
	]
  }
 
 time_zone_id = %[3]q

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, recurrenceMultiplier, timeZoneId))
}

func testAccRotationConfig_startTime(rName, startTime string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"01:00"
	]
  }

  start_time = %[2]q

  time_zone_id = "Australia/Sydney"

  depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, startTime))
}

func testAccRotationConfig_twoContacts(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		testAccRotationConfig_secondContact(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn,
	aws_ssmcontacts_contact.test_contact_two.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"01:00"
	]
  }

 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_threeContacts(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		testAccRotationConfig_secondContact(rName),
		testAccRotationConfig_thirdContact(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn,
	aws_ssmcontacts_contact.test_contact_two.arn,
	aws_ssmcontacts_contact.test_contact_three.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"01:00"
	]
  }

 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceDailySettings(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"18:00"
	]
  }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceOneMonthlySetting(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	monthly_settings {
		day_of_month = 20
		hand_off_time = "08:00"
	}
  }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceMultipleMonthlySetting(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	monthly_settings {
		day_of_month = 20
		hand_off_time = "08:00"
	}
	monthly_settings {
		day_of_month = 13
		hand_off_time = "12:34"
	}
  }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceOneWeeklySettings(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	weekly_settings {
		day_of_week = "MON"
		hand_off_time = "10:30"
	}
  }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceMultipleWeeklySettings(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	weekly_settings {
		day_of_week = "WED"
		hand_off_time = "04:25"
	}

	weekly_settings {
		day_of_week = "FRI"
		hand_off_time = "15:57"
	}
  }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceOneShiftCoverages(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"09:00"
	]
    shift_coverages {
		day_of_week = "MON"
		coverage_times {
		  start_time = "08:00"
		  end_time = "17:00"
		}
  	}
  }

 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_recurrenceMultipleShiftCoverages(rName string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"09:00"
	]
    shift_coverages {
		day_of_week = "MON"
		coverage_times {
		  start_time = "01:00"
		  end_time = "23:00"
		}
  	}
	shift_coverages {
		day_of_week = "WED"
		coverage_times {
		  start_time = "01:00"
		  end_time = "23:00"
		}
  	}
	shift_coverages {
		day_of_week = "FRI"
		coverage_times {
		  start_time = "01:00"
		  end_time = "23:00"
		}
  	}
  }

 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName))
}

func testAccRotationConfig_oneTag(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"18:00"
	]
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
		testAccRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [
	aws_ssmcontacts_contact.test_contact_one.arn
  ]

  name = %[1]q

  recurrence {
    number_of_on_calls = 1
	recurrence_multiplier = 1
	daily_settings = [
		"18:00"
	]
  }

 tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
 }
 
 time_zone_id = "Australia/Sydney"

 depends_on = [aws_ssmincidents_replication_set.test]
}`, rName, tagKey1, tagVal1, tagKey2, tagVal2))
}
