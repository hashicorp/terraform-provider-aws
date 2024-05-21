// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInsight_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					testAccCheckInsightARN(resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.aws_account_id.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison":    string(types.StringFilterComparisonEquals),
						names.AttrValue: "1234567890",
					}),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "AwsAccountId"),
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

func testAccInsight_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceInsight(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccInsight_DateFilters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	endDate := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	startDate := time.Now().Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_dateFiltersDateRange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.created_at.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.created_at.*", map[string]string{
						"date_range.#":       acctest.Ct1,
						"date_range.0.unit":  string(types.DateRangeUnitDays),
						"date_range.0.value": "5",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInsightConfig_dateFiltersStartEnd(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.created_at.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.created_at.*", map[string]string{
						"start": startDate,
						"end":   endDate,
					}),
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

func testAccInsight_IPFilters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_ipFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.network_destination_ipv4.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.network_destination_ipv4.*", map[string]string{
						"cidr": "10.0.0.0/16",
					}),
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

func testAccInsight_KeywordFilters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_keywordFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.keyword.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.keyword.*", map[string]string{
						names.AttrValue: rName,
					}),
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

func testAccInsight_MapFilters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_mapFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.product_fields.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.product_fields.*", map[string]string{
						"comparison":    string(types.MapFilterComparisonEquals),
						names.AttrKey:   acctest.CtKey1,
						names.AttrValue: acctest.CtValue1,
					}),
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

func testAccInsight_MultipleFilters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_multipleFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.aws_account_id.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison":    string(types.StringFilterComparisonEquals),
						names.AttrValue: "1234567890",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison":    string(types.StringFilterComparisonEquals),
						names.AttrValue: "09876543210",
					}),
					resource.TestCheckResourceAttr(resourceName, "filters.0.product_fields.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.product_fields.*", map[string]string{
						"comparison":    string(types.MapFilterComparisonEquals),
						names.AttrKey:   acctest.CtKey1,
						names.AttrValue: acctest.CtValue1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.product_fields.*", map[string]string{
						"comparison":    string(types.MapFilterComparisonEquals),
						names.AttrKey:   acctest.CtKey2,
						names.AttrValue: acctest.CtValue2,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInsightConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.aws_account_id.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison":    string(types.StringFilterComparisonEquals),
						names.AttrValue: "1234567890",
					}),
				),
			},
		},
	})
}

func testAccInsight_Name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")

	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					testAccCheckInsightARN(resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccInsightConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					testAccCheckInsightARN(resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
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

func testAccInsight_NumberFilters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_numberFilters(rName, "eq = 50.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.confidence.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.confidence.*", map[string]string{
						"eq": "50.5",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInsightConfig_numberFilters(rName, "gte = 50.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.confidence.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.confidence.*", map[string]string{
						"gte": "50.5",
					}),
				),
			},
			{
				Config: testAccInsightConfig_numberFilters(rName, "lte = 50.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.confidence.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.confidence.*", map[string]string{
						"lte": "50.5",
					}),
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

func testAccInsight_GroupByAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "AwsAccountId"),
				),
			},
			{
				Config: testAccInsightConfig_updateGroupByAttribute(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "CompanyName"),
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

func testAccInsight_WorkflowStatus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInsightDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInsightConfig_workflowStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInsightExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filters.0.workflow_status.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.workflow_status.*", map[string]string{
						"comparison":    string(types.StringFilterComparisonEquals),
						names.AttrValue: string(types.WorkflowStatusNew),
					}),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "WorkflowStatus"),
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

func testAccCheckInsightDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_insight" {
				continue
			}

			_, err := tfsecurityhub.FindInsightByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Insight (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInsightExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindInsightByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

// testAccCheckInsightARN checks the computed ARN value
// and accounts for differences in SecurityHub on GovCloud where the partition portion
// of the ARN is still "aws" while other services utilize the "aws-us-gov" partition
func testAccCheckInsightARN(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedArn := fmt.Sprintf(`^arn:aws[^:]*:securityhub:%s:%s:insight/%s/custom/.+$`, acctest.Region(), acctest.AccountID(), acctest.AccountID())
		//lintignore:AWSAT001
		return resource.TestMatchResourceAttr(resourceName, names.AttrARN, regexache.MustCompile(expectedArn))(s)
	}
}

func testAccInsightConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_dateFiltersDateRange(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    created_at {
      date_range {
        unit  = "DAYS"
        value = 5
      }
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_dateFiltersStartEnd(rName, startDate, endDate string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    created_at {
      start = %q
      end   = %q
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, startDate, endDate, rName)
}

func testAccInsightConfig_ipFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    network_destination_ipv4 {
      cidr = "10.0.0.0/16"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_keywordFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    keyword {
      value = %[1]q
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %[1]q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_mapFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    product_fields {
      comparison = "EQUALS"
      key        = "key1"
      value      = "value1"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %[1]q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_multipleFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }

    aws_account_id {
      comparison = "EQUALS"
      value      = "09876543210"
    }

    product_fields {
      comparison = "EQUALS"
      key        = "key1"
      value      = "value1"
    }

    product_fields {
      comparison = "EQUALS"
      key        = "key2"
      value      = "value2"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %[1]q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_numberFilters(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    confidence {
      %s
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, value, rName)
}

func testAccInsightConfig_updateGroupByAttribute(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "CompanyName"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccInsightConfig_workflowStatus(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    workflow_status {
      comparison = "EQUALS"
      value      = "NEW"
    }
  }

  group_by_attribute = "WorkflowStatus"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}
