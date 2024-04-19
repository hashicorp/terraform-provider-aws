// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports_test

import (
	"context"
	"errors"
	"fmt"
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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "export.0.data_query.0.query_statement"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export.0.destination_configurations.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export.0.refresh_cadence.0.frequency", "SYNCHRONOUS"),
					resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "HOURLY"),
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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDataExportsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExportConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExportConfig_tag2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccExportConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportExists(ctx, resourceName, &export),
					resource.TestCheckResourceAttr(resourceName, "export.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
					// resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "HOURLY"),
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
					// resource.TestCheckResourceAttr(resourceName, "export.0.data_query.0.table_configurations.COST_AND_USAGE_REPORT.TIME_GRANULARITY", "DAILY"),
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
