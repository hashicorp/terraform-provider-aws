// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMQBrokerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	dataSourceByIdName := "data.aws_mq_broker.by_id"
	dataSourceByNameName := "data.aws_mq_broker.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.MQEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				//Config: testAccBrokerDataSourceConfig_byID(rName),
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: map[string]config.Variable{
					"exclude_zone_ids":           config.ListVariable(config.StringVariable("usw2-az4"), config.StringVariable("usgw1-az2")),
					"state":                      config.StringVariable("available"),
					"name":                       config.StringVariable("opt-in-status"),
					"values":                     config.ListVariable(config.StringVariable("opt-in-not-required")),
					"cidr_block":                 config.StringVariable("10.0.0.0/16"),
					"random_name":                config.StringVariable(rName),
					"vcount":                     config.IntegerVariable(2),
					"cidr_block_2":               config.StringVariable("0.0.0.0/0"),
					"engine_type":                config.StringVariable("ActiveMQ"),
					"engine_version":             config.StringVariable("5.17.6"),
					"auto_minor_version_upgrade": config.BoolVariable(true),
					"apply_immediately":          config.BoolVariable(true),
					"deployment_mode":            config.StringVariable("ACTIVE_STANDBY_MULTI_AZ"),
					"host_instance_type":         config.StringVariable("mq.t2.micro"),
					"day_of_week":                config.StringVariable("TUESDAY"),
					"time_of_day":                config.StringVariable("02:00"),
					"time_zone":                  config.StringVariable("CET"),
					"publicly_accessible":        config.BoolVariable(true),
					"username":                   config.StringVariable("Ender"),
					"password":                   config.StringVariable("AndrewWiggin"),
					"username_2":                 config.StringVariable("Petra"),
					"password_2":                 config.StringVariable("PetraArkanian"),
					"console_access":             config.BoolVariable(true),
					"groups":                     config.ListVariable(config.StringVariable("dragon"), config.StringVariable("salamander"), config.StringVariable("leopard")),
				},
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
				//Config: testAccBrokerDataSourceConfig_byName(rName),
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: map[string]config.Variable{
					"exclude_zone_ids":           config.ListVariable(config.StringVariable("usw2-az4"), config.StringVariable("usgw1-az2")),
					"state":                      config.StringVariable("available"),
					"name":                       config.StringVariable("opt-in-status"),
					"values":                     config.ListVariable(config.StringVariable("opt-in-not-required")),
					"cidr_block":                 config.StringVariable("10.0.0.0/16"),
					"random_name":                config.StringVariable(rName),
					"vcount":                     config.IntegerVariable(2),
					"cidr_block_2":               config.StringVariable("0.0.0.0/0"),
					"engine_type":                config.StringVariable("ActiveMQ"),
					"engine_version":             config.StringVariable("5.17.6"),
					"auto_minor_version_upgrade": config.BoolVariable(true),
					"apply_immediately":          config.BoolVariable(true),
					"deployment_mode":            config.StringVariable("ACTIVE_STANDBY_MULTI_AZ"),
					"host_instance_type":         config.StringVariable("mq.t2.micro"),
					"day_of_week":                config.StringVariable("TUESDAY"),
					"time_of_day":                config.StringVariable("02:00"),
					"time_zone":                  config.StringVariable("CET"),
					"publicly_accessible":        config.BoolVariable(true),
					"username":                   config.StringVariable("Ender"),
					"password":                   config.StringVariable("AndrewWiggin"),
					"username_2":                 config.StringVariable("Petra"),
					"password_2":                 config.StringVariable("PetraArkanian"),
					"console_access":             config.BoolVariable(true),
					"groups":                     config.ListVariable(config.StringVariable("dragon"), config.StringVariable("salamander"), config.StringVariable("leopard")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceByNameName, "broker_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByNameName, "broker_name", resourceName, "broker_name"),
				),
			},
		},
	})
}

func testAccBrokerDataSourceConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccBrokerConfig_baseCustomVPC(rName), fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  name           = %[1]q
  engine_type    = "ActiveMQ"
  engine_version = "5.17.6"

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
  engine_version     = "5.17.6"
  host_instance_type = "mq.t2.micro"

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = true
  security_groups     = aws_security_group.test[*].id
  subnet_ids          = aws_subnet.test[*].id

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
`, rName))
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
