// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourceLFTags_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_basic(rName, []string{"copse"}, "copse"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "copse",
					}),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrCatalogID),
				),
			},
		},
	})
}

func testAccResourceLFTags_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource_lf_tags.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLFTagsConfig_basic(rName, []string{"copse"}, "copse"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceResourceLFTags(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourceLFTags_database(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_database(rName, []string{"copse"}, "copse"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "copse",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_database(rName, []string{"luffield"}, "luffield"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "luffield",
					}),
				),
			},
		},
	})
}

func testAccResourceLFTags_databaseMultipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_resource_lf_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_databaseMultipleTags(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "woodcote", "theloop"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "woodcote",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-2", rName),
						names.AttrValue: "theloop",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_databaseMultipleTags(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "stowe", "becketts"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "stowe",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-2", rName),
						names.AttrValue: "becketts",
					}),
				),
			},
		},
	})
}

func testAccResourceLFTags_hierarchy(t *testing.T) {
	ctx := acctest.Context(t)
	databaseResourceName := "aws_lakeformation_resource_lf_tags.database_tags"
	tableResourceName := "aws_lakeformation_resource_lf_tags.table_tags"
	columnResourceName := "aws_lakeformation_resource_lf_tags.column_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLFTagsConfig_hierarchy(rName,
					[]string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"},
					[]string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"},
					[]string{"one", "two", "three"},
					"woodcote",
					"theloop",
					"two",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, databaseResourceName),
					testAccCheckDatabaseLFTagsExists(ctx, tableResourceName),
					testAccCheckDatabaseLFTagsExists(ctx, columnResourceName),
					resource.TestCheckResourceAttr(databaseResourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(databaseResourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "woodcote",
					}),
					resource.TestCheckResourceAttr(tableResourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(tableResourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-2", rName),
						names.AttrValue: "theloop",
					}),
					resource.TestCheckResourceAttr(columnResourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(columnResourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-3", rName),
						names.AttrValue: "two",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_hierarchy(rName,
					[]string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"},
					[]string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"},
					[]string{"one", "two", "three"},
					"stowe",
					"becketts",
					"three",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, databaseResourceName),
					testAccCheckDatabaseLFTagsExists(ctx, tableResourceName),
					testAccCheckDatabaseLFTagsExists(ctx, columnResourceName),
					resource.TestCheckResourceAttr(databaseResourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(databaseResourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "stowe",
					}),
					resource.TestCheckResourceAttr(tableResourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(tableResourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-2", rName),
						names.AttrValue: "becketts",
					}),
					resource.TestCheckResourceAttr(columnResourceName, "lf_tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(columnResourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-3", rName),
						names.AttrValue: "three",
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_table(rName, []string{"copse", "abbey", "farm"}, "abbey"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "abbey",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_table(rName, []string{"copse", "abbey", "farm"}, "farm"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "farm",
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceLFTagsConfig_tableWithColumnsMultipleTags(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "luffield", "vale"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "luffield",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-2", rName),
						names.AttrValue: "vale",
					}),
				),
			},
			{
				Config: testAccResourceLFTagsConfig_tableWithColumnsMultipleTags(rName, []string{"abbey", "village", "luffield", "woodcote", "copse", "chapel", "stowe", "club"}, []string{"farm", "theloop", "aintree", "brooklands", "maggotts", "becketts", "vale"}, "copse", "aintree"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseLFTagsExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   rName,
						names.AttrValue: "copse",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lf_tag.*", map[string]string{
						names.AttrKey:   fmt.Sprintf("%s-2", rName),
						names.AttrValue: "aintree",
					}),
				),
			},
		},
	})
}

func testAccCheckDatabaseLFTagsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_resource_lf_tags" {
				continue
			}

			input := &lakeformation.GetResourceLFTagsInput{
				Resource:           &awstypes.Resource{},
				ShowAssignedLFTags: aws.Bool(true),
			}

			if v, ok := rs.Primary.Attributes[names.AttrCatalogID]; ok {
				input.CatalogId = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["database.0.name"]; ok {
				input.Resource.Database = &awstypes.DatabaseResource{
					Name: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["database.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.Database.CatalogId = aws.String(v)
				}
			}

			if v, ok := rs.Primary.Attributes["table.0.database_name"]; ok {
				input.Resource.Table = &awstypes.TableResource{
					DatabaseName: aws.String(v),
				}

				if v, ok := rs.Primary.Attributes["table.0.catalog_id"]; ok && len(v) > 1 {
					input.Resource.Table.CatalogId = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table.0.name"]; ok {
					input.Resource.Table.Name = aws.String(v)
				}

				if v, ok := rs.Primary.Attributes["table.0.wildcard"]; ok && v == acctest.CtTrue {
					input.Resource.Table.TableWildcard = &awstypes.TableWildcard{}
				}
			}

			if v, ok := rs.Primary.Attributes["table_with_columns.0.database_name"]; ok {
				input.Resource.TableWithColumns = &awstypes.TableWithColumnsResource{
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
					input.Resource.TableWithColumns.ColumnNames = cols
				}

				if v, ok := rs.Primary.Attributes["table_with_columns.0.wildcard"]; ok && v == acctest.CtTrue {
					input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{}
				}

				if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"]); err == nil && n > 0 {
					var cols []string
					for i := 0; i < n; i++ {
						cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)])
					}
					input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{
						ExcludedColumnNames: cols,
					}
				}
			}

			if _, err := conn.GetResourceLFTags(ctx, input); err != nil {
				if errs.IsA[*awstypes.EntityNotFoundException](err) {
					continue
				}

				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "not found") {
					continue
				}

				// If the lake formation admin has been revoked, there will be access denied instead of entity not found
				if errs.IsA[*awstypes.AccessDeniedException](err) {
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
			Resource:           &awstypes.Resource{},
			ShowAssignedLFTags: aws.Bool(true),
		}

		if v, ok := rs.Primary.Attributes[names.AttrCatalogID]; ok {
			input.CatalogId = aws.String(v)
		}

		if v, ok := rs.Primary.Attributes["database.0.name"]; ok {
			input.Resource.Database = &awstypes.DatabaseResource{
				Name: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["database.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.Database.CatalogId = aws.String(v)
			}
		}

		if v, ok := rs.Primary.Attributes["table.0.database_name"]; ok {
			input.Resource.Table = &awstypes.TableResource{
				DatabaseName: aws.String(v),
			}

			if v, ok := rs.Primary.Attributes["table.0.catalog_id"]; ok && len(v) > 1 {
				input.Resource.Table.CatalogId = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table.0.name"]; ok {
				input.Resource.Table.Name = aws.String(v)
			}

			if v, ok := rs.Primary.Attributes["table.0.wildcard"]; ok && v == acctest.CtTrue {
				input.Resource.Table.TableWildcard = &awstypes.TableWildcard{}
			}
		}

		if v, ok := rs.Primary.Attributes["table_with_columns.0.database_name"]; ok {
			input.Resource.TableWithColumns = &awstypes.TableWithColumnsResource{
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
				input.Resource.TableWithColumns.ColumnNames = cols
			}

			if v, ok := rs.Primary.Attributes["table_with_columns.0.wildcard"]; ok && v == acctest.CtTrue {
				input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{}
			}

			if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"]); err == nil && n > 0 {
				var cols []string
				for i := 0; i < n; i++ {
					cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)])
				}
				input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{
					ExcludedColumnNames: cols,
				}
			}
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)
		_, err := conn.GetResourceLFTags(ctx, input)

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

  # for consistency, ensure that admins are set up before testing
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

  # for consistency, ensure that admins are set up before testing
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

  # for consistency, ensure that admins are set up before testing
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values, `", "`)), value)
}

func testAccResourceLFTagsConfig_databaseMultipleTags(rName string, values1, values2 []string, value1, value2 string) string {
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "test2" {
  key    = "%[1]s-2"
  values = [%[3]s]

  # for consistency, ensure that admins are set up before testing
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values1, `", "`)), fmt.Sprintf(`"%s"`, strings.Join(values2, `", "`)), value1, value2)
}

func testAccResourceLFTagsConfig_hierarchy(rName string, values1, values2, values3 []string, value1, value2, value3 string) string {
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "test2" {
  key    = "%[1]s-2"
  values = [%[3]s]

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "column_tags" {
  key    = "%[1]s-3"
  values = [%[6]s]

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "database_tags" {
  database {
    name = aws_glue_catalog_database.test.name
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[4]q
  }

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "table_tags" {
  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test2.key
    value = %[5]q
  }

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_resource_lf_tags" "column_tags" {
  table_with_columns {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.column_tags.key
    value = %[7]q
  }

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values1, `", "`)), fmt.Sprintf(`"%s"`, strings.Join(values2, `", "`)), value1, value2, fmt.Sprintf(`"%s"`, strings.Join(values3, `", "`)), value3)
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

  # for consistency, ensure that admins are set up before testing
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values, `", "`)), value)
}

func testAccResourceLFTagsConfig_tableWithColumnsMultipleTags(rName string, values1, values2 []string, value1 string, value2 string) string {
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "test2" {
  key    = "%[1]s-2"
  values = [%[3]s]

  # for consistency, ensure that admins are set up before testing
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

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(values1, `", "`)), fmt.Sprintf(`"%s"`, strings.Join(values2, `", "`)), value1, value2)
}
