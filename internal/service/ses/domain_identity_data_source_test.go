package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSESDomainIdentityDataSource_basic(t *testing.T) {
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityDataSourceConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists("aws_ses_domain_identity.test"),
					testAccCheckDomainIdentityARN("data.aws_ses_domain_identity.test", domain),
				),
			},
		},
	})
}

func testAccDomainIdentityDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

data "aws_ses_domain_identity" "test" {
  depends_on = [aws_ses_domain_identity.test]
  domain     = "%s"
}
`, domain, domain)
}
