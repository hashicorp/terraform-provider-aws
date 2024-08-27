// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbcmdataexports "github.com/hashicorp/terraform-provider-aws/internal/service/bcmdataexports"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBCMDataExportsExport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var export bcmdataexports.GetExportOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "export.0.data_query.0.query_statement"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.overwrite", "OVERWRITE_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.format", "TEXT_OR_CSV"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.output_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.0.frequency", "SYNCHRONOUS"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "HOURLY"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_RESOURCES", acctest.CtFalseCaps),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY", acctest.CtFalseCaps),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_SPLIT_COST_ALLOCATION_DATA", acctest.CtFalseCaps),
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

func TestAccBCMDataExportsExport_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var export bcmdataexports.GetExportOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_update(rName, "OVERWRITE_REPORT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "export.0.data_query.0.query_statement"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.overwrite", "OVERWRITE_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.format", "TEXT_OR_CSV"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.output_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.0.frequency", "SYNCHRONOUS"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "HOURLY"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_RESOURCES", acctest.CtFalseCaps),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY", acctest.CtFalseCaps),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_SPLIT_COST_ALLOCATION_DATA", acctest.CtFalseCaps),
				),
			},
			{
				Config: testAccExportConfig_update(rName, "CREATE_NEW_REPORT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "export.0.data_query.0.query_statement"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.overwrite", "CREATE_NEW_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.format", "TEXT_OR_CSV"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.0.s3_output_configurations.0.output_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.0.frequency", "SYNCHRONOUS"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "HOURLY"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_RESOURCES", acctest.CtFalseCaps),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY", acctest.CtFalseCaps),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.INCLUDE_SPLIT_COST_ALLOCATION_DATA", acctest.CtFalseCaps),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/37126
func TestAccBCMDataExportsExport_curSubset(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var export bcmdataexports.GetExportOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	columns := []string{
		"bill_bill_type",
		"bill_billing_entity",
		"bill_billing_period_end_date",
		"bill_billing_period_start_date",
		"bill_invoice_id",
		"bill_invoicing_entity",
		"bill_payer_account_id",
		"bill_payer_account_name",
		"cost_category",
		"discount",
		"discount_bundled_discount",
		"discount_total_discount",
		"identity_line_item_id",
		"identity_time_interval",
		"line_item_availability_zone",
		"line_item_blended_cost",
		"line_item_blended_rate",
		"line_item_currency_code",
		"line_item_legal_entity",
		"line_item_line_item_description",
		"line_item_line_item_type",
		"line_item_net_unblended_cost",
		"line_item_net_unblended_rate",
		"line_item_normalization_factor",
		"line_item_normalized_usage_amount",
		"line_item_operation",
		"line_item_product_code",
		"line_item_resource_id",
		"line_item_tax_type",
		"line_item_unblended_cost",
		"line_item_unblended_rate",
		"line_item_usage_account_id",
		"line_item_usage_account_name",
		"line_item_usage_amount",
		"line_item_usage_end_date",
		"line_item_usage_start_date",
		"line_item_usage_type",
		"pricing_currency",
		"pricing_lease_contract_length",
		"pricing_offering_class",
		"pricing_public_on_demand_cost",
		"pricing_public_on_demand_rate",
		"pricing_purchase_option",
		"pricing_rate_code",
		"pricing_rate_id",
		"pricing_term",
		"pricing_unit",
		"product",
		"product_comment",
		"product_fee_code",
		"product_fee_description",
		"product_from_location",
		"product_from_location_type",
		"product_from_region_code",
		"product_instance_family",
		"product_instance_type",
		"product_instancesku",
		"product_location",
		"product_location_type",
		"product_operation",
		"product_pricing_unit",
		"product_product_family",
		"product_region_code",
		"product_servicecode",
		"product_sku",
		"product_to_location",
		"product_to_location_type",
		"product_to_region_code",
		"product_usagetype",
		"reservation_amortized_upfront_cost_for_usage",
		"reservation_amortized_upfront_fee_for_billing_period",
		"reservation_availability_zone",
		"reservation_effective_cost",
		"reservation_end_time",
		"reservation_modification_status",
		"reservation_net_amortized_upfront_cost_for_usage",
		"reservation_net_amortized_upfront_fee_for_billing_period",
		"reservation_net_effective_cost",
		"reservation_net_recurring_fee_for_usage",
		"reservation_net_unused_amortized_upfront_fee_for_billing_period",
		"reservation_net_unused_recurring_fee",
		"reservation_net_upfront_value",
		"reservation_normalized_units_per_reservation",
		"reservation_number_of_reservations",
		"reservation_recurring_fee_for_usage",
		"reservation_reservation_a_r_n",
		"reservation_start_time",
		"reservation_subscription_id",
		"reservation_total_reserved_normalized_units",
		"reservation_total_reserved_units",
		"reservation_units_per_reservation",
		"reservation_unused_amortized_upfront_fee_for_billing_period",
		"reservation_unused_normalized_unit_quantity",
		"reservation_unused_quantity",
		"reservation_unused_recurring_fee",
		"reservation_upfront_value",
		names.AttrResourceTags,
		"savings_plan_amortized_upfront_commitment_for_billing_period",
		"savings_plan_end_time",
		"savings_plan_instance_type_family",
		"savings_plan_net_amortized_upfront_commitment_for_billing_period",
		"savings_plan_net_recurring_commitment_for_billing_period",
		"savings_plan_net_savings_plan_effective_cost",
		"savings_plan_offering_type",
		"savings_plan_payment_option",
		"savings_plan_purchase_term",
		"savings_plan_recurring_commitment_for_billing_period",
		"savings_plan_region",
		"savings_plan_savings_plan_a_r_n",
		"savings_plan_savings_plan_effective_cost",
		"savings_plan_savings_plan_rate",
		"savings_plan_start_time",
		"savings_plan_total_commitment_to_date",
		"savings_plan_used_commitment",
	}
	query := fmt.Sprintf("SELECT %s FROM COST_AND_USAGE_REPORT", strings.Join(columns, ", "))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_curSubset(rName, query),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "export.0.data_query.0.query_statement"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.0.frequency", "SYNCHRONOUS"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "HOURLY"),
				),
			},
		},
	})
}

