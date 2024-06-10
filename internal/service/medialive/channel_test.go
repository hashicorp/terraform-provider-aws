// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaLiveChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_captionDescriptions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_caption_descriptions(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.caption_descriptions.*", map[string]string{
						"caption_selector_name": rName,
						names.AttrName:          "test-caption-name",
						"destination_settings.0.dvb_sub_destination_settings.0.font_resolution": "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_M2TS_settings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_m2tsSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.output_groups.0.outputs.0.output_settings.0.archive_output_settings.0.container_settings.0.m2ts_settings.*", map[string]string{
						"audio_buffer_model":        "ATSC",
						"buffer_model":              "MULTIPLEX",
						"rate_mode":                 "CBR",
						"audio_pids":                "200",
						"dvb_sub_pids":              "300",
						"arib_captions_pid":         "100",
						"arib_captions_pid_control": "AUTO",
						"video_pid":                 "101",
						"fragment_time":             "1.92",
						"program_num":               acctest.Ct1,
						"segmentation_time":         "1.92",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_UDP_outputSettings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_udpOutputSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.output_groups.0.outputs.0.output_settings.0.udp_output_settings.0.fec_output_settings.*", map[string]string{
						"include_fec":  "COLUMN_AND_ROW",
						"column_depth": "5",
						"row_length":   "5",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_MsSmooth_outputSettings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_msSmoothOutputSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.output_groups.0.outputs.0.output_settings.0.ms_smooth_output_settings.*", map[string]string{
						"name_modifier": rName,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_AudioDescriptions_codecSettings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_audioDescriptionCodecSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": "audio_1",
						names.AttrName:        "audio_1",
						"codec_settings.0.aac_settings.0.rate_control_mode": string(types.AacRateControlModeCbr),
						"codec_settings.0.aac_settings.0.bitrate":           "192000",
						"codec_settings.0.aac_settings.0.sample_rate":       "48000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name":                     "audio_2",
						names.AttrName:                            "audio_2",
						"codec_settings.0.ac3_settings.0.bitrate": "384000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
				),
			},
		},
	})
}

func TestAccMediaLiveChannel_VideoDescriptions_CodecSettings_h264Settings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_videoDescriptionCodecSettingsH264Settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName:     "test-video-name",
						"respond_to_afd":   "NONE",
						"scaling_behavior": "DEFAULT",
						"sharpness":        "100",
						"height":           "720",
						"width":            "1280",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.0.codec_settings.0.h264_settings.*", map[string]string{
						"adaptive_quantization":   "LOW",
						"afd_signaling":           "NONE",
						"bitrate":                 "5400000",
						"buf_fill_pct":            "90",
						"buf_size":                "10800000",
						"color_metadata":          "IGNORE",
						"entropy_encoding":        "CABAC",
						"filter_settings":         "",
						"fixed_afd":               "",
						"flicker_aq":              "ENABLED",
						"force_field_pictures":    "DISABLED",
						"framerate_control":       "SPECIFIED",
						"framerate_denominator":   acctest.Ct1,
						"framerate_numerator":     "50",
						"gop_b_reference":         "DISABLED",
						"gop_closed_cadence":      acctest.Ct1,
						"gop_num_b_frames":        acctest.Ct1,
						"gop_size":                "1.92",
						"gop_size_units":          "SECONDS",
						"level":                   "H264_LEVEL_AUTO",
						"look_ahead_rate_control": "HIGH",
						"max_bitrate":             acctest.Ct0,
						"min_i_interval":          acctest.Ct0,
						"num_ref_frames":          acctest.Ct3,
						"par_control":             "INITIALIZE_FROM_SOURCE",
						"par_denominator":         acctest.Ct0,
						"par_numerator":           acctest.Ct0,
						names.AttrProfile:         "HIGH",
						"quality_level":           "",
						"qvbr_quality_level":      acctest.Ct0,
						"rate_control_mode":       "CBR",
						"scan_type":               "PROGRESSIVE",
						"scene_change_detect":     "DISABLED",
						"slices":                  acctest.Ct1,
						"spatial_aq":              acctest.Ct0,
						"subgop_length":           "FIXED",
						"syntax":                  "DEFAULT",
						"temporal_aq":             "ENABLED",
						"timecode_insertion":      "PIC_TIMING_SEI",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_VideoDescriptions_CodecSettings_h265Settings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_videoDescriptionCodecSettingsH265Settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName:     "test-video-name",
						"respond_to_afd":   "NONE",
						"scaling_behavior": "DEFAULT",
						"sharpness":        "100",
						"height":           "720",
						"width":            "1280",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.0.codec_settings.0.h265_settings.*", map[string]string{
						"adaptive_quantization":   "LOW",
						"afd_signaling":           "FIXED",
						"bitrate":                 "5400000",
						"buf_size":                "20000000",
						"color_metadata":          "IGNORE",
						"fixed_afd":               "AFD_0000",
						"flicker_aq":              "ENABLED",
						"framerate_denominator":   acctest.Ct1,
						"framerate_numerator":     "50",
						"gop_closed_cadence":      acctest.Ct1,
						"gop_size":                "1.92",
						"gop_size_units":          "SECONDS",
						"level":                   "H265_LEVEL_AUTO",
						"look_ahead_rate_control": "HIGH",
						"min_i_interval":          "6",
						names.AttrProfile:         "MAIN_10BIT",
						"rate_control_mode":       "CBR",
						"scan_type":               "PROGRESSIVE",
						"scene_change_detect":     "ENABLED",
						"slices":                  acctest.Ct2,
						"tier":                    "HIGH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.0.codec_settings.0.h265_settings.0.color_space_settings.0.hdr10_settings.*", map[string]string{
						"max_cll":  "16",
						"max_fall": "16",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.0.codec_settings.0.h265_settings.0.filter_settings.0.temporal_filter_settings.*", map[string]string{
						"post_filter_sharpening": "AUTO",
						"strength":               "STRENGTH_1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.0.codec_settings.0.h265_settings.0.timecode_burnin_settings.*", map[string]string{
						"timecode_burnin_font_size": "SMALL_16",
						"timecode_burnin_position":  "BOTTOM_CENTER",
						names.AttrPrefix:            "terraform-test",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_channel"},
			},
		},
	})
}

