package mq_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/mq"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMQBrokerDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	dataSourceByIdName := "data.aws_mq_broker.by_id"
	dataSourceByNameName := "data.aws_mq_broker.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(mq.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerDataSourceConfig_byID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "broker_name", resourceName, "broker_name"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "authentication_strategy", resourceName, "authentication_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "auto_minor_version_upgrade", resourceName, "auto_minor_version_upgrade"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "deployment_mode", resourceName, "deployment_mode"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "configuration.#", resourceName, "configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "encryption_options.#", resourceName, "encryption_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "encryption_options.0.use_aws_owned_key", resourceName, "encryption_options.0.use_aws_owned_key"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "engine_type", resourceName, "engine_type"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "host_instance_type", resourceName, "host_instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "instances.#", resourceName, "instances.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "logs.#", resourceName, "logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "maintenance_window_start_time.#", resourceName, "maintenance_window_start_time.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "publicly_accessible", resourceName, "publicly_accessible"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "storage_type", resourceName, "storage_type"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "user.#", resourceName, "user.#"),
				),
			},
			{
				Config: testAccBrokerDataSourceConfig_byName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceByNameName, "broker_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByNameName, "broker_name", resourceName, "broker_name"),
				),
			},
		},
	})
}

func testAccBrokerDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = aws_subnet.test.*.id[count.index]
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_configuration" "test" {
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.12"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately          = true
  broker_name                = %[1]q

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.12"
  host_instance_type = "mq.t2.micro"

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = true
  security_groups     = aws_security_group.test.*.id
  subnet_ids          = aws_subnet.test.*.id

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

  depends_on = [aws_internet_gateway.test]
}
`, rName)
}

func testAccBrokerDataSourceConfig_byID(rName string) string {
	return acctest.ConfigCompose(testAccBrokerDataSourceConfig_base(rName), `
data "aws_mq_broker" "by_id" {
  broker_id = aws_mq_broker.test.id
}
`)
}

func testAccBrokerDataSourceConfig_byName(rName string) string {
	return acctest.ConfigCompose(testAccBrokerDataSourceConfig_base(rName), `
data "aws_mq_broker" "by_name" {
  broker_name = aws_mq_broker.test.broker_name
}
`)
}
