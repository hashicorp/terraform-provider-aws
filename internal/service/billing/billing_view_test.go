// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package billing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test description"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					// TIP: If the ARN can be partially or completely determined by the parameters passed, e.g. it contains the
					// value of `rName`, either include the values in the regex or check for an exact match using `acctest.CheckResourceAttrRegionalARN`
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "billing", regexache.MustCompile(`view:.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
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

			_, err := tfbilling.FindViewByARN(ctx, conn, &rs.Primary.ID)
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

		if rs.Primary.ID == "" {
			return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BillingClient(ctx)

		resp, err := tfbilling.FindViewByARN(ctx, conn, &rs.Primary.ID)
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

func testAccViewConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_billing_view" "test" {
  name         = "%s"
  description  = "Test description"
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]
}
`, rName)
}
