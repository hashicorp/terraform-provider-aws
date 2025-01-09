// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"

	tfmediapackagev2 "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackagev2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMediaPackageChannelGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_packagev2_channel_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackagev2.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackagev2.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackagev2", regexache.MustCompile(`channelGroup/.+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrDescription, regexache.MustCompile("Managed by Terraform")),
					resource.TestMatchResourceAttr(resourceName, "egress_domain", regexache.MustCompile(fmt.Sprintf("^[0-9a-z]+.egress.[0-9a-z]+.mediapackagev2.%s.amazonaws.com$", acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "tags.Foo", regexache.MustCompile("Bar")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMediaPackageChannelGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_packagev2_channel_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackagev2.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackagev2.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelGroupConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccMediaPackageChannelGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_packagev2_channel_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackagev2.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackagev2.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfmediapackagev2.ResourceChannelGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMediaPackageChannelGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_packagev2_channel_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackagev2.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackagev2.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelGroupConfig_tags(rName, "Environment", "nonprod"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "nonprod"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelGroupConfig_tags(rName, "Environment", "dev"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "dev"),
				),
			},
			{
				Config: testAccChannelGroupConfig_tags(rName, "Stage", "Prod"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Stage", "Prod"),
				),
			},
		},
	})
}

func testAccCheckChannelGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_packagev2_channel_group" {
				continue
			}

			_, err := tfmediapackagev2.FindChannelGroupByID(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if err == nil {
				return fmt.Errorf("MediaPackageV2 Channel Group: %s not deleted", rs.Primary.ID)
			}

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageV2Client(ctx)

		input := &mediapackagev2.GetChannelGroupInput{
			ChannelGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetChannelGroup(ctx, input)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageV2Client(ctx)

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
  channel_group_name = %[1]q
	tags = {
	  Foo = "Bar"
	}
}
`, rName)
}

func testAccChannelGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "test" {
  channel_group_name  = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccChannelGroupConfig_tags(rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_media_packagev2_channel_group" "test" {
  channel_group_name = %[1]q

  tags = {
    Name = %[1]q

    %[2]s = %[3]q
  }
}
`, rName, key, value)
}
