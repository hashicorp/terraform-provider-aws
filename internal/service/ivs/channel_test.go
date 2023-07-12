// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfivs "github.com/hashicorp/terraform-provider-aws/internal/service/ivs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVSChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel ivs.Channel

	resourceName := "aws_ivs_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVS)
			testAccChannelPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "ingest_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "playback_url"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ivs", regexp.MustCompile(`channel/.+`)),
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

func TestAccIVSChannel_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var channel ivs.Channel

	resourceName := "aws_ivs_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVS)
			testAccChannelPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccChannelConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIVSChannel_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ivs.Channel

	resourceName := "aws_ivs_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authorized := "true"
	latencyMode := "NORMAL"
	channelType := "BASIC"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVS)
			testAccChannelPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &v1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelConfig_update(rName, authorized, latencyMode, channelType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &v2),
					testAccCheckChannelNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "authorized", authorized),
					resource.TestCheckResourceAttr(resourceName, "latency_mode", latencyMode),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", channelType),
				),
			},
		},
	})
}

func TestAccIVSChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var channel ivs.Channel

	resourceName := "aws_ivs_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccChannelPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfivs.ResourceChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIVSChannel_recordingConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var channel ivs.Channel
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_channel.test"
	recordingConfigurationResourceName := "aws_ivs_recording_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccChannelPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_recordingConfiguration(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, "recording_configuration_arn", recordingConfigurationResourceName, "id"),
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

func testAccCheckChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ivs_channel" {
				continue
			}

			input := &ivs.GetChannelInput{
				Arn: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetChannelWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
					return nil
				}

				return err
			}

			return create.Error(names.IVS, create.ErrActionCheckingDestroyed, tfivs.ResNameChannel, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, name string, channel *ivs.Channel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameChannel, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameChannel, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

		output, err := tfivs.FindChannelByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNameChannel, rs.Primary.ID, err)
		}

		*channel = *output

		return nil
	}
}

func testAccChannelPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

	input := &ivs.ListChannelsInput{}
	_, err := conn.ListChannelsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckChannelNotRecreated(before, after *ivs.Channel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before != after {
			return create.Error(names.IVS, create.ErrActionCheckingNotRecreated, tfivs.ResNameChannel, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccChannelConfig_basic() string {
	return `
resource "aws_ivs_channel" "test" {
}
`
}

func testAccChannelConfig_update(rName, authorized, latencyMode, channelType string) string {
	return fmt.Sprintf(`
resource "aws_ivs_channel" "test" {
  name         = %[1]q
  authorized   = %[2]s
  latency_mode = %[3]q
  type         = %[4]q
}
`, rName, authorized, latencyMode, channelType)
}

func testAccChannelConfig_recordingConfiguration(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_ivs_recording_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}

resource "aws_ivs_channel" "test" {
  recording_configuration_arn = aws_ivs_recording_configuration.test.id
}
`, bucketName)
}

func testAccChannelConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ivs_channel" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccChannelConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ivs_channel" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
