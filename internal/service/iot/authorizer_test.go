package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIoTAuthorizer_basic(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enable_caching_for_http", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "signing_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "token_key_name", "Token-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "token_signing_public_keys.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token_signing_public_keys.Key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIoTAuthorizer_disappears(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourceAuthorizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTAuthorizer_signingDisabled(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_signingDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "signing_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", "INACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "token_key_name", ""),
					resource.TestCheckResourceAttr(resourceName, "token_signing_public_keys.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIoTAuthorizer_update(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enable_caching_for_http", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "signing_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "token_key_name", "Token-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "token_signing_public_keys.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token_signing_public_keys.Key1"),
				),
			},
			{
				Config: testAccAuthorizerConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(resourceName, &conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enable_caching_for_http", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "signing_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "INACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "token_key_name", "Token-Header-2"),
					resource.TestCheckResourceAttr(resourceName, "token_signing_public_keys.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "token_signing_public_keys.Key1"),
					resource.TestCheckResourceAttrSet(resourceName, "token_signing_public_keys.Key2"),
				),
			},
		},
	})
}

func testAccCheckAuthorizerExists(n string, v *iot.AuthorizerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Authorizer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		output, err := tfiot.FindAuthorizerByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAuthorizerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_authorizer" {
			continue
		}

		_, err := tfiot.FindAuthorizerByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IoT Authorizer %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAuthorizerBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
}
`, rName)
}

func testAccAuthorizerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                    = %[1]q
  authorizer_function_arn = aws_lambda_function.test.arn
  token_key_name          = "Token-Header-1"

  token_signing_public_keys = {
    Key1 = "${file("test-fixtures/iot-authorizer-signing-key.pem")}"
  }
}
`, rName))
}

func testAccAuthorizerConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                    = %[1]q
  authorizer_function_arn = aws_lambda_function.test.arn
  signing_disabled        = false
  token_key_name          = "Token-Header-2"
  status                  = "INACTIVE"
  enable_caching_for_http = true

  token_signing_public_keys = {
    Key1 = "${file("test-fixtures/iot-authorizer-signing-key.pem")}"
    Key2 = "${file("test-fixtures/iot-authorizer-signing-key.pem")}"
  }
}
`, rName))
}

func testAccAuthorizerConfig_signingDisabled(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                    = %[1]q
  authorizer_function_arn = aws_lambda_function.test.arn
  signing_disabled        = true
  status                  = "INACTIVE"
}
`, rName))
}
