package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAwsDataSourceElasticBeanstalkHostedZone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_currentRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elastic_beanstalk_hosted_zone.current", "id", "Z2PCDNR3VC2G1N"),
				),
			},
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_sydney,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elastic_beanstalk_hosted_zone.sydney", "id", "Z2PCDNR3VC2G1N"),
				),
			},
		},
	})
}

const testAccCheckAwsElasticBeanstalkHostedZoneDataSource_currentRegion = `
provider "aws" {
	region = "ap-southeast-2"
}
data "aws_elastic_beanstalk_hosted_zone" "current" {}
`

const testAccCheckAwsElasticBeanstalkHostedZoneDataSource_sydney = `
data "aws_elastic_beanstalk_hosted_zone" "sydney" {
	region = "ap-southeast-2"
}
`
