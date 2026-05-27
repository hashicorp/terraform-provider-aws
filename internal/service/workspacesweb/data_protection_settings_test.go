// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebDataProtectionSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "data_protection_settings_arn", "workspaces-web", regexache.MustCompile(`dataProtectionSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "data_protection_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "data_protection_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebDataProtectionSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceDataProtectionSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebDataProtectionSettings_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsConfig_complete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-complete"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Testing"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_confidence_level", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.0", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.1", "https://test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.0", "https://exempt.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.built_in_pattern_id", "emailAddress"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.confidence_level", "3"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.enforced_urls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.enforced_urls.0", "https://pattern1.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.exempt_urls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.exempt_urls.0", "https://exempt-pattern1.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.redaction_place_holder.0.redaction_place_holder_type", "CustomText"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.redaction_place_holder.0.redaction_place_holder_text", "REDACTED-SSN"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_name", "CustomPattern"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_regex", "/\\d{3}-\\d{2}-\\d{4}/g"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.keyword_regex", "/SSN|Social Security/gi"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_description", "Custom SSN pattern"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.redaction_place_holder.0.redaction_place_holder_type", "CustomText"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.redaction_place_holder.0.redaction_place_holder_text", "REDACTED-CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "dev"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "data_protection_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "data_protection_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebDataProtectionSettings_additionalEncryptionContextUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsConfig_additionalEncryptionContextBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Test"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "data_protection_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "data_protection_settings_arn",
			},
			{
				Config: testAccDataProtectionSettingsConfig_additionalEncryptionContextAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Live"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebDataProtectionSettings_customerManagedKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings.test"
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsConfig_customerManagedKeyBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "data_protection_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "data_protection_settings_arn",
			},
			{
				Config: testAccDataProtectionSettingsConfig_customerManagedKeyAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebDataProtectionSettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	var dataProtectionSettings awstypes.DataProtectionSettings
	resourceName := "aws_workspacesweb_data_protection_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionSettingsConfig_updateBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-update"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description before"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", "aws_kms_key.test_update", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Testing"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_confidence_level", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.0", "https://before.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.0", "https://exempt-before.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.built_in_pattern_id", "ssn"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.confidence_level", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.enforced_urls.0", "https://pattern-before.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.exempt_urls.0", "https://exempt-pattern-before.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.redaction_place_holder.0.redaction_place_holder_type", "CustomText"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.redaction_place_holder.0.redaction_place_holder_text", "REDACTED-CC"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_name", "CustomPatternBefore"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_regex", "/\\d{4}-\\d{4}-\\d{4}-\\d{4}/g"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.keyword_regex", "/ssn/gi"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_description", "SSN pattern"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.redaction_place_holder.0.redaction_place_holder_type", "CustomText"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.redaction_place_holder.0.redaction_place_holder_text", "REDACTED-CUSTOM-BEF"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "before"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "dev"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "data_protection_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "data_protection_settings_arn",
			},
			{
				Config: testAccDataProtectionSettingsConfig_updateAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProtectionSettingsExists(ctx, t, resourceName, &dataProtectionSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-update-after"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description after"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", "aws_kms_key.test_update", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Testing"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_confidence_level", "3"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.0", "https://after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_enforced_urls.1", "https://second-after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.0", "https://exempt-after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.global_exempt_urls.1", "https://second-exempt-after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.built_in_pattern_id", "phoneNum"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.confidence_level", "3"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.enforced_urls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.enforced_urls.0", "https://pattern-after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.enforced_urls.1", "https://second-pattern-after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.exempt_urls.0", "https://exempt-pattern-after.example.com"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.redaction_place_holder.0.redaction_place_holder_type", "CustomText"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.0.redaction_place_holder.0.redaction_place_holder_text", "REDACTED-PHONE"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_name", "CustomPatternAfter"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_regex", "/\\d{3}-\\d{3}-\\d{4}/g"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.keyword_regex", "/phone|number/gi"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.custom_pattern.0.pattern_description", "Custom phone number pattern"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.redaction_place_holder.0.redaction_place_holder_type", "CustomText"),
					resource.TestCheckResourceAttr(resourceName, "inline_redaction_configuration.0.inline_redaction_pattern.1.redaction_place_holder.0.redaction_place_holder_text", "REDACTED-CUSTOM-AFT"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "after"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "prod"),
					resource.TestCheckResourceAttr(resourceName, "tags.Owner", "team"),
				),
			},
		},
	})
}

func testAccCheckDataProtectionSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_data_protection_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindDataProtectionSettingsByARN(ctx, conn, rs.Primary.Attributes["data_protection_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Data Protection Settings %s still exists", rs.Primary.Attributes["data_protection_settings_arn"])
		}

		return nil
	}
}

