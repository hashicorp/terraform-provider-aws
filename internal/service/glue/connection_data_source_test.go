package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGlueConnectionDataSource_basic(t *testing.T) {
	resourceName := "aws_glue_connection.test"
	datasourceName := "data.aws_glue_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccConnectionCheckDataSource(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "catalog_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties", resourceName, "connection_properties"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements", resourceName, "physical_connection_requirements"),
					resource.TestCheckResourceAttrPair(datasourceName, "match_criteria", resourceName, "match_criteria"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func testAccConnectionCheckDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccConnectionDataSourceConfig_basic(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

}

data "aws_glue_connection" "test" {
  id = aws_glue_connection.test.id
}
`, rName, jdbcConnectionUrl)
}
