package outposts_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsSiteDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_outposts_site.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSites(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(dataSourceName, "account_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^os-.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccOutpostsSiteDataSource_name(t *testing.T) {
	sourceDataSourceName := "data.aws_outposts_site.source"
	dataSourceName := "data.aws_outposts_site.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSites(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteNameDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", sourceDataSourceName, "account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", sourceDataSourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", sourceDataSourceName, "name"),
				),
			},
		},
	})
}

func testAccSiteIDDataSourceConfig() string {
	return `
data "aws_outposts_sites" "test" {}

data "aws_outposts_site" "test" {
  id = tolist(data.aws_outposts_sites.test.ids)[0]
}
`
}

func testAccSiteNameDataSourceConfig() string {
	return `
data "aws_outposts_sites" "test" {}

data "aws_outposts_site" "source" {
  id = tolist(data.aws_outposts_sites.test.ids)[0]
}

data "aws_outposts_site" "test" {
  name = data.aws_outposts_site.source.name
}
`
}
