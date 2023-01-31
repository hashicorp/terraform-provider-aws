package lakeformation_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccResourceLFTags_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_basic(rName, []string{"copse"}, "copse"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "copse",
					}),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
		},
	})
}

func testAccResourceLFTags_database(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_database(rName, []string{"copse"}, "copse"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "copse",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_database(rName, []string{"luffield"}, "luffield"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "luffield",
					}),
				),
			},
		},
	})
}

func testAccResourceLFTags_databaseMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_databaseMultiple(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "woodcote", "theloop"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "woodcote",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   fmt.Sprintf("%s-2", rName),
						"value": "theloop",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_databaseMultiple(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "stowe", "becketts"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "stowe",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   fmt.Sprintf("%s-2", rName),
						"value": "becketts",
					}),
				),
			},
		},
	})
}

func testAccResourceLFTags_table(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_table(rName, []string{"copse", "abbey", "farm"}, "abbey"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "abbey",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_table(rName, []string{"copse", "abbey", "farm"}, "farm"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "farm",
					}),
				),
			},
		},
	})
}

func testAccResourceLFTags_tableWithColumns(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_tableWithColumnsMultiple(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "luffield", "vale"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "luffield",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   fmt.Sprintf("%s-2", rName),
						"value": "vale",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_tableWithColumnsMultiple(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "copse", "aintree"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   rName,
						"value": "copse",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						"key":   fmt.Sprintf("%s-2", rName),
						"value": "aintree",
					}),
				),
			},
		},
	})
}

func testAccCheckDatabaseLFTagsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_resource_lf_tags" {
				continue
			}

			input := &lakeformation.GetResourceLFTagsInput{
				Resource:           &lakeformation.Resource{},
				ShowAssignedLFTags: aws.Bool(true),
			}

			if v, ok := rs.Primary.Attributes["catalog_id"]; ok {
				input.CatalogId = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["database.0.name"]; ok {
				input.Resource.Database = &lakeformation.DatabaseResource{
					Name: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["database.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.Database.CatalogId = aws.String(v)
				}
			}

			if v, ok := rs.Primary.Attributes["table.0.database_name"]; ok {
				input.Resource.Table = &lakeformation.TableResource{
					DatabaseName: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["table.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.Table.CatalogId = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table.0.name"]; ok {
					input.Resource.Table.Name = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table.0.wildcard"]; ok && v == "true" {
					input.Resource.Table.TableWildcard = &lakeformation.TableWildcard{}
				}
			}

			if v, ok := rs.Primary.Attributes["table_with_columns.0.database_name"]; ok {
				input.Resource.TableWithColumns = &lakeformation.TableWithColumnsResource{
					DatabaseName: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["table_with_columns.0.name"]; ok {
					input.Resource.TableWithColumns.Name = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table_with_columns.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.TableWithColumns.CatalogId = aws.String(v)
				}

				if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_names.#"]); err == nil && n > 0 {
					var cols []string
					for i := 0; i < n; i++ {
						cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.column_names.%d", i)])
					}
					input.Resource.TableWithColumns.ColumnNames = aws.StringSlice(cols)
				}

				if v, ok := rs.Primary.Attributes["table_with_columns.0.wildcard"]; ok && v == "true" {
					input.Resource.TableWithColumns.ColumnWildcard = &lakeformation.ColumnWildcard{}
				}

				if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"]); err == nil && n > 0 {
					var cols []string
					for i := 0; i < n; i++ {
						cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)])
					}
					input.Resource.TableWithColumns.ColumnWildcard = &lakeformation.ColumnWildcard{
						ExcludedColumnNames: aws.StringSlice(cols),
					}
				}
			}

			if _, err := conn.GetResourceLFTagsWithContext(ctx, input); err != nil {
				if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
					continue
				}

				if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "not found") {
					continue
				}

				// If the lake formation admin has been revoked, there will be access denied instead of entity not found
				if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeAccessDeniedException) {
					continue
				}
				return err
			}
			return fmt.Errorf("Lake Formation Resource LF Tag (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDatabaseLFTagsExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("acceptance test: resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		input := &lakeformation.GetResourceLFTagsInput{
			Resource:           &lakeformation.Resource{},
			ShowAssignedLFTags: aws.Bool(true),
		}

		if v, ok := rs.Primary.Attributes["catalog_id"]; ok {
			input.CatalogId = aws.String(v)
		}

		if v, ok := rs.Primary.Attributes["database.0.name"]; ok {
			input.Resource.Database = &lakeformation.DatabaseResource{
				Name: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["database.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.Database.CatalogId = aws.String(v)
			}
		}

		if v, ok := rs.Primary.Attributes["table.0.database_name"]; ok {
			input.Resource.Table = &lakeformation.TableResource{
				DatabaseName: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["table.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.Table.CatalogId = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table.0.name"]; ok {
				input.Resource.Table.Name = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table.0.wildcard"]; ok && v == "true" {
				input.Resource.Table.TableWildcard = &lakeformation.TableWildcard{}
			}
		}

		if v, ok := rs.Primary.Attributes["table_with_columns.0.database_name"]; ok {
			input.Resource.TableWithColumns = &lakeformation.TableWithColumnsResource{
				DatabaseName: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["table_with_columns.0.name"]; ok {
				input.Resource.TableWithColumns.Name = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table_with_columns.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.TableWithColumns.CatalogId = aws.String(v)
			}

			if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_names.#"]); err == nil && n > 0 {
				var cols []string
				for i := 0; i < n; i++ {
					cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.column_names.%d", i)])
				}
				input.Resource.TableWithColumns.ColumnNames = aws.StringSlice(cols)
			}

			if v, ok := rs.Primary.Attributes["table_with_columns.0.wildcard"]; ok && v == "true" {
				input.Resource.TableWithColumns.ColumnWildcard = &lakeformation.ColumnWildcard{}
			}

			if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"]); err == nil && n > 0 {
				var cols []string
				for i := 0; i < n; i++ {
					cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)])
				}
				input.Resource.TableWithColumns.ColumnWildcard = &lakeformation.ColumnWildcard{
					ExcludedColumnNames: aws.StringSlice(cols),
				}
			}
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn()
		_, err := conn.GetResourceLFTagsWithContext(ctx, input)

		return err
	}
}