func TestAccBCMDataExportsExport_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var export bcmdataexports.GetExportOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbcmdataexports.ResourceExport, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBCMDataExportsExport_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var export bcmdataexports.GetExportOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExportConfig_tag2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccExportConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccBCMDataExportsExport_updateTable(t *testing.T) {
	ctx := acctest.Context(t)
	var export bcmdataexports.GetExportOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bcmdataexports_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_updateTableConfigs(rName, "SELECT identity_line_item_id, identity_time_interval, line_item_usage_amount, line_item_unblended_cost FROM COST_AND_USAGE_REPORT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.query_statement", "SELECT identity_line_item_id, identity_time_interval, line_item_usage_amount, line_item_unblended_cost FROM COST_AND_USAGE_REPORT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExportConfig_updateTableConfigs(rName, "SELECT identity_line_item_id, identity_time_interval, line_item_usage_amount, line_item_unblended_cost, cost_category FROM COST_AND_USAGE_REPORT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.query_statement", "SELECT identity_line_item_id, identity_time_interval, line_item_usage_amount, line_item_unblended_cost, cost_category FROM COST_AND_USAGE_REPORT"),
				),
			},
		},
	})
}

func testAccCheckExportDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BCMDataExportsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bcmdataexports_export" {
				continue
			}

			_, err := tfbcmdataexports.FindExportByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.BCMDataExports, create.ErrActionCheckingDestroyed, tfbcmdataexports.ResNameExport, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckExportExists(ctx context.Context, name string, export *bcmdataexports.GetExportOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingExistence, tfbcmdataexports.ResNameExport, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingExistence, tfbcmdataexports.ResNameExport, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BCMDataExportsClient(ctx)
		resp, err := tfbcmdataexports.FindExportByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.BCMDataExports, create.ErrActionCheckingExistence, tfbcmdataexports.ResNameExport, rs.Primary.ID, err)
		}

		*export = *resp

		return nil
	}
}

func testAccExportConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:PutObject",
        "s3:GetBucketPolicy"
      ]
      Effect = "Allow"
      Sid    = "EnableAWSDataExportsToWriteToS3AndCheckPolicy"
      Principal = {
        Service = [
          "billingreports.amazonaws.com",
          "bcm-data-exports.amazonaws.com"
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccExportConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccExportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_bcmdataexports_export" "test" {
  export {
    name = %[1]q
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "FALSE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = "OVERWRITE_REPORT"
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }

    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
}
`, rName))
}

func testAccExportConfig_update(rName, overwrite string) string {
	return acctest.ConfigCompose(
		testAccExportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_bcmdataexports_export" "test" {
  export {
    name = %[1]q
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "FALSE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = %[2]q
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }

    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
}
`, rName, overwrite))
}

func testAccExportConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccExportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_bcmdataexports_export" "test" {
  export {
    name = %[1]q
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "FALSE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = "OVERWRITE_REPORT"
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }
    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccExportConfig_tag2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccExportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_bcmdataexports_export" "test" {
  export {
    name = %[1]q
    data_query {
      query_statement = "SELECT identity_line_item_id, identity_time_interval, line_item_product_code,line_item_unblended_cost FROM COST_AND_USAGE_REPORT"
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "FALSE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = "OVERWRITE_REPORT"
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }
    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccExportConfig_updateTableConfigs(rName, queryStatement string) string {
	return acctest.ConfigCompose(
		testAccExportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_bcmdataexports_export" "test" {
  export {
    name = %[1]q
    data_query {
      query_statement = %[2]q
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "FALSE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }
    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          overwrite   = "OVERWRITE_REPORT"
          format      = "TEXT_OR_CSV"
          compression = "GZIP"
          output_type = "CUSTOM"
        }
      }
    }

    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
}
`, rName, queryStatement))
}

func testAccExportConfig_curSubset(rName, query string) string {
	return acctest.ConfigCompose(
		testAccExportConfigBase(rName),
		fmt.Sprintf(`
resource "aws_bcmdataexports_export" "test" {
  export {
    name = %[1]q
    data_query {
      query_statement = %[2]q
      table_configurations = {
        "COST_AND_USAGE_REPORT" = {
          "TIME_GRANULARITY"                      = "HOURLY",
          "INCLUDE_RESOURCES"                     = "TRUE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY" = "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA"    = "FALSE",
        }
      }
    }

    destination_configurations {
      s3_destination {
        s3_bucket = aws_s3_bucket.test.bucket
        s3_prefix = aws_s3_bucket.test.bucket_prefix
        s3_region = aws_s3_bucket.test.region
        s3_output_configurations {
          compression = "PARQUET"
          format      = "PARQUET"
          output_type = "CUSTOM"
          overwrite   = "OVERWRITE_REPORT"
        }
      }
    }

    refresh_cadence {
      frequency = "SYNCHRONOUS"
    }
  }
}
`, rName, query))
}
