package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserId_basic(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_log_delivery_canonical_user_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserIdConfig(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserId_default(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_log_delivery_canonical_user_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserIdConfig(endpoints.UsWest2RegionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserId_cn(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_log_delivery_canonical_user_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserIdConfig(endpoints.CnNorthwest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "a52cb28745c0c06e84ec548334e44bfa7fc2a85c54af20cd59e4969344b7af56"),
				),
			},
		},
	})
}

func testAccDataSourceAWSCloudFrontLogDeliveryCanonicalUserIdConfig(region string) string {
	if region == "" {
		region = "null"
	}

	return fmt.Sprintf(`
data "aws_cloudfront_log_delivery_canonical_user_id" "test" {
  region = %[1]q
}
`, region)
}
