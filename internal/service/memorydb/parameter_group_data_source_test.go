package memorydb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMemoryDBParameterGroupDataSource_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"
	dataSourceName := "data.aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "family", resourceName, "family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "parameter.*", map[string]string{
						"name":  "active-defrag-cycle-max",
						"value": "70",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "parameter.*", map[string]string{
						"name":  "active-defrag-cycle-min",
						"value": "10",
					}),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Test", resourceName, "tags.Test"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccParameterGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  parameter {
    name  = "active-defrag-cycle-max"
    value = "70"
  }

  parameter {
    name  = "active-defrag-cycle-min"
    value = "10"
  }

  tags = {
    Test = "test"
  }
}

data "aws_memorydb_parameter_group" "test" {
  name = aws_memorydb_parameter_group.test.name
}
`, rName)
}
