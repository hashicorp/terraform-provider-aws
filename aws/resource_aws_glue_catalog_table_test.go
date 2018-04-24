package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSGlueCatalogTable_full(t *testing.T) {
	rInt := acctest.RandInt()
	description := "A test table from terraform"
	tableName := "aws_glue_catalog_table.test"

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
						tableName,
						"database_name",
						fmt.Sprintf("my_test_catalog_database_%d", rInt),
					),
				),
			},
			{
				Config:  testAccGlueCatalogTable_full(rInt, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(tableName),
					resource.TestCheckResourceAttr(tableName, "name", fmt.Sprintf("my_test_catalog_table_%d", rInt)),
					resource.TestCheckResourceAttr(tableName, "database_name", fmt.Sprintf("my_test_catalog_database_%d", rInt)),
					resource.TestCheckResourceAttr(tableName, "description", description),
					resource.TestCheckResourceAttr(tableName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(tableName, "retention", "1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(tableName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(tableName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(tableName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(tableName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(tableName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(tableName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(tableName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(tableName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(tableName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(tableName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(tableName, "parameters.param1", "param1_val"),
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
  name     = "my_test_catalog_table_%d"
  database_name = "${aws_glue_catalog_database.test.name}"
}
`, rInt, rInt)
}

func testAccGlueCatalogTable_full(rInt int, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "my_test_catalog_database_%d"
}

resource "aws_glue_catalog_table" "test" {
  name = "my_test_catalog_table_%d"
  database_name = "${aws_glue_catalog_database.test.name}"
  description = "%s"
  owner = "my_owner"
  retention = 1
  storage_descriptor {
    columns = [
     {
       name = "my_column_1"
       type = "int"
       comment = "my_column1_comment"
     },
     {
       name = "my_column_2"
       type = "string"
       comment = "my_column2_comment"
     }
    ]
	location = "my_location"
	input_format = "SequenceFileInputFormat"
	output_format = "SequenceFileInputFormat"
	compressed = false
	number_of_buckets = 1
	ser_de_info {
      name = "ser_de_name"
      parameters {
        param1 = "param_val_1"
      }
      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
	}
	bucket_columns = ["bucket_column_1"]
	sort_columns = [
      {
        column = "my_column_1"
        sort_order = 1
      }
	]
	parameters {
      param1 = "param1_val"
	}
	skewed_info {
      skewed_column_names = [
        "my_column_1"
      ]
      skewed_column_value_location_maps {
        my_column_1 = "my_column_1_val_loc_map"
      }
      skewed_column_values = [
        "skewed_val_1"
      ]
	}
	stored_as_sub_directories = false
  }
  partition_keys = [
    {
      name = "my_column_1"
      type = "int"
      comment = "my_column_1_comment"
    },
    {
      name = "my_column_2"
      type = "string"
      comment = "my_column_2_comment"
    }
  ]
  view_original_text = "view_original_text_1"
  view_expanded_text = "view_expanded_text_1"
  table_type = "VIRTUAL_VIEW"
  parameters {
  	param1 = "param1_val"
  }
}
`, rInt, rInt, desc)
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
			return fmt.Errorf("No Glue Table Found")
		}

		if *out.Table.Name != tableName {
			return fmt.Errorf("Glue Table Mismatch - existing: %q, state: %q",
				*out.Table.Name, tableName)
		}

		return nil
	}
}
