package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func init() {
	resource.AddTestSweepers("aws_lambda_function", &resource.Sweeper{
		Name: "aws_lambda_function",
		F:    testSweepLambdaFunctions,
	})
}

func testSweepLambdaFunctions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	lambdaconn := client.(*AWSClient).lambdaconn

	resp, err := lambdaconn.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		return fmt.Errorf("Error retrieving Lambda functions: %s", err)
	}

	if len(resp.Functions) == 0 {
		log.Print("[DEBUG] No aws lambda functions to sweep")
		return nil
	}

	for _, f := range resp.Functions {
		var testOptGroup bool
		for _, testName := range []string{"tf_test"} {
			if strings.HasPrefix(*f.FunctionName, testName) {
				testOptGroup = true
			}
		}

		if !testOptGroup {
			continue
		}

		_, err := lambdaconn.DeleteFunction(
			&lambda.DeleteFunctionInput{
				FunctionName: f.FunctionName,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAccAWSLambdaFunction_importLocalFile(t *testing.T) {
	resourceName := "aws_lambda_function.lambda_function_test"

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("tf_test_%s", rSt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSLambdaConfigBasic(rName, rSt),
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_importLocalFile_VPC(t *testing.T) {
	resourceName := "aws_lambda_function.lambda_function_test"

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("tf_test_%s", rSt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSLambdaConfigWithVPC(rName, rSt),
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_importS3(t *testing.T) {
	resourceName := "aws_lambda_function.lambda_function_s3test"

	rSt := acctest.RandString(5)
	rName := fmt.Sprintf("tf_test_%s", rSt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSLambdaConfigS3(rName, rSt),
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "publish"},
			},
		},
	})
}
