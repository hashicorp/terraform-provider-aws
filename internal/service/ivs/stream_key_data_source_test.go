// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVSStreamKeyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ivs_stream_key.test"
	channelResourceName := "aws_ivs_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamKeyDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamKeyDataSource(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "channel_arn", channelResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrValue),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "ivs", regexache.MustCompile(`stream-key/.+`)),
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
