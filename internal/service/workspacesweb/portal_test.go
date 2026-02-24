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

func TestAccWorkSpacesWebPortal_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var portal awstypes.Portal
	resourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, string(awstypes.InstanceTypeStandardRegular)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "portal_arn", "workspaces-web", regexache.MustCompile(`portal/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "portal_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "portal_status", string(awstypes.PortalStatusIncomplete)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_arn"),
				ImportStateVerifyIdentifierAttribute: "portal_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebPortal_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var portal awstypes.Portal
	resourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &portal),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourcePortal, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebPortal_update(t *testing.T) {
	ctx := acctest.Context(t)
	var portal awstypes.Portal
	resourceName := "aws_workspacesweb_portal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalConfig_updateBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-before"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, string(awstypes.InstanceTypeStandardRegular)),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_sessions", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", string(awstypes.AuthenticationTypeStandard)),
					resource.TestCheckResourceAttr(resourceName, "portal_status", string(awstypes.PortalStatusIncomplete)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_arn"),
				ImportStateVerifyIdentifierAttribute: "portal_arn",
			},
			{
				Config: testAccPortalConfig_updateAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-after"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, string(awstypes.InstanceTypeStandardLarge)),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_sessions", "2"),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", string(awstypes.AuthenticationTypeIamIdentityCenter)),
					resource.TestCheckResourceAttr(resourceName, "portal_status", string(awstypes.PortalStatusActive)),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebPortal_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var portal awstypes.Portal
	resourceName := "aws_workspacesweb_portal.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalConfig_complete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalExists(ctx, t, resourceName, &portal),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "test-complete"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, string(awstypes.InstanceTypeStandardLarge)),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_sessions", "2"),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", string(awstypes.AuthenticationTypeStandard)),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "portal_status", string(awstypes.PortalStatusIncomplete)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_arn"),
				ImportStateVerifyIdentifierAttribute: "portal_arn",
			},
		},
	})
}

func testAccCheckPortalDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_portal" {
				continue
			}

			_, err := tfworkspacesweb.FindPortalByARN(ctx, conn, rs.Primary.Attributes["portal_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Portal %s still exists", rs.Primary.Attributes["portal_arn"])
		}

		return nil
	}
}

func testAccCheckPortalExists(ctx context.Context, t *testing.T, n string, v *awstypes.Portal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindPortalByARN(ctx, conn, rs.Primary.Attributes["portal_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPortalConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_portal" "test" {
  display_name  = "test"
  instance_type = %q
}
`, string(awstypes.InstanceTypeStandardRegular))
}

func testAccPortalConfig_updateBefore() string {
	return testAccPortalConfig_template("test-before", string(awstypes.InstanceTypeStandardRegular), 1, string(awstypes.AuthenticationTypeStandard))
}

func testAccPortalConfig_updateAfter() string {
	return testAccPortalConfig_template("test-after", string(awstypes.InstanceTypeStandardLarge), 2, string(awstypes.AuthenticationTypeIamIdentityCenter))
}

func testAccPortalConfig_template(displayName, instanceType string, maxConcurrentSessions int, authenticationType string) string {
	return fmt.Sprintf(`
resource "aws_workspacesweb_portal" "test" {
  display_name            = %q
  instance_type           = %q
  max_concurrent_sessions = %d
  authentication_type     = %q
}
`, displayName, instanceType, maxConcurrentSessions, authenticationType)
}

func testAccPortalConfig_complete() string {
	return fmt.Sprintf(`

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
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

resource "aws_workspacesweb_portal" "test" {
  display_name            = "test-complete"
  instance_type           = %q
  max_concurrent_sessions = 2
  authentication_type     = %q
  customer_managed_key    = aws_kms_key.test.arn

  additional_encryption_context = {
    Environment = "Production"
  }
}
`, string(awstypes.InstanceTypeStandardLarge), string(awstypes.AuthenticationTypeStandard))
}
