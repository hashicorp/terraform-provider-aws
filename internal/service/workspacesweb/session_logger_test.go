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

func TestAccWorkSpacesWebSessionLogger_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerConfig_basic(rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_configuration.0.s3.0.bucket", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.folder_structure", string(awstypes.FolderStructureFlat)),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.log_file_format", string(awstypes.LogFileFormatJson)),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.all.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "session_logger_arn", "workspaces-web", regexache.MustCompile(`sessionLogger/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "session_logger_arn"),
				ImportStateVerifyIdentifierAttribute: "session_logger_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebSessionLogger_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerConfig_complete(rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson), string(awstypes.EventSessionStart)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "customer_managed_key"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.test", names.AttrValue),
					resource.TestCheckResourceAttrSet(resourceName, "log_configuration.0.s3.0.bucket_owner"),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.key_prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.include.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_filter.0.include.*", string(awstypes.EventSessionStart)),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebSessionLogger_update(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerConfig_basic(rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.folder_structure", string(awstypes.FolderStructureFlat)),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.log_file_format", string(awstypes.LogFileFormatJson)),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.all.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.include.#", "0"),
				),
			},
			{
				Config: testAccSessionLoggerConfig_update(rName2, string(awstypes.FolderStructureNestedByDate), string(awstypes.LogFileFormatJsonLines), string(awstypes.EventSessionStart), string(awstypes.EventSessionEnd)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName2),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.folder_structure", string(awstypes.FolderStructureNestedByDate)),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.log_file_format", string(awstypes.LogFileFormatJsonLines)),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3.0.key_prefix", "updated-logs/"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.all.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "event_filter.0.include.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_filter.0.include.*", string(awstypes.EventSessionStart)),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_filter.0.include.*", string(awstypes.EventSessionEnd)),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebSessionLogger_customerManagedKey(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerConfig_customerManagedKey(rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", "aws_kms_key.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebSessionLogger_additionalEncryptionContext(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerConfig_additionalEncryptionContext(rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.test", names.AttrValue),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebSessionLogger_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerConfig_basic(rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerExists(ctx, t, resourceName, &sessionLogger),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceSessionLogger, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSessionLoggerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_session_logger" {
				continue
			}

			_, err := tfworkspacesweb.FindSessionLoggerByARN(ctx, conn, rs.Primary.Attributes["session_logger_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Session Logger %s still exists", rs.Primary.Attributes["session_logger_arn"])
		}

		return nil
	}
}

func testAccCheckSessionLoggerExists(ctx context.Context, t *testing.T, n string, v *awstypes.SessionLogger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindSessionLoggerByARN(ctx, conn, rs.Primary.Attributes["session_logger_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSessionLoggerConfig_s3Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "allow_write_access" {
  statement {
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }

    actions = [
      "s3:PutObject"
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_s3_bucket_policy" "allow_write_access" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.allow_write_access.json
}
`, rName)
}

func testAccSessionLoggerConfig_kmsBase(rName string) string {
	return testAccSessionLoggerConfig_s3Base(rName) + `
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "kms_key_policy" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    actions   = ["kms:*"]
    resources = ["*"]
  }

  statement {
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }
    actions = [
      "kms:Encrypt",
      "kms:GenerateDataKey*",
      "kms:ReEncrypt*",
      "kms:Decrypt"
    ]
    resources = ["*"]
  }
}

resource "aws_kms_key" "test" {
  description = "Test key for session logger"
  policy      = data.aws_iam_policy_document.kms_key_policy.json
}
`
}

func testAccSessionLoggerConfig_basic(rName, folderStructureType, logFileFormat string) string {
	return testAccSessionLoggerConfig_s3Base(rName) + fmt.Sprintf(`
resource "aws_workspacesweb_session_logger" "test" {
  display_name = %[1]q

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.id
      folder_structure = %[2]q
      log_file_format  = %[3]q
    }
  }

  event_filter {
    all {}
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]
}
`, rName, folderStructureType, logFileFormat)
}

func testAccSessionLoggerConfig_complete(rName, folderStructureType, logFileFormat, event string) string {
	return testAccSessionLoggerConfig_kmsBase(rName) + fmt.Sprintf(`
resource "aws_workspacesweb_session_logger" "test" {
  display_name         = %[1]q
  customer_managed_key = aws_kms_key.test.arn
  additional_encryption_context = {
    test = "value"
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.id
      bucket_owner     = data.aws_caller_identity.current.account_id
      folder_structure = %[2]q
      key_prefix       = "logs/"
      log_file_format  = %[3]q
    }
  }

  event_filter {
    include = [%[4]q]
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]
}
`, rName, folderStructureType, logFileFormat, event)
}

func testAccSessionLoggerConfig_update(rName, folderStructureType, logFileFormat, event1, event2 string) string {
	return testAccSessionLoggerConfig_s3Base(rName) + fmt.Sprintf(`
resource "aws_workspacesweb_session_logger" "test" {
  display_name = %[1]q

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.id
      folder_structure = %[2]q
      key_prefix       = "updated-logs/"
      log_file_format  = %[3]q
    }
  }

  event_filter {
    include = [%[4]q, %[5]q]
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]
}
`, rName, folderStructureType, logFileFormat, event1, event2)
}

func testAccSessionLoggerConfig_customerManagedKey(rName, folderStructureType, logFileFormat string) string {
	return testAccSessionLoggerConfig_kmsBase(rName) + fmt.Sprintf(`
resource "aws_workspacesweb_session_logger" "test" {
  display_name         = %[1]q
  customer_managed_key = aws_kms_key.test.arn

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.id
      folder_structure = %[2]q
      log_file_format  = %[3]q
    }
  }

  event_filter {
    all {}
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]
}
`, rName, folderStructureType, logFileFormat)
}

func testAccSessionLoggerConfig_additionalEncryptionContext(rName, folderStructureType, logFileFormat string) string {
	return testAccSessionLoggerConfig_s3Base(rName) + fmt.Sprintf(`
resource "aws_workspacesweb_session_logger" "test" {
  display_name = %[1]q
  additional_encryption_context = {
    test = "value"
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.id
      folder_structure = %[2]q
      log_file_format  = %[3]q
    }
  }

  event_filter {
    all {}
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]
}
`, rName, folderStructureType, logFileFormat)
}
