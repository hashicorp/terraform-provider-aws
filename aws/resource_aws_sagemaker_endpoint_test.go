package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_endpoint", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint",
		Dependencies: []string{
			"aws_sagemaker_model",
			"aws_sagemaker_endpoint_configuration",
		},
		F: testSweepSagemakerEndpoints,
	})
}

func testSweepSagemakerEndpoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListEndpointsInput{
		NameContains: aws.String("tf-acc-test"),
	}
	resp, err := conn.ListEndpoints(req)
	if err != nil {
		return fmt.Errorf("error listing endpoints: %s", err)
	}

	if len(resp.Endpoints) == 0 {
		log.Print("[DEBUG] No SageMaker Endpoint to sweep")
		return nil
	}

	for _, endpoint := range resp.Endpoints {
		_, err := conn.DeleteEndpoint(&sagemaker.DeleteEndpointInput{
			EndpointName: endpoint.EndpointName,
		})
		if err != nil {
			return fmt.Errorf(
				"error deleting SageMaker Endpoint (%s): %s",
				*endpoint.EndpointName, err)
		}
	}

	return nil
}

func TestAccAWSSagemakerEndpoint_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config_name", rName),
				),
				ExpectError: regexp.MustCompile(`ResourceNotReady: failed waiting for successful resource state`),
			},
			// Does not work with failing resource
			//{
			//	ResourceName:      resourceName,
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
		},
	})
}

func TestAccAWSSagemakerEndpoint_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
				ExpectError: regexp.MustCompile(`ResourceNotReady: failed waiting for successful resource state`),
			},
			{
				Config: testAccSagemakerEndpointConfig_Tags_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.bar", "baz"),
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

func TestAccAWSSagemakerEndpoint_update(t *testing.T) {
	// Cannot update failed endpoint
	t.Skip()

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config_name", rName),
				),
				ExpectError: regexp.MustCompile(`ResourceNotReady: failed waiting for successful resource state`),
			},
			{
				Config: testAccSagemakerEndpointConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config_name", rNameUpdated),
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

func testAccCheckSagemakerEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint" {
			continue
		}

		describeInput := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeEndpoint(describeInput)

		if isAWSErr(err, "ValidationException", "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Endpoint (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckSagemakerEndpointExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker Endpoint ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeEndpoint(opts)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccSagemakerEndpointConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
	name = %q

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.test.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "test" {
	name = %q
	execution_role_arn = "${aws_iam_role.test.arn}"

	primary_container {
		image = %q
	}
}

resource "aws_iam_role" "test" {
	name = %q
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, rName, rName, image, rName)
}

func testAccSagemakerEndpointConfig(rName string) string {
	return testAccSagemakerEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
	name = %q
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.test.name}"
}
`, rName)
}

func testAccSagemakerEndpointConfig_Update(rName string) string {
	return testAccSagemakerEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
	name = %q
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.test_updated.name}"
}
`, rName)
}

func testAccSagemakerEndpointConfig_Tags(rName string) string {
	return testAccSagemakerEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
	name = %q
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.test.name}"

	tags = {
		foo = "bar"
	}
}
`, rName)
}

func testAccSagemakerEndpointConfig_Tags_Update(rName string) string {
	return testAccSagemakerEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
	name = %q
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.test.name}"

	tags = {
		bar = "baz"
	}
}
`, rName)
}
