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
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.id", regexache.MustCompile("1")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.id", regexache.MustCompile("2")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "input_type", regexache.MustCompile("HLS")),
					resource.TestMatchResourceAttr(resourceName, "input_switch_configuration.mqcs_input_switching", regexache.MustCompile(acctest.CtFalse)),
					resource.TestMatchResourceAttr(resourceName, "output_header_configuration.publish_mqcs", regexache.MustCompile(acctest.CtFalse)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccChannelImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccMediaPackageV2Channel_cmaf(t *testing.T) {
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
				Config: testAccChannelConfig_cmaf(channelGroupRName, channelRName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/[a-zA-Z0-9_-]+/channel/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.id", regexache.MustCompile("1")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.id", regexache.MustCompile("2")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "input_type", regexache.MustCompile("CMAF")),
					resource.TestMatchResourceAttr(resourceName, "input_switch_configuration.mqcs_input_switching", regexache.MustCompile(acctest.CtTrue)),
					resource.TestMatchResourceAttr(resourceName, "output_header_configuration.publish_mqcs", regexache.MustCompile(acctest.CtTrue)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccChannelImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccChannelConfig_cmaf(channelGroupRName, channelRName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/[a-zA-Z0-9_-]+/channel/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.id", regexache.MustCompile("1")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.id", regexache.MustCompile("2")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "input_type", regexache.MustCompile("CMAF")),
					resource.TestMatchResourceAttr(resourceName, "input_switch_configuration.mqcs_input_switching", regexache.MustCompile(acctest.CtFalse)),
					resource.TestMatchResourceAttr(resourceName, "output_header_configuration.publish_mqcs", regexache.MustCompile(acctest.CtFalse)),
				),
			},
		},
	})
}

func testAccMediaPackageV2Channel_description(t *testing.T) {
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
				Config: testAccChannelConfig_description(channelGroupRName, channelRName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/[a-zA-Z0-9_-]+/channel/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.id", regexache.MustCompile("1")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.id", regexache.MustCompile("2")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "input_type", regexache.MustCompile("HLS")),
					resource.TestMatchResourceAttr(resourceName, "input_switch_configuration.mqcs_input_switching", regexache.MustCompile(acctest.CtFalse)),
					resource.TestMatchResourceAttr(resourceName, "output_header_configuration.publish_mqcs", regexache.MustCompile(acctest.CtFalse)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccChannelImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccChannelConfig_description(channelGroupRName, channelRName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/[a-zA-Z0-9_-]+/channel/[a-zA-Z0-9_-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.id", regexache.MustCompile("1")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.0.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.id", regexache.MustCompile("2")),
					resource.TestMatchResourceAttr(resourceName, "ingest_endpoints.1.url", regexache.MustCompile(fmt.Sprintf("^(https://[0-9a-z-]+.ingest.[0-9a-z-]+.mediapackagev2.%s.amazonaws.com(/.*)?)$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "input_type", regexache.MustCompile("HLS")),
					resource.TestMatchResourceAttr(resourceName, "input_switch_configuration.mqcs_input_switching", regexache.MustCompile(acctest.CtFalse)),
					resource.TestMatchResourceAttr(resourceName, "output_header_configuration.publish_mqcs", regexache.MustCompile(acctest.CtFalse)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccMediaPackageV2Channel_disappears(t *testing.T) {
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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfmediapackagev2.ResourceChannel, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccChannelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["channel_group_name"], rs.Primary.Attributes[acctest.CtName]), nil
	}
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
  name               = %[2]q
}
`, channelGroupName, channelName)
}

func testAccChannelConfig_cmaf(channelGroupName string, channelName string, mqcsInputSwitching bool, publishMqcs bool) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "channelGroup" {
  name = %[1]q
}

resource "aws_media_packagev2_channel" "channel" {
  channel_group_name = aws_media_packagev2_channel_group.channelGroup.name
  name               = %[2]q
  input_type         = "CMAF"

  input_switch_configuration = {
    mqcs_input_switching = %[3]t
  }

  output_header_configuration = {
    publish_mqcs = %[4]t
  }
}
`, channelGroupName, channelName, mqcsInputSwitching, publishMqcs)
}

func testAccChannelConfig_description(channelGroupName string, channelName string, description string) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "channelGroup" {
  name = %[1]q
}

resource "aws_media_packagev2_channel" "channel" {
  channel_group_name = aws_media_packagev2_channel_group.channelGroup.name
  name               = %[2]q
  description        = %[3]q
}
`, channelGroupName, channelName, description)
}
