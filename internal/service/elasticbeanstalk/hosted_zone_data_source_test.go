// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkHostedZoneDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elastic_beanstalk_hosted_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneDataSourceConfig_currentRegion,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedZone(dataSourceName, acctest.Region()),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkHostedZoneDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elastic_beanstalk_hosted_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneDataSourceConfig_byRegion("ap-southeast-2"), //lintignore:AWSAT003 // passes in GovCloud
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedZone(dataSourceName, "ap-southeast-2"), //lintignore:AWSAT003 // passes in GovCloud
				),
			},
			{
				Config: testAccHostedZoneDataSourceConfig_byRegion("eu-west-1"), //lintignore:AWSAT003 // passes in GovCloud
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedZone(dataSourceName, "eu-west-1"), //lintignore:AWSAT003 // passes in GovCloud
				),
			},
			{
				Config:      testAccHostedZoneDataSourceConfig_byRegion("ss-pluto-1"),
				ExpectError: regexache.MustCompile("Unsupported region"),
			},
		},
	})
}

func testAccCheckHostedZone(resourceName string, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue, ok := tfelasticbeanstalk.HostedZoneIDs[region]

		if !ok {
			return fmt.Errorf("Unsupported region: %s", region)
		}

		return resource.TestCheckResourceAttr(resourceName, names.AttrID, expectedValue)(s)
	}
}

const testAccHostedZoneDataSourceConfig_currentRegion = `
data "aws_elastic_beanstalk_hosted_zone" "test" {}
`

func testAccHostedZoneDataSourceConfig_byRegion(r string) string {
	return fmt.Sprintf(`
data "aws_elastic_beanstalk_hosted_zone" "test" {
  region = "%s"
}
`, r)
}
