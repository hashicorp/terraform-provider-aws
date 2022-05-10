package cloudformation_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
)

func TestAccCloudFormationStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "on_failure"),
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

func TestAccCloudFormationStack_CreationFailure_doNothing(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackCreationFailureConfig(rName, cloudformation.OnFailureDoNothing),
				ExpectError: regexp.MustCompile(`failed to create CloudFormation stack \(CREATE_FAILED\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_CreationFailure_delete(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackCreationFailureConfig(rName, cloudformation.OnFailureDelete),
				ExpectError: regexp.MustCompile(`failed to create CloudFormation stack, delete requested \(DELETE_COMPLETE\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_CreationFailure_rollback(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackCreationFailureConfig(rName, cloudformation.OnFailureRollback),
				ExpectError: regexp.MustCompile(`failed to create CloudFormation stack, rollback requested \(ROLLBACK_COMPLETE\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_updateFailure(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	vpcCidrInitial := "10.0.0.0/16"
	vpcCidrInvalid := "1000.0.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_withParams(rName, vpcCidrInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
			{
				Config:      testAccStackConfig_withParams(rName, vpcCidrInvalid),
				ExpectError: regexp.MustCompile(`failed to update CloudFormation stack \(UPDATE_ROLLBACK_COMPLETE\).*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_disappears(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudformation.ResourceStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStack_yaml(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_yaml(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
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

func TestAccCloudFormationStack_defaultParams(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_defaultParams(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		},
	})
}

func TestAccCloudFormationStack_allAttributes(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"
	expectedPolicyBody := "{\"Statement\":[{\"Action\":\"Update:*\",\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"LogicalResourceId/StaticVPC\"},{\"Action\":\"Update:*\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"*\"}]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_allAttributesWithBodies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.First", "Mickey"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", "10"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "policy_body"},
			},
			{
				Config: testAccStackConfig_allAttributesWithBodies_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.First", "Mickey"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", "10"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4332
func TestAccCloudFormationStack_withParams(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	vpcCidrInitial := "10.0.0.0/16"
	vpcCidrUpdated := "12.0.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_withParams(rName, vpcCidrInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", vpcCidrInitial),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters"},
			},
			{
				Config: testAccStackConfig_withParams(rName, vpcCidrUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", vpcCidrUpdated),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4534
func TestAccCloudFormationStack_WithURL_withParams(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateURL_withParams(rName, "tf-cf-stack.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "template_url"},
			},
			{
				Config: testAccStackConfig_templateURL_withParams(rName, "tf-cf-stack.json", "13.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
		},
	})
}

func TestAccCloudFormationStack_WithURLWithParams_withYAML(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateURL_withParams_withYAML(rName, "tf-cf-stack.test", "13.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "template_url"},
			},
		},
	})
}

// Test for https://github.com/hashicorp/terraform/issues/5653
func TestAccCloudFormationStack_WithURLWithParams_noUpdate(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateURL_withParams(rName, "tf-cf-stack-1.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "template_url"},
			},
			{
				Config: testAccStackConfig_templateURL_withParams(rName, "tf-cf-stack-2.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
		},
	})
}

func TestAccCloudFormationStack_withTransform(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_withTransform(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
			{
				PlanOnly: true,
				Config:   testAccStackConfig_withTransform(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
				),
			},
		},
	})
}

// TestAccCloudFormationStack_onFailure verifies https://github.com/hashicorp/terraform-provider-aws/issues/5204
func TestAccCloudFormationStack_onFailure(t *testing.T) {
	var stack cloudformation.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_onFailure(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "on_failure", cloudformation.OnFailureDoNothing),
				),
			},
		},
	})
}

func testAccCheckStackExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
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
		if len(resp.Stacks) == 0 || resp.Stacks[0] == nil {
			return fmt.Errorf("CloudFormation stack not found")
		}

		*stack = *resp.Stacks[0]

		return nil
	}
}

func testAccCheckDestroy(s *terraform.State) error {
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

func testAccCheckStackDisappears(stack *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

		input := &cloudformation.DeleteStackInput{
			StackName: stack.StackName,
		}

		_, err := conn.DeleteStack(input)

		if err != nil {
			return err
		}

		// Use the AWS Go SDK waiter until the resource is refactored
		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: stack.StackName,
		}

		return conn.WaitUntilStackDeleteComplete(describeStacksInput)
	}
}

func testAccStackConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}
`, rName)
}

func testAccStackCreationFailureConfig(rName, onFailure string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name       = %[1]q
  on_failure = %[2]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "1000.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}
`, rName, onFailure)
}

func testAccStackConfig_yaml(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
Resources:
  MyVPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        -
          Key: Name
          Value: Primary_CF_VPC

Outputs:
  DefaultSgId:
    Description: The ID of default security group
    Value: !GetAtt MyVPC.DefaultSecurityGroup
  VpcID:
    Description: The VPC ID
    Value: !Ref MyVPC
STACK
}
`, rName)
}

func testAccStackConfig_defaultParams(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<BODY
{
    "Parameters": {
        "TopicName": {
            "Type": "String"
        },
        "VPCCIDR": {
            "Type": "String",
            "Default": "10.10.0.0/16"
        }
    },
    "Resources": {
        "NotificationTopic": {
            "Type": "AWS::SNS::Topic",
            "Properties": {
                "TopicName": {
                    "Ref": "TopicName"
                }
            }
        },
        "MyVPC": {
            "Type": "AWS::EC2::VPC",
            "Properties": {
                "CidrBlock": {
                    "Ref": "VPCCIDR"
                },
                "Tags": [
                    {
                        "Key": "Name",
                        "Value": "Primary_CF_VPC"
                    }
                ]
            }
        }
    },
    "Outputs": {
        "VPCCIDR": {
            "Value": {
                "Ref": "VPCCIDR"
            }
        }
    }
}
BODY


  parameters = {
    TopicName = %[1]q
  }
}
`, rName)
}

