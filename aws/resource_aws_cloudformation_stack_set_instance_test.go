package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudformation/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_cloudformation_stack_set_instance", &resource.Sweeper{
		Name: "aws_cloudformation_stack_set_instance",
		F:    testSweepCloudformationStackSetInstances,
	})
}

func testSweepCloudformationStackSetInstances(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cfconn
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*testSweepResource, 0)

	err = conn.ListStackSetsPages(input, func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, summary := range page.Summaries {
			input := &cloudformation.ListStackInstancesInput{
				StackSetName: summary.StackSetName,
			}

			err = conn.ListStackInstancesPages(input, func(page *cloudformation.ListStackInstancesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, summary := range page.Summaries {
					r := resourceAwsCloudFormationStackSetInstance()
					d := r.Data(nil)
					id := tfcloudformation.StackSetInstanceCreateResourceID(
						aws.StringValue(summary.StackSetId),
						aws.StringValue(summary.Account),
						aws.StringValue(summary.Region),
					)
					d.SetId(id)

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if testSweepSkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFormation StackSet Instances (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation StackSet Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CloudFormation StackSet Instances (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFormationStackSetInstance_basic(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "stack_id"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_disappears(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsCloudFormationStackSetInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_disappears_StackSet(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	stackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(stackSetResourceName, &stackSet1),
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsCloudFormationStackSetInstance(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsCloudFormationStackSet(), stackSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_ParameterOverrides(t *testing.T) {
	var stackInstance1, stackInstance2, stackInstance3, stackInstance4 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides1(rName, "overridevalue1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides2(rName, "overridevalue1updated", "overridevalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance2),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter2", "overridevalue2"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides1(rName, "overridevalue1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance3),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance4),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance3, &stackInstance4),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
				),
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_RetainStack verifies retain_stack = true
// This acceptance test performs the following steps:
//  * Trigger a Terraform destroy of the resource, which should only remove the instance from the StackSet
//  * Check it still exists outside Terraform
//  * Destroy for real outside Terraform
func TestAccAWSCloudFormationStackSetInstance_RetainStack(t *testing.T) {
	var stack1 cloudformation.Stack
	var stackInstance1, stackInstance2, stackInstance3 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance2),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "false"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance3),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "true"),
				),
			},
			{
				Config:  testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, true),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceStackExists(&stackInstance3, &stack1),
					testAccCheckCloudFormationStackDisappears(&stack1),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_DeploymentTargets(t *testing.T) {
	acctest.Skip(t, "API does not support enabling Organizations access (in particular, creating the Stack Sets IAM Service-Linked Role)")

	var stackInstance cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAWSCloudFormationStackSet(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID, "organizations"),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigDeploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"deployment_targets",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig_ServiceManagedStackSet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackSetInstanceExists(resourceName string, v *cloudformation.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*AWSClient).cfconn

		stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := finder.StackInstanceByName(conn, stackSetName, accountID, region)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCloudFormationStackSetInstanceStackExists(stackInstance *cloudformation.StackInstance, v *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*AWSClient).cfconn

		output, err := finder.StackByID(conn, aws.StringValue(stackInstance.StackId))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSCloudFormationStackSetInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).cfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack_set_instance" {
			continue
		}

		stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.StackInstanceByName(conn, stackSetName, accountID, region)

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

func testAccCheckCloudFormationStackSetInstanceNotRecreated(i, j *cloudformation.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackId) != aws.StringValue(j.StackId) {
			return fmt.Errorf("CloudFormation StackSet Instance (%s,%s,%s) recreated", aws.StringValue(i.StackSetId), aws.StringValue(i.Account), aws.StringValue(i.Region))
		}

		return nil
	}
}

func testAccAWSCloudFormationStackSetInstanceConfigBase(rName string) string {
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

func testAccAWSCloudFormationStackSetInstanceConfig(rName string) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`
}

func testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides1(rName, value1 string) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1)
}

func testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides2(rName, value1, value2 string) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
    Parameter2 = %[2]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1, value2)
}

func testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName string, retainStack bool) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  retain_stack   = %[1]t
  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, retainStack)
}

func testAccAWSCloudFormationStackSetInstanceConfigBase_ServiceManagedStackSet(rName string) string {
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
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName))
}

func testAccAWSCloudFormationStackSetInstanceConfigDeploymentTargets(rName string) string {
	return acctest.ConfigCompose(
		testAccAWSCloudFormationStackSetInstanceConfigBase_ServiceManagedStackSet(rName),
		`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccAWSCloudFormationStackSetInstanceConfig_ServiceManagedStackSet(rName string) string {
	return acctest.ConfigCompose(
		testAccAWSCloudFormationStackSetInstanceConfigBase_ServiceManagedStackSet(rName),
		`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}
