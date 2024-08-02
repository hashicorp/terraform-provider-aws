// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESDomainIdentity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, "aws_ses_domain_identity.test"),
					testAccCheckDomainIdentityARN("aws_ses_domain_identity.test", domain),
				),
			},
		},
	})
}

func TestAccSESDomainIdentity_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, "aws_ses_domain_identity.test"),
					testAccCheckDomainIdentityDisappears(ctx, domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccSESDomainIdentity_trailingPeriod updated in 3.0 to account for domain plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccSESDomainIdentity_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDomainIdentityConfig_basic(domain),
				ExpectError: regexache.MustCompile(`invalid value for domain \(cannot end with a period\)`),
			},
		},
	})
}

func testAccCheckDomainIdentityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_domain_identity" {
				continue
			}

			domain := rs.Primary.ID
			params := &ses.GetIdentityVerificationAttributesInput{
				Identities: []*string{
					aws.String(domain),
				},
			}

			response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, params)
			if err != nil {
				return err
			}

			if response.VerificationAttributes[domain] != nil {
				return fmt.Errorf("SES Domain Identity %s still exists. Failing!", domain)
			}
		}

		return nil
	}
}

func testAccCheckDomainIdentityExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityVerificationAttributesWithContext(ctx, params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[domain] == nil {
			return fmt.Errorf("SES Domain Identity %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckDomainIdentityDisappears(ctx context.Context, identity string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		input := &ses.DeleteIdentityInput{
			Identity: aws.String(identity),
		}

		_, err := conn.DeleteIdentityWithContext(ctx, input)

		return err
	}
}

func testAccCheckDomainIdentityARN(n string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		awsClient := acctest.Provider.Meta().(*conns.AWSClient)

		expected := arn.ARN{
			AccountID: awsClient.AccountID,
			Partition: awsClient.Partition,
			Region:    awsClient.Region,
			Resource:  fmt.Sprintf("identity/%s", strings.TrimSuffix(domain, ".")),
			Service:   "ses",
		}

		if rs.Primary.Attributes[names.AttrARN] != expected.String() {
			return fmt.Errorf("Incorrect ARN: expected %q, got %q", expected, rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

	input := &ses.ListIdentitiesInput{}

	_, err := conn.ListIdentitiesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDomainIdentityConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}
`, domain)
}
