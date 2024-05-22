// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cleanrooms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcleanrooms "github.com/hashicorp/terraform-provider-aws/internal/service/cleanrooms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCleanRoomsConfiguredTable_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConfiguredTable(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, TEST_NAME),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, TEST_DESCRIPTION),
					resource.TestCheckResourceAttr(resourceName, "analysis_method", TEST_ANALYSIS_METHOD),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.1", "my_column_2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "table_reference.*", map[string]string{
						names.AttrDatabaseName: rName,
						names.AttrTableName:    rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.Project", TEST_TAG),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTable_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcleanrooms.ResourceConfiguredTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTable_mutableProperties(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
				),
			},
			{
				Config: testAccConfiguredTableConfig_basic(rName, "updated description", "updated tag", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableIsTheSame(resourceName, &configuredTable),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
					resource.TestCheckResourceAttr(resourceName, "tags.Project", "updated tag"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTable_updateAllowedColumns(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_allowedColumns(TEST_ALLOWED_COLUMNS, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.1", "my_column_2"),
				),
			},
			{
				Config: testAccConfiguredTableConfig_allowedColumns("[\"my_column_1\"]", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableRecreated(resourceName, &configuredTable),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "allowed_columns.0", "my_column_1"),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTable_updateTableReference(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	firstDatabaseName := rName
	secondDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_additionalTables(rName, firstDatabaseName, secondDatabaseName, firstDatabaseName, TEST_FIRST_ADDITIONAL_TABLE_NAME),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
				),
			},
			{
				Config: testAccConfiguredTableConfig_additionalTables(rName, firstDatabaseName, secondDatabaseName, secondDatabaseName, TEST_SECOND_ADDITIONAL_TABLE_NAME),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableRecreated(resourceName, &configuredTable),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "table_reference.*", map[string]string{
						names.AttrDatabaseName: secondDatabaseName,
						names.AttrTableName:    TEST_SECOND_ADDITIONAL_TABLE_NAME,
					}),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTable_updateTableReference_onlyDatabase(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	firstDatabaseName := rName
	secondDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_additionalTables(rName, firstDatabaseName, secondDatabaseName, firstDatabaseName, TEST_FIRST_ADDITIONAL_TABLE_NAME),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
				),
			},
			{
				Config: testAccConfiguredTableConfig_additionalTables(rName, firstDatabaseName, secondDatabaseName, secondDatabaseName, TEST_FIRST_ADDITIONAL_TABLE_NAME),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableRecreated(resourceName, &configuredTable),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "table_reference.*", map[string]string{
						names.AttrDatabaseName: secondDatabaseName,
						names.AttrTableName:    TEST_FIRST_ADDITIONAL_TABLE_NAME,
					}),
				),
			},
		},
	})
}

func TestAccCleanRoomsConfiguredTable_updateTableReference_onlyTable(t *testing.T) {
	ctx := acctest.Context(t)

	var configuredTable cleanrooms.GetConfiguredTableOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	firstDatabaseName := rName
	secondDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_configured_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfiguredTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfiguredTableConfig_additionalTables(rName, firstDatabaseName, secondDatabaseName, firstDatabaseName, TEST_FIRST_ADDITIONAL_TABLE_NAME),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableExists(ctx, resourceName, &configuredTable),
				),
			},
			{
				Config: testAccConfiguredTableConfig_additionalTables(rName, firstDatabaseName, secondDatabaseName, firstDatabaseName, TEST_SECOND_ADDITIONAL_TABLE_NAME),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfiguredTableRecreated(resourceName, &configuredTable),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "table_reference.*", map[string]string{
						names.AttrDatabaseName: firstDatabaseName,
						names.AttrTableName:    TEST_SECOND_ADDITIONAL_TABLE_NAME,
					}),
				),
			},
		},
	})
}

func testAccPreCheckConfiguredTable(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)

	input := &cleanrooms.ListConfiguredTablesInput{}
	_, err := conn.ListConfiguredTables(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckConfiguredTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != tfcleanrooms.ResNameConfiguredTable {
				continue
			}

			_, err := conn.GetConfiguredTable(ctx, &cleanrooms.GetConfiguredTableInput{
				ConfiguredTableIdentifier: aws.String(rs.Primary.ID),
			})

			if err == nil {
				return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckConfiguredTableExists(ctx context.Context, name string, configuredTable *cleanrooms.GetConfiguredTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, name, errors.New("not set"))
		}

		client := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)
		resp, err := client.GetConfiguredTable(ctx, &cleanrooms.GetConfiguredTableInput{
			ConfiguredTableIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, rs.Primary.ID, err)
		}

		*configuredTable = *resp

		return nil
	}
}

