package outposts_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_outposts.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostsAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostsAttributes(dataSourceName string) resource.TestCheckFunc {
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

func testAccOutpostsDataSourceConfig_basic() string {
	return `
data "aws_outposts_outposts" "test" {}
`
}
