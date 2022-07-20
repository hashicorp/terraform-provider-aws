package ssm_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSMInstancesDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssm_instances.test"
	resourceName := "aws_instance.test"

	registrationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow SSM Agent to register EC2 instance as a managed node.")
			time.Sleep(1 * time.Minute)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_filterInstance(rName),
			},
			{
				Config: testAccInstancesDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					registrationSleep(),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resourceName, "id"),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig_filterInstance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro", "t3.micro"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

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
          "ec2.${data.aws_partition.current.dns_suffix}"
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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_main_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  vpc_id         = aws_vpc.test.id
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

  vpc_security_group_ids      = [aws_vpc.test.default_security_group_id]
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  depends_on                  = [aws_main_route_table_association.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstancesDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(
		testAccInstancesDataSourceConfig_filterInstance(rName),
		`
data "aws_ssm_instances" "test" {
  filter {
    name   = "InstanceIds"
    values = [aws_instance.test.id]
  }
}
`)
}
