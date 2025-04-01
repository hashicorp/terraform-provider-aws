package glue_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSourceCatalogTables_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_glue_catalog_tables.test"
	databaseName := "test_db"
	tableName1 := "table1"
	tableName2 := "table2"
	fmt.Println("Starting")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTablesDataSourceConfig_basic(databaseName, tableName1, tableName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "database_name", databaseName),
					resource.TestCheckResourceAttr(resourceName, "ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ids.0", tableName1),
					resource.TestCheckResourceAttr(resourceName, "ids.1", tableName2),
				),
			},
		},
	})
}

func testAccCatalogTablesDataSourceConfig_basic(databaseName, tableName1, tableName2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test1" {
  name          = %[2]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "col1"
      type = "string"
    }

    location = "s3://my-bucket/test1/"
    input_format = "org.apache.hadoop.mapred.TextInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"
    compressed = false
    number_of_buckets = -1
    ser_de_info {
      name = "test1"
      serialization_library = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"
    }
    stored_as_sub_directories = false
  }
}

resource "aws_glue_catalog_table" "test2" {
  name          = %[3]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "col1"
      type = "string"
    }

    location = "s3://my-bucket/test2/"
    input_format = "org.apache.hadoop.mapred.TextInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"
    compressed = false
    number_of_buckets = -1
    ser_de_info {
      name = "test2"
      serialization_library = "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe"
    }
    stored_as_sub_directories = false
  }
}

data "aws_glue_catalog_tables" "test" {
  database_name = aws_glue_catalog_database.test.name
  depends_on = [aws_glue_catalog_table.test1, aws_glue_catalog_table.test2]
}
`, databaseName, tableName1, tableName2)
}
