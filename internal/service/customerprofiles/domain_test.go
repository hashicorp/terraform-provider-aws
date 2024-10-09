// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customerprofiles_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/customerprofiles"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCustomerProfilesDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_customerprofiles_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_base(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_expiration_days", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_base(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_expiration_days", "365"),
				),
			},
		},
	})
}

func TestAccCustomerProfilesDomain_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_customerprofiles_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_expiration_days", "120"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "matching.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.conflict_resolution.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.conflict_resolution.0.conflict_resolving_model", "SOURCE"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.conflict_resolution.0.source_name", "FirstName"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.consolidation.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.consolidation.0.matching_attributes_list.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.min_allowed_confidence_score_for_merging", "0.1"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.exporting_config.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "matching.0.exporting_config.0.s3_exporting.0.%", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "matching.0.exporting_config.0.s3_exporting.0.s3_bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.exporting_config.0.s3_exporting.0.s3_key_name", "example/"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.job_schedule.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "matching.0.job_schedule.0.day_of_the_week", "MONDAY"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.job_schedule.0.time", "18:00"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.%", "8"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.address.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.attribute_matching_model", "ONE_TO_ONE"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.email_address.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.phone_number.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.conflict_resolution.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.conflict_resolution.0.conflict_resolving_model", "SOURCE"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.conflict_resolution.0.source_name", "FirstName"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.exporting_config.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.exporting_config.0.s3_exporting.0.%", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "rule_based_matching.0.exporting_config.0.s3_exporting.0.s3_bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.exporting_config.0.s3_exporting.0.s3_key_name", "example/"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.matching_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.max_allowed_rule_level_for_matching", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.max_allowed_rule_level_for_merging", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "rule_based_matching.0.status"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_fullUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_expiration_days", "365"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "matching.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.conflict_resolution.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.conflict_resolution.0.conflict_resolving_model", "RECENCY"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.conflict_resolution.0.source_name", ""),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.consolidation.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.consolidation.0.matching_attributes_list.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "matching.0.auto_merging.0.min_allowed_confidence_score_for_merging", "0.8"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.exporting_config.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "matching.0.exporting_config.0.s3_exporting.0.%", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "matching.0.exporting_config.0.s3_exporting.0.s3_bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.exporting_config.0.s3_exporting.0.s3_key_name", "exampleupdated/"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.job_schedule.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "matching.0.job_schedule.0.day_of_the_week", "SUNDAY"),
					resource.TestCheckResourceAttr(resourceName, "matching.0.job_schedule.0.time", "20:00"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.%", "8"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.address.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.attribute_matching_model", "MANY_TO_MANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.email_address.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.attribute_types_selector.0.phone_number.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.conflict_resolution.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.conflict_resolution.0.conflict_resolving_model", "RECENCY"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.conflict_resolution.0.source_name", ""),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.exporting_config.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.exporting_config.0.s3_exporting.0.%", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "rule_based_matching.0.exporting_config.0.s3_exporting.0.s3_bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.exporting_config.0.s3_exporting.0.s3_key_name", "exampleupdated/"),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.matching_rules.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.max_allowed_rule_level_for_matching", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_based_matching.0.max_allowed_rule_level_for_merging", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "rule_based_matching.0.status"),
				),
			},
		},
	})
}

func TestAccCustomerProfilesDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_customerprofiles_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfig_tags(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCustomerProfilesDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_customerprofiles_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_base(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, customerprofiles.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CustomerProfilesClient(ctx)

		_, err := customerprofiles.FindDomainByDomainName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CustomerProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_customerprofiles_domain" {
				continue
			}

			_, err := customerprofiles.FindDomainByDomainName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Customer Profiles Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainConfig_base(rName string, defaultExpirationDays int) string {
	return fmt.Sprintf(`
resource "aws_customerprofiles_domain" "test" {
  domain_name             = %[1]q
  default_expiration_days = %[2]d
}
`, rName, defaultExpirationDays)
}

func testAccDomainConfig_tags(rName string, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_customerprofiles_domain" "test" {
  domain_name             = %[1]q
  default_expiration_days = 365

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey, tagValue)
}

