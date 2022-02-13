package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSMInstancesDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssm_instances.test"
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstancesDataSourceConfig_filter_instance(rName),
			},
			{
				Config: testAccCheckInstancesDataSourceConfig_filter_dataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckInstancesDataSourceConfig_filter_instance(rName string) string {
	return acctest.ConfigCompose(
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name = "AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role" "test" {
  name                = %[1]q
  managed_policy_arns = [data.aws_iam_policy.test.arn]

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_instance" "test" {
  ami                  = data.aws_ami.test.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  iam_instance_profile = aws_iam_instance_profile.test.name
}
`, rName))
}

func testAccCheckInstancesDataSourceConfig_filter_dataSource(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckInstancesDataSourceConfig_filter_instance(rName), `
data "aws_ssm_instances" "test" {
  filter {
    name   = "InstanceIds"
    values = [aws_instance.test.id]
  }
}
`)
}
