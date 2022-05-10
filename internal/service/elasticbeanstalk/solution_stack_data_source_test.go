package elasticbeanstalk_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccElasticBeanstalkSolutionStackDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSElasticBeanstalkSolutionStackDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSolutionStackIDDataSource("data.aws_elastic_beanstalk_solution_stack.multi_docker"),
					resource.TestMatchResourceAttr("data.aws_elastic_beanstalk_solution_stack.multi_docker", "name", regexp.MustCompile("^64bit Amazon Linux (.*) Multi-container Docker (.*)$")),
				),
			},
		},
	})
}

func testAccCheckSolutionStackIDDataSource(n string) resource.TestCheckFunc {
	// Wait for solution stacks
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find solution stack data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Solution stack data source ID not set")
		}
		return nil
	}
}

const testAccCheckAWSElasticBeanstalkSolutionStackDataSourceConfig = `
data "aws_elastic_beanstalk_solution_stack" "multi_docker" {
  most_recent = true
  name_regex  = "^64bit Amazon Linux (.*) Multi-container Docker (.*)$"
}
`
