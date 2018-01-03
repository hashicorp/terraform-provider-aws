package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSKinesisFirehoseDeliveryStream_importBasic(t *testing.T) {
	resName := "aws_kinesis_firehose_delivery_stream.test_stream"
	rInt := acctest.RandInt()
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		rInt, os.Getenv("AWS_ACCOUNT_ID"), rInt, rInt, rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
