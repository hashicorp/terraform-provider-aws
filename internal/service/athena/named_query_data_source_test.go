package athena_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/athena"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAthenaNamedQueryDataSource_basic(t *testing.T) {
	var output athena.NamedQuery
	resourceName := "aws_athena_named_query.test"
	dataSourceName := "data.aws_athena_named_query.test"
	rName := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaNamedQueryDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamedQueryExists(resourceName, &output),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccAthenaNamedQueryDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccNamedQueryConfig(sdkacctest.RandInt(), rName),
		`
data "aws_athena_named_query" "test" {
  name = aws_athena_named_query.test.name
}
`)
}
