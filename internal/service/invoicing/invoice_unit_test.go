// Copyright IBM Corp. 2014, 2025
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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinvoicing "github.com/hashicorp/terraform-provider-aws/internal/service/invoicing"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInvoicingInvoiceUnit_basic(t *testing.T) {
	ctx := acctest.Context(t)

	invoiceReceiver := acctest.SkipIfEnvVarNotSet(t, "INVOICING_INVOICE_RECEIVER_ACCOUNT_ID")
	linkedAccount := acctest.SkipIfEnvVarNotSet(t, "INVOICING_INVOICE_LINKED_ACCOUNT_ID")

	var invoiceUnit invoicing.GetInvoiceUnitOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_invoicing_invoice_unit.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InvoicingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInvoiceUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceUnitConfig_basic(rName, invoiceReceiver, linkedAccount),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "invoice_receiver", invoiceReceiver),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.linked_accounts.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule.0.linked_accounts.*", linkedAccount),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "invoicing", regexache.MustCompile(`invoice-unit/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "tax_inheritance_disabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccInvoiceUnitConfig_description(rName, invoiceReceiver, linkedAccount, "test"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvoiceUnitExists(ctx, resourceName, &invoiceUnit),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "invoice_receiver", invoiceReceiver),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.linked_accounts.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule.0.linked_accounts.*", linkedAccount),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "invoicing", regexache.MustCompile(`invoice-unit/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "tax_inheritance_disabled", acctest.CtFalse),
				),
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

func testAccInvoiceUnitConfig_basic(rName, invoiceReceiver, linkedAccount string) string {
	return fmt.Sprintf(`
resource "aws_invoicing_invoice_unit" "test" {
  name             = %[1]q
  invoice_receiver = %[2]q

  rule {
    linked_accounts = [%[3]q]
  }
}
`, rName, invoiceReceiver, linkedAccount)
}

func testAccInvoiceUnitConfig_description(rName, invoiceReceiver, linkedAccount, description string) string {
	return fmt.Sprintf(`
resource "aws_invoicing_invoice_unit" "test" {
  name             = %[1]q
  invoice_receiver = %[2]q

  rule {
    linked_accounts = [%[3]q]
  }
  description = %[4]q
}
`, rName, invoiceReceiver, linkedAccount, description)
}
