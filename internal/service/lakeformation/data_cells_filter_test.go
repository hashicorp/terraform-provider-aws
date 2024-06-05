// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataCellsFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var datacellsfilter awstypes.DataCellsFilter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_data_cells_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccDataCellsFilterPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellsFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCellsFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.table_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "table_data.0.version_id"),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.column_names.#", acctest.Ct1),
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

func testAccDataCellsFilter_columnWildcard(t *testing.T) {
	ctx := acctest.Context(t)

	var datacellsfilter awstypes.DataCellsFilter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_data_cells_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccDataCellsFilterPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellsFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCellsFilterConfig_columnWildcard(rName, "my_column_12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.table_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "table_data.0.version_id"),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.column_wildcard.0.excluded_column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.column_wildcard.0.excluded_column_names.0", "my_column_12"),
				),
			},
		},
	})
}

func testAccDataCellsFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var datacellsfilter awstypes.DataCellsFilter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_data_cells_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccDataCellsFilterPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellsFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCellsFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceDataCellsFilter, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataCellsFilter_rowFilter(t *testing.T) {
	ctx := acctest.Context(t)

	var datacellsfilter awstypes.DataCellsFilter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_data_cells_filter.test"

	filterExpression := `
  filter_expression = "my_column_23='testing'"
`
	allRowsildcard := `
  all_rows_wildcard {}
`
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccDataCellsFilterPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellsFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCellsFilterConfig_rowFilter(rName, filterExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.table_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "table_data.0.version_id"),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.row_filter.0.filter_expression", "my_column_23='testing'"),
				),
			},
			{
				Config: testAccDataCellsFilterConfig_rowFilter(rName, allRowsildcard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellsFilterExists(ctx, resourceName, &datacellsfilter),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.table_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "table_data.0.version_id"),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.column_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "table_data.0.row_filter.0.all_rows_wildcard.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckDataCellsFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_data_cells_filter" {
				continue
			}

			_, err := tflakeformation.FindDataCellsFilterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameDataCellsFilter, rs.Primary.ID, err)
			}

			return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameDataCellsFilter, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataCellsFilterExists(ctx context.Context, name string, datacellsfilter *awstypes.DataCellsFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameDataCellsFilter, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameDataCellsFilter, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)
		resp, err := tflakeformation.FindDataCellsFilterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameDataCellsFilter, rs.Primary.ID, err)
		}

		*datacellsfilter = *resp

		return nil
	}
}

func testAccDataCellsFilterPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

	input := &lakeformation.ListDataCellsFilterInput{}
	_, err := conn.ListDataCellsFilter(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDataCellsFilterConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}


data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name    = "my_column_12"
      type    = "date"
      comment = "my_column1_comment2"
    }

    columns {
      name    = "my_column_22"
      type    = "timestamp"
      comment = "my_column2_comment2"
    }

    columns {
      name    = "my_column_23"
      type    = "string"
      comment = "my_column23_comment2"
    }
  }
}`, rName)
}

func testAccDataCellsFilterConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDataCellsFilterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_lakeformation_data_cells_filter" "test" {
  table_data {
    database_name    = aws_glue_catalog_database.test.name
    name             = %[1]q
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.test.name

    column_names = ["my_column_22"]

    row_filter {
      filter_expression = "my_column_23='testing'"
    }
  }

  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName))
}

func testAccDataCellsFilterConfig_columnWildcard(rName, column string) string {
	return acctest.ConfigCompose(
		testAccDataCellsFilterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_lakeformation_data_cells_filter" "test" {
  table_data {
    database_name    = aws_glue_catalog_database.test.name
    name             = %[1]q
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.test.name

    column_wildcard {
      excluded_column_names = [%[2]q]
    }

    row_filter {
      filter_expression = "my_column_23='testing'"
    }
  }

  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, column))
}

func testAccDataCellsFilterConfig_rowFilter(rName, rowFilter string) string {
	return acctest.ConfigCompose(
		testAccDataCellsFilterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_lakeformation_data_cells_filter" "test" {
  table_data {
    database_name    = aws_glue_catalog_database.test.name
    name             = %[1]q
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.test.name

    column_names = ["my_column_22"]

    row_filter {
%[2]s
    }
  }

  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, rowFilter))
}
