// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package invoicing_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/invoicing"
	"github.com/aws/aws-sdk-go-v2/service/invoicing/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInvoceUnit_basic(t *testing.T) {
	ctx := acctest.Context(t)

	invoiceReceiver := os.Getenv("INVOICING_INVOICE_RECEIVER_ACCOUNT_ID")
	if invoiceReceiver == "" {
		t.Skip("Environment variable INVOICING_INVOICE_RECEIVER_ACCOUNT_ID is not set. ")
	}

	linkedAccount := os.Getenv("INVOICING_INVOICE_LINKED_ACCOUNT_ID")
	if linkedAccount == "" {
		t.Skip("Environment variable INVOICING_INVOICE_LINKED_ACCOUNT_ID is not set. ")
	}

	resourceName := "aws_invoicing_invoice_unit.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InvoicingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInvoiceUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceUnitConfig_base(rName, rDescription1, invoiceReceiver, linkedAccount),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription1),
					resource.TestCheckResourceAttr(resourceName, "invoice_receiver", invoiceReceiver),
					resource.TestCheckResourceAttr(resourceName, "tax_inheritance_disabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInvoiceUnitConfig_base(rName, rDescription2, invoiceReceiver, linkedAccount),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription2),
					resource.TestCheckResourceAttr(resourceName, "invoice_receiver", invoiceReceiver),
					resource.TestCheckResourceAttr(resourceName, "tax_inheritance_disabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccCheckInvoiceUnitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InvoicingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_invoicing_invoice_unit" {
				continue
			}

			input := invoicing.GetInvoiceUnitInput{
				InvoiceUnitArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetInvoiceUnit(ctx, &input)

			if errs.IsA[*types.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading Invoice Unit (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Invoice Unit (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccInvoiceUnitConfig_base(rName string, rDescription string, invoiceReceiver string, linkedAccount string) string {
	return fmt.Sprintf(`
resource "aws_invoicing_invoice_unit" "test" {
  name = %[1]q
  description = %[2]q
  invoice_receiver = %[3]q
  tax_inheritance_disabled = false
  linked_accounts = [
   %[4]q
  ]
}
`, rName, rDescription, invoiceReceiver, linkedAccount)
}
