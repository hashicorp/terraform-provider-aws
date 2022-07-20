package serverlessrepo_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	serverlessrepo "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfserverlessrepo "github.com/hashicorp/terraform-provider-aws/internal/service/serverlessrepo"
)

// Since aws_serverlessapplicationrepository_cloudformation_stack creates CloudFormation stacks,
// the aws_cloudformation_stack sweeper will clean these up as well.

func TestAccServerlessRepoCloudFormationStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, serverlessrepo.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_basic(stackName, appARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "application_id", "serverlessrepo", "applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(resourceName, "semantic_version"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					acctest.CheckResourceAttrRegionalHostnameService(resourceName, "parameters.endpoint", "secretsmanager"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCloudFormationStackNameImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCloudFormationStackNameNoPrefixImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_disappears(t *testing.T) {
	var stack cloudformation.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, serverlessrepo.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_basic(stackName, appARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					acctest.CheckResourceDisappears(acctest.Provider, tfserverlessrepo.ResourceCloudFormationStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_versioned(t *testing.T) {
	var stack1, stack2, stack3 cloudformation.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	const (
		version1 = "1.1.36"
		version2 = "1.1.88"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, serverlessrepo.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_versioned(stackName, appARN, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack1),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudFormationStackConfig_versioned2(stackName, appARN, version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack2),
					testAccCheckCloudFormationStackNotRecreated(&stack1, &stack2),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version2),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
			{
				// Confirm removal of "CAPABILITY_RESOURCE_POLICY" is handled properly
				Config: testAccCloudFormationStackConfig_versioned(stackName, appARN, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack3),
					testAccCheckCloudFormationStackNotRecreated(&stack2, &stack3),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_paired(t *testing.T) {
	var stack cloudformation.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	const version = "1.1.36"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, serverlessrepo.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_versionedPaired(stackName, appARN, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_tags(t *testing.T) {
	var stack cloudformation.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, serverlessrepo.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_tags1(stackName, appARN, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudFormationStackConfig_tags2(stackName, appARN, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCloudFormationStackConfig_tags1(stackName, appARN, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_update(t *testing.T) {
	var stack cloudformation.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	initialName := sdkacctest.RandomWithPrefix("FuncName1")
	updatedName := sdkacctest.RandomWithPrefix("FuncName2")
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, serverlessrepo.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_updateInitial(stackName, appARN, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "application_id", "serverlessrepo", "applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", initialName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
			{
				Config: testAccCloudFormationStackConfig_updateUpdated(stackName, appARN, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", updatedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn
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

func testAccCloudFormationStackNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s%s", tfserverlessrepo.CloudFormationStackNamePrefix, rs.Primary.Attributes["name"]), nil
	}
}

func testAccCloudFormationStackNameNoPrefixImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["name"], nil
	}
}

func testAccCloudFormationApplicationID() string {
	arnRegion := endpoints.UsEast1RegionID
	arnAccountID := "297356227824"
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		arnRegion = endpoints.UsGovWest1RegionID
		arnAccountID = "023102451235"
	}

	return arn.ARN{
		Partition: acctest.Partition(),
		Service:   serverlessrepo.ServiceName,
		Region:    arnRegion,
		AccountID: arnAccountID,
		Resource:  "applications/SecretsManagerRDSPostgreSQLRotationSingleUser",
	}.String()
}

func testAccCloudFormationStackConfig_basic(stackName, appARN string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = %[1]q
  application_id = %[2]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}
`, stackName, appARN)
}

func testAccCloudFormationStackConfig_updateInitial(stackName, appARN, functionName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = %[1]q
  application_id = %[2]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  parameters = {
    functionName = %[3]q
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }

  tags = {
    key = "value"
  }
}
`, stackName, appARN, functionName)
}

func testAccCloudFormationStackConfig_updateUpdated(stackName, appARN, functionName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = %[1]q
  application_id = %[2]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  parameters = {
    functionName = %[3]q
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }

  tags = {
    key = "value"
  }
}
`, stackName, appARN, functionName)
}

func testAccCloudFormationStackConfig_versioned(stackName, appARN, version string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name             = %[1]q
  application_id   = %[2]q
  semantic_version = %[3]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}
`, stackName, appARN, version)
}

func testAccCloudFormationStackConfig_versioned2(stackName, appARN, version string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = %[1]q
  application_id = %[2]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  semantic_version = %[3]q

  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}
`, stackName, appARN, version)
}

func testAccCloudFormationStackConfig_versionedPaired(stackName, appARN, version string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name             = %[1]q
  application_id   = data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator.application_id
  semantic_version = data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator.semantic_version
  capabilities     = data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator.required_capabilities

  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_serverlessapplicationrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = %[2]q
  semantic_version = %[3]q
}
`, stackName, appARN, version)
}

func testAccCloudFormationStackConfig_tags1(rName, appARN, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = %[1]q
  application_id = %[2]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, appARN, tagKey1, tagValue1)
}

func testAccCloudFormationStackConfig_tags2(rName, appARN, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = %[1]q
  application_id = %[2]q

  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]

  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, appARN, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCheckAMIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami" {
			continue
		}

		// Try to find the AMI
		log.Printf("AMI-ID: %s", rs.Primary.ID)
		DescribeAmiOpts := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeImages(DescribeAmiOpts)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidAMIID", "NotFound") {
				log.Printf("[DEBUG] AMI not found, passing")
				return nil
			}
			return err
		}

		if len(resp.Images) > 0 {
			state := resp.Images[0].State
			return fmt.Errorf("AMI %s still exists in the state: %s.", aws.StringValue(resp.Images[0].ImageId),
				aws.StringValue(state))
		}
	}
	return nil
}

func testAccCheckCloudFormationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack" {
			continue
		}

		params := cloudformation.DescribeStacksInput{
			StackName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeStacks(&params)

		if err != nil {
			return err
		}

		for _, s := range resp.Stacks {
			if aws.StringValue(s.StackId) == rs.Primary.ID && aws.StringValue(s.StackStatus) != cloudformation.StackStatusDeleteComplete {
				return fmt.Errorf("CloudFormation stack still exists: %q", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckCloudFormationStackNotRecreated(i, j *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackId) != aws.StringValue(j.StackId) {
			return fmt.Errorf("CloudFormation stack recreated")
		}

		return nil
	}
}
