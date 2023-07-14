package ssmcontacts_test

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
	"regexp"
	"testing"
	"time"
)

func TestAccSSMContactsRotationDataSource_basic(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testRotationDataSourceConfig_basic(rName, startTime),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.number_of_on_calls", dataSourceName, "recurrence.0.number_of_on_calls"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.recurrence_multiplier", dataSourceName, "recurrence.0.recurrence_multiplier"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.#", dataSourceName, "recurrence.0.weekly_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.0.day_of_week", dataSourceName, "recurrence.0.weekly_settings.0.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.0.hand_off_time", dataSourceName, "recurrence.0.weekly_settings.0.hand_off_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.1.day_of_week", dataSourceName, "recurrence.0.weekly_settings.1.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.1.hand_off_time", dataSourceName, "recurrence.0.weekly_settings.1.hand_off_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.2.day_of_week", dataSourceName, "recurrence.0.weekly_settings.2.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.weekly_settings.2.hand_off_time", dataSourceName, "recurrence.0.weekly_settings.2.hand_off_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.#", dataSourceName, "recurrence.0.shift_coverages.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.0.day_of_week", dataSourceName, "recurrence.0.shift_coverages.0.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start_time", dataSourceName, "recurrence.0.shift_coverages.0.coverage_times.0.start_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end_time", dataSourceName, "recurrence.0.shift_coverages.0.coverage_times.0.end_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.1.day_of_week", dataSourceName, "recurrence.0.shift_coverages.1.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start_time", dataSourceName, "recurrence.0.shift_coverages.1.coverage_times.0.start_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end_time", dataSourceName, "recurrence.0.shift_coverages.1.coverage_times.0.end_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.2.day_of_week", dataSourceName, "recurrence.0.shift_coverages.2.day_of_week"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start_time", dataSourceName, "recurrence.0.shift_coverages.2.coverage_times.0.start_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end_time", dataSourceName, "recurrence.0.shift_coverages.2.coverage_times.0.end_time"),
					resource.TestCheckResourceAttrPair(resourceName, "start_time", dataSourceName, "start_time"),
					resource.TestCheckResourceAttrPair(resourceName, "time_zone_id", dataSourceName, "time_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_ids.#", dataSourceName, "contact_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_ids.0", dataSourceName, "contact_ids.0"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_ids.1", dataSourceName, "contact_ids.1"),
					resource.TestCheckResourceAttrPair(resourceName, "contact_ids.2", dataSourceName, "contact_ids.2"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.key1", dataSourceName, "tags.key1"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.key2", dataSourceName, "tags.key2"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssm-contacts", regexp.MustCompile(`rotation\/+.`)),
				),
			},
		},
	})
}

func TestAccSSMContactsRotationDataSource_dailySettings(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testRotationDataSourceConfig_dailySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.daily_settings.#", dataSourceName, "recurrence.0.daily_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.daily_settings.0", dataSourceName, "recurrence.0.daily_settings.0"),
				),
			},
		},
	})
}

func TestAccSSMContactsRotationDataSource_monthlySettings(t *testing.T) {
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testRotationDataSourceConfig_monthlySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.#", dataSourceName, "recurrence.0.monthly_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.0.day_of_month", dataSourceName, "recurrence.0.monthly_settings.0.day_of_month"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.0.hand_off_time", dataSourceName, "recurrence.0.monthly_settings.0.hand_off_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.1.day_of_month", dataSourceName, "recurrence.0.monthly_settings.1.day_of_month"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.1.hand_off_time", dataSourceName, "recurrence.0.monthly_settings.1.hand_off_time"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.2.day_of_month", dataSourceName, "recurrence.0.monthly_settings.2.day_of_month"),
					resource.TestCheckResourceAttrPair(resourceName, "recurrence.0.monthly_settings.2.hand_off_time", dataSourceName, "recurrence.0.monthly_settings.2.hand_off_time"),
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

resource "aws_ssmcontacts_contact" "test_contact_one" {
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
resource "aws_ssmcontacts_contact" "test_contact_two" {
  alias = "test-contact-two-for-%[1]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test_contact_three" {
  alias = "test-contact-three-for-%[1]s"
  type  = "PERSONAL"
  
  depends_on = [aws_ssmincidents_replication_set.test]
}

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
	weekly_settings {
		day_of_week = "MON"
		hand_off_time = "04:25"
	}
	weekly_settings {
		day_of_week = "WED"
		hand_off_time = "07:34"
	}
	weekly_settings {
		day_of_week = "FRI"
		hand_off_time = "15:57"
	}
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
	aws_ssmcontacts_contact.test_contact_one.arn,
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
	aws_ssmcontacts_contact.test_contact_one.arn,
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
	monthly_settings {
		day_of_month = 1
		hand_off_time = "04:58"
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
