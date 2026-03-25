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

func TestAccWorkSpacesWebIPAccessSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.ip_range", "10.0.0.0/16"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "ip_access_settings_arn", "workspaces-web", regexache.MustCompile(`ipAccessSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "ip_access_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "ip_access_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebIPAccessSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceIPAccessSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebIPAccessSettings_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsConfig_complete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-complete"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.ip_range", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.description", "Main office"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.1.ip_range", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.1.description", "Branch office"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "ip_access_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "ip_access_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebIPAccessSettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsConfig_updateBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-update"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description before"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.ip_range", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.description", "Main office"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "ip_access_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "ip_access_settings_arn",
			},
			{
				Config: testAccIPAccessSettingsConfig_updateAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-update-after"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description after"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.ip_range", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.0.description", "Main office updated"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.1.ip_range", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "ip_rule.1.description", "Branch office"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebIPAccessSettings_additionalEncryptionContextUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAccessSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsConfig_additionalEncryptionContextBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Test"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "ip_access_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "ip_access_settings_arn",
			},
			{
				Config: testAccIPAccessSettingsConfig_additionalEncryptionContextAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Live"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebIPAccessSettings_customerManagedKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var ipAccessSettings awstypes.IpAccessSettings
	resourceName := "aws_workspacesweb_ip_access_settings.test"
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
		CheckDestroy:             testAccCheckIPAccessSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAccessSettingsConfig_customerManagedKeyBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "ip_access_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "ip_access_settings_arn",
			},
			{
				Config: testAccIPAccessSettingsConfig_customerManagedKeyAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAccessSettingsExists(ctx, t, resourceName, &ipAccessSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckIPAccessSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_ip_access_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindIPAccessSettingsByARN(ctx, conn, rs.Primary.Attributes["ip_access_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web IP Access Settings %s still exists", rs.Primary.Attributes["ip_access_settings_arn"])
		}

		return nil
	}
}

func testAccCheckIPAccessSettingsExists(ctx context.Context, t *testing.T, n string, v *awstypes.IpAccessSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindIPAccessSettingsByARN(ctx, conn, rs.Primary.Attributes["ip_access_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIPAccessSettingsConfig_basic() string {
	return `
resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name = "test"

  ip_rule {
    ip_range = "10.0.0.0/16"
  }
}
`
}

func testAccIPAccessSettingsConfig_complete() string {
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

resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name         = "test-complete"
  description          = "test description"
  customer_managed_key = aws_kms_key.test.arn

  additional_encryption_context = {
    Environment = "Production"
  }

  ip_rule {
    ip_range    = "10.0.0.0/16"
    description = "Main office"
  }

  ip_rule {
    ip_range    = "192.168.0.0/24"
    description = "Branch office"
  }

  tags = {
    Name = "test"
  }
}
`
}

func testAccIPAccessSettingsConfig_updateBefore() string {
	return `
resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name = "test-update"
  description  = "test description before"
  ip_rule {
    ip_range    = "10.0.0.0/16"
    description = "Main office"
  }
}
`
}

func testAccIPAccessSettingsConfig_updateAfter() string {
	return `
resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name = "test-update-after"
  description  = "test description after"

  ip_rule {
    ip_range    = "10.0.0.0/16"
    description = "Main office updated"
  }

  ip_rule {
    ip_range    = "192.168.0.0/24"
    description = "Branch office"
  }
}
`
}

func testAccIPAccessSettingsConfig_additionalEncryptionContextBefore() string {
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

resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name         = "test-encryption-context"
  customer_managed_key = aws_kms_key.test.arn

  ip_rule {
    ip_range = "10.0.0.0/16"
  }

  additional_encryption_context = {
    Environment = "Development"
    Project     = "Test"
  }
}
`
}

func testAccIPAccessSettingsConfig_additionalEncryptionContextAfter() string {
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

resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name         = "test-encryption-context"
  customer_managed_key = aws_kms_key.test.arn

  ip_rule {
    ip_range = "10.0.0.0/16"
  }

  additional_encryption_context = {
    Environment = "Production"
    Project     = "Live"
  }
}
`
}

func testAccIPAccessSettingsConfig_customerManagedKeyBefore() string {
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

resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name         = "test-cmk"
  customer_managed_key = aws_kms_key.test1.arn
  ip_rule {
    ip_range = "10.0.0.0/16"
  }
}
`
}

func testAccIPAccessSettingsConfig_customerManagedKeyAfter() string {
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

resource "aws_workspacesweb_ip_access_settings" "test" {
  display_name         = "test-cmk"
  customer_managed_key = aws_kms_key.test2.arn
  ip_rule {
    ip_range = "10.0.0.0/16"
  }
}
`
}
