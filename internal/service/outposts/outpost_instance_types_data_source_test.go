package outposts_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsOutpostInstanceTypesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_outpost_instance_types.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostInstanceTypesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostInstanceTypesAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostInstanceTypesAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		return nil
	}
}

func testAccOutpostInstanceTypesDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}
