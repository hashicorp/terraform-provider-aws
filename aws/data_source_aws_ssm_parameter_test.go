package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAwsSsmParameterDataSource_basic(t *testing.T) {
	name := "test.parameter"
	with_decryption := []string{"true", "false"}[acctest.RandIntRange(0, 2)]
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, with_decryption),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_ssm_parameter.test", "arn"),
					resource.TestCheckResourceAttr("data.aws_ssm_parameter.test", "name", name),
					resource.TestCheckResourceAttr("data.aws_ssm_parameter.test", "type", "String"),
					resource.TestCheckResourceAttr("data.aws_ssm_parameter.test", "value", "TestValue"),
					resource.TestCheckResourceAttr("data.aws_ssm_parameter.test", "with_decryption", with_decryption),
				),
			},
		},
	})
}

func testAccCheckAwsSsmParameterDataSourceConfig(name string, with_decryption string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
	name = "%s"
	type = "String"
	value = "TestValue"
}

data "aws_ssm_parameter" "test" {
	name = "${aws_ssm_parameter.test.name}"
	with_decryption = %s
}
`, name, with_decryption)
}
