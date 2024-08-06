// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMQBrokerInstanceTypeOfferingsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.MQEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerInstanceTypeOfferingsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.empty", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.empty", "broker_instance_options.*", map[string]string{
						"engine_type": "ACTIVEMQ",
					}),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.engine", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.engine", "broker_instance_options.*", map[string]string{
						"engine_type": "ACTIVEMQ",
					}),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.storage", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.storage", "broker_instance_options.*", map[string]string{
						names.AttrStorageType: "ebs",
					}),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.instance", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.instance", "broker_instance_options.*", map[string]string{
						"host_instance_type": "mq.m5.large",
					}),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.all", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.instance", "broker_instance_options.*", map[string]string{
						"host_instance_type":  "mq.m5.large",
						names.AttrStorageType: "ebs",
						"engine_type":         "ACTIVEMQ",
					}),
				),
			},
		},
	})
}

func testAccBrokerInstanceTypeOfferingsDataSourceConfig_basic() string {
	return `
data "aws_mq_broker_instance_type_offerings" "empty" {}

data "aws_mq_broker_instance_type_offerings" "engine" {
  engine_type = "ACTIVEMQ"
}

data "aws_mq_broker_instance_type_offerings" "storage" {
  storage_type = "EBS"
}

data "aws_mq_broker_instance_type_offerings" "instance" {
  host_instance_type = "mq.m5.large"
}

data "aws_mq_broker_instance_type_offerings" "all" {
  host_instance_type = "mq.m5.large"
  storage_type       = "EBS"
  engine_type        = "ACTIVEMQ"
}
`
}
