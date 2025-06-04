// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebUserAccessLoggingSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userAccessLoggingSettings awstypes.UserAccessLoggingSettings
	resourceName := "aws_workspacesweb_user_access_logging_settings.test"
	kinesisStreamResourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserAccessLoggingSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserAccessLoggingSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserAccessLoggingSettingsExists(ctx, resourceName, &userAccessLoggingSettings),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_stream_arn", kinesisStreamResourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "user_access_logging_settings_arn", "workspaces-web", regexache.MustCompile(`userAccessLoggingSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_access_logging_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_access_logging_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserAccessLoggingSettings_updateKinesisStreamARN(t *testing.T) {
	ctx := acctest.Context(t)
	var userAccessLoggingSettings awstypes.UserAccessLoggingSettings
	resourceName := "aws_workspacesweb_user_access_logging_settings.test"
	kinesisStreamResourceName1 := "aws_kinesis_stream.test1"
	kinesisStreamResourceName2 := "aws_kinesis_stream.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserAccessLoggingSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserAccessLoggingSettingsConfig_kinesisStreamBefore(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserAccessLoggingSettingsExists(ctx, resourceName, &userAccessLoggingSettings),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_stream_arn", kinesisStreamResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_access_logging_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_access_logging_settings_arn",
			},
			{
				Config: testAccUserAccessLoggingSettingsConfig_kinesisStreamAfter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserAccessLoggingSettingsExists(ctx, resourceName, &userAccessLoggingSettings),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_stream_arn", kinesisStreamResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebUserAccessLoggingSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userAccessLoggingSettings awstypes.UserAccessLoggingSettings
	resourceName := "aws_workspacesweb_user_access_logging_settings.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserAccessLoggingSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserAccessLoggingSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserAccessLoggingSettingsExists(ctx, resourceName, &userAccessLoggingSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkspacesweb.ResourceUserAccessLoggingSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserAccessLoggingSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_user_access_logging_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindUserAccessLoggingSettingsByARN(ctx, conn, rs.Primary.Attributes["user_access_logging_settings_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web User Access Logging Settings %s still exists", rs.Primary.Attributes["user_access_logging_settings_arn"])
		}

		return nil
	}
}

func testAccCheckUserAccessLoggingSettingsExists(ctx context.Context, n string, v *awstypes.UserAccessLoggingSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindUserAccessLoggingSettingsByARN(ctx, conn, rs.Primary.Attributes["user_access_logging_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccUserAccessLoggingSettingsConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "amazon-workspaces-web-%[1]s"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "test" {
  kinesis_stream_arn = aws_kinesis_stream.test.arn
}
`, rName)
}

func testAccUserAccessLoggingSettingsConfig_kinesisStreamBefore(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test1" {
  name        = "amazon-workspaces-web-%[1]s-1"
  shard_count = 1
}

resource "aws_kinesis_stream" "test2" {
  name        = "amazon-workspaces-web-%[1]s-2"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "test" {
  kinesis_stream_arn = aws_kinesis_stream.test1.arn
}
`, rName)
}

func testAccUserAccessLoggingSettingsConfig_kinesisStreamAfter(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test1" {
  name        = "amazon-workspaces-web-%[1]s-1"
  shard_count = 1
}

resource "aws_kinesis_stream" "test2" {
  name        = "amazon-workspaces-web-%[1]s-2"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "test" {
  kinesis_stream_arn = aws_kinesis_stream.test2.arn
}
`, rName)
}
