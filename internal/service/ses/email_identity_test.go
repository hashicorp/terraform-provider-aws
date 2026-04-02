// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESEmailIdentity_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:             testAccEmailIdentity_basic,
		acctest.CtDisappears:        testAccEmailIdentity_disappears,
		"trailingPeriod":            testAccEmailIdentity_trailingPeriod,
		"dataSource_basic":          testAccEmailIdentityDataSource_basic,
		"dataSource_trailingPeriod": testAccEmailIdentityDataSource_trailingPeriod,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEmailIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	email := acctest.DefaultEmailAddress
	resourceName := "aws_ses_email_identity.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(email)))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccEmailIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	email := acctest.DefaultEmailAddress
	resourceName := "aws_ses_email_identity.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceEmailIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccEmailIdentity_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(email, "."))))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckEmailIdentityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_email_identity" {
				continue
			}

			_, err := tfses.FindIdentityVerificationAttributesByIdentity(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Email Identity %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEmailIdentityExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		_, err := tfses.FindIdentityVerificationAttributesByIdentity(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEmailIdentityConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %[1]q
}
`, email)
}
