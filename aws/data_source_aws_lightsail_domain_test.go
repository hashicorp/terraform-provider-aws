package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsLightsailDomainDataSource_domain_name(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_lightsail_domain.test"
	dataSourceName := "data.aws_lightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLightsailDomainDomainName(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLightsailDomainDomainName(rInt int) string {
	return fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = "terraformtestacchz-%[1]d.com."
}
data "aws_lightsail_domain" "test" {
  domain_name = aws_lightsail_domain.test.domain_name
}
`, rInt)
}
