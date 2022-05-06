package mq_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMQBrokerInstanceTypeOfferingsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(mq.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerDataSourceInstanceTypeOfferingsConfig(),
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
						"storage_type": "ebs",
					}),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.instance", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.instance", "broker_instance_options.*", map[string]string{
						"host_instance_type": "mq.m5.large",
					}),
					resource.TestCheckResourceAttrSet("data.aws_mq_broker_instance_type_offerings.all", "broker_instance_options.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.aws_mq_broker_instance_type_offerings.instance", "broker_instance_options.*", map[string]string{
						"host_instance_type": "mq.m5.large",
						"storage_type":       "ebs",
						"engine_type":        "ACTIVEMQ",
					}),
				),
			},
		},
	})
}

func testAccBrokerDataSourceInstanceTypeOfferingsConfig() string {
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
