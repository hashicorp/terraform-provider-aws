package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectUserHierarchyStructureDataSource_instanceID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"
	datasourceName := "data.aws_connect_user_hierarchy_structure.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureDataSourceConfig_instanceID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.#", resourceName, "hierarchy_structure.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_one.#", resourceName, "hierarchy_structure.0.level_one.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_one.0.name", resourceName, "hierarchy_structure.0.level_one.0.name"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_two.#", resourceName, "hierarchy_structure.0.level_two.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_two.0.name", resourceName, "hierarchy_structure.0.level_two.0.name"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_three.#", resourceName, "hierarchy_structure.0.level_three.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_three.0.name", resourceName, "hierarchy_structure.0.level_three.0.name"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_four.#", resourceName, "hierarchy_structure.0.level_four.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_four.0.name", resourceName, "hierarchy_structure.0.level_four.0.name"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_five.#", resourceName, "hierarchy_structure.0.level_five.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_structure.0.level_five.0.name", resourceName, "hierarchy_structure.0.level_five.0.name"),
				),
			},
		},
	})
}

func testAccUserHierarchyStructureBaseDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = "levelone"
    }

    level_two {
      name = "leveltwo"
    }

    level_three {
      name = "levelthree"
    }

    level_four {
      name = "levelfour"
    }

    level_five {
      name = "levelfive"
    }
  }
}
`, rName)
}

func testAccUserHierarchyStructureDataSourceConfig_instanceID(rName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseDataSourceConfig(rName),
		`
data "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}
`)
}
