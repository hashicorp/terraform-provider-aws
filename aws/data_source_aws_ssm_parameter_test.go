package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSSsmParameterDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_parameter.test"
	name := "test.parameter"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:parameter/%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
				),
			},
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:parameter/%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "true"),
				),
			},
		},
	})
}

func TestAccAWSSsmParameterDataSource_defaultValue(t *testing.T) {
	resourceName := "data.aws_ssm_parameter.test"
	name := "doesnotexist"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfigNotExist(name, "false", "true", "foo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arn", ""),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "foo"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "with_default", "true"),
				),
			},
		},
	})
}

func TestAccAWSSsmParameterDataSource_fullPath(t *testing.T) {
	resourceName := "data.aws_ssm_parameter.test"
	name := "/path/parameter"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:parameter%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsSsmParameterDataSourceConfig(name string, withDecryption string) string {
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
`, name, withDecryption)
}

func testAccCheckAwsSsmParameterDataSourceConfigNotExist(name string, withDecryption string, withDefault string, defaultValue string) string {
	return fmt.Sprintf(`
data "aws_ssm_parameter" "test" {
	name = "%s"
	with_decryption = %s
	with_default = %s
	default = "%s"
}
`, name, withDecryption, withDefault, defaultValue)
}
