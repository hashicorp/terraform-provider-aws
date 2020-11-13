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

// Since aws_serverlessrepository_stack creates CloudFormation stacks,
// the aws_cloudformation_stack sweeper will clean these up as well.

func TestAccAwsServerlessRepositoryStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

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
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "application_id", "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(resourceName, "semantic_version"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					resource.TestCheckResourceAttr(resourceName, "parameters.endpoint", "secretsmanager.us-west-2.amazonaws.com"), // FIXME
					resource.TestCheckResourceAttr(resourceName, "outputs.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_disappears(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServerlessRepositoryStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_versioned(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	const (
		version1 = "1.0.15"
		version2 = "1.1.78"
	)

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versioned2(stackName, version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version2),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
			{
				// Confirm removal of "CAPABILITY_RESOURCE_POLICY" is handled properly
				Config: testAccAWSServerlessRepositoryStackConfig_versioned(stackName, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_paired(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	const version = "1.1.78"

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessRepositoryStackConfig_versionedPaired(stackName, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_Tags(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_serverlessrepository_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessRepositoryStackConfigTags1(stackName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAwsServerlessRepositoryStackConfigTags2(stackName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsServerlessRepositoryStackConfigTags1(stackName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsServerlessRepositoryStack_update(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	initialName := acctest.RandomWithPrefix("FuncName1")
	updatedName := acctest.RandomWithPrefix("FuncName2")

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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
			{
				Config: testAccAWSServerlessRepositoryStackConfig_updateUpdated(stackName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessRepositoryStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", updatedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
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

func testAccAwsServerlessRepositoryStackConfig(stackName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, stackName)
}

func testAccAWSServerlessRepositoryStackConfig_updateInitial(stackName, functionName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    key = "value"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, stackName, functionName)
}

func testAccAWSServerlessRepositoryStackConfig_updateUpdated(stackName, functionName string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    key = "value"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, stackName, functionName)
}

func testAccAWSServerlessRepositoryStackConfig_versioned(stackName, version string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name             = "%[1]s"
  application_id   = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  semantic_version = "%[2]s"
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, stackName, version)
}

func testAccAWSServerlessRepositoryStackConfig_versioned2(stackName, version string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  semantic_version = "%[2]s"
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, stackName, version)
}

func testAccAWSServerlessRepositoryStackConfig_versionedPaired(stackName, version string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name             = "%[1]s"
  application_id   = data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator.application_id
  semantic_version = data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator.semantic_version
  capabilities     = data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator.required_capabilities
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  semantic_version = "%[2]s"
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, stackName, version)
}

func testAccAwsServerlessRepositoryStackConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    %[2]q = %[3]q
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName, tagKey1, tagValue1)
}

func testAccAwsServerlessRepositoryStackConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_serverlessrepository_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
