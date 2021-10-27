package elasticbeanstalk_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccElasticBeanstalkEnvironmentDataSource_basic(t *testing.T) {
	dataSourceResourceName := "data.aws_elastic_beanstalk_environment.test"
	resourceName := "aws_elastic_beanstalk_environment.tftest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfig_Basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceResourceName, "description"),
				),
			},
		},
	})
}

const testAccEnvironmentDataSourceConfig_Basic = `
data "aws_elastic_beanstalk_environment" "test" {
	name = aws_elastic_beanstalk_environment.tftest.name
  }
`
