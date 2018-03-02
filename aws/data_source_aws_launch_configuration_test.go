package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSLaunchConfigurationDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	rName := "data.aws_launch_configuration.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rName, "image_id"),
					resource.TestCheckResourceAttrSet(rName, "instance_type"),
					resource.TestCheckResourceAttrSet(rName, "associate_public_ip_address"),
					resource.TestCheckResourceAttrSet(rName, "user_data"),
					resource.TestCheckResourceAttr(rName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(rName, "ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(rName, "ephemeral_block_device.#", "1"),
				),
			},
		},
	})
}

func testAccLaunchConfigurationDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_configuration" "foo" {
  name = "terraform-test-%d"
  image_id = "ami-21f78e11"
  instance_type = "m1.small"
  associate_public_ip_address = true
  user_data = "foobar-user-data"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
	}
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }
  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops = 100
  }
  ephemeral_block_device {
    device_name = "/dev/sde"
    virtual_name = "ephemeral0"
  }
}

data "aws_launch_configuration" "foo" {
  name = "${aws_launch_configuration.foo.name}"
}
`, rInt)
}
