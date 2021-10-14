package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/iot/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSIoTAuthorizer_basic(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAuthorizerConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
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

func TestAccAWSIoTAuthorizer_disappears(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAuthorizerConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIoTAuthorizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIoTAuthorizer_SigningDisabled(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAuthorizerConfigSigningDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
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

func TestAccAWSIoTAuthorizer_Update(t *testing.T) {
	var conf iot.AuthorizerDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAuthorizerConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "signing_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "token_key_name", "Token-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "token_signing_public_keys.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token_signing_public_keys.Key1"),
				),
			},
			{
				Config: testAccAWSIoTAuthorizerConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("authorizer/%s", rName)),
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

func testAccCheckAWSIoTAuthorizerExists(n string, v *iot.AuthorizerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Authorizer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn

		output, err := finder.AuthorizerByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSIoTAuthorizerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_authorizer" {
			continue
		}

		_, err := finder.AuthorizerByName(conn, rs.Primary.ID)

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

func testAccAWSIoTAuthorizerConfigBase(rName string) string {
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

func testAccAWSIoTAuthorizerConfigBasic(rName string) string {
	return composeConfig(testAccAWSIoTAuthorizerConfigBase(rName), fmt.Sprintf(`
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

func testAccAWSIoTAuthorizerConfigUpdated(rName string) string {
	return composeConfig(testAccAWSIoTAuthorizerConfigBase(rName), fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                    = %[1]q
  authorizer_function_arn = aws_lambda_function.test.arn
  signing_disabled        = false
  token_key_name          = "Token-Header-2"
  status                  = "INACTIVE"

  token_signing_public_keys = {
    Key1 = "${file("test-fixtures/iot-authorizer-signing-key.pem")}"
    Key2 = "${file("test-fixtures/iot-authorizer-signing-key.pem")}"
  }
}
`, rName))
}

func testAccAWSIoTAuthorizerConfigSigningDisabled(rName string) string {
	return composeConfig(testAccAWSIoTAuthorizerConfigBase(rName), fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                    = %[1]q
  authorizer_function_arn = aws_lambda_function.test.arn
  signing_disabled        = true
  status                  = "INACTIVE"
}
`, rName))
}
