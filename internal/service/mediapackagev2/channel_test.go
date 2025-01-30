// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmediapackagev2 "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackagev2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMediaPackageV2Channel_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var channel mediapackagev2.GetChannelOutput

	channelGroupRName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	channelRName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_media_packagev2_channel.channel"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(channelGroupRName, channelRName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/[a-zA-Z0-9_-]+/channel/[a-zA-Z0-9_-]+$`)),
					//resource.TestMatchResourceAttr(resourceName, "egress_domain", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+.egress.[0-9a-z]+.mediapackagev2.%s.amazonaws.com$", acctest.Region()))),
				),
			},
			//{
			//	ResourceName:                         resourceName,
			//	ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
			//	ImportStateVerifyIdentifierAttribute: names.AttrName,
			//	ImportState:                          true,
			//	ImportStateVerify:                    true,
			//},
		},
	})
}

func testAccChannelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrImportStateIdFunc(resourceName, names.AttrName)
}

func testAccCheckChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_packagev2_channel" {
				continue
			}

			_, err := tfmediapackagev2.FindChannelByID(ctx, conn, rs.Primary.Attributes["channel_group_name"], rs.Primary.Attributes[names.AttrARN])
			if err == nil {
				return fmt.Errorf("MediaPackageV2 Channel Group: %s not deleted", rs.Primary.ID)
			}

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaPackageV2, create.ErrActionCheckingDestroyed, tfmediapackagev2.ResNameChannelGroup, rs.Primary.Attributes[names.AttrName], err)
			}
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, channelName string, channel *mediapackagev2.GetChannelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[channelName]
		if !ok {
			return fmt.Errorf("Not found: %s", channelName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageV2Client(ctx)
		resp, err := tfmediapackagev2.FindChannelByID(ctx, conn, rs.Primary.Attributes["channel_group_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return create.Error(names.MediaPackageV2, create.ErrActionCheckingExistence, tfmediapackagev2.ResNameChannelGroup, rs.Primary.Attributes[names.AttrName], err)
		}

		*channel = *resp

		return nil
	}
}

func testAccChannelConfig_basic(channelGroupName string, channelName string) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "channelGroup" {
  name = %[1]q
}

resource "aws_media_packagev2_channel" "channel" {
  channel_group_name = aws_media_packagev2_channel_group.channelGroup.name
  name        = %[2]q
  input_type = "HLS"
}
`, channelGroupName, channelName)
}
