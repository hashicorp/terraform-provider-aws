package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsNetworkInterfaceSGAttachment(t *testing.T) {
	cases := []struct {
		Name                      string
		CheckPrimaryInterfaceAttr bool
		Config                    func(bool) string
	}{
		{
			Name: "instance primary interface",
			CheckPrimaryInterfaceAttr: false,
			Config: testAccAwsNetworkInterfaceSGAttachmentViaInstance,
		},
		{
			Name: "externally supplied instance through data source",
			CheckPrimaryInterfaceAttr: true,
			Config: testAccAwsNetworkInterfaceSGAttachmentViaDataSource,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					resource.TestStep{
						Config: tc.Config(true),
						Check:  checkSecurityGroupAttachment(tc.CheckPrimaryInterfaceAttr, true),
					},
					resource.TestStep{
						Config: tc.Config(false),
						Check:  checkSecurityGroupAttachment(tc.CheckPrimaryInterfaceAttr, false),
					},
				},
			})
		})
	}
}

func testAccAwsNetworkInterfaceSGAttachmentViaInstance(attachmentEnabled bool) string {
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

func testAccAwsNetworkInterfaceSGAttachmentViaDataSource(attachmentEnabled bool) string {
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

func checkSecurityGroupAttachment(checkPrimaryInterfaceAttr bool, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		var ifAttr string
		if checkPrimaryInterfaceAttr {
			ifAttr = "network_interface_id"
		} else {
			ifAttr = "primary_network_interface_id"
		}
		interfaceID := s.Modules[0].Resources["aws_instance.instance"].Primary.Attributes[ifAttr]
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
