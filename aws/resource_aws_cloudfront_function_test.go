package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	RegisterServiceErrorCheckFunc(cloudfront.EndpointsID, testAccErrorCheckSkipFunction)

	resource.AddTestSweepers("aws_cloudfront_function", &resource.Sweeper{
		Name: "aws_cloudfront_function",
		F:    testSweepCloudfrontFunctions,
	})
}

func testAccErrorCheckSkipFunction(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"InvalidParameterValueException: Unsupported source arn",
	)
}

func testSweepCloudfrontFunctions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	cloudfrontconn := client.(*AWSClient).cloudfrontconn

	resp, err := cloudfrontconn.ListFunctions(&cloudfront.ListFunctionsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Cloudfront Function sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Cloudfront Functions: %s", err)
	}

	if len(resp.FunctionList.Items) == 0 {
		log.Print("[DEBUG] No aws Cloudfront Functions to sweep")
		return nil
	}

	for _, f := range resp.FunctionList.Items {
		describeParams := &cloudfront.DescribeFunctionInput{
			Name: f.Name,
		}
		DescribeFunctionOutput, err := cloudfrontconn.DescribeFunction(describeParams)
		if err != nil {
			return err
		}
		_, delerr := cloudfrontconn.DeleteFunction(
			&cloudfront.DeleteFunctionInput{
				Name:    f.Name,
				IfMatch: DescribeFunctionOutput.ETag,
			})
		if delerr != nil {
			return delerr
		}
	}

	return nil
}

func TestAccAWSCloudfrontFunction_basic(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	var getconf cloudfront.GetFunctionOutput
	resourceName := "aws_cloudfront_function.test"

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_cloudfront_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, funcName, &conf, &getconf),
					testAccCheckAwsCloudfrontFunctionName(&conf, funcName),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", funcName)),
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

func TestAccAWSCloudfrontFunction_disappears(t *testing.T) {
	var function cloudfront.DescribeFunctionOutput
	var getconf cloudfront.GetFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, rName, &function, &getconf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_codeUpdate(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	var getconf cloudfront.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_cloudfront_func_code_upd_%s", rString)
	resourceName := "aws_cloudfront_function.test"

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, funcName, &conf, &getconf),
					testAccCheckAwsCloudfrontFunctionName(&conf, funcName),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", funcName)),
					testAccCheckAwsCloudfrontFunctionETag(&getconf, "ETVPDKIKX0DER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudfrontConfigCodeUpdate(funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, funcName, &conf, &getconf),
					testAccCheckAwsCloudfrontFunctionName(&conf, funcName),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", funcName)),
					testAccCheckAwsCloudfrontFunctionETag(&getconf, "E3UN6WX5RRO2AG"),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Update_nameOnly(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	var getconf cloudfront.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_cloudfront_func_local_upd_name_%s", rString)
	funcNameUpdate := funcName + "_update"
	resourceName := "aws_cloudfront_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, funcName, &conf, &getconf),
					testAccCheckAwsCloudfrontFunctionName(&conf, funcName),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", funcName)),
					testAccCheckAwsCloudfrontFunctionETag(&getconf, "ETVPDKIKX0DER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudfrontConfigBasic(funcNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, funcNameUpdate, &conf, &getconf),
					testAccCheckAwsCloudfrontFunctionName(&conf, funcNameUpdate),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", funcNameUpdate)),
					testAccCheckAwsCloudfrontFunctionETag(&getconf, "ETVPDKIKX0DER"),
				),
			},
		},
	})
}

func testAccCheckCloudfrontFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_function" {
			continue
		}

		_, err := conn.GetFunction(&cloudfront.GetFunctionInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Cloudfront Function still exists")
		}

	}

	return nil

}

func testAccCheckAwsCloudfrontFunctionExists(res, funcName string, function *cloudfront.DescribeFunctionOutput, getfunction *cloudfront.GetFunctionOutput) resource.TestCheckFunc {
	// Wait for IAM role
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Cloudfront Function not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Cloudfront Function ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		params := &cloudfront.DescribeFunctionInput{
			Name: aws.String(funcName),
		}

		describeFunction, err := conn.DescribeFunction(params)
		if err != nil {
			return err
		}

		*function = *describeFunction

		getparams := &cloudfront.GetFunctionInput{
			Name: aws.String(funcName),
		}

		getFunction, geterr := conn.GetFunction(getparams)
		if geterr != nil {
			return geterr
		}

		*getfunction = *getFunction
		return nil
	}
}

func testAccCheckAwsCloudfrontFunctionName(function *cloudfront.DescribeFunctionOutput, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.FunctionSummary
		if *c.Name != expectedName {
			return fmt.Errorf("Expected function name %s, got %s", expectedName, *c.Name)
		}

		return nil
	}
}

func testAccCheckAwsCloudfrontFunctionETag(function *cloudfront.GetFunctionOutput, expectedETag string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *function.ETag != expectedETag {
			return fmt.Errorf("Expected code ETag %s, got %s", expectedETag, *function.ETag)
		}

		return nil
	}
}

func testAccAWSCloudfrontConfigBasic(funcName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = "%s"
  runtime = "cloudfront-js-1.0"
  comment = "%s"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, funcName, funcName)
}

func testAccAWSCloudfrontConfigCodeUpdate(funcName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = "%s"
  runtime = "cloudfront-js-1.0"
  comment = "%s"
  code    = <<-EOT
function handler(event) {
	// updated code
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, funcName, funcName)
}
