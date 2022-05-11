package elasticbeanstalk_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
)

func TestAccElasticBeanstalkHostedZoneDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_elastic_beanstalk_hosted_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSElasticBeanstalkHostedZoneDataSource_currentRegion,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedZone(dataSourceName, acctest.Region()),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkHostedZoneDataSource_region(t *testing.T) {
	dataSourceName := "data.aws_elastic_beanstalk_hosted_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHostedZoneDataSource_byRegion("ap-southeast-2"), //lintignore:AWSAT003 // passes in GovCloud
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedZone(dataSourceName, "ap-southeast-2"), //lintignore:AWSAT003 // passes in GovCloud
				),
			},
			{
				Config: testAccCheckHostedZoneDataSource_byRegion("eu-west-1"), //lintignore:AWSAT003 // passes in GovCloud
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedZone(dataSourceName, "eu-west-1"), //lintignore:AWSAT003 // passes in GovCloud
				),
			},
			{
				Config:      testAccCheckHostedZoneDataSource_byRegion("ss-pluto-1"),
				ExpectError: regexp.MustCompile("Unsupported region"),
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

		return resource.TestCheckResourceAttr(resourceName, "id", expectedValue)(s)
	}
}

const testAccCheckAWSElasticBeanstalkHostedZoneDataSource_currentRegion = `
data "aws_elastic_beanstalk_hosted_zone" "test" {}
`

func testAccCheckHostedZoneDataSource_byRegion(r string) string {
	return fmt.Sprintf(`
data "aws_elastic_beanstalk_hosted_zone" "test" {
  region = "%s"
}
`, r)
}
