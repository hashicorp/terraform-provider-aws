package lakeformation_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

func testAccResourceDataCellsFilter_basic(t *testing.T) {
	resourceName := "aws_lakeformation_data_cells_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	filterExpression := fmt.Sprintf("event = '%s'", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellFiltersDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceDataCellFiltersConfig_basic(rName, rName, filterExpression),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellFiltersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "row_filter.0.filter_expression", filterExpression),
					acctest.CheckResourceAttrAccountID(resourceName, "table_catalog_id"),
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

func testAccResourceDataCellsFilter_disappears(t *testing.T) {
	resourceName := "aws_lakeformation_data_cells_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	filterExpression := fmt.Sprintf("event = '%s'", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellFiltersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataCellFiltersConfig_basic(rName, rName, filterExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellFiltersExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflakeformation.ResourceDataCellsFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourceDataCellsFilter_excludeColumns(t *testing.T) {
	resourceName := "aws_lakeformation_data_cells_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	filterExpression := fmt.Sprintf("event = '%s'", rName)
	excludeColumn := "timestamp"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellFiltersDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceDataCellFiltersConfig_excludeColumns(rName, rName, filterExpression, excludeColumn),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellFiltersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "column_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.0.excluded_column_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.0.excluded_column_names.0", excludeColumn),
					acctest.CheckResourceAttrAccountID(resourceName, "table_catalog_id"),
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

func testAccResourceDataCellsFilter_includeColumns(t *testing.T) {
	resourceName := "aws_lakeformation_data_cells_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	filterExpression := fmt.Sprintf("event = '%s'", rName)
	includeColumns := []string{"event", "timestamp"}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellFiltersDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceDataCellFiltersConfig_includeColumns(rName, rName, filterExpression, includeColumns),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellFiltersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.#", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "table_catalog_id"),
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

func testAccResourceDataCellsFilter_allRowsExcludeColumns(t *testing.T) {
	resourceName := "aws_lakeformation_data_cells_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	excludeColumn := "timestamp"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCellFiltersDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceDataCellFiltersConfig_allRowsExcludeColumns(rName, rName, excludeColumn),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCellFiltersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "column_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.0.excluded_column_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "column_wildcard.0.excluded_column_names.0", excludeColumn),
					resource.TestCheckResourceAttr(resourceName, "row_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "row_filter.0.all_rows_wildcard", "true"),
					acctest.CheckResourceAttrAccountID(resourceName, "table_catalog_id"),
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

func testAccCheckDataCellFiltersDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_data_cells_filter" {
			continue
		}

		catalogID, databaseName, tableName, dataCellsFilterName, err := tflakeformation.ReadDataCellsFilterID(rs.Primary.ID)
		if err != nil {
			return err

		}

		// there is no api to get a single data cell filter. the only option is to iterate through a list
		inDCFId := fmt.Sprintf("%s:%s:%s:%s", catalogID, databaseName, tableName, dataCellsFilterName)

		in := &lakeformation.ListDataCellsFilterInput{
			Table: &lakeformation.TableResource{
				CatalogId:    aws.String(catalogID),
				DatabaseName: aws.String(databaseName),
				Name:         aws.String(tableName),
				//Table.TableWildcard
			},
			//MaxResults: aws.Int64(maxResults),
		}

		bMatch := false
		continueIteration := true
		pageNum := 0
		//recNum := 0
		errList := conn.ListDataCellsFilterPages(in, func(page *lakeformation.ListDataCellsFilterOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			pageNum++

			for recNum, dcf := range page.DataCellsFilters {
				recNum++
				outDCFId := fmt.Sprintf("%s:%s:%s:%s", *dcf.TableCatalogId, *dcf.DatabaseName, *dcf.TableName, *dcf.Name)
				if inDCFId == outDCFId {
					bMatch = true
					continueIteration = false
					log.Printf("[INFO] testAccCheckDataCellFiltersDestroy Matched DCF. item: %v on page: %v", recNum, pageNum)
					break
				}
			}

			// return false to stop the function from iterating through pages
			return continueIteration
		})

		if tfawserr.ErrCodeEquals(errList, lakeformation.ErrCodeEntityNotFoundException) {
			continue
		}

		if tfawserr.ErrMessageContains(errList, lakeformation.ErrCodeInvalidInputException, "not found") {
			continue
		}

		// If the lake formation admin has been revoked, there will be access denied instead of entity not found
		if tfawserr.ErrCodeEquals(errList, lakeformation.ErrCodeAccessDeniedException) {
			continue
		}

		if bMatch {
			return fmt.Errorf("Lake Formation Resource Data Cell Filter (%s) still exists", rs.Primary.ID)
		} else {
			log.Printf("[INFO] testAccCheckDataCellFiltersDestroy found no resource")
			return nil
		}
	}

	return nil
}

func testAccCheckDataCellFiltersExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("acceptance test: resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		catalogID, databaseName, tableName, dataCellsFilterName, err := tflakeformation.ReadDataCellsFilterID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn

		// there is no api to get a single data cell filter. the only option is to iterate through a list
		inDCFId := fmt.Sprintf("%s:%s:%s:%s", catalogID, databaseName, tableName, dataCellsFilterName)

		in := &lakeformation.ListDataCellsFilterInput{
			Table: &lakeformation.TableResource{
				CatalogId:    aws.String(catalogID),
				DatabaseName: aws.String(databaseName),
				Name:         aws.String(tableName),
				//Table.TableWildcard
			},
			//MaxResults: aws.Int64(maxResults),
		}

		bMatch := false
		continueIteration := true
		pageNum := 0
		errList := conn.ListDataCellsFilterPages(in, func(page *lakeformation.ListDataCellsFilterOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			pageNum++

			for recNum, dcf := range page.DataCellsFilters {
				recNum++
				outDCFId := fmt.Sprintf("%s:%s:%s:%s", *dcf.TableCatalogId, *dcf.DatabaseName, *dcf.TableName, *dcf.Name)
				if inDCFId == outDCFId {
					bMatch = true
					continueIteration = false
					break
				}
			}

			// return false to stop the function from iterating through pages
			return continueIteration
		})

		if bMatch {
			log.Printf("[INFO] testAccCheckDataCellFiltersExists errList %v", errList)
			return nil
		} else {
			return fmt.Errorf("Lake Formation Resource Data Cell Filter (%s) does not exist", rs.Primary.ID)
		}
	}
}

// copied glue configuration from resource_lf_tags_test
func testAccResourceDataCellFiltersConfig_basic(rName string, dcfName string, filterExpression string) string {
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

resource "aws_lakeformation_data_cells_filter" "test" {
	table_catalog_id = data.aws_caller_identity.current.account_id
	database_name    = aws_glue_catalog_database.test.name
	table_name       = aws_glue_catalog_table.test.name
	name             = %[2]q

	row_filter {
		filter_expression = %[3]q
	}

	column_wildcard {}

	# for consistency, ensure that admins are setup before testing
	# depends_on = [aws_lakeformation_data_lake_settings.test]
}

`, rName, dcfName, filterExpression)
}

func testAccResourceDataCellFiltersConfig_excludeColumns(rName string, dcfName string, filterExpression string, excludeColumns string) string {
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

resource "aws_lakeformation_data_cells_filter" "test" {
  table_catalog_id = data.aws_caller_identity.current.account_id
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  name             = %[2]q

  row_filter {
    filter_expression = %[3]q
  }

  column_wildcard {
    excluded_column_names = [ %[4]q ]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

`, rName, dcfName, filterExpression, excludeColumns)
}

func testAccResourceDataCellFiltersConfig_includeColumns(rName string, dcfName string, filterExpression string, includeColumns []string) string {
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

resource "aws_lakeformation_data_cells_filter" "test" {
  table_catalog_id = data.aws_caller_identity.current.account_id
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  name             = %[2]q

  row_filter {
    filter_expression = %[3]q
  }

  column_names  = [ %[4]s ]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

`, rName, dcfName, filterExpression, fmt.Sprintf(`"%s"`, strings.Join(includeColumns, `", "`)))
}

func testAccResourceDataCellFiltersConfig_allRowsExcludeColumns(rName string, dcfName string, excludeColumns string) string {
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

resource "aws_lakeformation_data_cells_filter" "test" {
  table_catalog_id = data.aws_caller_identity.current.account_id
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  name             = %[2]q

  row_filter {
    all_rows_wildcard = "true"
  }

  column_wildcard {
    excluded_column_names = [ %[3]q ]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

`, rName, dcfName, excludeColumns)
}
