package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSSsmParametersDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_parameters.test"
	path := "/tf-acc-test" + acctest.RandString(10) + "/"
	name := path + "TestKey"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParametersDataSourceConfig(name, path),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "values", name),
				),
			},
			//{
			//Config: testAccCheckAwsSsmParametersDataSourceConfig(path, "na"),
			//Check: resource.ComposeAggregateTestCheckFunc(
			//resource.TestCheckResourceAttr(resourceName, "value", ""),
			//),
			//},
		},
	})
}

func testAccCheckAwsSsmParametersDataSourceConfig(name string, path string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = "%s"
  type  = "String"
  value = "TestValue"
}

data "aws_ssm_parameters" "test" {
	depends_on = [aws_ssm_parameter.test]
	path     = "%s"
}
`, name, path)
}
