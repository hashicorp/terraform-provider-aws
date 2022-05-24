package outposts_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsOutpostDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(dataSourceName, "arn", "outposts", regexp.MustCompile(`outpost/op-.+$`).String()),
					resource.TestMatchResourceAttr(dataSourceName, "availability_zone", regexp.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "availability_zone_id", regexp.MustCompile(`^.+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^op-.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(`^.+$`)),
					acctest.MatchResourceAttrAccountID(dataSourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccOutpostsOutpostDataSource_name(t *testing.T) {
	sourceDataSourceName := "data.aws_outposts_outpost.source"
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostNameDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", sourceDataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", sourceDataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", sourceDataSourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", sourceDataSourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", sourceDataSourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", sourceDataSourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccOutpostsOutpostDataSource_arn(t *testing.T) {
	sourceDataSourceName := "data.aws_outposts_outpost.source"
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostARNDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", sourceDataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", sourceDataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", sourceDataSourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", sourceDataSourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", sourceDataSourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", sourceDataSourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccOutpostsOutpostDataSource_ownerID(t *testing.T) {
	sourceDataSourceName := "data.aws_outposts_outpost.source"
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostOwnerIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", sourceDataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", sourceDataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", sourceDataSourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", sourceDataSourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", sourceDataSourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", sourceDataSourceName, "owner_id"),
				),
			},
		},
	})
}

func testAccOutpostIDDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}
`
}

func testAccOutpostNameDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "source" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_outposts_outpost" "test" {
  name = data.aws_outposts_outpost.source.name
}
`
}

func testAccOutpostARNDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "source" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_outpost" "test" {
  arn = data.aws_outposts_outpost.source.arn
}
`
}

func testAccOutpostOwnerIDDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "source" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_outposts_outpost" "test" {
  owner_id = data.aws_outposts_outpost.source.owner_id
}
`
}