func TestAccMediaLiveChannel_hls(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_hls(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						names.AttrName:        rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.output_groups.0.outputs.0.output_settings.0.hls_output_settings.0.h265_packaging_type", "HVC1"),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.output_groups.0.output_group_settings.0.hls_group_settings.0.client_cache", "ENABLED"),
				),
			},
		},
	})
}

func TestAccMediaLiveChannel_noAudio(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_noAudio(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
					resource.TestCheckNoResourceAttr(resourceName, "encoder_settings.0.audio_descriptions.*"),
				),
			},
		},
	})
}

func TestAccMediaLiveChannel_status(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_start(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					testAccCheckChannelStatus(ctx, resourceName, types.ChannelStateRunning),
				),
			},
			{
				Config: testAccChannelConfig_start(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					testAccCheckChannelStatus(ctx, resourceName, types.ChannelStateIdle),
				),
			},
		},
	})
}

func TestAccMediaLiveChannel_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_update(rName, rName, "AVC", "HD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: "destination1",
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": "test-audio-selector",
						names.AttrName:        "test-audio-description",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
				),
			},
			{
				Config: testAccChannelConfig_update(rName, rNameUpdated, "AVC", "HD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						names.AttrID: "destination1",
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": "test-audio-selector",
						names.AttrName:        "test-audio-description",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						names.AttrName: "test-video-name",
					}),
				),
			},
		},
	})
}

func TestAccMediaLiveChannel_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccChannelConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccChannelConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMediaLiveChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmedialive.ResourceChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_medialive_channel" {
				continue
			}

			_, err := tfmedialive.FindChannelByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameChannel, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckChannelExists(ctx context.Context, name string, channel *medialive.DescribeChannelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameChannel, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameChannel, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		resp, err := tfmedialive.FindChannelByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameChannel, rs.Primary.ID, err)
		}

		*channel = *resp

		return nil
	}
}

