package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "code"),
					resource.TestCheckResourceAttr(resourceName, "comment", ""),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "runtime", "cloudfront-js-1.0"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_disappears(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Publish(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNPUBLISHED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
			{
				Config: testAccAWSCloudfrontConfigBasic(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Update_Code(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
				),
			},
			{
				Config: testAccAWSCloudfrontConfigCodeUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "E3UN6WX5RRO2AG"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Update_Comment(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigComment(rName, "test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "comment", "test 1"),
				),
			},
			{
				Config: testAccAWSCloudfrontConfigComment(rName, "test 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "comment", "test 2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
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

		_, err := finder.FunctionByNameAndStage(conn, rs.Primary.ID, cloudfront.FunctionStageDevelopment)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Function %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckAwsCloudfrontFunctionExists(n string, v *cloudfront.DescribeFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Cloudfront Function not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Cloudfront Function ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		output, err := finder.FunctionByNameAndStage(conn, rs.Primary.ID, cloudfront.FunctionStageDevelopment)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSCloudfrontConfigBasic(rName, publish string) string {
	if publish == "" {
		publish = "null"
	}

	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
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

  publish = %[2]s
}
`, rName, publish)
}

func testAccAWSCloudfrontConfigCodeUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
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
`, rName)
}

func testAccAWSCloudfrontConfigComment(rName, comment string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  comment = %[2]q
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
`, rName, comment)
}
