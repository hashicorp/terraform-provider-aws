package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccAwsConnectLambdaFunctionAssociation_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		//"basic":      testAccAwsConnectLambdaFunctionAssociation_basic,
		"disappears": testAccAwsConnectLambdaFunctionAssociation_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAwsConnectLambdaFunctionAssociation_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_lambda_function_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectLambdaFunctionAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectLambdaFunctionAssociationConfigBasic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectLambdaFunctionAssociationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "function_arn"),
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

func testAccAwsConnectLambdaFunctionAssociation_disappears(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_lambda_function_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectLambdaFunctionAssociationConfigBasic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectLambdaFunctionAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConnectLambdaFunctionAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsConnectLambdaFunctionAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Lambda Function Association not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("error Connect Lambda Function Association ID not set")
		}
		instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		LambdaFunction, err := finder.LambdaFunctionAssociationByArnWithContext(context.Background(), conn, instanceID, functionArn)

		if err != nil {
			return fmt.Errorf("error finding LambdaFunction Association by Function Arn (%s): %w", functionArn, err)
		}

		if LambdaFunction == "" {
			return fmt.Errorf("error finding LambdaFunction Association by Function Arn (%s): not found", functionArn)
		}

		return nil
	}
}

func testAccCheckAwsConnectLambdaFunctionAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_lambda_function_association" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Lambda Function Association ID not set")
		}
		instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		LambdaFunction, err := finder.LambdaFunctionAssociationByArnWithContext(context.Background(), conn, instanceID, functionArn)
		if err == nil {
			return fmt.Errorf("error LambdaFunction Association by Function Arn (%s): still exists", functionArn)
		}

		if LambdaFunction != "" {
			return fmt.Errorf("error LambdaFunction Association by Function Arn (%s): still exists", functionArn)
		}
	}
	return nil
}

func testAccAwsConnectLambdaFunctionAssociationConfigBase(rName string, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
}

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

resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[2]q
  outbound_calls_enabled   = true
}
`, rName, rName2)
}

func testAccAwsConnectLambdaFunctionAssociationConfigBasic(rName string, rName2 string) string {
	return composeConfig(
		testAccAwsConnectLambdaFunctionAssociationConfigBase(rName, rName2),
		`data "aws_region" "current" {}

resource "aws_connect_lambda_function_association" "test" {
  instance_id   = aws_connect_instance.test.id
  function_arn  = aws_lambda_function.test.arn
}
`)
}
