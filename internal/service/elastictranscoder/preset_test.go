// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elastictranscoder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elastictranscoder/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfet "github.com/hashicorp/terraform-provider-aws/internal/service/elastictranscoder"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticTranscoderPreset_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elastictranscoder", regexache.MustCompile(`preset/.+`)),
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

func TestAccElasticTranscoderPreset_video_noCodec(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_videoNoCodec(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
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

// https://github.com/terraform-providers/terraform-provider-aws/issues/14090
func TestAccElasticTranscoderPreset_audio_noBitRate(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_noBitRate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
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

func TestAccElasticTranscoderPreset_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfet.ResourcePreset(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14087
func TestAccElasticTranscoderPreset_AudioCodecOptions_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_audioCodecOptionsEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"audio_codec_options"}, // Due to incorrect schema (should be nested under audio)
			},
		},
	})
}

func TestAccElasticTranscoderPreset_description(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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

// Tests all configuration blocks
func TestAccElasticTranscoderPreset_full(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_full1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "audio.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "audio_codec_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "thumbnails.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "video.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "video_codec_options.%", "5"),
					resource.TestCheckResourceAttr(resourceName, "video_watermarks.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPresetConfig_full2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "audio.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "audio_codec_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "thumbnails.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "video.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "video_codec_options.%", "5"),
					resource.TestCheckResourceAttr(resourceName, "video_watermarks.#", acctest.Ct1),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/695
func TestAccElasticTranscoderPreset_Video_frameRate(t *testing.T) {
	ctx := acctest.Context(t)
	var preset awstypes.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticTranscoderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPresetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig_videoFrameRate(rName, "29.97"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPresetExists(ctx, resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "video.0.frame_rate", "29.97"),
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

func testAccCheckPresetExists(ctx context.Context, name string, preset *awstypes.Preset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderClient(ctx)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Preset ID is set")
		}

		out, err := conn.ReadPreset(ctx, &elastictranscoder.ReadPresetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*preset = *out.Preset

		return nil
	}
}

func testAccCheckPresetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastictranscoder_preset" {
				continue
			}

			out, err := conn.ReadPreset(ctx, &elastictranscoder.ReadPresetInput{
				Id: aws.String(rs.Primary.ID),
			})

			if err == nil {
				if out.Preset != nil && aws.ToString(out.Preset.Id) == rs.Primary.ID {
					return fmt.Errorf("Elastic Transcoder Preset still exists")
				}
			}

			if !errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return fmt.Errorf("unexpected error: %s", err)
			}
		}
		return nil
	}
}

func testAccPresetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }
}
`, rName)
}

func testAccPresetConfig_audioCodecOptionsEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }

  audio_codec_options {}
}
`, rName)
}

func testAccPresetConfig_description(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container   = "mp4"
  description = %[2]q
  name        = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }
}
`, rName, description)
}

func testAccPresetConfig_full1(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 128
    channels           = 2
    codec              = "AAC"
    sample_rate        = 48000
  }

  audio_codec_options {
    profile = "auto"
  }

  video {
    bit_rate             = "auto"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "true"
    frame_rate           = "auto"
    keyframes_max_dist   = 90
    max_height           = 1080
    max_width            = 1920
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "4.1"
    MaxReferenceFrames       = 4
    InterlacedMode           = "Auto"
    ColorSpaceConversionMode = "None"
  }

  thumbnails {
    format         = "jpg"
    interval       = 5
    max_width      = 960
    max_height     = 540
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }
}
`, rName)
}

func testAccPresetConfig_full2(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 96
    channels           = 2
    codec              = "AAC"
    sample_rate        = 44100
  }

  audio_codec_options {
    profile = "AAC-LC"
  }

  video {
    bit_rate             = "1600"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "false"
    frame_rate           = "auto"
    max_frame_rate       = "60"
    keyframes_max_dist   = 240
    max_height           = "auto"
    max_width            = "auto"
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "2.2"
    MaxReferenceFrames       = 3
    InterlacedMode           = "Progressive"
    ColorSpaceConversionMode = "None"
  }

  video_watermarks {
    id                = "Terraform Test"
    max_width         = "20%%"
    max_height        = "20%%"
    sizing_policy     = "ShrinkToFit"
    horizontal_align  = "Right"
    horizontal_offset = "10px"
    vertical_align    = "Bottom"
    vertical_offset   = "10px"
    opacity           = "55.5"
    target            = "Content"
  }

  thumbnails {
    format         = "png"
    interval       = 120
    max_width      = "auto"
    max_height     = "auto"
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }
}
`, rName)
}

func testAccPresetConfig_videoFrameRate(rName string, frameRate string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  thumbnails {
    format         = "png"
    interval       = 120
    max_width      = "auto"
    max_height     = "auto"
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }

  video {
    bit_rate             = "auto"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "true"
    frame_rate           = %[2]q
    keyframes_max_dist   = 90
    max_height           = 1080
    max_width            = 1920
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "4.1"
    MaxReferenceFrames       = 4
    InterlacedMode           = "Auto"
    ColorSpaceConversionMode = "None"
  }
}
`, rName, frameRate)
}

func testAccPresetConfig_videoNoCodec(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "webm"
  type      = "Custom"
  name      = %[1]q

  audio {
    codec              = "vorbis"
    sample_rate        = 44100
    bit_rate           = 128
    channels           = 2
    audio_packing_mode = "SingleTrack"
  }

  thumbnails {
    format         = "png"
    interval       = 300
    max_width      = 640
    max_height     = 360
    sizing_policy  = "ShrinkToFit"
    padding_policy = "NoPad"
  }

  video {
    codec                = "vp9"
    keyframes_max_dist   = 90
    fixed_gop            = false
    bit_rate             = 600
    frame_rate           = 30
    max_width            = 640
    max_height           = 360
    display_aspect_ratio = "auto"
    sizing_policy        = "Fit"
    padding_policy       = "NoPad"
  }
}
`, rName)
}

func testAccPresetConfig_noBitRate(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "wav"
  name      = %[1]q
  audio {
    audio_packing_mode = "SingleTrack"
    channels           = 2
    codec              = "pcm"
    sample_rate        = 44100
  }
}
`, rName)
}
