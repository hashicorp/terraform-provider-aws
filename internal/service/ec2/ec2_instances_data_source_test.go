package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2InstancesDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_ids(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_instances.test", "private_ips.#", "3"),
					// Public IP values are flakey for new EC2 instances due to eventual consistency
					resource.TestCheckResourceAttrSet("data.aws_instances.test", "public_ips.#"),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_instanceStateNames(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_instanceStateNames(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_instances.test", "private_ips.#", "0"),
					resource.TestCheckResourceAttr("data.aws_instances.test", "public_ips.#", "0"),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig_ids(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 3
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  filter {
    name   = "instance-id"
    values = aws_instance.test[*].id
  }
}
`, rName))
}

func testAccInstancesDataSourceConfig_tags(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name      = %[1]q
    SecondTag = "%[1]s-2"
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name      = aws_instance.test[0].tags["Name"]
    SecondTag = aws_instance.test[0].tags["SecondTag"]
  }

  depends_on = [aws_instance.test]
}
`, rName))
}

func testAccInstancesDataSourceConfig_instanceStateNames(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name = aws_instance.test[0].tags["Name"]
  }

  instance_state_names = ["pending", "running"]
  depends_on           = [aws_instance.test]
}
`, rName))
}

func testAccInstancesDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_instances" "test" {
  instance_tags = {
    Name = %[1]q
  }
}
`, rName)
}
