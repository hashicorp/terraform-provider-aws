// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESIdentityPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyConfig_domain(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(ctx, resourceName),
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

func TestAccSESIdentityPolicy_Identity_email(t *testing.T) {
	ctx := acctest.Context(t)
	emailPrefix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := fmt.Sprintf("%s@%s", emailPrefix, acctest.RandomDomainName())
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyConfig_email(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(ctx, resourceName),
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

func TestAccSESIdentityPolicy_policy(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyConfig_1(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(ctx, resourceName),
				),
			},
			{
				Config: testAccIdentityPolicyConfig_2(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(ctx, resourceName),
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

func TestAccSESIdentityPolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyConfig_equivalent(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(ctx, resourceName),
				),
			},
			{
				Config:   testAccIdentityPolicyConfig_equivalent2(rName, domain),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckIdentityPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_identity_policy" {
				continue
			}

			identityARN, policyName, err := tfses.IdentityPolicyParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &ses.GetIdentityPoliciesInput{
				Identity:    aws.String(identityARN),
				PolicyNames: []string{policyName},
			}

			output, err := conn.GetIdentityPolicies(ctx, input)

			if err != nil {
				return err
			}

			if output != nil && len(output.Policies) > 0 && output.Policies[policyName] != "" {
				return fmt.Errorf("SES Identity (%s) Policy (%s) still exists", identityARN, policyName)
			}
		}

		return nil
	}
}

func testAccCheckIdentityPolicyExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("SES Identity Policy not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Identity Policy ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		identityARN, policyName, err := tfses.IdentityPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &ses.GetIdentityPoliciesInput{
			Identity:    aws.String(identityARN),
			PolicyNames: []string{policyName},
		}

		output, err := conn.GetIdentityPolicies(ctx, input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Policies) == 0 {
			return fmt.Errorf("SES Identity (%s) Policy (%s) not found", identityARN, policyName)
		}

		return nil
	}
}

func testAccIdentityPolicyConfig_domain(domain string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccIdentityPolicyConfig_email(email string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_email_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_email_identity" "test" {
  email = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_email_identity.test.email
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, email)
}

func testAccIdentityPolicyConfig_1(domain string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccIdentityPolicyConfig_2(domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccIdentityPolicyConfig_equivalent(rName, domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = %[2]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[2]q
    Statement = [{
      Sid    = %[2]q
      Effect = "Allow"
      Principal = {
        AWS = [data.aws_caller_identity.current.account_id]
      }
      Action = [
        "SES:SendEmail",
        "SES:SendRawEmail",
      ]
      Resource = [aws_ses_domain_identity.test.arn]
    }]
  })
}
`, domain, rName)
}

func testAccIdentityPolicyConfig_equivalent2(rName, domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = %[2]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[2]q
    Statement = [{
      Sid    = %[2]q
      Effect = "Allow"
      Principal = {
        AWS = data.aws_caller_identity.current.account_id
      }
      Action = [
        "SES:SendRawEmail",
        "SES:SendEmail",
      ]
      Resource = aws_ses_domain_identity.test.arn
    }]
  })
}
`, domain, rName)
}
