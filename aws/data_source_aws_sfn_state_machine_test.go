package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsSfnStateMachine(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "data.aws_sfn_state_machine.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSfnStateMachineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSfnStateMachineCheck(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSfnStateMachineCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		sfnStateMachineRs, ok := s.RootModule().Resources["aws_sfn_state_machine.foo"]
		if !ok {
			return fmt.Errorf("can't find aws_sfn_state_machine.foo in state")
		}

		attr := rs.Primary.Attributes

		if attr["name"] != sfnStateMachineRs.Primary.Attributes["name"] {
			return fmt.Errorf(
				"name is %s; want %s",
				attr["name"],
				sfnStateMachineRs.Primary.Attributes["name"],
			)
		}

		return nil
	}
}

func testAccDataSourceAwsSfnStateMachineConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iam_role" "iam_for_sfn" {
  name = "iam_for_sfn_%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_sfn_state_machine" "foo" {
  name     = "test_sfn_%s"
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

  definition = <<EOF
{
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Succeed"
    }
  }
}
EOF
}

data "aws_sfn_state_machine" "test" {
  name = "${aws_sfn_state_machine.foo.name}"
}
`, rName, rName)
}
