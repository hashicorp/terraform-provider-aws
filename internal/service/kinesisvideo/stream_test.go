// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisvideo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesisvideo "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisvideo"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisVideoStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	rInt1 := sdkacctest.RandInt()
	rInt2 := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kinesisvideo.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("terraform-kinesis-video-stream-test-%d", rInt1)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisvideo", regexache.MustCompile(fmt.Sprintf("stream/terraform-kinesis-video-stream-test-%d/.+", rInt1))),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("terraform-kinesis-video-stream-test-%d", rInt2)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisvideo", regexache.MustCompile(fmt.Sprintf("stream/terraform-kinesis-video-stream-test-%d/.+", rInt2))),
				),
			},
		},
	})
}

func TestAccKinesisVideoStream_options(t *testing.T) {
	ctx := acctest.Context(t)
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	kmsResourceName := "aws_kms_key.default"
	rInt := sdkacctest.RandInt()
	rName1 := sdkacctest.RandString(8)
	rName2 := sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kinesisvideo.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_options(rInt, rName1, "video/h264"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisvideo", regexache.MustCompile(fmt.Sprintf("stream/terraform-kinesis-video-stream-test-%d/.+", rInt))),
					resource.TestCheckResourceAttr(resourceName, "data_retention_in_hours", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "media_type", "video/h264"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, fmt.Sprintf("kinesis-video-device-name-%s", rName1)),
					resource.TestCheckResourceAttrPair(
						resourceName, names.AttrKMSKeyID,
						kmsResourceName, names.AttrID),
				),
			},
			{
				Config: testAccStreamConfig_options(rInt, rName2, "video/h120"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "media_type", "video/h120"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeviceName, fmt.Sprintf("kinesis-video-device-name-%s", rName2)),
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
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kinesisvideo.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_tags1(rInt, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccStreamConfig_tags2(rInt, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
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
				Config: testAccStreamConfig_tags1(rInt, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKinesisVideoStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kinesisvideo.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisVideoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &stream),
					testAccCheckStreamDisappears(ctx, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStreamDisappears(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisVideoConn(ctx)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		input := &kinesisvideo.DeleteStreamInput{
			StreamARN:      aws.String(rs.Primary.ID),
			CurrentVersion: aws.String(rs.Primary.Attributes[names.AttrVersion]),
		}

		if _, err := conn.DeleteStreamWithContext(ctx, input); err != nil {
			return err
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{kinesisvideo.StatusDeleting},
			Target:     []string{"DELETED"},
			Refresh:    tfkinesisvideo.StreamStateRefresh(ctx, conn, rs.Primary.ID),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckStreamExists(ctx context.Context, n string, stream *kinesisvideo.StreamInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisVideoConn(ctx)
		describeOpts := &kinesisvideo.DescribeStreamInput{
			StreamARN: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeStreamWithContext(ctx, describeOpts)
		if err != nil {
			return err
		}

		*stream = *resp.StreamInfo

		return nil
	}
}

func testAccCheckStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_video_stream" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisVideoConn(ctx)
			describeOpts := &kinesisvideo.DescribeStreamInput{
				StreamARN: aws.String(rs.Primary.ID),
			}
			resp, err := conn.DescribeStreamWithContext(ctx, describeOpts)
			if err == nil {
				if resp.StreamInfo != nil && aws.StringValue(resp.StreamInfo.Status) != "DELETING" {
					return fmt.Errorf("Error Kinesis Video Stream still exists")
				}
			}

			return nil
		}

		return nil
	}
}

func testAccStreamConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%d"
}
`, rInt)
}

func testAccStreamConfig_options(rInt int, rName, mediaType string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "default" {
  description             = "KMS key 1"
  deletion_window_in_days = 7
}

resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%[1]d"

  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name-%[2]s"
  kms_key_id              = aws_kms_key.default.id
  media_type              = "%[3]s"
}
`, rInt, rName, mediaType)
}

func testAccStreamConfig_tags1(rInt int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%d"

  tags = {
    %[2]q = %[3]q
  }
}
`, rInt, tagKey1, tagValue1)
}

func testAccStreamConfig_tags2(rInt int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%d"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rInt, tagKey1, tagValue1, tagKey2, tagValue2)
}
