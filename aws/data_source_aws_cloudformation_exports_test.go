package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudformationExports_dataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsCloudformationExportsJson,
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudformation_exports.waiter", "value", "waiter"),
					resource.TestMatchResourceAttr("data.aws_cloudformation_exports.vpc", "value",
						regexp.MustCompile("^vpc-[a-z0-9]{8}$")),
					resource.TestMatchResourceAttr("data.aws_cloudformation_exports.vpc", "exporting_stack_id",
						regexp.MustCompile("^arn:aws:cloudformation")),
				),
			},
		},
	})
}

const testAccCheckCfnExport = `
data "aws_cloudformation_exports" "here" {
	name = "Intuit-vpc-1:vpc:id"
}
`
const testAccCheckAwsCloudformationExportsJson = `
resource "aws_cloudformation_stack" "cfs" {
  name = "tf-waiter-stack"
  timeout_in_minutes = 6
  template_body = <<STACK
{
  "Resources": {
    "waiter": {
      "Type": "AWS::CloudFormation::WaitConditionHandle",
      "Properties": { }
    }
  },
  "Outputs": {
    "waiter": {
      "Value": "waiter" ,
      "Description": "VPC ID",
      "Export": {
        "Name": "waiter" 
      }
    }
  }
}
STACK
  tags {
    TestExport = "waiter"
    Second = "meh"
  }
}
resource "aws_cloudformation_stack" "yaml" {
  name = "tf-acc-ds-yaml-stack"
  parameters {
    CIDR = "10.10.10.0/24"
  }
  timeout_in_minutes = 6
  template_body = <<STACK
Parameters:
  CIDR:
    Type: String

Resources:
  myvpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: !Ref CIDR
      Tags:
        -
          Key: Name
          Value: Primary_CF_VPC

Outputs:
  MyVpcId:
    Value: !Ref myvpc
    Description: VPC ID
    Export:
      Name: MyVpcId
STACK
  tags {
    TestExport = "MyVpcId"
    Second = "meh"
  }
}
data "aws_cloudformation_exports" "vpc" {
	name = "${aws_cloudformation_stack.yaml.tags["TestExport"]}"
}
data "aws_cloudformation_exports" "waiter" {
	name = "${aws_cloudformation_stack.cfs.tags["TestExport"]}"
}
`
