// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package billing_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/billing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/billing/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbilling "github.com/hashicorp/terraform-provider-aws/internal/service/billing"
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbilling.ResourceView, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBillingView_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var view1, view2, view3 awstypes.BillingViewElement
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
				Config: testAccViewConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccViewConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccViewConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccBillingView_dataFilterExpressionTags(t *testing.T) {
	ctx := acctest.Context(t)

	var view1, view2 awstypes.BillingViewElement
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
				Config: testAccViewConfig_dataFilterExpressionTags(rName, "Environment", []string{"production"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view1),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.key", "Environment"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.values.0", "production"),
				),
			},
			{
				Config: testAccViewConfig_dataFilterExpressionTags(rName, "CostCenter", []string{"engineering", "finance"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view2),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.key", "CostCenter"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.values.0", "engineering"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.tags.0.values.1", "finance"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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

func TestAccBillingView_dataFilterExpressionDimensions(t *testing.T) {
	ctx := acctest.Context(t)

	var view1, view2 awstypes.BillingViewElement
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
				Config: testAccViewConfig_dataFilterExpressionDimensions(rName, "LINKED_ACCOUNT", []string{acctest.Ct12Digit}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view1),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.0.key", "LINKED_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.0.values.#", "1"),
				),
			},
			{
				Config: testAccViewConfig_dataFilterExpressionDimensions(rName, "LINKED_ACCOUNT", []string{"111222333444", "999999999912"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view2),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.0.key", "LINKED_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.0.values.#", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccViewConfig_dataFilterExpressionDimensions(rName, "LINKED_ACCOUNT", []string{"999999999912", "111222333444"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckViewExists(ctx, resourceName, &view2),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.0.key", "LINKED_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "data_filter_expression.0.dimensions.0.values.#", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
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

func testAccCheckViewDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BillingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_billing_view" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]
			if arn == "" {
				return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, arn, errors.New("no ARN is set"))
			}

			_, err := tfbilling.FindViewByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Billing, create.ErrActionCheckingDestroyed, tfbilling.ResNameView, arn, err)
			}

			return create.Error(names.Billing, create.ErrActionCheckingDestroyed, tfbilling.ResNameView, arn, errors.New("not destroyed"))
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
			return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, arn, errors.New("no ARN is set"))
		}

		resp, err := tfbilling.FindViewByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.Billing, create.ErrActionCheckingExistence, tfbilling.ResNameView, arn, err)
		}

		*view = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BillingClient(ctx)

	input := billing.ListBillingViewsInput{}

	_, err := conn.ListBillingViews(ctx, &input)

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

func testAccViewConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccViewConfig_base(), fmt.Sprintf(`
resource "aws_billing_view" "test" {
  name         = %[1]q
  description  = "Test description"
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccViewConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccViewConfig_base(), fmt.Sprintf(`
resource "aws_billing_view" "test" {
  name         = %[1]q
  description  = "Test description"
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccViewConfig_dataFilterExpressionTags(rName, tagKey string, tagValues []string) string {
	var tagValuesStr strings.Builder
	for i, v := range tagValues {
		if i > 0 {
			tagValuesStr.WriteString(", ")
		}
		fmt.Fprintf(&tagValuesStr, "%q", v)
	}
	return acctest.ConfigCompose(testAccViewConfig_base(), fmt.Sprintf(`
resource "aws_billing_view" "test" {
  name         = %[1]q
  description  = "Test with data_filter_expression tags"
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]
  data_filter_expression {
    tags {
      key    = %[2]q
      values = [%[3]s]
    }
  }
}
`, rName, tagKey, tagValuesStr.String()))
}

func testAccViewConfig_dataFilterExpressionDimensions(rName, dimensionKey string, dimensionValues []string) string {
	var dimensionValuesStr strings.Builder
	for i, v := range dimensionValues {
		if i > 0 {
			dimensionValuesStr.WriteString(", ")
		}
		fmt.Fprintf(&dimensionValuesStr, "%q", v)
	}
	return acctest.ConfigCompose(testAccViewConfig_base(), fmt.Sprintf(`
resource "aws_billing_view" "test" {
  name         = %[1]q
  description  = "Test with data_filter_expression dimensions"
  source_views = ["arn:${data.aws_partition.current.partition}:billing::${data.aws_caller_identity.current.account_id}:billingview/primary"]
  data_filter_expression {
    dimensions {
      key    = %[2]q
      values = [%[3]s]
    }
  }
}
`, rName, dimensionKey, dimensionValuesStr.String()))
}
