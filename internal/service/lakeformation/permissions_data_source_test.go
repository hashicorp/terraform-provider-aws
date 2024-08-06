// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPermissionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "catalog_resource", dataSourceName, "catalog_resource"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_dataCellsFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_dataCellsFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "data_cells_filter.#", dataSourceName, "data_cells_filter.#"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_dataLocation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_dataLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.#", dataSourceName, "data_location.#"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.0.arn", dataSourceName, "data_location.0.arn"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_database(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "database.#", dataSourceName, "database.#"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dataSourceName, "database.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.1", dataSourceName, "permissions.1"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.2", dataSourceName, "permissions.2"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.#", dataSourceName, "permissions_with_grant_option.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.0", dataSourceName, "permissions_with_grant_option.0"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_lfTag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_lfTag(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.#", dataSourceName, "lf_tag.#"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.0.key", dataSourceName, "lf_tag.0.key"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag.0.values", dataSourceName, "lf_tag.0.values"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.1", dataSourceName, "permissions.1"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.#", dataSourceName, "permissions_with_grant_option.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.0", dataSourceName, "permissions_with_grant_option.0"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.1", dataSourceName, "permissions_with_grant_option.1"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_lfTagPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_lfTagPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.#", dataSourceName, "lf_tag_policy.#"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.resource_type", dataSourceName, "lf_tag_policy.0.resource_type"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.#", dataSourceName, "lf_tag_policy.0.expression.#"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.0.key", dataSourceName, "lf_tag_policy.0.expression.0.key"),
					resource.TestCheckResourceAttrPair(resourceName, "lf_tag_policy.0.expression.0.values", dataSourceName, "lf_tag_policy.0.expression.0.values"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.1", dataSourceName, "permissions.1"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.2", dataSourceName, "permissions.2"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.#", dataSourceName, "permissions_with_grant_option.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.0", dataSourceName, "permissions_with_grant_option.0"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_table(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_table(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "table.#", dataSourceName, "table.#"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dataSourceName, "table.0.database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dataSourceName, "table.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_tableWithColumns(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_tableWithColumns(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, dataSourceName, names.AttrPrincipal),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.#", dataSourceName, "table_with_columns.#"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", dataSourceName, "table_with_columns.0.database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", dataSourceName, "table_with_columns.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.column_names.#", dataSourceName, "table_with_columns.0.column_names.#"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.column_names.0", dataSourceName, "table_with_columns.0.column_names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.column_names.1", dataSourceName, "table_with_columns.0.column_names.1"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
				),
			},
		},
	})
}

func testAccPermissionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

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

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal        = aws_iam_role.test.arn
  permissions      = ["CREATE_DATABASE"]
  catalog_resource = true

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal        = aws_lakeformation_permissions.test.principal
  catalog_resource = true
}
`, rName)
}

func testAccPermissionsDataSourceConfig_dataCellsFilter(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

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

data "aws_caller_identity" "current" {}

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
}

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

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DESCRIBE"]

  data_cells_filter {
    database_name    = aws_lakeformation_data_cells_filter.test.table_data[0].database_name
    name             = aws_lakeformation_data_cells_filter.test.table_data[0].name
    table_catalog_id = aws_lakeformation_data_cells_filter.test.table_data[0].table_catalog_id
    table_name       = aws_lakeformation_data_cells_filter.test.table_data[0].table_name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  data_cells_filter {
    database_name    = aws_lakeformation_data_cells_filter.test.table_data[0].database_name
    name             = aws_lakeformation_data_cells_filter.test.table_data[0].name
    table_catalog_id = aws_lakeformation_data_cells_filter.test.table_data[0].table_catalog_id
    table_name       = aws_lakeformation_data_cells_filter.test.table_data[0].table_name
  }
}
`, rName)
}

func testAccPermissionsDataSourceConfig_dataLocation(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

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

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_lakeformation_resource" "test" {
  arn = aws_s3_bucket.test.arn
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DATA_LOCATION_ACCESS"]

  data_location {
    arn = aws_s3_bucket.test.arn
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  data_location {
    arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

func testAccPermissionsDataSourceConfig_database(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

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

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  database {
    name = aws_glue_catalog_database.test.name
  }
}
`, rName)
}

func testAccPermissionsDataSourceConfig_lfTag(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = ["value1", "value2"]
  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ASSOCIATE", "DESCRIBE"]
  permissions_with_grant_option = ["ASSOCIATE", "DESCRIBE"]
  principal                     = aws_iam_role.test.arn

  lf_tag {
    key    = aws_lakeformation_lf_tag.test.key
    values = aws_lakeformation_lf_tag.test.values
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  lf_tag {
    key    = aws_lakeformation_lf_tag.test.key
    values = aws_lakeformation_lf_tag.test.values
  }
}
`, rName)
}

func testAccPermissionsDataSourceConfig_lfTagPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = ["value1", "value2"]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  lf_tag_policy {
    resource_type = "DATABASE"

    expression {
      key    = aws_lakeformation_lf_tag.test.key
      values = aws_lakeformation_lf_tag.test.values
    }
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [
    aws_lakeformation_data_lake_settings.test,
    aws_lakeformation_lf_tag.test,
  ]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  lf_tag_policy {
    resource_type = "DATABASE"

    expression {
      key    = aws_lakeformation_lf_tag.test.key
      values = aws_lakeformation_lf_tag.test.values
    }
  }
}
`, rName)
}

func testAccPermissionsDataSourceConfig_table(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

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

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }
}
`, rName)
}

func testAccPermissionsDataSourceConfig_tableWithColumns(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

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

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }
    columns {
      name = "timestamp"
      type = "date"
    }
    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }
}
`, rName)
}
