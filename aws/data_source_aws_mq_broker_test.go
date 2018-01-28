package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSMqBroker_basic(t *testing.T) {
	rString := acctest.RandString(7)
	prefix := "tf-acctest-d-aws-mq-broker"
	brokerName := fmt.Sprintf("%s-%s", prefix, rString)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSMqBrokerConfigBasic(brokerName, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "broker_name", brokerName),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "configuration.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "instances.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "publicly_accessible", "true"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "security_groups.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_id", "user.#", "2"),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker.by_id", "arn"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "broker_name", brokerName),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "configuration.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "instances.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "publicly_accessible", "true"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "security_groups.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_mq_broker.by_name", "user.#", "2"),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker.by_name", "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAWSMqBrokerConfigBasic(brokerName, prefix string) string {
	return fmt.Sprintf(`
variable "prefix" {
	default = "%s"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "acctest" {
	cidr_block = "10.0.0.0/16"

	tags {
		Name = "${var.prefix}"
	}
}

resource "aws_internet_gateway" "acctest" {
	vpc_id = "${aws_vpc.acctest.id}"
}

resource "aws_route_table" "acctest" {
	vpc_id = "${aws_vpc.acctest.id}"

	route {
		cidr_block = "0.0.0.0/0"
		gateway_id = "${aws_internet_gateway.acctest.id}"
	}
}

resource "aws_subnet" "acctest" {
	count             = 2
	cidr_block        = "10.0.${count.index}.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
	vpc_id            = "${aws_vpc.acctest.id}"

	tags {
		Name = "${var.prefix}"
	}
}

resource "aws_route_table_association" "acctest" {
	count          = 2
	subnet_id      = "${aws_subnet.acctest.*.id[count.index]}"
	route_table_id = "${aws_route_table.acctest.id}"
}

resource "aws_security_group" "acctest" {
	count  = 2
	name   = "${var.prefix}-${count.index}"
	vpc_id = "${aws_vpc.acctest.id}"
}

resource "aws_mq_configuration" "acctest" {
	name           = "${var.prefix}"
	engine_type    = "ActiveMQ"
	engine_version = "5.15.0"

	data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}

resource "aws_mq_broker" "acctest" {
	auto_minor_version_upgrade = true
	apply_immediately          = true
	broker_name                = "%s"

	configuration {
		id       = "${aws_mq_configuration.acctest.id}"
		revision = "${aws_mq_configuration.acctest.latest_revision}"
	}

	deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
	engine_type        = "ActiveMQ"
	engine_version     = "5.15.0"
	host_instance_type = "mq.t2.micro"

	maintenance_window_start_time {
		day_of_week = "TUESDAY"
		time_of_day = "02:00"
		time_zone   = "CET"
	}

	publicly_accessible = true
	security_groups     = ["${aws_security_group.acctest.*.id}"]
	subnet_ids          = ["${aws_subnet.acctest.*.id}"]

	user {
		username = "Ender"
		password = "AndrewWiggin"
	}

	user {
		username       = "Petra"
		password       = "PetraArkanian"
		console_access = true
		groups         = ["dragon", "salamander", "leopard"]
	}

	depends_on = ["aws_internet_gateway.acctest"]
}

data "aws_mq_broker" "by_id" {
	broker_id = "${aws_mq_broker.acctest.id}"
}

data "aws_mq_broker" "by_name" {
	broker_name = "${aws_mq_broker.acctest.broker_name}"
}
`, prefix, brokerName)
}
