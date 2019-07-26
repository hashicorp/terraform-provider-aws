package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSLambdaAlias_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rAlias := acctest.RandomWithPrefix("tf-acc-test")
	rDescription := acctest.RandomWithPrefix("tf-acc-test")

	dataSourceName := "data.aws_lambda_alias.test"
	lambdaAliasResourceName := "aws_lambda_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaAliasConfigBasic(rName, rAlias, rDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", lambdaAliasResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "invoke_arn", lambdaAliasResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "description", rDescription),
					resource.TestCheckResourceAttr(dataSourceName, "function_version", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLambdaAliasConfigBasic(rName, rAlias, rDescription string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = "%s"
  handler          = "exports.example"
  runtime          = "nodejs8.10"
  source_code_hash = "${filebase64hash256("test-fixtures/lambdatest.zip")}"
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = "%s"
  description      = "%s"
  function_name    = "${aws_lambda_function.test.function_name}"
  function_version = "1"
}

data "aws_lambda_alias" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
  alias         = "${aws_lambda_alias.test.arn}"
}
  `, rName, rAlias, rDescription)
}
