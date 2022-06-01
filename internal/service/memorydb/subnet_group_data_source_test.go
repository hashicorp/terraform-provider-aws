package memorydb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMemoryDBSubnetGroupDataSource_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"
	dataSourceName := "data.aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "subnet_ids.*", resourceName, "subnet_ids.0"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "subnet_ids.*", resourceName, "subnet_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Test", resourceName, "tags.Test"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccSubnetGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id

  tags = {
    Test = "test"
  }
}

data "aws_memorydb_subnet_group" "test" {
  name = aws_memorydb_subnet_group.test.name
}
`, rName),
	)
}