func testAccCheckConfiguredTableIsTheSame(name string, configuredTable *cleanrooms.GetConfiguredTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return checkConfiguredTableIsTheSame(name, configuredTable, s)
	}
}

func testAccCheckConfiguredTableRecreated(name string, configuredTable *cleanrooms.GetConfiguredTableOutput) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		err := checkConfiguredTableIsTheSame(name, configuredTable, state)
		if err == nil {
			return fmt.Errorf("Configured Table was expected to be recreated but was updated")
		}
		return nil
	}
}

func checkConfiguredTableIsTheSame(name string, configuredTable *cleanrooms.GetConfiguredTableOutput, s *terraform.State) error {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, name, errors.New("not found"))
	}

	if rs.Primary.ID == "" {
		return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameConfiguredTable, name, errors.New("not set"))
	}

	if rs.Primary.ID != *configuredTable.ConfiguredTable.Id {
		return fmt.Errorf("New configured table: %s created instead of updating: %s", rs.Primary.ID, *configuredTable.ConfiguredTable.Id)
	}

	return nil
}

const TEST_ALLOWED_COLUMNS = "[\"my_column_1\",\"my_column_2\"]"
const TEST_ANALYSIS_METHOD = "DIRECT_QUERY"
const TEST_DATABASE_NAME = names.AttrDatabase
const TEST_TABLE_NAME = "table"

func testAccConfiguredTableConfig_basic(name string, description string, tagValue string, rName string) string {
	return testAccConfiguredTableConfig(rName, name, description, tagValue, TEST_ALLOWED_COLUMNS,
		TEST_ANALYSIS_METHOD, rName, rName)
}

func testAccConfiguredTableConfig_allowedColumns(allowedColumns string, rName string) string {
	return testAccConfiguredTableConfig(rName, TEST_NAME, TEST_DESCRIPTION, TEST_TAG, allowedColumns,
		TEST_ANALYSIS_METHOD, rName, rName)
}

const TEST_FIRST_ADDITIONAL_TABLE_NAME = "table_1"
const TEST_SECOND_ADDITIONAL_TABLE_NAME = "table_2"

func testAccConfiguredTableConfig_additionalTables(rName string, firstDatabaseName string, secondDatabaseName string, databaseName string, tableName string) string {
	storageDescriptor := `
storage_descriptor {
  location = "s3://${aws_s3_bucket.test.bucket}"

  columns {
    name = "my_column_1"
    type = "string"
  }
}`
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test_1" {
  name = %[2]q
}

resource "aws_glue_catalog_database" "test_2" {
  name = %[3]q
}

resource "aws_glue_catalog_table" "test_1" {
  name          = %[7]q
  database_name = aws_glue_catalog_database.test_1.name

  %[6]s
}

resource "aws_glue_catalog_table" "test_2" {
  name          = %[8]q
  database_name = aws_glue_catalog_database.test_1.name

  %[6]s
}

resource "aws_glue_catalog_table" "test_3" {
  name          = %[7]q
  database_name = aws_glue_catalog_database.test_2.name

  %[6]s
}

resource "aws_glue_catalog_table" "test_4" {
  name          = %[8]q
  database_name = aws_glue_catalog_database.test_2.name

  %[6]s
}

resource "aws_cleanrooms_configured_table" "test" {
  name            = "test name"
  description     = "test description"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = ["my_column_1"]

  table_reference {
    database_name = %[4]q
    table_name    = %[5]q
  }

  tags = {
    Project = "Terraform"
  }

  depends_on = [aws_glue_catalog_table.test_1, aws_glue_catalog_table.test_2, aws_glue_catalog_table.test_3, aws_glue_catalog_table.test_4]
}
	`, rName, firstDatabaseName, secondDatabaseName, databaseName, tableName, storageDescriptor, TEST_FIRST_ADDITIONAL_TABLE_NAME, TEST_SECOND_ADDITIONAL_TABLE_NAME)
}

func testAccConfiguredTableConfig(rName string, name string, description string, tagValue string, allowedColumns string,
	analysisMethod string, databaseName string, tableName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = %[1]q

  storage_descriptor {
    location = "s3://${aws_s3_bucket.test.bucket}"

    columns {
      name = "my_column_1"
      type = "string"
    }

    columns {
      name = "my_column_2"
      type = "string"
    }
  }
}

resource "aws_cleanrooms_configured_table" "test" {
  name            = %[2]q
  description     = %[3]q
  analysis_method = %[6]q
  allowed_columns = %[5]s

  table_reference {
    database_name = %[7]q
    table_name    = %[8]q
  }

  tags = {
    Project = %[4]q
  }

  depends_on = [aws_glue_catalog_table.test]
}
	`, rName, name, description, tagValue, allowedColumns, analysisMethod, databaseName, tableName)
}
