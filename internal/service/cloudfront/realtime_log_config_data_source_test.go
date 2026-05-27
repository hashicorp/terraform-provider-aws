// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontRealtimeLogConfigDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	samplingRate := acctest.RandIntRange(t, 1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	dataSourceName := "data.aws_cloudfront_realtime_log_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfigDataSourceConfig_basic(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.#", resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.stream_type", resourceName, "endpoint.0.stream_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.kinesis_stream_config.#", resourceName, "endpoint.0.kinesis_stream_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "sampling_rate", resourceName, "sampling_rate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fields.#", resourceName, "fields.#"),
				),
			},
		},
	})
}

func testAccRealtimeLogConfigDataSourceConfig_basic(rName string, samplingRate int) string {
	return acctest.ConfigCompose(testAccRealtimeLogConfigConfig_basic(rName, samplingRate), `
data "aws_cloudfront_realtime_log_config" "test" {
  name = aws_cloudfront_realtime_log_config.test.name
}
`,
	)
}
