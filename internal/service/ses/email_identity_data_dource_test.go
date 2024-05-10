// Copyright (c) HashiCorp, Inc.
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

func TestAccSESEmailIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	email := acctest.DefaultEmailAddress

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataDourceConfig_source(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, "aws_ses_email_identity.test"),
					acctest.MatchResourceAttrRegionalARN("data.aws_ses_email_identity.test", names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(email)))),
				),
			},
		},
	})
}

func TestAccSESEmailIdentityDataSource_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataDourceConfig_source(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, "aws_ses_email_identity.test"),
					acctest.MatchResourceAttrRegionalARN("data.aws_ses_email_identity.test", names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(email, "."))))),
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