func testAccCheckChannelStatus(ctx context.Context, name string, state types.ChannelState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionChecking, tfmedialive.ResNameChannel, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionChecking, tfmedialive.ResNameChannel, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		resp, err := tfmedialive.FindChannelByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionChecking, tfmedialive.ResNameChannel, rs.Primary.ID, err)
		}

		if resp.State != state {
			return create.Error(names.MediaLive, create.ErrActionChecking, tfmedialive.ResNameChannel, rs.Primary.ID, fmt.Errorf("not (%s) got: %s", state, resp.State))
		}

		return nil
	}
}

func testAccChannelsPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

	input := &medialive.ListChannelsInput{}
	_, err := conn.ListChannels(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "medialive.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:*",
          "s3:*",
          "mediastore:*",
          "mediaconnect:*",
          "cloudwatch:*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
`, rName)
}

func testAccChannelConfig_baseS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  bucket = "%[1]s-1"
}

resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}
`, rName)
}

func testAccChannelConfig_baseMultiplex(rName string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_medialive_input" "test" {
  name                  = %[1]q
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"

  tags = {
    Name = %[1]q
  }
}

`, rName)
}

func testAccChannelConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_udpOutputSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "rtp://localhost:8000"
    }

    settings {
      url = "rtp://localhost:8001"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    video_descriptions {
      name = "test-video-name"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    output_groups {
      output_group_settings {
        udp_group_settings {
          input_loss_action = "DROP_TS"
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          udp_output_settings {
            destination {
              destination_ref_id = %[1]q
            }

            fec_output_settings {
              include_fec  = "COLUMN_AND_ROW"
              column_depth = 5
              row_length   = 5
            }

            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_msSmoothOutputSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "http://localhost:8000/path"
    }

    settings {
      url = "http://localhost:8001/path"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    video_descriptions {
      name = "test-video-name"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    output_groups {
      output_group_settings {
        ms_smooth_group_settings {
          audio_only_timecode_control = "USE_CONFIGURED_CLOCK"
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          ms_smooth_output_settings {
            name_modifier = %[1]q
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_m2tsSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
      codec_settings {
        aac_settings {
          rate_control_mode = "CBR"
        }
      }
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model        = "ATSC"
                buffer_model              = "MULTIPLEX"
                rate_mode                 = "CBR"
                audio_pids                = 200
                dvb_sub_pids              = 300
                arib_captions_pid         = 100
                arib_captions_pid_control = "AUTO"
                video_pid                 = 101
                fragment_time             = 1.92
                program_num               = 1
                segmentation_time         = 1.92
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_audioDescriptionCodecSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = "audio_1"
      name                = "audio_1"
      codec_settings {
        aac_settings {
          rate_control_mode = "CBR"
          bitrate           = 192000
          sample_rate       = 48000
        }
      }
    }

    audio_descriptions {
      audio_selector_name = "audio_2"
      name                = "audio_2"

      codec_settings {
        ac3_settings {
          bitrate = 384000
        }
      }
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = ["audio_1", "audio_2"]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_videoDescriptionCodecSettingsH264Settings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
      codec_settings {
        aac_settings {
          rate_control_mode = "CBR"
        }
      }
    }

    video_descriptions {
      name             = "test-video-name"
      respond_to_afd   = "NONE"
      sharpness        = 100
      scaling_behavior = "DEFAULT"
      width            = 1280
      height           = 720
      codec_settings {
        h264_settings {
          afd_signaling           = "NONE"
          color_metadata          = "IGNORE"
          adaptive_quantization   = "LOW"
          bitrate                 = "5400000"
          buf_size                = "10800000"
          buf_fill_pct            = 90
          entropy_encoding        = "CABAC"
          flicker_aq              = "ENABLED"
          force_field_pictures    = "DISABLED"
          framerate_control       = "SPECIFIED"
          framerate_numerator     = 50
          framerate_denominator   = 1
          gop_b_reference         = "DISABLED"
          gop_closed_cadence      = 1
          gop_num_b_frames        = 1
          gop_size                = 1.92
          gop_size_units          = "SECONDS"
          subgop_length           = "FIXED"
          scan_type               = "PROGRESSIVE"
          level                   = "H264_LEVEL_AUTO"
          look_ahead_rate_control = "HIGH"
          num_ref_frames          = 3
          par_control             = "INITIALIZE_FROM_SOURCE"
          profile                 = "HIGH"
          rate_control_mode       = "CBR"
          syntax                  = "DEFAULT"
          scene_change_detect     = "ENABLED"
          slices                  = 1
          spatial_aq              = "ENABLED"
          temporal_aq             = "ENABLED"
          timecode_insertion      = "PIC_TIMING_SEI"
        }
      }
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_videoDescriptionCodecSettingsH265Settings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
      codec_settings {
        aac_settings {
          rate_control_mode = "CBR"
        }
      }
    }

    video_descriptions {
      name             = "test-video-name"
      respond_to_afd   = "NONE"
      sharpness        = 100
      scaling_behavior = "DEFAULT"
      width            = 1280
      height           = 720
      codec_settings {
        h265_settings {
          bitrate  = "5400000"
          buf_size = "20000000"

          framerate_numerator   = 50
          framerate_denominator = 1

          color_metadata        = "IGNORE"
          adaptive_quantization = "LOW"

          flicker_aq = "ENABLED"

          afd_signaling = "FIXED"
          fixed_afd     = "AFD_0000"

          gop_closed_cadence = 1
          gop_size           = 1.92
          gop_size_units     = "SECONDS"
          min_i_interval     = 6
          scan_type          = "PROGRESSIVE"

          level                   = "H265_LEVEL_AUTO"
          look_ahead_rate_control = "HIGH"
          profile                 = "MAIN_10BIT"

          rate_control_mode   = "CBR"
          scene_change_detect = "ENABLED"

          slices = 2
          tier   = "HIGH"

          timecode_insertion = "DISABLED"

          color_space_settings {
            hdr10_settings {
              max_cll  = 16
              max_fall = 16
            }
          }

          filter_settings {
            temporal_filter_settings {
              post_filter_sharpening = "AUTO"
              strength               = "STRENGTH_1"
            }
          }

          timecode_burnin_settings {
            timecode_burnin_font_size = "SMALL_16"
            timecode_burnin_position  = "BOTTOM_CENTER"
            prefix                    = "terraform-test"
          }
        }
      }
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = %[1]q
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_hls(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        hls_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
          client_cache = "ENABLED"
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          hls_output_settings {
            name_modifier       = "_1"
            h265_packaging_type = "HVC1"
            hls_settings {
              standard_hls_settings {
                m3u8_settings {
                  audio_frames_per_pes = 4
                }
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_noAudio(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name            = "test-output-name"
        video_description_name = "test-video-name"
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName))
}

func testAccChannelConfig_caption_descriptions(rName string, fontResolution int) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id

    input_settings {
      caption_selector {
        name = %[1]q
      }

      audio_selector {
        name = "test-audio-selector"
      }
    }
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      name                = "test-audio-name"
      audio_selector_name = "test-audio-selector"
    }


    video_descriptions {
      name = "test-video-name"
    }

    caption_descriptions {
      name                  = "test-caption-name"
      caption_selector_name = aws_medialive_input.test.name

      destination_settings {
        dvb_sub_destination_settings {
          font_resolution = %[2]d
        }
      }
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name               = "test-output-name"
        video_description_name    = "test-video-name"
        audio_description_names   = ["test-audio-name"]
        caption_description_names = ["test-caption-name"]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName, fontResolution))
}

func testAccChannelConfig_start(rName string, start bool) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn
  start_channel = %[2]t

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName, start))
}

func testAccChannelConfig_update(rName, rNameUpdated, codec, inputResolution string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[2]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = %[3]q
    input_resolution = %[4]q
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = "destination1"

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = "test-audio-selector"
      name                = "test-audio-description"
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = "destination1"
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = ["test-audio-description"]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
`, rName, rNameUpdated, codec, inputResolution))
}

func testAccChannelConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccChannelConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccChannelConfig_base(rName),
		testAccChannelConfig_baseS3(rName),
		testAccChannelConfig_baseMultiplex(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input1"
    input_id              = aws_medialive_input.test.id
  }

  destinations {
    id = %[1]q

    settings {
      url = "s3://${aws_s3_bucket.test1.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.test2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = %[1]q
      name                = %[1]q
    }

    video_descriptions {
      name = "test-video-name"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = %[1]q
          }
        }
      }

      outputs {
        output_name             = "test-output-name"
        video_description_name  = "test-video-name"
        audio_description_names = [%[1]q]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
