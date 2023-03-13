package ses_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSESEmailIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	email := acctest.DefaultEmailAddress

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ses.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataDourceConfig_source(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, "aws_ses_email_identity.test"),
					acctest.MatchResourceAttrRegionalARN("data.aws_ses_email_identity.test", "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(email)))),
				),
			},
		},
	})
}

func TestAccSESEmailIdentityDataSource_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ses.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataDourceConfig_source(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, "aws_ses_email_identity.test"),
					acctest.MatchResourceAttrRegionalARN("data.aws_ses_email_identity.test", "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(email, "."))))),
				),
			},
		},
	})
}

func testAccEmailIdentityDataDourceConfig_source(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %q
}
data "aws_ses_email_identity" "test" {
  depends_on = [aws_ses_email_identity.test]
  email      = %q
}
`, email, email)
}
