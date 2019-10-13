package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSKinesisStreamConsumerDataSource(t *testing.T) {
	var stream kinesis.StreamDescription
	var consumer kinesis.ConsumerDescription
	config := createAccKinesisStreamConsumerConfig()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: config.data(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(config.stream.getName(), &stream),
					testAccCheckKinesisStreamConsumerExists(config, -1, &consumer),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("data.%s", config.getName()), "arn"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("data.%s", config.getName()), "stream_arn"),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", config.getName()), "name", config.getConsumerName()),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s", config.getName()), "status", "ACTIVE"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("data.%s", config.getName()), "creation_timestamp"),
				),
			},
		},
	})
}
