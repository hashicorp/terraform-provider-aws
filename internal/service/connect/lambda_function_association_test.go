package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectLambdaFunctionAssociation_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccLambdaFunctionAssociation_basic,
		"disappears": testAccLambdaFunctionAssociation_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccLambdaFunctionAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_lambda_function_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLambdaFunctionAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLambdaFunctionAssociationConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaFunctionAssociationExists(resourceName),
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

func testAccLambdaFunctionAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_lambda_function_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLambdaFunctionAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLambdaFunctionAssociationConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaFunctionAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceLambdaFunctionAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLambdaFunctionAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_lambda_function_association" {
			continue
		}

		instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		lfaArn, err := tfconnect.FindLambdaFunctionAssociationByARNWithContext(context.Background(), conn, instanceID, functionArn)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if lfaArn != "" {
			return fmt.Errorf("Connect Lambda Function Association (%s): still exists", functionArn)
		}
	}
	return nil
}

func testAccCheckLambdaFunctionAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Lambda Function Association not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("error Connect Lambda Function Association ID not set")
		}

		instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		lfaArn, err := tfconnect.FindLambdaFunctionAssociationByARNWithContext(context.Background(), conn, instanceID, functionArn)

		if err != nil {
			return fmt.Errorf("error finding Connect Lambda Function Association by Function Arn (%s): %w", functionArn, err)
		}

		if lfaArn == "" {
			return fmt.Errorf("error finding Connect Lambda Function Association by Function Arn (%s): not found", functionArn)
		}

		return nil
	}
}

func testAccLambdaFunctionAssociationConfigBase(rName string, rName2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs14.x"
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
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
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

func testAccLambdaFunctionAssociationConfig_basic(rName string, rName2 string) string {
	return acctest.ConfigCompose(
		testAccLambdaFunctionAssociationConfigBase(rName, rName2), `
resource "aws_connect_lambda_function_association" "test" {
  instance_id  = aws_connect_instance.test.id
  function_arn = aws_lambda_function.test.arn
}
`)
}
