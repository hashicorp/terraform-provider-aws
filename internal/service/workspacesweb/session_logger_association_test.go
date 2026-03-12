// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebSessionLoggerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger_association.test"
	sessionLoggerResourceName := "aws_workspacesweb_session_logger.test"
	portalResourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerAssociationConfig_basic(rName, rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerAssociationExists(ctx, t, resourceName, &sessionLogger),
					resource.TestCheckResourceAttrPair(resourceName, "session_logger_arn", sessionLoggerResourceName, "session_logger_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "portal_arn", portalResourceName, "portal_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccSessionLoggerAssociationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "session_logger_arn",
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
			},
			{
				Config: testAccSessionLoggerAssociationConfig_basic(rName, rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					//The following checks are for the SessionLogger Resource and the PortalResource (and not for the association resource).
					resource.TestCheckResourceAttr(sessionLoggerResourceName, "associated_portal_arns.#", "1"),
					resource.TestCheckResourceAttrPair(sessionLoggerResourceName, "associated_portal_arns.0", portalResourceName, "portal_arn"),
					resource.TestCheckResourceAttrPair(portalResourceName, "session_logger_arn", sessionLoggerResourceName, "session_logger_arn"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebSessionLoggerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sessionLogger awstypes.SessionLogger
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workspacesweb_session_logger_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSessionLoggerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSessionLoggerAssociationConfig_basic(rName, rName, string(awstypes.FolderStructureFlat), string(awstypes.LogFileFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionLoggerAssociationExists(ctx, t, resourceName, &sessionLogger),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceSessionLoggerAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSessionLoggerAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_session_logger_association" {
				continue
			}

			sessionLogger, err := tfworkspacesweb.FindSessionLoggerByARN(ctx, conn, rs.Primary.Attributes["session_logger_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if the portal is still associated
			portalARN := rs.Primary.Attributes["portal_arn"]
			if slices.Contains(sessionLogger.AssociatedPortalArns, portalARN) {
				return fmt.Errorf("WorkSpaces Web Session Logger Association %s still exists", rs.Primary.Attributes["session_logger_arn"])
			}
		}

		return nil
	}
}

func testAccCheckSessionLoggerAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.SessionLogger) resource.TestCheckFunc {
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

		// Check if the portal is associated
		portalARN := rs.Primary.Attributes["portal_arn"]
		if !slices.Contains(output.AssociatedPortalArns, portalARN) {
			return fmt.Errorf("Association not found")
		}

		*v = *output

		return nil
	}
}

func testAccSessionLoggerAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["session_logger_arn"], rs.Primary.Attributes["portal_arn"]), nil
	}
}

func testAccSessionLoggerAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_portal" "test" {
  display_name = %[1]q
}

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

func testAccSessionLoggerAssociationConfig_basic(rName, sessionLoggerName, folderStructureType, logFileFormat string) string {
	return testAccSessionLoggerAssociationConfig_base(rName) + fmt.Sprintf(`
resource "aws_workspacesweb_session_logger" "test" {
  display_name = %[1]q

  event_filter {
    all {}
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.id
      folder_structure = %[2]q
      log_file_format  = %[3]q
    }
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]
}

resource "aws_workspacesweb_session_logger_association" "test" {
  portal_arn         = aws_workspacesweb_portal.test.portal_arn
  session_logger_arn = aws_workspacesweb_session_logger.test.session_logger_arn
}
`, sessionLoggerName, folderStructureType, logFileFormat)
}
