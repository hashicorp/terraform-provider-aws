package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNetworkInterfaceSGAttachment(t *testing.T) {
	cases := []struct {
		Name         string
		ResourceAttr string
		Config       func(bool) string
	}{
		{
			Name:         "instance primary interface",
			ResourceAttr: "primary_network_interface_id",
			Config:       testAccAwsNetworkInterfaceSGAttachmentConfigViaInstance,
		},
		{
			Name:         "externally supplied instance through data source",
			ResourceAttr: "network_interface_id",
			Config:       testAccAwsNetworkInterfaceSGAttachmentConfigViaDataSource,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: tc.Config(true),
						Check:  checkSecurityGroupAttached(tc.ResourceAttr, true),
					},
					{
						Config: tc.Config(false),
						Check:  checkSecurityGroupAttached(tc.ResourceAttr, false),
					},
				},
			})
		})
	}
}

func checkSecurityGroupAttached(attr string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		interfaceID := s.Modules[0].Resources["aws_instance.instance"].Primary.Attributes[attr]
		sgID := s.Modules[0].Resources["aws_security_group.sg"].Primary.ID

		iface, err := fetchNetworkInterface(conn, interfaceID)
		if err != nil {
			return err
		}
		actual := sgExistsInENI(sgID, iface)
		if expected != actual {
			return fmt.Errorf("expected existence of security group in ENI to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccAwsNetworkInterfaceSGAttachmentConfigViaInstance(attachmentEnabled bool) string {
	return fmt.Sprintf(`
variable "sg_attachment_enabled" {
  type    = "string"
  default = "%t"
}

data "aws_ami" "ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*"]
  }

  owners = ["amazon"]
}

resource "aws_instance" "instance" {
  instance_type = "t2.micro"
  ami           = "${data.aws_ami.ami.id}"

  tags = {
    "type" = "terraform-test-instance"
  }
}

resource "aws_security_group" "sg" {
  tags = {
    "type" = "terraform-test-security-group"
  }
}

resource "aws_network_interface_sg_attachment" "sg_attachment" {
  count                = "${var.sg_attachment_enabled == "true" ? 1 : 0}"
  security_group_id    = "${aws_security_group.sg.id}"
  network_interface_id = "${aws_instance.instance.primary_network_interface_id}"
}
`, attachmentEnabled)
}

func testAccAwsNetworkInterfaceSGAttachmentConfigViaDataSource(attachmentEnabled bool) string {
	return fmt.Sprintf(`
variable "sg_attachment_enabled" {
  type    = "string"
  default = "%t"
}

data "aws_ami" "ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*"]
  }

  owners = ["amazon"]
}

resource "aws_instance" "instance" {
  instance_type = "t2.micro"
  ami           = "${data.aws_ami.ami.id}"

  tags = {
    "type" = "terraform-test-instance"
  }
}

data "aws_instance" "external_instance" {
  instance_id = "${aws_instance.instance.id}"
}

resource "aws_security_group" "sg" {
  tags = {
    "type" = "terraform-test-security-group"
  }
}

resource "aws_network_interface_sg_attachment" "sg_attachment" {
  count                = "${var.sg_attachment_enabled == "true" ? 1 : 0}"
  security_group_id    = "${aws_security_group.sg.id}"
  network_interface_id = "${data.aws_instance.external_instance.network_interface_id}"
}
`, attachmentEnabled)
}

func TestAccAWSNetworkInterfaceSGAttachmentRaceCheck(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkInterfaceSGAttachmentRaceCheckConfig(),
				Check:  checkSecurityGroupAttachmentRace(),
			},
		},
	})
}

// sgRaceCheckCount specifies the amount of security groups to create in the
// race check. This should be the maximum amount of security groups that can be
// attached to an interface at once, minus the default (we don't remove it in
// the config).
const sgRaceCheckCount = 4

func checkSecurityGroupAttachmentRace() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		interfaceID := s.Modules[0].Resources["aws_network_interface.interface"].Primary.ID
		for i := 0; i < sgRaceCheckCount; i++ {
			sgID := s.Modules[0].Resources["aws_security_group.sg."+strconv.Itoa(i)].Primary.ID
			iface, err := fetchNetworkInterface(conn, interfaceID)
			if err != nil {
				return err
			}
			if !sgExistsInENI(sgID, iface) {
				return fmt.Errorf("security group ID %s was not attached to ENI ID %s", sgID, interfaceID)
			}
		}
		return nil
	}
}

func testAccAwsNetworkInterfaceSGAttachmentRaceCheckConfig() string {
	return fmt.Sprintf(`
variable "security_group_count" {
  type    = "string"
  default = "%d"
}

data "aws_availability_zones" "available" {}

data "aws_subnet" "subnet" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  default_for_az    = "true"
}

resource "aws_network_interface" "interface" {
  subnet_id = "${data.aws_subnet.subnet.id}"

  tags = {
    "type" = "terraform-test-network-interface"
  }
}

resource "aws_security_group" "sg" {
  count = "${var.security_group_count}"

  tags = {
    "type" = "terraform-test-security-group"
  }
}

resource "aws_network_interface_sg_attachment" "sg_attachment" {
  count                = "${var.security_group_count}"
  security_group_id    = "${aws_security_group.sg.*.id[count.index]}"
  network_interface_id = "${aws_network_interface.interface.id}"
}
`, sgRaceCheckCount)
}
