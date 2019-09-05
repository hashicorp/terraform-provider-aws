package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsServerlessRepositoryApplication_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-basic-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryApplicationConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryApplicationExists("aws_serverlessrepository_application.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet("aws_serverlessrepository_application.postgres-rotator", "semantic_version"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "parameters.endpoint", "secretsmanager.us-east-2.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "outputs.%", "1"),
					resource.TestCheckResourceAttrSet("aws_serverlessrepository_application.postgres-rotator", "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "capabilities.#", "1"),
					//resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "capabilities.0", "CAPABILITY_NAMED_IAM"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryApplication_versioned(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-versioned-%s", acctest.RandString(10))
	const version = "1.0.15"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryApplicationConfig_versioned(stackName, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryApplicationExists("aws_serverlessrepository_application.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "semantic_version", version),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryApplication_updateVersion(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-update-%s", acctest.RandString(10))
	const initialVersion = "1.0.15"
	const updateVersion = "1.0.36"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryApplicationConfig_versioned(stackName, initialVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryApplicationExists("aws_serverlessrepository_application.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "semantic_version", initialVersion),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryApplicationConfig_versioned(stackName, updateVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryApplicationExists("aws_serverlessrepository_application.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "semantic_version", updateVersion),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryApplication_updateFunctionName(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-update-name-%s", acctest.RandString(10))
	const initialName = "FuncName1"
	const updatedName = "FuncName2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryApplicationConfig_functionName(stackName, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryApplicationExists("aws_serverlessrepository_application.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "parameters.functionName", initialName),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryApplicationConfig_functionName(stackName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryApplicationExists("aws_serverlessrepository_application.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_application.postgres-rotator", "parameters.functionName", updatedName),
				),
			},
		},
	})
}

func testAccAwsServerlessRepositoryApplicationConfig(stackName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_application" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.us-east-2.amazonaws.com"
  }
}`, stackName)
}

func testAccAWSServerlessRepositoryApplicationConfig_functionName(stackName, functionName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_application" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.us-east-2.amazonaws.com"
  }
}`, stackName, functionName)
}

func testAccAWSServerlessRepositoryApplicationConfig_versioned(stackName, version string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_application" "postgres-rotator" {
  name             = "%[1]s"
  application_id   = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  semantic_version = "%[2]s"
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.us-east-2.amazonaws.com"
  }
}`, stackName, version)
}

func testAccCheckerverlessRepositoryApplicationExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cfconn
		params := &cloudformation.DescribeStacksInput{
			StackName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeStacks(params)
		if err != nil {
			return err
		}
		if len(resp.Stacks) == 0 {
			return fmt.Errorf("CloudFormation stack not found")
		}

		return nil
	}
}
