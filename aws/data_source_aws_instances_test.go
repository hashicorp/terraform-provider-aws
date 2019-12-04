package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSInstancesDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_instances.test"
	rName := fmt.Sprintf("tf-testacc-instance-%s", acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_ids(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "3"),
					resource.TestCheckResourceAttr(datasourceName, "private_ips.#", "3"),
					// Public IP values are flakey for new EC2 instances due to eventual consistency
					resource.TestCheckResourceAttrSet(datasourceName, "public_ips.#"),
				),
			},
		},
	})
}

func TestAccAWSInstancesDataSource_tags(t *testing.T) {
	datasourceName := "data.aws_instances.test"
	rName := fmt.Sprintf("tf-testacc-instance-%s", acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSInstancesDataSource_instance_state_names(t *testing.T) {
	datasourceName := "data.aws_instances.test"
	rName := fmt.Sprintf("tf-testacc-instance-%s", acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_instance_state_names(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "2"),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig_ids(rName string) string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
resource "aws_instance" "test" {
  count = 3

  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  filter {
    name   = "instance-id"
    values = ["${aws_instance.test.*.id}"]
  }
}
`, rName)
}

func testAccInstancesDataSourceConfig_tags(rName string) string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
resource "aws_instance" "test" {
  count = 2

  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"

  tags = {
    Name      = %[1]q
    SecondTag = %[1]q
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name      = "${aws_instance.test.0.tags["Name"]}"
    SecondTag = "${aws_instance.test.1.tags["Name"]}"
  }
}
`, rName)
}

func testAccInstancesDataSourceConfig_instance_state_names(rName string) string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
resource "aws_instance" "test" {
  count = 2

  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name = "${aws_instance.test.0.tags["Name"]}"
  }

  instance_state_names = ["pending", "running"]
}
`, rName)
}
