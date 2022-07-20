package memorydb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMemoryDBACLDataSource_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_acl.test"
	dataSourceName := "data.aws_memorydb_acl.test"
	userName1 := "tf-" + sdkacctest.RandString(8)
	userName2 := "tf-" + sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLDataSourceConfig_basic(rName, userName1, userName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "minimum_engine_version", resourceName, "minimum_engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "user_names.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "user_names.*", resourceName, "user_names.0"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "user_names.*", resourceName, "user_names.1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Test", resourceName, "tags.Test"),
				),
			},
		},
	})
}

func testAccACLDataSourceConfig_basic(rName, userName1, userName2 string) string {
	return acctest.ConfigCompose(
		testAccACLConfigUsers(userName1, userName2),
		fmt.Sprintf(`
resource "aws_memorydb_acl" "test" {
  depends_on = [aws_memorydb_user.test]

  name       = %[1]q
  user_names = [%[2]q, %[3]q]

  tags = {
    Test = "test"
  }
}

data "aws_memorydb_acl" "test" {
  name = aws_memorydb_acl.test.name
}
`, rName, userName1, userName2),
	)
}
