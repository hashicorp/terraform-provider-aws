package cloudfront_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontRealtimeLogConfigDataSource_basic(t *testing.T) {
	var v cloudfront.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	dataSourceName := "data.aws_cloudfront_realtime_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfigDataSource(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.#", resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.stream_type", resourceName, "endpoint.0.stream_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.kinesis_stream_config.#", resourceName, "endpoint.0.kinesis_stream_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sampling_rate", resourceName, "sampling_rate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fields.#", resourceName, "fields.#"),
				),
			},
		},
	})
}

func testAccRealtimeLogConfigDataSource(rName string, samplingRate int) string {
	return acctest.ConfigCompose(
		testAccRealtimeLogConfig(rName, samplingRate), `
data "aws_cloudfront_realtime_log_config" "test" {
  name = aws_cloudfront_realtime_log_config.test.name
}
`,
	)
}
