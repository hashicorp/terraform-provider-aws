// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customerprofiles_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/customerprofiles"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCustomerProfilesProfile_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_customerprofiles_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accountNumber := sdkacctest.RandString(8)
	accountNumberUpdated := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()
	email := acctest.RandomEmailAddress(domain)
	emailUpdated := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_full(rName, accountNumber, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "account_number", accountNumber),
					resource.TestCheckResourceAttr(resourceName, "additional_information", "Low Profile Customer"),
					resource.TestCheckResourceAttr(resourceName, "address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_1", "123 Sample Street"),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_2", "Apt 5"),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_3", "null"),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_4", "null"),
					resource.TestCheckResourceAttr(resourceName, "address.0.city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "address.0.country", "USA"),
					resource.TestCheckResourceAttr(resourceName, "address.0.county", "King"),
					resource.TestCheckResourceAttr(resourceName, "address.0.postal_code", "98110"),
					resource.TestCheckResourceAttr(resourceName, "address.0.province", "null"),
					resource.TestCheckResourceAttr(resourceName, "address.0.state", "WA"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "attributes.SSN", "123-44-0000"),
					resource.TestCheckResourceAttr(resourceName, "attributes.LoyaltyPoints", "30000"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_1", "789 Sample St"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_2", "Apt 1"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_3", "null"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_4", "null"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.country", "USA"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.county", "King"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.postal_code", "98011"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.province", "null"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.state", "WA"),
					resource.TestCheckResourceAttr(resourceName, "birth_date", "07/12/1980"),
					resource.TestCheckResourceAttr(resourceName, "business_email_address", email),
					resource.TestCheckResourceAttr(resourceName, "business_name", "My Awesome Company"),
					resource.TestCheckResourceAttr(resourceName, "business_phone_number", "555-334-3389"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "Doe"),
					resource.TestCheckResourceAttr(resourceName, "gender_string", "MALE"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_1", "234 Home St"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_2", "Apt 5"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_3", "null"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_4", "null"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.country", "USA"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.county", "King"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.postal_code", "98011"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.province", "null"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.state", "WA"),
					resource.TestCheckResourceAttr(resourceName, "middle_name", "Ex"),
					resource.TestCheckResourceAttr(resourceName, "mobile_phone_number", "555-334-7777"),
					resource.TestCheckResourceAttr(resourceName, "party_type_string", "INDIVIDUAL"),
					resource.TestCheckResourceAttr(resourceName, "personal_email_address", email),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "555-334-6666"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_1", "555 A St"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_2", "Suite 100"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_3", "null"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_4", "null"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.country", "USA"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.county", "King"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.postal_code", "98011"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.province", "null"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.state", "WA"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccProfileImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccProfileConfig_fullUpdated(rName, accountNumberUpdated, emailUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "additional_information", "High Profile Customer"),
					resource.TestCheckResourceAttr(resourceName, "address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_1", "123 Sample St"),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_2", "Apt 4"),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_3", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.address_4", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.city", "Seattle-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.country", "USA-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.county", "King-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.postal_code", "98011-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.province", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "address.0.state", "WA-updated"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "attributes.SSN", "123-44-3433"),
					resource.TestCheckResourceAttr(resourceName, "attributes.LoyaltyPoints", "3000"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_1", "789 Sample St-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_2", "Apt 1-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_3", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.address_4", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.city", "Seattle-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.country", "USA-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.county", "King-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.postal_code", "98011-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.province", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "billing_address.0.state", "WA-updated"),
					resource.TestCheckResourceAttr(resourceName, "birth_date", "07/12/1980-updated"),
					resource.TestCheckResourceAttr(resourceName, "business_email_address", emailUpdated),
					resource.TestCheckResourceAttr(resourceName, "business_name", "My Awesome Company-updated"),
					resource.TestCheckResourceAttr(resourceName, "business_phone_number", "555-334-3389-updated"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "Doe-updated"),
					resource.TestCheckResourceAttr(resourceName, "gender_string", "MALE-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_1", "234 Home St-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_2", "Apt 5-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_3", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.address_4", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.city", "Seattle-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.country", "USA-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.county", "King-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.postal_code", "98011-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.province", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "mailing_address.0.state", "WA-updated"),
					resource.TestCheckResourceAttr(resourceName, "middle_name", "Ex-updated"),
					resource.TestCheckResourceAttr(resourceName, "mobile_phone_number", "555-334-7777-updated"),
					resource.TestCheckResourceAttr(resourceName, "party_type_string", "INDIVIDUAL-updated"),
					resource.TestCheckResourceAttr(resourceName, "personal_email_address", emailUpdated),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "555-334-6666-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.%", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_1", "555 A St-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_2", "Suite 100-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_3", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.address_4", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.city", "Seattle-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.country", "USA-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.county", "King-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.postal_code", "98011-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.province", "null-updated"),
					resource.TestCheckResourceAttr(resourceName, "shipping_address.0.state", "WA-updated"),
				),
			},
		},
	})
}

func TestAccCustomerProfilesProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_customerprofiles_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accountNumber := sdkacctest.RandString(8)
	domain := acctest.RandomDomainName()
	email := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_full(rName, accountNumber, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, customerprofiles.ResourceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProfileExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CustomerProfilesClient(ctx)

		_, err := customerprofiles.FindProfileByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrDomainName])

		return err
	}
}

func testAccCheckProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CustomerProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_customerprofiles_profile" {
				continue
			}

			_, err := customerprofiles.FindProfileByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Customer Profiles Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccProfileImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrDomainName], rs.Primary.ID), nil
	}
}

func testAccProfileConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_customerprofiles_domain" "test" {
  domain_name             = %[1]q
  default_expiration_days = 7
}
`, rName)
}

func testAccProfileConfig_full(rName, accountNumber, email string) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_customerprofiles_profile" "test" {
  domain_name            = aws_customerprofiles_domain.test.domain_name
  account_number         = %[2]q
  additional_information = "Low Profile Customer"

  address {
    address_1   = "123 Sample Street"
    address_2   = "Apt 5"
    address_3   = "null"
    address_4   = "null"
    city        = "Seattle"
    country     = "USA"
    county      = "King"
    postal_code = "98110"
    province    = "null"
    state       = "WA"
  }

  attributes = {
    SSN           = "123-44-0000"
    LoyaltyPoints = "30000"
  }

  billing_address {
    address_1   = "789 Sample St"
    address_2   = "Apt 1"
    address_3   = "null"
    address_4   = "null"
    city        = "Seattle"
    country     = "USA"
    county      = "King"
    postal_code = "98011"
    province    = "null"
    state       = "WA"
  }

  birth_date             = "07/12/1980"
  business_email_address = %[3]q
  business_name          = "My Awesome Company"
  business_phone_number  = "555-334-3389"
  email_address          = %[3]q
  first_name             = "John"
  gender_string          = "MALE"
  home_phone_number      = "555-334-3344"
  last_name              = "Doe"

  mailing_address {
    address_1   = "234 Home St"
    address_2   = "Apt 5"
    address_3   = "null"
    address_4   = "null"
    city        = "Seattle"
    country     = "USA"
    county      = "King"
    postal_code = "98011"
    province    = "null"
    state       = "WA"
  }

  middle_name            = "Ex"
  mobile_phone_number    = "555-334-7777"
  party_type_string      = "INDIVIDUAL"
  personal_email_address = %[3]q
  phone_number           = "555-334-6666"

  shipping_address {
    address_1   = "555 A St"
    address_2   = "Suite 100"
    address_3   = "null"
    address_4   = "null"
    city        = "Seattle"
    country     = "USA"
    county      = "King"
    postal_code = "98011"
    province    = "null"
    state       = "WA"
  }
}
`, rName, accountNumber, email))
}

func testAccProfileConfig_fullUpdated(rName, accountNumberUpdated, emailUpdated string) string {
	return acctest.ConfigCompose(
		testAccProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_customerprofiles_profile" "test" {
  domain_name            = aws_customerprofiles_domain.test.domain_name
  account_number         = %[2]q
  additional_information = "High Profile Customer"

  address {
    address_1   = "123 Sample St"
    address_2   = "Apt 4"
    address_3   = "null-updated"
    address_4   = "null-updated"
    city        = "Seattle-updated"
    country     = "USA-updated"
    county      = "King-updated"
    postal_code = "98011-updated"
    province    = "null-updated"
    state       = "WA-updated"
  }

  attributes = {
    SSN           = "123-44-3433"
    LoyaltyPoints = "3000"
  }

  billing_address {
    address_1   = "789 Sample St-updated"
    address_2   = "Apt 1-updated"
    address_3   = "null-updated"
    address_4   = "null-updated"
    city        = "Seattle-updated"
    country     = "USA-updated"
    county      = "King-updated"
    postal_code = "98011-updated"
    province    = "null-updated"
    state       = "WA-updated"
  }

  birth_date             = "07/12/1980-updated"
  business_email_address = %[3]q
  business_name          = "My Awesome Company-updated"
  business_phone_number  = "555-334-3389-updated"
  email_address          = %[3]q
  first_name             = "John-updated"
  gender_string          = "MALE-updated"
  home_phone_number      = "555-334-3344-updated"
  last_name              = "Doe-updated"

  mailing_address {
    address_1   = "234 Home St-updated"
    address_2   = "Apt 5-updated"
    address_3   = "null-updated"
    address_4   = "null-updated"
    city        = "Seattle-updated"
    country     = "USA-updated"
    county      = "King-updated"
    postal_code = "98011-updated"
    province    = "null-updated"
    state       = "WA-updated"
  }

  middle_name            = "Ex-updated"
  mobile_phone_number    = "555-334-7777-updated"
  party_type_string      = "INDIVIDUAL-updated"
  personal_email_address = %[3]q
  phone_number           = "555-334-6666-updated"

  shipping_address {
    address_1   = "555 A St-updated"
    address_2   = "Suite 100-updated"
    address_3   = "null-updated"
    address_4   = "null-updated"
    city        = "Seattle-updated"
    country     = "USA-updated"
    county      = "King-updated"
    postal_code = "98011-updated"
    province    = "null-updated"
    state       = "WA-updated"
  }
}
`, rName, accountNumberUpdated, emailUpdated))
}
