// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFormationStackSetInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1 awstypes.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance1),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_id", ""),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "stack_id"),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
			// Test import with call_as.
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}

					return fmt.Sprintf("%s,SELF", rs.Primary.ID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1 awstypes.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackSetInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_Disappears_stackSet(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1 awstypes.StackInstance
	var stackSet1 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	stackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, stackSetResourceName, &stackSet1),
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackSetInstance(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackSet(), stackSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_parameterOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1, stackInstance2, stackInstance3, stackInstance4 awstypes.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_parameterOverrides1(rName, "overridevalue1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
			{
				Config: testAccStackSetInstanceConfig_parameterOverrides2(rName, "overridevalue1updated", "overridevalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance2),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter2", "overridevalue2"),
				),
			},
			{
				Config: testAccStackSetInstanceConfig_parameterOverrides1(rName, "overridevalue1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance3),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
				),
			},
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, resourceName, &stackInstance4),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance3, &stackInstance4),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_deploymentTargets(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_deploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.account_filter_type", "INTERSECTION"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.accounts_url", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
					"deployment_targets",
				},
			},
			{
				Config: testAccStackSetInstanceConfig_deploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_DeploymentTargets_emptyOU(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_DeploymentTargets_emptyOU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
					"deployment_targets",
				},
			},
			{
				Config: testAccStackSetInstanceConfig_DeploymentTargets_emptyOU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_operationPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_operationPreferences(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.concurrency_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_concurrencyMode(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_concurrencyMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.concurrency_mode", "SOFT_FAILURE_TOLERANCE"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/32536.
func TestAccCloudFormationStackSetInstance_delegatedAdministrator(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/member.org.stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccStackSetConfig_delegatedAdministratorInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
					acctest.PreCheckIAMServiceLinkedRoleWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers), "/aws-service-role/stacksets.cloudformation.amazonaws.com")
				},
				Config: testAccStackSetInstanceConfig_delegatedAdministrator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}

					return fmt.Sprintf("%s,DELEGATED_ADMIN", rs.Primary.ID), nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
		},
	})
}

func testAccCheckStackSetInstanceExists(ctx context.Context, resourceName string, v *awstypes.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
		if err != nil {
			return err
		}

		stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
		callAs := rs.Primary.Attributes["call_as"]
		output, err := tfcloudformation.FindStackInstanceByFourPartKey(ctx, conn, stackSetName, accountOrOrgID, region, callAs)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckStackSetInstanceForOrganizationalUnitExists is a variant of the
// standard CheckExistsFunc which expects the resource ID to contain organizational
// unit IDs rather than an account ID
func testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx context.Context, resourceName string, v []awstypes.StackInstanceSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
		if err != nil {
			return err
		}

		stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
		orgIDs := strings.Split(accountOrOrgID, "/")
		callAs := rs.Primary.Attributes["call_as"]
		output, err := tfcloudformation.FindStackInstanceSummariesByFourPartKey(ctx, conn, stackSetName, region, callAs, orgIDs)

		if err != nil {
			return err
		}

		v = output

		return nil
	}
}

// testAccCheckStackSetInstanceForOrganizationalUnitDestroy is a variant of the
// standard CheckDestroyFunc which expects the resource ID to contain organizational
// unit IDs rather than an account ID
func testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set_instance" {
				continue
			}

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
			if err != nil {
				return err
			}

			stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
			orgIDs := strings.Split(accountOrOrgID, "/")
			callAs := rs.Primary.Attributes["call_as"]
			output, err := tfcloudformation.FindStackInstanceSummariesByFourPartKey(ctx, conn, stackSetName, region, callAs, orgIDs)

			if tfresource.NotFound(err) {
				continue
			}
			if len(output) == 0 {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation StackSet Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackSetInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set_instance" {
				continue
			}

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
			if err != nil {
				return err
			}

			stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
			callAs := rs.Primary.Attributes["call_as"]
			_, err = tfcloudformation.FindStackInstanceByFourPartKey(ctx, conn, stackSetName, accountOrOrgID, region, callAs)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation StackSet Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackSetInstanceNotRecreated(i, j *awstypes.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StackId) != aws.ToString(j.StackId) {
			return fmt.Errorf("CloudFormation StackSet Instance (%s,%s,%s) recreated", aws.ToString(i.StackSetId), aws.ToString(i.Account), aws.ToString(i.Region))
		}

		return nil
	}
}

func testAccStackSetInstanceBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "Administration" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Administration"
}

resource "aws_iam_role_policy" "Administration" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Administration.name
}

resource "aws_iam_role" "Execution" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "${aws_iam_role.Administration.arn}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Execution"
}

resource "aws_iam_role_policy" "Execution" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "*"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Execution.name
}

resource "aws_cloudformation_stack_set" "test" {
  depends_on = [aws_iam_role_policy.Execution]

  administration_role_arn = aws_iam_role.Administration.arn
  execution_role_name     = aws_iam_role.Execution.name
  name                    = %[1]q

  parameters = {
    Parameter1 = "stacksetvalue1"
    Parameter2 = "stacksetvalue2"
  }

  template_body = <<TEMPLATE
Parameters:
  Parameter1:
    Type: String
  Parameter2:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Parameter2Value:
    Value: !Ref Parameter2
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
TEMPLATE
}
`, rName)
}

func testAccStackSetInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_parameterOverrides1(rName, value1 string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1))
}

func testAccStackSetInstanceConfig_parameterOverrides2(rName, value1, value2 string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
    Parameter2 = %[2]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1, value2))
}

func testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "Administration" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Administration"
}

resource "aws_iam_role_policy" "Administration" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Administration.name
}

resource "aws_iam_role" "Execution" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "${aws_iam_role.Administration.arn}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Execution"
}

resource "aws_iam_role_policy" "Execution" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "*"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Execution.name
}

data "aws_organizations_organization" "test" {}

resource "aws_cloudformation_stack_set" "test" {
  depends_on = [data.aws_organizations_organization.test]

  name             = %[1]q
  permission_model = "SERVICE_MANAGED"

  auto_deployment {
    enabled                          = true
    retain_stacks_on_account_removal = false
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE

  lifecycle {
    ignore_changes = [administration_role_arn]
  }
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetInstanceConfig_deploymentTargets(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
    account_filter_type     = "INTERSECTION"
    accounts                = [data.aws_organizations_organization.test.non_master_accounts[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_DeploymentTargets_emptyOU(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), fmt.Sprintf(`
resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [aws_organizations_organizational_unit.test.id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, rName))
}

func testAccStackSetInstanceConfig_operationPreferences(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  operation_preferences {
    failure_tolerance_count = 1
    max_concurrent_count    = 10
  }

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_concurrencyMode(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  operation_preferences {
    failure_tolerance_count = 1
    max_concurrent_count    = 10
    concurrency_mode        = "SOFT_FAILURE_TOLERANCE"
  }

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_delegatedAdministrator(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_delegatedAdministrator(rName), `
data "aws_organizations_organization" "test" {}

resource "aws_cloudformation_stack_set_instance" "test" {
  call_as = "DELEGATED_ADMIN"

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}