func testAccDomainConfig_tags2(rName string, tagKey, tagValue, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_customerprofiles_domain" "test" {
  domain_name             = %[1]q
  default_expiration_days = 365

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey, tagValue, tagKey2, tagValue2)
}

func testAccDomainConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Customer Profiles SQS policy"
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = "*"
        Principal = {
          Service = "profile.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_kms_key" "test" {
  description             = "test"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "example" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Customer Profiles S3 policy"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
        ]
        Resource = [
          aws_s3_bucket.example.arn,
          "${aws_s3_bucket.example.arn}/*",
        ]
        Principal = {
          Service = "profile.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_customerprofiles_domain" "test" {
  domain_name             = %[1]q
  dead_letter_queue_url   = aws_sqs_queue.test.id
  default_encryption_key  = aws_kms_key.test.arn
  default_expiration_days = 120

  matching {
    enabled = true

    auto_merging {
      enabled = true

      conflict_resolution {
        conflict_resolving_model = "SOURCE"
        source_name              = "FirstName"
      }

      consolidation {
        matching_attributes_list = [["PhoneNumber"], ["FirstName"]]
      }

      min_allowed_confidence_score_for_merging = 0.1
    }

    exporting_config {
      s3_exporting {
        s3_bucket_name = aws_s3_bucket.example.id
        s3_key_name    = "example/"
      }
    }

    job_schedule {
      day_of_the_week = "MONDAY"
      time            = "18:00"
    }
  }

  rule_based_matching {
    enabled = true

    attribute_types_selector {
      attribute_matching_model = "ONE_TO_ONE"
      address                  = ["Address"]
      email_address            = ["EmailAddress"]
      phone_number             = ["PhoneNumber"]
    }

    conflict_resolution {
      conflict_resolving_model = "SOURCE"
      source_name              = "FirstName"
    }

    exporting_config {
      s3_exporting {
        s3_bucket_name = aws_s3_bucket.example.id
        s3_key_name    = "example/"
      }
    }

    matching_rules {
      rule = ["Address.City", "Address.Country"]
    }

    max_allowed_rule_level_for_matching = 1
    max_allowed_rule_level_for_merging  = 1
  }
}
`, rName)
}

func testAccDomainConfig_fullUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Customer Profiles SQS policy"
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = "*"
        Principal = {
          Service = "profile.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_kms_key" "test" {
  description             = "test"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "example" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Customer Profiles S3 policy"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
        ]
        Resource = [
          aws_s3_bucket.example.arn,
          "${aws_s3_bucket.example.arn}/*",
        ]
        Principal = {
          Service = "profile.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_customerprofiles_domain" "test" {
  domain_name             = %[1]q
  dead_letter_queue_url   = aws_sqs_queue.test.id
  default_encryption_key  = aws_kms_key.test.arn
  default_expiration_days = 365

  matching {
    enabled = true

    auto_merging {
      enabled = true

      conflict_resolution {
        conflict_resolving_model = "RECENCY"
      }

      consolidation {
        matching_attributes_list = [["EmailAddress"]]
      }

      min_allowed_confidence_score_for_merging = 0.8
    }

    exporting_config {
      s3_exporting {
        s3_bucket_name = aws_s3_bucket.example.id
        s3_key_name    = "exampleupdated/"
      }
    }

    job_schedule {
      day_of_the_week = "SUNDAY"
      time            = "20:00"
    }
  }

  rule_based_matching {
    enabled = true

    attribute_types_selector {
      attribute_matching_model = "MANY_TO_MANY"
      address                  = ["Address"]
      email_address            = ["EmailAddress", "BusinessEmailAddress"]
      phone_number             = ["PhoneNumber", "HomePhoneNumber"]
    }

    conflict_resolution {
      conflict_resolving_model = "RECENCY"
    }

    exporting_config {
      s3_exporting {
        s3_bucket_name = aws_s3_bucket.example.id
        s3_key_name    = "exampleupdated/"
      }
    }

    matching_rules {
      rule = ["Address.City", "Address.Country"]
    }

    matching_rules {
      rule = ["EmailAddress"]
    }

    matching_rules {
      rule = ["PhoneNumber"]
    }

    max_allowed_rule_level_for_matching = 2
    max_allowed_rule_level_for_merging  = 2
  }
}
`, rName)
}
