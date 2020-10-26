package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAwsServerlessRepositoryStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test-basic")

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(resourceName, "semantic_version"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					resource.TestCheckResourceAttr(resourceName, "parameters.endpoint", "secretsmanager.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_versioned(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test-versioned")
	const version = "1.0.15"

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_tagged(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test-tagged")

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryStackConfig_tagged(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.MyTag", "My value"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_versionUpdate(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test-update")
	const initialVersion = "1.0.15"
	const updateVersion = "1.0.36"

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, initialVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", initialVersion),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, updateVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", updateVersion),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_update(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test-update-name")
	const initialName = "FuncName1"
	const updatedName = "FuncName2"

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryStackConfig_updateInitial(stackName, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", initialName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.ToDelete", "ToBeDeleted"),
					resource.TestCheckResourceAttr(resourceName, "tags.ToUpdate", "InitialValue"),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryStackConfig_updateUpdated(stackName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", updatedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.ToUpdate", "UpdatedValue"),
					resource.TestCheckResourceAttr(resourceName, "tags.ToAdd", "AddedValue"),
				),
			},
		},
	})
}

func testAccAwsServerlessRepositoryStackConfig(stackName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities   = ["CAPABILITY_IAM"]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
}`, stackName)
}

func testAccAWSServerlessRepositoryStackConfig_updateInitial(stackName, functionName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities   = ["CAPABILITY_IAM"]
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
  tags = {
    ToDelete = "ToBeDeleted"
    ToUpdate = "InitialValue"
  }
}`, stackName, functionName)
}

func testAccAWSServerlessRepositoryStackConfig_updateUpdated(stackName, functionName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities   = ["CAPABILITY_IAM"]
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
  tags = {
    ToUpdate = "UpdatedValue"
    ToAdd    = "AddedValue"
  }
}`, stackName, functionName)
}

func testAccAWSServerlessRepositoryStackConfig_versioned(stackName, version string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name             = "%[1]s"
  application_id   = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  semantic_version = "%[2]s"
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
}`, stackName, version)
}

func testAccAwsServerlessRepositoryStackConfig_tagged(stackName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
  tags = {
    MyTag = "My value"
  }
}`, stackName)
}

func testAccCheckServerlessRepositoryStackExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
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
			return fmt.Errorf("CloudFormation stack (%s) not found", rs.Primary.ID)
		}

		*stack = *resp.Stacks[0]

		return nil
	}
}
