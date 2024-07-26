// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfserverlessrepo "github.com/hashicorp/terraform-provider-aws/internal/service/serverlessrepo"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Since aws_serverlessapplicationrepository_cloudformation_stack creates CloudFormation stacks,
// the aws_cloudformation_stack sweeper will clean these up as well.

func TestAccServerlessRepoCloudFormationStack_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stack cloudformationtypes.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCloudFormationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_basic(stackName, appARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, stackName),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, names.AttrApplicationID, "serverlessrepo", "applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(resourceName, "semantic_version"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					acctest.CheckResourceAttrRegionalHostnameService(resourceName, "parameters.endpoint", "secretsmanager"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var stack cloudformationtypes.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_basic(stackName, appARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfserverlessrepo.ResourceCloudFormationStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_versioned(t *testing.T) {
	ctx := acctest.Context(t)
	var stack1, stack2, stack3 cloudformationtypes.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	const (
		version1 = "1.1.465"
		version2 = "1.1.88"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCloudFormationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_versioned(stackName, appARN, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack1),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct2),
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
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack2),
					testAccCheckCloudFormationStackNotRecreated(&stack1, &stack2),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version2),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
			{
				// Confirm removal of "CAPABILITY_RESOURCE_POLICY" is handled properly
				Config: testAccCloudFormationStackConfig_versioned(stackName, appARN, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack3),
					testAccCheckCloudFormationStackNotRecreated(&stack2, &stack3),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_paired(t *testing.T) {
	ctx := acctest.Context(t)
	var stack cloudformationtypes.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	const version = "1.1.465"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCloudFormationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_versionedPaired(stackName, appARN, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var stack cloudformationtypes.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCloudFormationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_tags1(stackName, appARN, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudFormationStackConfig_tags2(stackName, appARN, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCloudFormationStackConfig_tags1(stackName, appARN, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccServerlessRepoCloudFormationStack_update(t *testing.T) {
	ctx := acctest.Context(t)
	var stack cloudformationtypes.Stack
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	initialName := sdkacctest.RandomWithPrefix("FuncName1")
	updatedName := sdkacctest.RandomWithPrefix("FuncName2")
	appARN := testAccCloudFormationApplicationID()
	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServerlessRepoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCloudFormationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudFormationStackConfig_updateInitial(stackName, appARN, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, names.AttrApplicationID, "serverlessrepo", "applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", initialName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.key", names.AttrValue),
				),
			},
			{
				Config: testAccCloudFormationStackConfig_updateUpdated(stackName, appARN, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", updatedName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.key", names.AttrValue),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackExists(ctx context.Context, n string, stack *cloudformationtypes.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)
		params := &cloudformation.DescribeStacksInput{
			StackName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeStacks(ctx, params)
		if err != nil {
			return err
		}
		if len(resp.Stacks) == 0 {
			return fmt.Errorf("CloudFormation stack (%s) not found", rs.Primary.ID)
		}

		*stack = resp.Stacks[0]

		return nil
	}
}

func testAccCloudFormationStackNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s%s", tfserverlessrepo.CloudFormationStackNamePrefix, rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccCloudFormationStackNameNoPrefixImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrName], nil
	}
}

func testAccCloudFormationApplicationID() string {
	arnRegion := names.USEast1RegionID
	arnAccountID := "297356227824"
	if acctest.Partition() == names.USGovCloudPartitionID {
		arnRegion = names.USGovWest1RegionID
		arnAccountID = "023102451235"
	}

	return arn.ARN{
		Partition: acctest.Partition(),
		Service:   names.ServerlessRepo,
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

func testAccCheckAMIDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ami" {
				continue
			}

			// Try to find the AMI
			log.Printf("AMI-ID: %s", rs.Primary.ID)
			DescribeAmiOpts := &ec2.DescribeImagesInput{
				ImageIds: []string{rs.Primary.ID},
			}
			resp, err := conn.DescribeImages(ctx, DescribeAmiOpts)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidAMIID", "NotFound") {
					log.Printf("[DEBUG] AMI not found, passing")
					return nil
				}
				return err
			}

			if len(resp.Images) > 0 {
				state := resp.Images[0].State
				return fmt.Errorf("AMI %s still exists in the state: %s.", aws.ToString(resp.Images[0].ImageId),
					string(state))
			}
		}
		return nil
	}
}

func testAccCheckCloudFormationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack" {
				continue
			}

			params := cloudformation.DescribeStacksInput{
				StackName: aws.String(rs.Primary.ID),
			}

			resp, err := conn.DescribeStacks(ctx, &params)

			if err != nil {
				return err
			}

			for _, s := range resp.Stacks {
				if aws.ToString(s.StackId) == rs.Primary.ID && s.StackStatus != cloudformationtypes.StackStatusDeleteComplete {
					return fmt.Errorf("CloudFormation stack still exists: %q", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckCloudFormationStackNotRecreated(i, j *cloudformationtypes.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StackId) != aws.ToString(j.StackId) {
			return fmt.Errorf("CloudFormation stack recreated")
		}

		return nil
	}
}
