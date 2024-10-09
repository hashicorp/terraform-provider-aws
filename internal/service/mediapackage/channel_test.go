// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmediapackage "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackage"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaPackageChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_package_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackage.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackage.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mediapackage", regexache.MustCompile(`channels/.+`)),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.0.password", regexache.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.0.url", regexache.MustCompile("^https://")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.0.username", regexache.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.1.password", regexache.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.1.url", regexache.MustCompile("^https://")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.1.username", regexache.MustCompile("^[0-9a-f]*$")),
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

func TestAccMediaPackageChannel_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_package_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackage.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackage.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_description(rName, "description1"),
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
				Config: testAccChannelConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccMediaPackageChannel_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_package_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackage.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackage.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_tags(rName, "Environment", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelConfig_tags(rName, "Environment", "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test1"),
				),
			},
			{
				Config: testAccChannelConfig_tags(rName, "Update", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Update", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMediaPackageChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_package_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(mediapackage.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(mediapackage.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediapackage.ResourceChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_package_channel" {
				continue
			}

			_, err := tfmediapackage.FindChannelByID(ctx, conn, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("MediaPackage Channel (%s) not deleted", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		input := &mediapackage.DescribeChannelInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(ctx, input)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

	input := &mediapackage.ListChannelsInput{}

	_, err := conn.ListChannels(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = %[1]q
}
`, rName)
}

func testAccChannelConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id  = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccChannelConfig_tags(rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = %[1]q

  tags = {
    Name = %[1]q

    %[2]s = %[3]q
  }
}
`, rName, key, value)
}
