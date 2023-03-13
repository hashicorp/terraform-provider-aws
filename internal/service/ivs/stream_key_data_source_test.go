package ivs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIVSStreamKeyDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ivs_stream_key.test"
	channelResourceName := "aws_ivs_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamKeyDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamKeyDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "channel_arn", channelResourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "value"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "ivs", regexp.MustCompile(`stream-key/.+`)),
				),
			},
		},
	})
}

func testAccCheckStreamKeyDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Stream Key data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Stream Key data source ID not set")
		}
		return nil
	}
}

func testAccStreamKeyDataSourceConfig_basic() string {
	return `
resource "aws_ivs_channel" "test" {
}

data "aws_ivs_stream_key" "test" {
  channel_arn = aws_ivs_channel.test.arn
}
`
}
