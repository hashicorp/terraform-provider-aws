package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSOutpostsOutpostsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_outposts.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:   acctest.ErrorCheck(t, outposts.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsOutpostsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostsOutpostsAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostsOutpostsAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["arns.#"]; v == "0" {
			return fmt.Errorf("expected at least one arns result, got none")
		}

		if v := rs.Primary.Attributes["ids.#"]; v == "0" {
			return fmt.Errorf("expected at least one ids result, got none")
		}

		return nil
	}
}

func testAccAWSOutpostsOutpostsDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}
`
}
