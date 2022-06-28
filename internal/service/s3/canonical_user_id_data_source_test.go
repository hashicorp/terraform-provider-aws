package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3CanonicalUserIDDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCanonicalUserIDDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCanonicalUserIdCheckExistsDataSource("data.aws_canonical_user_id.current"),
				),
			},
		},
	})
}

func testAccCanonicalUserIdCheckExistsDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Canonical User ID resource: %s", name)
		}

		if rs.Primary.Attributes["id"] == "" {
			return fmt.Errorf("Missing Canonical User ID")
		}
		if rs.Primary.Attributes["display_name"] == "" {
			return fmt.Errorf("Missing Display Name")
		}

		return nil
	}
}

const testAccCanonicalUserIDDataSourceConfig_basic = `
data "aws_canonical_user_id" "current" {}
`