func testAccResourceLFTagsConfig_basic(rName string, values []string, value string) string {
	return fmt.Sprintf(`
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

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = [%[2]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "test" {
  catalog_id = data.aws_caller_identity.current.account_id

  database {
    name = aws_glue_catalog_database.test.name
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[3]q
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values, `", "`)), value)
}

func testAccResourceLFTagsConfig_database(rName string, values []string, value string) string {
	return fmt.Sprintf(`
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

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = [%[2]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "test" {
  database {
    name = aws_glue_catalog_database.test.name
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[3]q
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values, `", "`)), value)
}

func testAccResourceLFTagsConfig_databaseMultiple(rName string, values1, values2 []string, value1, value2 string) string {
	return fmt.Sprintf(`
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

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = [%[2]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "test2" {
  key    = "%[1]s-2"
  values = [%[3]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "test" {
  database {
    name = aws_glue_catalog_database.test.name
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[4]q
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test2.key
    value = %[5]q
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values1, `", "`)), fmt.Sprintf(`"%s"`, strings.Join(values2, `", "`)), value1, value2)
}

func testAccResourceLFTagsConfig_table(rName string, values []string, value string) string {
	return fmt.Sprintf(`
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

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = [%[2]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "test" {
  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[3]q
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values, `", "`)), value)
}

func testAccResourceLFTagsConfig_tableWithColumnsMultiple(rName string, values1, values2 []string, value1 string, value2 string) string {
	return fmt.Sprintf(`
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
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = [%[2]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "test2" {
  key    = "%[1]s-2"
  values = [%[3]s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "test" {
  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[4]q
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test2.key
    value = %[5]q
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values1, `", "`)), fmt.Sprintf(`"%s"`, strings.Join(values2, `", "`)), value1, value2)
}
