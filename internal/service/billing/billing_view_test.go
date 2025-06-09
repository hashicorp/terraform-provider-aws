// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package billing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/billing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/billing/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbilling "github.com/hashicorp/terraform-provider-aws/internal/service/billing"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBillingView_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var view awstypes.BillingViewElement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_billing_view.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BillingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test description"),
					acctest.CheckResourceAttrContains(resourceName, names.AttrARN, "billingview/custom-"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccBillingView_update(t *testing.T) {
	ctx := acctest.Context(t)

	var view awstypes.BillingViewElement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_billing_view.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BillingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_update(rName, "Test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccViewConfig_update(fmt.Sprintf("%s-updated", rName), "Test description updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-updated", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test description updated"),
					acctest.CheckResourceAttrContains(resourceName, names.AttrARN, "billingview/custom-"),
				),
			},
		},
	})
}

func TestAccBillingView_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var view awstypes.BillingViewElement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_billing_view.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BillingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckViewDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccViewConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbilling.ResourceView, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckViewDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BillingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_billing_view" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]
			if arn == "" {
				return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, rs.Primary.ID, errors.New("no ARN is set"))
			}

			_, err := tfbilling.FindViewByARN(ctx, conn, arn)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Billing, create.ErrActionCheckingDestroyed, tfbilling.ResNameView, rs.Primary.ID, err)
			}

			return create.Error(names.Billing, create.ErrActionCheckingDestroyed, tfbilling.ResNameView, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckViewExists(ctx context.Context, name string, view *awstypes.BillingViewElement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BillingClient(ctx)

		arn := rs.Primary.Attributes[names.AttrARN]
		if arn == "" {
			return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, rs.Primary.ID, errors.New("no ARN is set"))
		}

		resp, err := tfbilling.FindViewByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, rs.Primary.ID, err)
		}

		*view = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BillingClient(ctx)

	input := &billing.ListBillingViewsInput{}

	_, err := conn.ListBillingViews(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccViewConfig_base() string {
	return `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
`
}

func testAccViewConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccViewConfig_base(), fmt.Sprintf(`
resource "aws_billing_view" "test" {
  name         = "%s"
  description  = "Test description"
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]
}
`, rName))
}

func testAccViewConfig_update(rName, description string) string {
	return acctest.ConfigCompose(testAccViewConfig_base(), fmt.Sprintf(`
resource "aws_billing_view" "test" {
  name         = %[1]q
  description  = %[2]q
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]
}
`, rName, description))
}
