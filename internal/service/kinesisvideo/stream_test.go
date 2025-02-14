// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisvideo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisvideo/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesisvideo "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisvideo"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisVideoStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stream awstypes.StreamInfo
	resourceName := "aws_kinesis_video_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KinesisVideoEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kinesisvideo", regexache.MustCompile(`stream/`+rName+`/\d+$`)), // TODO: Last component is Unix timestamp of `creation_time`
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, "data_retention_in_hours", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, ""),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrKMSKeyID, "kms", "alias/aws/kinesisvideo"),
					resource.TestCheckResourceAttr(resourceName, "media_type", ""),
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

func TestAccKinesisVideoStream_options(t *testing.T) {
	ctx := acctest.Context(t)
	var stream awstypes.StreamInfo
	resourceName := "aws_kinesis_video_stream.test"
	kmsResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	deviceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	deviceNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KinesisVideoEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_options(rName, deviceName, "video/h264"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kinesisvideo", regexache.MustCompile(`stream/`+rName+`/\d+$`)), // TODO: Last component is Unix timestamp of `creation_time`
					resource.TestCheckResourceAttr(resourceName, "data_retention_in_hours", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_type", "video/h264"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, deviceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsResourceName, names.AttrID),
				),
			},
			{
				Config: testAccStreamConfig_options(rName, deviceNameUpdated, "video/h120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "media_type", "video/h120"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, deviceNameUpdated),
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

func TestAccKinesisVideoStream_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var stream awstypes.StreamInfo
	resourceName := "aws_kinesis_video_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KinesisVideoEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccStreamConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKinesisVideoStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stream awstypes.StreamInfo
	resourceName := "aws_kinesis_video_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KinesisVideoEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesisvideo.ResourceStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStreamExists(ctx context.Context, n string, v *awstypes.StreamInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisVideoClient(ctx)

		output, err := tfkinesisvideo.FindStreamByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_video_stream" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisVideoClient(ctx)

			_, err := tfkinesisvideo.FindStreamByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Video Stream %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStreamConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "test" {
  name = %[1]q
}
`, rName)
}

func testAccStreamConfig_options(rName, deviceName, mediaType string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "test" {
  name = %[1]q

  data_retention_in_hours = 1
  device_name             = %[2]q
  kms_key_id              = aws_kms_key.test.id
  media_type              = %[3]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName, deviceName, mediaType)
}

func testAccStreamConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccStreamConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
