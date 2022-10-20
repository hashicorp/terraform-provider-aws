package outposts_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsAssetsDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_outposts_assets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:               acctest.ErrorCheck(t, outposts.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostAssetsDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "outposts", regexp.MustCompile(`outpost/.+`)),
				),
			},
			{
				Config: testAccOutpostAssetsDataSourceConfig_statusFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "status_id_filter.0", "ACTIVE"),
				),
			},
		},
	})
}

func testAccOutpostAssetsDataSourceConfig_id() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

`
}

func testAccOutpostAssetsDataSourceConfig_statusFilter() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "source" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_assets" "test" {
  arn = data.aws_outposts_assets.source.arn
  status_id_filter = ["ACTIVE"]
}
`
}
