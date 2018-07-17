package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSStorageGatewayGatewayActivationKeyDataSource_FileGateway(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_storagegateway_gateway_activation_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayActivationKeyDataSourceConfig_FileGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "activation_key"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayGatewayActivationKeyDataSource_TapeAndVolumeGateway(t *testing.T) {
	t.Skip("Currently the EC2 instance webserver is never reachable, its likely an instance configuration error.")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_storagegateway_gateway_activation_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayActivationKeyDataSourceConfig_TapeAndVolumeGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "activation_key"),
				),
			},
		},
	})
}

// testAccAWSStorageGateway_VPCBase provides a publicly accessible subnet
// and security group, suitable for Storage Gateway EC2 instances of any type
func testAccAWSStorageGateway_VPCBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags {
    Name = %q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags {
    Name = %q
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags {
    Name = %q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = "${aws_subnet.test.id}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "test" {
  name   = %q
  vpc_id = "${aws_vpc.test.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = %q
  }
}
`, rName, rName, rName, rName, rName, rName)
}

// testAccAWSStorageGateway_FileGatewayBase uses the "thinstaller" Storage
// Gateway AMI for File Gateways
func testAccAWSStorageGateway_FileGatewayBase(rName string) string {
	return testAccAWSStorageGateway_VPCBase(rName) + fmt.Sprintf(`
data "aws_ami" "aws-thinstaller" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true
  instance_type               = "t2.micro"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags {
    Name = %q
  }
}
`, rName)
}

// testAccAWSStorageGateway_TapeAndVolumeGatewayBase uses the Storage Gateway
// AMI for either Tape or Volume Gateways
func testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName string) string {
	return testAccAWSStorageGateway_VPCBase(rName) + fmt.Sprintf(`
data "aws_ami" "aws-storage-gateway-2" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["aws-storage-gateway-2.*"]
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-storage-gateway-2.id}"
  associate_public_ip_address = true
  instance_type               = "t2.micro"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags {
    Name = %q
  }
}
`, rName)
}

func testAccAWSStorageGatewayGatewayActivationKeyDataSourceConfig_FileGateway(rName string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + `
data "aws_storagegateway_gateway_activation_key" "test" {
  ip_address = "${aws_instance.test.public_ip}"
}
`
}

func testAccAWSStorageGatewayGatewayActivationKeyDataSourceConfig_TapeAndVolumeGateway(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + `
data "aws_storagegateway_gateway_activation_key" "test" {
  ip_address = "${aws_instance.test.public_ip}"
}
`
}
