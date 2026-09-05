// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package taxsettings_test

import (
	"context"
	"errors"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/taxsettings/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftaxsettings "github.com/hashicorp/terraform-provider-aws/internal/service/taxsettings"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Tax registration is a singleton resource per AWS account — there is no random
// name, and create/update both call PutTaxRegistration (upsert). Tests use a
// fixed but clearly synthetic GB VAT number so the registration lands in
// "Pending" status and does not affect real billing.

func TestAccTaxSettingsTaxRegistration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var taxregistration awstypes.TaxRegistration
	resourceName := "aws_taxsettings_tax_registration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TaxSettingsServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TaxSettingsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaxRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTaxRegistrationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTaxRegistrationExists(ctx, t, resourceName, &taxregistration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "registration_id", "GB123456789"),
					resource.TestCheckResourceAttr(resourceName, "registration_type", "VAT"),
					resource.TestCheckResourceAttr(resourceName, "legal_name", "Test Company Ltd"),
					resource.TestCheckResourceAttr(resourceName, "legal_address.0.address_line1", "123 Test Street"),
					resource.TestCheckResourceAttr(resourceName, "legal_address.0.city", "London"),
					resource.TestCheckResourceAttr(resourceName, "legal_address.0.country_code", "GB"),
					resource.TestCheckResourceAttr(resourceName, "legal_address.0.postal_code", "EC1A 1BB"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
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

func TestAccTaxSettingsTaxRegistration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var taxregistration awstypes.TaxRegistration
	resourceName := "aws_taxsettings_tax_registration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TaxSettingsServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TaxSettingsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaxRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTaxRegistrationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaxRegistrationExists(ctx, t, resourceName, &taxregistration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tftaxsettings.ResourceTaxRegistration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTaxSettingsTaxRegistration_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var taxregistration awstypes.TaxRegistration
	resourceName := "aws_taxsettings_tax_registration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TaxSettingsServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TaxSettingsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaxRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTaxRegistrationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTaxRegistrationExists(ctx, t, resourceName, &taxregistration),
					resource.TestCheckResourceAttr(resourceName, "registration_id", "GB123456789"),
					resource.TestCheckResourceAttr(resourceName, "legal_name", "Test Company Ltd"),
				),
			},
			{
				Config: testAccTaxRegistrationConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTaxRegistrationExists(ctx, t, resourceName, &taxregistration),
					resource.TestCheckResourceAttr(resourceName, "registration_id", "GB987654321"),
					resource.TestCheckResourceAttr(resourceName, "legal_name", "Updated Company Ltd"),
				),
			},
		},
	})
}

func testAccCheckTaxRegistrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TaxSettingsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_taxsettings_tax_registration" {
				continue
			}

			_, err := tftaxsettings.FindTaxRegistrationByAccountID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.TaxSettings, create.ErrActionCheckingDestroyed, tftaxsettings.ResNameTaxRegistration, rs.Primary.ID, err)
			}

			return create.Error(names.TaxSettings, create.ErrActionCheckingDestroyed, tftaxsettings.ResNameTaxRegistration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTaxRegistrationExists(ctx context.Context, t *testing.T, name string, taxregistration *awstypes.TaxRegistration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxRegistration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxRegistration, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).TaxSettingsClient(ctx)
		resp, err := tftaxsettings.FindTaxRegistrationByAccountID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.TaxSettings, create.ErrActionCheckingExistence, tftaxsettings.ResNameTaxRegistration, rs.Primary.ID, err)
		}

		*taxregistration = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).TaxSettingsClient(ctx)

	_, err := conn.GetTaxRegistration(ctx, nil)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	// ResourceNotFoundException is expected for accounts without a registration — that's fine.
	if err != nil && !isNoTaxRegistrationError(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// isNoTaxRegistrationError returns true for the expected "no registration found" error
// so testAccPreCheck can distinguish API access errors from missing registrations.
func isNoTaxRegistrationError(err error) bool {
	var rnf *awstypes.ResourceNotFoundException
	return errors.As(err, &rnf)
}

func testAccTaxRegistrationConfig_basic() string {
	return `
resource "aws_taxsettings_tax_registration" "test" {
  registration_id   = "GB123456789"
  registration_type = "VAT"
  legal_name        = "Test Company Ltd"

  legal_address {
    address_line1 = "123 Test Street"
    city          = "London"
    country_code  = "GB"
    postal_code   = "EC1A 1BB"
  }
}
`
}

func testAccTaxRegistrationConfig_updated() string {
	return `
resource "aws_taxsettings_tax_registration" "test" {
  registration_id   = "GB987654321"
  registration_type = "VAT"
  legal_name        = "Updated Company Ltd"

  legal_address {
    address_line1 = "456 New Street"
    city          = "Manchester"
    country_code  = "GB"
    postal_code   = "M1 1AE"
  }
}
`
}
