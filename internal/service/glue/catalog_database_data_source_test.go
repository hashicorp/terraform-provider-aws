package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGlueCatalogDatabaseDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	datasourceName := "data.aws_glue_catalog_database.test"

	dbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogDatabaseDataSourceConfig_basic(dbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "catalog_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "create_table_default_permission", resourceName, "create_table_default_permission"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "location_uri", resourceName, "location_uri"),
					resource.TestCheckResourceAttrPair(datasourceName, "parameters", resourceName, "parameters"),
					resource.TestCheckResourceAttrPair(datasourceName, "target_database", resourceName, "target_database"),
				),
			},
		},
	})
}

func testAccCatalogDatabaseDataSourceConfig_basic(dbName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_glue_catalog_database" "test" {
  name = aws_glue_catalog_database.test.name
}
`, dbName)
}
