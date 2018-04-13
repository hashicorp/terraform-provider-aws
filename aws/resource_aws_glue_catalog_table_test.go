package aws

import (
	"testing"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"fmt"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/aws"
)

func TestAccAWSGlueCatalogTable_full(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_basic(rInt),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists("aws_glue_catalog_table.test"),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"name",
						fmt.Sprintf("my_test_catalog_table_%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"database",
						fmt.Sprintf("my_test_catalog_database_%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"description",
						"",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_table.test",
						"owner",
						"",
					),
				),
			},
		},
	})
}

func testAccGlueCatalogTable_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "my_test_catalog_database_%d"
}

resource "aws_glue_catalog_table" "test" {
  name     = "my_test_table_%d"
  database = "${aws_glue_catalog_database.test.name}"
}
`, rInt, rInt)
}

func testAccGlueCatalogTable_full(rInt int, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "my_test_catalog_database_%d"
}
# @TODO Fill out parameters from https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor
resource "aws_glue_catalog_table" "test" {
  name = "my_test_table_%d"
  database = "${aws_glue_catalog_database.test.name}"
  description = "%s"
  owner = "my_owner"
  retetention = "%s",
  storage {
    location = "my_location"
    columns = [{
      name = "my_column_1"
      type = "int"
      comment = "my_column_comment"
    },{
      name = "my_column_1"
      type = "string"
      comment = "my_column_comment"
    }]
    ...
  }
  partition_keys = [{
    name = "my_column_1"
    type = "int"
    comment = "my_column_comment"
  }]
  ...
}
`, rInt, rInt, rInt, desc)
}

func testAccCheckGlueTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_table" {
			continue
		}

		catalogId, dbName, tableName := readAwsGlueTableID(rs.Primary.ID)

		input := &glue.GetTableInput{
			DatabaseName: aws.String(dbName),
			CatalogId:    aws.String(catalogId),
			Name:         aws.String(tableName),
		}
		if _, err := conn.GetTable(input); err != nil {
			//Verify the error is what we want
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCheckGlueCatalogTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, tableName := readAwsGlueTableID(rs.Primary.ID)

		glueconn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueconn.GetTable(&glue.GetTableInput{
			CatalogId:    aws.String(catalogId),
			DatabaseName: aws.String(dbName),
			Name:         aws.String(tableName),
		})

		if err != nil {
			return err
		}

		if out.Table == nil {
			return fmt.Errorf("No Glue Database Found")
		}

		if *out.Table.Name != dbName {
			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
				*out.Table.Name, tableName)
		}

		return nil
	}
}