func testAccCheckDataProtectionSettingsExists(ctx context.Context, t *testing.T, n string, v *awstypes.DataProtectionSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindDataProtectionSettingsByARN(ctx, conn, rs.Primary.Attributes["data_protection_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDataProtectionSettingsConfig_basic() string {
	return `
resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name = "test"
}
`
}

func testAccDataProtectionSettingsConfig_complete() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-complete"
  description          = "test description"
  customer_managed_key = aws_kms_key.test.arn

  additional_encryption_context = {
    Environment = "Production"
    Project     = "Testing"
  }

  inline_redaction_configuration {
    global_confidence_level = 2
    global_enforced_urls    = ["https://example.com", "https://test.example.com"]
    global_exempt_urls      = ["https://exempt.example.com"]

    inline_redaction_pattern {
      built_in_pattern_id = "emailAddress"
      confidence_level    = 3
      enforced_urls       = ["https://pattern1.example.com"]
      exempt_urls         = ["https://exempt-pattern1.example.com"]
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-SSN"
      }
    }

    inline_redaction_pattern {
      custom_pattern {
        pattern_name        = "CustomPattern"
        pattern_regex       = "/\\d{3}-\\d{2}-\\d{4}/g"
        keyword_regex       = "/SSN|Social Security/gi"
        pattern_description = "Custom SSN pattern"
      }
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-CUSTOM"
      }
    }
  }

  tags = {
    Name        = "test"
    Environment = "dev"
  }
}
`
}

func testAccDataProtectionSettingsConfig_updateBefore() string {
	return `
resource "aws_kms_key" "test_update" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-update"
  description          = "test description before"
  customer_managed_key = aws_kms_key.test_update.arn

  additional_encryption_context = {
    Environment = "Development"
    Project     = "Testing"
  }

  inline_redaction_configuration {
    global_confidence_level = 2
    global_enforced_urls    = ["https://before.example.com"]
    global_exempt_urls      = ["https://exempt-before.example.com"]

    inline_redaction_pattern {
      built_in_pattern_id = "ssn"
      confidence_level    = 2
      enforced_urls       = ["https://pattern-before.example.com"]
      exempt_urls         = ["https://exempt-pattern-before.example.com"]
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-CC"
      }
    }

    inline_redaction_pattern {
      custom_pattern {
        pattern_name        = "CustomPatternBefore"
        pattern_regex       = "/\\d{4}-\\d{4}-\\d{4}-\\d{4}/g"
        keyword_regex       = "/ssn/gi"
        pattern_description = "SSN pattern"
      }
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-CUSTOM-BEF"
      }
    }
  }

  tags = {
    Name        = "before"
    Environment = "dev"
  }
}
`
}

func testAccDataProtectionSettingsConfig_additionalEncryptionContextBefore() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-encryption-context"
  customer_managed_key = aws_kms_key.test.arn
  additional_encryption_context = {
    Environment = "Development"
    Project     = "Test"
  }
}
`
}

func testAccDataProtectionSettingsConfig_additionalEncryptionContextAfter() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-encryption-context"
  customer_managed_key = aws_kms_key.test.arn
  additional_encryption_context = {
    Environment = "Production"
    Project     = "Live"
  }
}
`
}

func testAccDataProtectionSettingsConfig_customerManagedKeyBefore() string {
	return `
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-cmk"
  customer_managed_key = aws_kms_key.test1.arn
}
`
}

func testAccDataProtectionSettingsConfig_customerManagedKeyAfter() string {
	return `
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-cmk"
  customer_managed_key = aws_kms_key.test2.arn
}
`
}

func testAccDataProtectionSettingsConfig_updateAfter() string {
	return `
resource "aws_kms_key" "test_update" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_data_protection_settings" "test" {
  display_name         = "test-update-after"
  description          = "test description after"
  customer_managed_key = aws_kms_key.test_update.arn

  additional_encryption_context = {
    Environment = "Development"
    Project     = "Testing"
  }

  inline_redaction_configuration {
    global_confidence_level = 3
    global_enforced_urls    = ["https://after.example.com", "https://second-after.example.com"]
    global_exempt_urls      = ["https://exempt-after.example.com", "https://second-exempt-after.example.com"]

    inline_redaction_pattern {
      built_in_pattern_id = "phoneNum"
      confidence_level    = 3
      enforced_urls       = ["https://pattern-after.example.com", "https://second-pattern-after.example.com"]
      exempt_urls         = ["https://exempt-pattern-after.example.com"]
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-PHONE"
      }
    }

    inline_redaction_pattern {
      custom_pattern {
        pattern_name        = "CustomPatternAfter"
        pattern_regex       = "/\\d{3}-\\d{3}-\\d{4}/g"
        keyword_regex       = "/phone|number/gi"
        pattern_description = "Custom phone number pattern"
      }
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-CUSTOM-AFT"
      }
    }
  }

  tags = {
    Name        = "after"
    Environment = "prod"
    Owner       = "team"
  }
}
`
}
