package nas_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccDataSourceAwsCanonicalUserId_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCanonicalUserIdConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsCanonicalUserIdCheckExists("data.aws_canonical_user_id.current"),
				),
			},
		},
	})
}

func testAccDataSourceAwsCanonicalUserIdCheckExists(name string) resource.TestCheckFunc {
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

const testAccDataSourceAwsCanonicalUserIdConfig = `
data "aws_canonical_user_id" "current" {}
`
