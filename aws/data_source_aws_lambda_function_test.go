package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceLambdaFunction_basic(t *testing.T) {
	rString := acctest.RandString(5)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_data_source_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_data_source_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_data_source_%s", rString)
	rName := fmt.Sprintf("tf_test_%s", rString)
	arnRegexp := regexp.MustCompile("^arn:aws:lambda:")

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceLambdaFunctionConfig_basic(rName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.func", "arn"),
					resource.TestMatchResourceAttr("data.aws_lambda_function.func", "arn", arnRegexp),
					resource.TestCheckResourceAttr("data.aws_lambda_function.func", "function_name", rName),
				),
			},
		},
	})
}

func TestAccDataSourceLambdaFunction_alias(t *testing.T) {
	rString := acctest.RandString(5)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_data_source_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_data_source_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_data_source_%s", rString)
	rName := fmt.Sprintf("tf_test_%s", rString)
	arnRegexp := regexp.MustCompile("^arn:aws:lambda:")

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceLambdaFunctionConfig_alias(rName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.func", "arn"),
					resource.TestMatchResourceAttr("data.aws_lambda_function.func", "arn", arnRegexp),
					resource.TestCheckResourceAttr("data.aws_lambda_function.func", "function_name", rName),
					resource.TestCheckResourceAttr("data.aws_lambda_function.func", "version", "1"),
				),
			},
		},
	})
}

func testAccAWSDataSourceLambdaFunctionConfig_basic(rName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs4.3"
}

data "aws_lambda_function" "func" {
	function_name = "${aws_lambda_function.lambda_function_test.function_name}"
	qualifier = "${aws_lambda_function.lambda_function_test.version}"
}
`, rName)
}

func testAccAWSDataSourceLambdaFunctionConfig_alias(rName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
		publish = true
    handler = "exports.example"
    runtime = "nodejs4.3"
}

resource "aws_lambda_alias" "lambda_alias" {
  name             = "testalias"
  function_name    = "${aws_lambda_function.lambda_function_test.arn}"
  function_version = "1"

  depends_on       = ["aws_lambda_function.lambda_function_test"]
}

data "aws_lambda_function" "func" {
	function_name = "${aws_lambda_function.lambda_function_test.function_name}"
	qualifier = "${aws_lambda_alias.lambda_alias.name}"
}
`, rName)
}
