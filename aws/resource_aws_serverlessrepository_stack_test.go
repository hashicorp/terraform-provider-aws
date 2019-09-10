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

func TestAccAwsServerlessRepositoryStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-basic-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet("aws_serverlessrepository_stack.postgres-rotator", "semantic_version"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "parameters.%", "2"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "parameters.endpoint", "secretsmanager.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "outputs.%", "1"),
					resource.TestCheckResourceAttrSet("aws_serverlessrepository_stack.postgres-rotator", "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "capabilities.#", "1"),
					//resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "capabilities.0", "CAPABILITY_NAMED_IAM"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_versioned(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-versioned-%s", acctest.RandString(10))
	const version = "1.0.15"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "semantic_version", version),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_tagged(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-tagged-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryStackConfig_tagged(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.MyTag", "My value"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_versionUpdate(t *testing.T) {
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
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, initialVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "semantic_version", initialVersion),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, updateVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "semantic_version", updateVersion),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_update(t *testing.T) {
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
				Config: testAccAWSServerlessRepositoryStackConfig_updateInitial(stackName, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "parameters.functionName", initialName),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.ToDelete", "ToBeDeleted"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.ToUpdate", "InitialValue"),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryStackConfig_updateUpdated(stackName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckerverlessRepositoryStackExists("aws_serverlessrepository_stack.postgres-rotator", &stack),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "parameters.functionName", updatedName),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.ToUpdate", "UpdatedValue"),
					resource.TestCheckResourceAttr("aws_serverlessrepository_stack.postgres-rotator", "tags.ToAdd", "AddedValue"),
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
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
  tags = {
    MyTag = "My value"
  }
}`, stackName)
}

func testAccCheckerverlessRepositoryStackExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
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
