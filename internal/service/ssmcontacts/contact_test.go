// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmcontacts "github.com/hashicorp/terraform-provider-aws/internal/service/ssmcontacts"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccContact_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_contact.contact_one"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "PERSONAL"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile(`contact/.+$`)),
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
				Config: testAccContactConfig_none(),
				Check:  testAccCheckContactDestroy(ctx),
			},
		},
	})
}

func testAccContact_updateAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	oldAlias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	newAlias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_ssmcontacts_contact.contact_one"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_alias(oldAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, oldAlias),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_alias(newAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, newAlias),
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

func testAccContact_updateType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	personalType := "PERSONAL"
	escalationType := "ESCALATION"

	resourceName := "aws_ssmcontacts_contact.contact_one"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_type(name, personalType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, personalType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_type(name, escalationType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, escalationType),
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

func testAccContact_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_contact.contact_one"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssmcontacts.ResourceContact(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContact_updateDisplayName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	oldDisplayName := sdkacctest.RandString(26)
	newDisplayName := sdkacctest.RandString(26)
	resourceName := "aws_ssmcontacts_contact.contact_one"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_displayName(rName, oldDisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, oldDisplayName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_displayName(rName, newDisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, newDisplayName),
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

func testAccContact_oncallSchedule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_contact.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactConfig_oncallSchedule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ONCALL_SCHEDULE"),
					resource.TestCheckResourceAttr(resourceName, "rotation_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile(`contact/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContactConfig_oncallScheduleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ONCALL_SCHEDULE"),
					resource.TestCheckResourceAttr(resourceName, "rotation_ids.#", "2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm-contacts", regexache.MustCompile(`contact/.+$`)),
				),
			},
			{
				Config: testAccContactConfig_none(),
				Check:  testAccCheckContactDestroy(ctx),
			},
		},
	})
}


func testAccCheckContactDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmcontacts_contact" {
				continue
			}

			input := &ssmcontacts.GetContactInput{
				ContactId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetContact(ctx, input)

			if err != nil {
				// Getting resources may return validation exception when the replication set has been destroyed
				var ve *types.ValidationException
				if errors.As(err, &ve) {
					continue
				}

				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					continue
				}

				return err
			}

			return create.Error(names.SSMContacts, create.ErrActionCheckingDestroyed, tfssmcontacts.ResNameContact, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckContactExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContact, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContact, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

		_, err := conn.GetContact(ctx, &ssmcontacts.GetContactInput{
			ContactId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.SSMContacts, create.ErrActionCheckingExistence, tfssmcontacts.ResNameContact, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContactPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMContactsClient(ctx)

	input := &ssmcontacts.ListContactsInput{}
	_, err := conn.ListContacts(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccContactConfig_base() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}

func testAccContactConfig_basic(alias string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias = %[1]q
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}

func testAccContactConfig_none() string {
	return testAccContactConfig_base()
}

func testAccContactConfig_alias(alias string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias = %[1]q
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}

func testAccContactConfig_type(name, typeValue string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias = %[1]q
  type  = %[2]q

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, name, typeValue))
}

func testAccContactConfig_displayName(alias, displayName string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias        = %[1]q
  display_name = %[2]q
  type         = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias, displayName))
}

func testAccContactConfig_oncallSchedule(alias string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test_contact" {
  alias = "%[1]s-contact"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [aws_ssmcontacts_contact.test_contact.arn]
  name        = %[1]q
  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 9
      minute_of_hour = 0
    }
  }
  time_zone_id = "America/Los_Angeles"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test" {
  alias        = %[1]q
  display_name = %[1]q
  type         = "ONCALL_SCHEDULE"
  rotation_ids = [aws_ssmcontacts_rotation.test.arn]

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}

func testAccContactConfig_oncallScheduleUpdated(alias string) string {
	return acctest.ConfigCompose(
		testAccContactConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "test_contact" {
  alias = "%[1]s-contact"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test_contact_2" {
  alias = "%[1]s-contact-2"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_rotation" "test" {
  contact_ids = [aws_ssmcontacts_contact.test_contact.arn]
  name        = %[1]q
  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 9
      minute_of_hour = 0
    }
  }
  time_zone_id = "America/Los_Angeles"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_rotation" "test_2" {
  contact_ids = [aws_ssmcontacts_contact.test_contact_2.arn]
  name        = "%[1]s-2"
  recurrence {
    number_of_on_calls    = 1
    recurrence_multiplier = 1
    daily_settings {
      hour_of_day    = 14
      minute_of_hour = 30
    }
  }
  time_zone_id = "America/New_York"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact" "test" {
  alias        = %[1]q
  display_name = %[1]q
  type         = "ONCALL_SCHEDULE"
  rotation_ids = [
    aws_ssmcontacts_rotation.test.arn,
    aws_ssmcontacts_rotation.test_2.arn
  ]

  depends_on = [aws_ssmincidents_replication_set.test]
}
`, alias))
}
