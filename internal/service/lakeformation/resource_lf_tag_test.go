// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourceLFTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcelftag lakeformation.GetResourceLFTagsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceLFTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLFTagConfig_basic(rName, []string{names.AttrValue}, names.AttrValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceLFTagExists(ctx, resourceName, &resourcelftag),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.0.value", names.AttrValue),
				),
			},
		},
	})
}

func testAccResourceLFTag_table(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcelftag lakeformation.GetResourceLFTagsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceLFTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLFTagConfig_table(rName, []string{names.AttrValue}, names.AttrValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceLFTagExists(ctx, resourceName, &resourcelftag),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.0.value", names.AttrValue),
				),
			},
		},
	})
}

func testAccResourceLFTag_tableWithColumns(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcelftag lakeformation.GetResourceLFTagsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceLFTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLFTagConfig_tableWithColumns(rName, []string{names.AttrValue}, names.AttrValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceLFTagExists(ctx, resourceName, &resourcelftag),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, "lf_tag.0.value", names.AttrValue),
				),
			},
		},
	})
}

func testAccResourceLFTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcelftag lakeformation.GetResourceLFTagsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource_lf_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceLFTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLFTagConfig_basic(rName, []string{names.AttrValue}, names.AttrValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceLFTagExists(ctx, resourceName, &resourcelftag),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, acctest.Provider, tflakeformation.ResourceResourceLFTag, resourceName, lfTagsDisappearsStateFunc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func lfTagsDisappearsStateFunc(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	var lfdata tflakeformation.ResourceResourceLFTagData
	var lt tflakeformation.LFTag

	if v, ok := is.Attributes[names.AttrCatalogID]; ok {
		lfdata.CatalogID = fwflex.StringValueToFramework(ctx, v)
	}

	if v, ok := is.Attributes["database.0.name"]; ok {
		lfdata.Database = fwtypes.NewListNestedObjectValueOfPtrMust[tflakeformation.Database](ctx, &tflakeformation.Database{
			Name: fwflex.StringValueToFramework(ctx, v),
		})
	}

	if v, ok := is.Attributes["lf_tag.0.key"]; ok {
		lt.Key = fwflex.StringValueToFramework(ctx, v)
	}

	if v, ok := is.Attributes["lf_tag.0.value"]; ok {
		lt.Value = fwflex.StringValueToFramework(ctx, v)
	}

	lfdata.LFTag = fwtypes.NewListNestedObjectValueOfPtrMust[tflakeformation.LFTag](ctx, &lt)

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(names.AttrCatalogID), lfdata.Database)); err != nil {
		log.Printf("[WARN] %s", err)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(names.AttrDatabase), lfdata.Database)); err != nil {
		log.Printf("[WARN] %s", err)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("lf_tag"), lfdata.LFTag)); err != nil {
		log.Printf("[WARN] %s", err)
	}

	return nil
}

func testAccCheckResourceLFTagDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_resource_lf_tag" {
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

				if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_wildcard.#"]); err == nil && n > 0 {
					input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{}
				}

				if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_wildcard.0.excluded_column_names.#"]); err == nil && n > 0 {
					var cols []string
					for i := 0; i < n; i++ {
						cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.column_wildcard.0.excluded_column_names.%d", i)])
					}
					input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{
						ExcludedColumnNames: cols,
					}
				}
			}
			_, err := conn.GetResourceLFTags(ctx, input)

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

			if err != nil {
				return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameResourceLFTag, rs.Primary.ID, err)
			}

			return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameResourceLFTag, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourceLFTagExists(ctx context.Context, name string, resourcelftag *lakeformation.GetResourceLFTagsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameResourceLFTag, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameResourceLFTag, name, errors.New("not set"))
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

			if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_wildcard.#"]); err == nil && n > 0 {
				input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{}
			}

			if n, err := strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_wildcard.0.excluded_column_names.#"]); err == nil && n > 0 {
				var cols []string
				for i := 0; i < n; i++ {
					cols = append(cols, rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.column_wildcard.0.excluded_column_names.%d", i)])
				}
				input.Resource.TableWithColumns.ColumnWildcard = &awstypes.ColumnWildcard{
					ExcludedColumnNames: cols,
				}
			}
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)
		resp, err := conn.GetResourceLFTags(ctx, input)

		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameResourceLFTag, rs.Primary.ID, err)
		}

		*resourcelftag = *resp

		return nil
	}
}

func testAccResourceLFTagConfig_basic(rName string, values []string, value string) string {
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

resource "aws_lakeformation_resource_lf_tag" "test" {
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

func testAccResourceLFTagConfig_table(rName string, values []string, value string) string {
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

resource "aws_lakeformation_resource_lf_tag" "test" {
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

func testAccResourceLFTagConfig_tableWithColumns(rName string, valuesList []string, value1 string) string {
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


resource "aws_lakeformation_resource_lf_tag" "test" {
  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }

  lf_tag {
    key   = aws_lakeformation_lf_tag.test.key
    value = %[3]q
  }

  # for consistency, ensure that admins are set up before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, fmt.Sprintf(`"%s"`, strings.Join(valuesList, `", "`)), value1)
}
