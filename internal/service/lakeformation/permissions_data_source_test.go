package lakeformation_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccPermissionsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "catalog_resource", dataSourceName, "catalog_resource"),
				),
			},
		},
	})
}

func testAccPermissionsDataSource_dataLocation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_dataLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
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

func testAccPermissionsDataSource_table(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_table(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsDataSourceConfig_tableWithColumns(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