var testAccStackConfig_allAttributesWithBodies_tpl = `
data "aws_partition" "current" {}

resource "aws_cloudformation_stack" "test" {
  name          = %[1]q
  template_body = <<STACK
{
  "Parameters" : {
    "VpcCIDR" : {
      "Description" : "CIDR to be used for the VPC",
      "Type" : "String"
    }
  },
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": %[1]q}
        ]
      }
    },
    "StaticVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "%[1]s-2"}
        ]
      }
    },
    "InstanceRole" : {
      "Type" : "AWS::IAM::Role",
      "Properties" : {
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [ {
            "Effect": "Allow",
            "Principal": { "Service": "ec2.${data.aws_partition.current.dns_suffix}" },
            "Action": "sts:AssumeRole"
          } ]
        },
        "Path" : "/",
        "Policies" : [ {
          "PolicyName": "terraformtest",
          "PolicyDocument": {
            "Version": "2012-10-17",
            "Statement": [ {
              "Effect": "Allow",
              "Action": [ "ec2:DescribeSnapshots" ],
              "Resource": [ "*" ]
            } ]
          }
        } ]
      }
    }
  }
}
STACK
  parameters = {
    VpcCIDR = "10.0.0.0/16"
  }

  policy_body        = <<POLICY
%[2]s
POLICY
  capabilities       = ["CAPABILITY_IAM"]
  notification_arns  = [aws_sns_topic.test.arn]
  on_failure         = "DELETE"
  timeout_in_minutes = 10
  tags = {
    First  = "Mickey"
    Second = "Mouse"
  }
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`

var policyBody = `
{
  "Statement" : [
    {
      "Effect" : "Deny",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "LogicalResourceId/StaticVPC"
    },
    {
      "Effect" : "Allow",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "*"
    }
  ]
}
`

func testAccStackConfig_allAttributesWithBodies(rName string) string {
	return fmt.Sprintf(
		testAccStackConfig_allAttributesWithBodies_tpl,
		rName,
		policyBody)
}

func testAccStackConfig_allAttributesWithBodies_modified(rName string) string {
	return fmt.Sprintf(
		testAccStackConfig_allAttributesWithBodies_tpl,
		rName,
		policyBody)
}

func testAccStackConfig_withParams(rName, cidr string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q
  parameters = {
    VpcCIDR = %[2]q
  }
  template_body = <<STACK
{
  "Parameters" : {
    "VpcCIDR" : {
      "Description" : "CIDR to be used for the VPC",
      "Type" : "String"
    }
  },
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  }
}
STACK

  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, cidr)
}

func testAccStackConfig_templateURL_withParams(rName, bucketKey, vpcCidr string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_s3_bucket" "b" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "b" {
  bucket = aws_s3_bucket.b.id
  acl    = "public-read"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.b.id
  policy = <<POLICY
{
  "Version":"2008-10-17",
  "Statement": [
    {
      "Sid":"AllowPublicRead",
      "Effect":"Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.b.id
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.b.id
  key    = %[2]q
  source = "test-fixtures/cloudformation-template.json"
}

resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  parameters = {
    VpcCIDR = %[3]q
  }

  template_url       = "https://${aws_s3_bucket.b.id}.s3-${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${aws_s3_object.object.key}"
  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr)
}

func testAccStackConfig_templateURL_withParams_withYAML(rName, bucketKey, vpcCidr string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_s3_bucket" "b" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "b" {
  bucket = aws_s3_bucket.b.id
  acl    = "public-read"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.b.id
  policy = <<POLICY
{
  "Version":"2008-10-17",
  "Statement": [
    {
      "Sid":"AllowPublicRead",
      "Effect":"Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.b.id
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.b.id
  key    = %[2]q
  source = "test-fixtures/cloudformation-template.yaml"
}

resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  parameters = {
    VpcCIDR = %[3]q
  }

  template_url       = "https://${aws_s3_bucket.b.id}.s3-${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${aws_s3_object.object.key}"
  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr)
}

func testAccStackConfig_withTransform(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Transform": "AWS::Serverless-2016-10-31",
  "Resources": {
    "Api": {
      "Type": "AWS::Serverless::Api",
      "Properties": {
        "StageName": "Prod",
        "EndpointConfiguration": "REGIONAL",
        "DefinitionBody": {
          "swagger": "2.0",
          "paths": {
            "/": {
              "get": {
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "responses": {
                  "200": {
                    "description": "200 response"
                  }
                },
                "x-amazon-apigateway-integration": {
                  "responses": {
                    "default": {
                      "statusCode": "200"
                    }
                  },
                  "requestTemplates": {
                    "application/json": "{\"statusCode\": 200}"
                  },
                  "passthroughBehavior": "when_no_match",
                  "type": "mock"
                }
              }
            }
          }
        }
      }
    }
  }
}
STACK

  capabilities       = ["CAPABILITY_AUTO_EXPAND"]
  on_failure         = "DELETE"
  timeout_in_minutes = 10
}
`, rName)
}

func testAccStackConfig_onFailure(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudformation_stack" "test" {
  name       = %[1]q
  on_failure = "DO_NOTHING"

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Resources = {
      S3Bucket = {
        Type = "AWS::S3::Bucket"
      }
    }
  })
}
`, rName)
}
