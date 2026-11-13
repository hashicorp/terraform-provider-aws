// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEmailIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	email := acctest.DefaultEmailAddress

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataDourceConfig_source(email),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(ctx, "data.aws_ses_email_identity.test", names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(email)))),
				),
			},
		},
	})
}

func testAccEmailIdentityDataSource_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataDourceConfig_source(email),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(ctx, "data.aws_ses_email_identity.test", names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(email, "."))))),
				),
			},
		},
	})
}

func testAccEmailIdentityDataDourceConfig_source(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %[1]q
}

data "aws_ses_email_identity" "test" {
  depends_on = [aws_ses_email_identity.test]
  email      = %[1]q
}
`, email)
}
