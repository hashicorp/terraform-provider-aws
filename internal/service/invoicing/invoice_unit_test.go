// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package invoicing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinvoicing "github.com/hashicorp/terraform-provider-aws/internal/service/invoicing"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInvoicingInvoiceUnit_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccInvoicingInvoiceUnit_basic,
		// acctest.CtDisappears: testAccInvoicingInvoiceUnit_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccInvoicingInvoiceUnit_basic(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.SkipIfEnvVarNotSet(t, "INVOICING_INVOICE_TESTS_ENABLED")

	var invoiceUnit invoicing.GetInvoiceUnitOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_invoicing_invoice_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InvoicingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInvoiceUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceUnitConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "invoice_receiver"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.linked_accounts.#", "1"),
					// resource.TestCheckTypeSetElemAttr(resourceName, "rule.0.linked_accounts.*", linkedAccount),
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
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
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

func testAccCheckInvoiceUnitExists(ctx context.Context, n string, v *invoicing.GetInvoiceUnitOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).InvoicingClient(ctx)

		output, err := tfinvoicing.FindInvoiceUnitByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckInvoiceUnitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InvoicingClient(ctx)

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
