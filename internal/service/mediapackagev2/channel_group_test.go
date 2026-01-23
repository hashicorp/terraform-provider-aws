// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmediapackagev2 "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackagev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMediaPackageV2ChannelGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var channelGroup mediapackagev2.GetChannelGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_media_packagev2_channel_group.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelGroupExists(ctx, t, resourceName, &channelGroup),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/.+`)),
					resource.TestMatchResourceAttr(resourceName, "egress_domain", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+.egress.[0-9a-z]+.mediapackagev2.%s.amazonaws.com$", acctest.Region()))),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccMediaPackageV2ChannelGroup_description(t *testing.T) {
	ctx := acctest.Context(t)

	var channelGroup mediapackagev2.GetChannelGroupOutput
	resourceName := "aws_media_packagev2_channel_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelGroupExists(ctx, t, resourceName, &channelGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccChannelGroupConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelGroupExists(ctx, t, resourceName, &channelGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccMediaPackageV2ChannelGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var channelGroup mediapackagev2.GetChannelGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_media_packagev2_channel_group.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelGroupExists(ctx, t, resourceName, &channelGroup),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmediapackagev2.ResourceChannelGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccChannelGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrImportStateIdFunc(resourceName, names.AttrName)
}

func testAccCheckChannelGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MediaPackageV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_packagev2_channel_group" {
				continue
			}

			_, err := tfmediapackagev2.FindChannelGroupByID(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if err == nil {
				return fmt.Errorf("MediaPackageV2 Channel Group: %s not deleted", rs.Primary.ID)
			}

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaPackageV2, create.ErrActionCheckingDestroyed, tfmediapackagev2.ResNameChannelGroup, rs.Primary.Attributes[names.AttrName], err)
			}
		}

		return nil
	}
}

func testAccCheckChannelGroupExists(ctx context.Context, t *testing.T, name string, channelGroup *mediapackagev2.GetChannelGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).MediaPackageV2Client(ctx)
		resp, err := tfmediapackagev2.FindChannelGroupByID(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return create.Error(names.MediaPackageV2, create.ErrActionCheckingExistence, tfmediapackagev2.ResNameChannelGroup, rs.Primary.Attributes[names.AttrName], err)
		}

		*channelGroup = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MediaPackageV2Client(ctx)

	input := &mediapackagev2.ListChannelGroupsInput{}

	_, err := conn.ListChannelGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccChannelGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}
