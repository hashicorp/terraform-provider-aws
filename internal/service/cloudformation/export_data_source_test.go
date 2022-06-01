package cloudformation_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFormationExportDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudformation_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:                    testAccExportDataSourceConfig_staticValue(rName),
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "value", "waiter"),
				),
			},
		},
	})
}

func TestAccCloudFormationExportDataSource_resourceReference(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudformation_export.test"
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:                    testAccExportDataSourceConfig_resourceReference(rName),
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "exporting_stack_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "value", resourceName, "outputs.MyVpcId"),
				),
			},
		},
	})
}

func testAccExportDataSourceConfig_staticValue(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

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
        "Name": %[1]q
      }
    }
  }
}
STACK

  tags = {
    TestExport = %[1]q
    Second     = "meh"
  }
}

data "aws_cloudformation_export" "test" {
  name = aws_cloudformation_stack.test.tags["TestExport"]
}
`, rName)
}

func testAccExportDataSourceConfig_resourceReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  parameters = {
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
      Name: %[1]q
STACK

  tags = {
    TestExport = %[1]q
    Second     = "meh"
  }
}

data "aws_cloudformation_export" "test" {
  name = aws_cloudformation_stack.test.tags["TestExport"]
}
`, rName)
}
