// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESEmailIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	email := acctest.DefaultEmailAddress
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(email)))),
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

func TestAccSESEmailIdentity_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	email := fmt.Sprintf("%s.", acctest.DefaultEmailAddress)
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", regexache.MustCompile(fmt.Sprintf("identity/%s$", regexp.QuoteMeta(strings.TrimSuffix(email, "."))))),
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

func testAccCheckEmailIdentityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_email_identity" {
				continue
			}

			email := rs.Primary.ID
			params := &ses.GetIdentityVerificationAttributesInput{
				Identities: []*string{
					aws.String(email),
				},
			}

			response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, params)
			if err != nil {
				return err
			}

			if response.VerificationAttributes[email] != nil {
				return fmt.Errorf("SES Email Identity %s still exists. Failing!", email)
			}
		}

		return nil
	}
}

func testAccCheckEmailIdentityExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Email Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Email Identity name not set")
		}

		email := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(email),
			},
		}

		response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[email] == nil {
			return fmt.Errorf("SES Email Identity %s not found in AWS", email)
		}

		return nil
	}
}

func testAccEmailIdentityConfig_basic(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %q
}
`, email)
}
