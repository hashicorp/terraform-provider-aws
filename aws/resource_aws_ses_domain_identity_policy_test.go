package aws

import (
	"regexp"
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSSESDomainIdentityPolicy_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com.",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESDomainIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESDomainIdentityConfig_withPolicy(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainIdentityExists("aws_ses_domain_identity.test"),
					resource.TestMatchResourceAttr("aws_ses_domain_identity_policy.custom", "policy",
						regexp.MustCompile("^{\"Version\":\"2012-10-17\".+")),
				),
			},
		},
	})
}

func testAccAWSSESDomainIdentityConfig_withPolicy(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
    name = "%s"
}

resource "aws_ses_domain_identity_policy" "custom" {
	arn = "${aws_ses_domain_identity.test.arn}"
	name = "test"
	policy = <<POLICY
{
   "Version":"2012-10-17",
   "Id": "default",
   "Statement":[{
   	"Sid":"default",
   	"Effect":"Allow",
   	"Principal":{"AWS":"*"},
   	"Action":["SES:SendEmail","SES:SendRawEmail"],
   	"Resource":"${aws_ses_domain_identity.test.arn}"
  }]
}
POLICY
}
`, domain)
}
