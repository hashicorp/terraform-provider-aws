package cloudwatchevidently_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccCheckProjectExists(name string, project *cloudwatchevidently.GetProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEvidentlyConn
		input := &cloudwatchevidently.GetProjectInput{
			Project: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetProject(input)

		if err != nil {
			return err
		}

		*project = *resp

		return nil
	}
}
