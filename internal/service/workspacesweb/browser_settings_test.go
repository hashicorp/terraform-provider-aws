// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebBrowserSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	browserPolicy1 := `{
    "chromePolicies":
    {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
} `
	browserPolicy2 := `{
    "chromePolicies":
    {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles2"
        }
    }
} `
	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsConfig_basic(browserPolicy1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "browser_settings_arn", "workspaces-web", regexache.MustCompile(`browserSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
			{
				Config: testAccBrowserSettingsConfig_basic(browserPolicy2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy2),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "browser_settings_arn", "workspaces-web", regexache.MustCompile(`browserSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebBrowserSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	browserPolicy := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
}`
	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsConfig_basic(browserPolicy),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceBrowserSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebBrowserSettings_additionalEncryptionContext(t *testing.T) {
	ctx := acctest.Context(t)
	browserPolicy1 := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
}`
	browserPolicy2 := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles2"
        }
    }
}`
	encryptionContext1 := map[string]string{
		"department": "finance",
		"project":    "alpha",
	}
	encryptionContext2 := map[string]string{
		"department": "hr",
		"project":    "beta",
	}

	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsConfig_additionalEncryptionContext(browserPolicy1, encryptionContext1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy1),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.department", "finance"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.project", "alpha"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
			{
				Config: testAccBrowserSettingsConfig_additionalEncryptionContext(browserPolicy2, encryptionContext2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy2),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.department", "hr"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.project", "beta"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebBrowserSettings_customerManagedKey(t *testing.T) {
	ctx := acctest.Context(t)
	browserPolicy1 := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
}`
	browserPolicy2 := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles2"
        }
    }
}`
	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings.test"

	customerManagedKey1 := "aws_kms_key.test1"
	customerManagedKey2 := "aws_kms_key.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsConfig_customerManagedKey(browserPolicy1, customerManagedKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy1),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", customerManagedKey1, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
			{
				Config: testAccBrowserSettingsConfig_customerManagedKey(browserPolicy2, customerManagedKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy2),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", customerManagedKey2, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebBrowserSettings_complete(t *testing.T) {
	ctx := acctest.Context(t)
	browserPolicy1 := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
}`
	browserPolicy2 := `{
    "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles2"
        }
    }
}`
	encryptionContext := map[string]string{
		"department": "finance",
		"project":    "alpha",
	}
	var browserSettings awstypes.BrowserSettings
	resourceName := "aws_workspacesweb_browser_settings.test"
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserSettingsConfig_complete(browserPolicy1, encryptionContext),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy1),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.department", "finance"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.project", "alpha"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", keyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
			{
				Config: testAccBrowserSettingsConfig_complete(browserPolicy2, encryptionContext),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserSettingsExists(ctx, t, resourceName, &browserSettings),
					resource.TestCheckResourceAttr(resourceName, "browser_policy", browserPolicy2),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.department", "finance"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.project", "alpha"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", keyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "browser_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "browser_settings_arn",
			},
		},
	})
}

func testAccCheckBrowserSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_browser_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindBrowserSettingsByARN(ctx, conn, rs.Primary.Attributes["browser_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Browser Settings %s still exists", rs.Primary.Attributes["browser_settings_arn"])
		}

		return nil
	}
}

func testAccCheckBrowserSettingsExists(ctx context.Context, t *testing.T, n string, v *awstypes.BrowserSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindBrowserSettingsByARN(ctx, conn, rs.Primary.Attributes["browser_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

	input := workspacesweb.ListBrowserSettingsInput{}

	_, err := conn.ListBrowserSettings(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBrowserSettingsConfig_basic(browserPolicy string) string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = %[1]q
}
`, browserPolicy)
}

func testAccBrowserSettingsConfig_additionalEncryptionContext(browserPolicy string, encryptionContext map[string]string) string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = %[1]q
  additional_encryption_context = {
%[2]s
  }
}
`, browserPolicy, testAccBrowserSettingsAdditionalEncryptionContextString(encryptionContext))
}

func testAccBrowserSettingsAdditionalEncryptionContextString(m map[string]string) string {
	var items []string
	for k, v := range m {
		items = append(items, fmt.Sprintf("    %s = %q", k, v))
	}
	return strings.Join(items, "\n")
}

func testAccBrowserSettingsConfig_customerManagedKey(browserPolicy, customerManagedKey string) string {
	return fmt.Sprintf(`
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

resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy       = %[1]q
  customer_managed_key = %[2]s.arn
}
`, browserPolicy, customerManagedKey)
}

func testAccBrowserSettingsConfig_complete(browserPolicy string, encryptionContext map[string]string) string {
	return fmt.Sprintf(`
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

resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy       = %[1]q
  customer_managed_key = aws_kms_key.test.arn
  additional_encryption_context = {
%[2]s
  }
}
`, browserPolicy, testAccBrowserSettingsAdditionalEncryptionContextString(encryptionContext))
}
