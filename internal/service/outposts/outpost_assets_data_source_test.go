package outposts_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsAssetsDataSource_id(t *testing.T) { // nosemgrep: outposts-in-func-name
	dataSourceName := "data.aws_outposts_assets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostAssetsDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^op-.+$`)),
				),
			},
		},
	})
}

func testAccOutpostAssetsDataSourceConfig_id() string { // nosemgrep: outposts-in-func-name
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

`
}
