// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package invoicing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinvoicing "github.com/hashicorp/terraform-provider-aws/internal/service/invoicing"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInvoicingInvoiceUnit_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccInvoicingInvoiceUnit_basic,
		// acctest.CtDisappears: testAccInvoicingInvoiceUnit_disappears,
		"regionOverride": testAccInvoicingInvoiceUnit_regionOverride,
		"Identity":       testAccInvoicingInvoiceUnit_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccInvoicingInvoiceUnit_basic(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.SkipIfEnvVarNotSet(t, "INVOICING_INVOICE_TESTS_ENABLED")

	var invoiceUnit invoicing.GetInvoiceUnitOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_invoicing_invoice_unit.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InvoicingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInvoiceUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceUnitConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, t, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "invoicing", regexache.MustCompile(`invoice-unit/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "tax_inheritance_disabled", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invoice_receiver"), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"linked_accounts": knownvalue.SetExact([]knownvalue.Check{
								tfknownvalue.AccountID(),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccInvoiceUnitConfig_description(rName, "test"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, t, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "invoicing", regexache.MustCompile(`invoice-unit/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "tax_inheritance_disabled", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("invoice_receiver"), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"linked_accounts": knownvalue.SetExact([]knownvalue.Check{
								tfknownvalue.AccountID(),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func testAccInvoicingInvoiceUnit_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.SkipIfEnvVarNotSet(t, "INVOICING_INVOICE_TESTS_ENABLED")

	var invoiceUnit invoicing.GetInvoiceUnitOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_invoicing_invoice_unit.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InvoicingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInvoiceUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceUnitConfig_region(rName, acctest.AlternateRegion()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, t, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.CrossRegionAttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				// This test step succeeds because `aws_invoicing_invoice_unit` is global
				// Import assigns the default region when not set
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIgnore:              []string{names.AttrRegion},
			},
			{
				Config: testAccInvoiceUnitConfig_region(rName, acctest.Region()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, t, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
				),
			},
		},
	})
}

func testAccCheckInvoiceUnitExists(ctx context.Context, t *testing.T, n string, v *invoicing.GetInvoiceUnitOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).InvoicingClient(ctx)

		output, err := tfinvoicing.FindInvoiceUnitByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckInvoiceUnitDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).InvoicingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_invoicing_invoice_unit" {
				continue
			}

			_, err := tfinvoicing.FindInvoiceUnitByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Invoice Unit %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccInvoiceUnitConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_invoicing_invoice_unit" "test" {
  name             = %[1]q
  invoice_receiver = data.aws_caller_identity.current.account_id

  rule {
    linked_accounts = [data.aws_caller_identity.current.account_id]
  }
}

data "aws_caller_identity" "current" {}
`, rName)
}

func testAccInvoiceUnitConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_invoicing_invoice_unit" "test" {
  name             = %[1]q
  invoice_receiver = data.aws_caller_identity.current.account_id

  rule {
    linked_accounts = [data.aws_caller_identity.current.account_id]
  }
  description = %[2]q
}

data "aws_caller_identity" "current" {}
`, rName, description)
}

func testAccInvoiceUnitConfig_region(rName, region string) string {
	return fmt.Sprintf(`
resource "aws_invoicing_invoice_unit" "test" {
  name             = %[1]q
  region           = %[2]q
  invoice_receiver = data.aws_caller_identity.current.account_id

  rule {
    linked_accounts = [data.aws_caller_identity.current.account_id]
  }
}

data "aws_caller_identity" "current" {}
`, rName, region)
}
