package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						"id": rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						"name":                rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						"name": "test-video-name",
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

func TestAccMediaLiveChannel_audioDescriptions_codecSettings(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var channel medialive.DescribeChannelOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_audioDescriptionCodecSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						"id": rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						"name":                rName,
						"codec_settings.0.aac_settings.0.rate_control_mode": string(types.AacRateControlModeCbr),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						"name": "test-video-name",
					}),
				),
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_hls(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						"id": rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": rName,
						"name":                rName,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						"name": "test-video-name",
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.output_groups.0.outputs.0.output_settings.0.hls_output_settings.0.h265_packaging_type", "HVC1"),
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_update(rName, "AVC", "HD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						"id": "destination1",
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": "test-audio-selector",
						"name":                "test-audio-description",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						"name": "test-video-name",
					}),
				),
			},
			{
				Config: testAccChannelConfig_update(rNameUpdated, "AVC", "HD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "channel_id"),
					resource.TestCheckResourceAttr(resourceName, "channel_class", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.codec", "AVC"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.input_resolution", "HD"),
					resource.TestCheckResourceAttr(resourceName, "input_specification.0.maximum_bitrate", "MAX_20_MBPS"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_attachments.*", map[string]string{
						"input_attachment_name": "example-input1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "destinations.*", map[string]string{
						"id": "destination1",
					}),
					resource.TestCheckResourceAttr(resourceName, "encoder_settings.0.timecode_config.0.source", "EMBEDDED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.audio_descriptions.*", map[string]string{
						"audio_selector_name": "test-audio-selector",
						"name":                "test-audio-description",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encoder_settings.0.video_descriptions.*", map[string]string{
						"name": "test-video-name",
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccChannelConfig_tags2(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccChannelConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
			testAccChannelsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient()

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient()

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient()

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
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient()

	input := &medialive.ListChannelsInput{}
	_, err := conn.ListChannels(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelBaseConfig(rName string) string {
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

func testAccChannelBaseS3Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  bucket = "%[1]s-1"
}

resource "aws_s3_bucket_acl" "test1" {
  bucket = aws_s3_bucket.test1.id
  acl    = "private"
}

resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}

resource "aws_s3_bucket_acl" "test2" {
  bucket = aws_s3_bucket.test2.id
  acl    = "private"
}
`, rName)
}

func testAccChannelBaseMultiplexConfig(rName string) string {
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
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
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

func testAccChannelConfig_audioDescriptionCodecSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
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
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
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

func testAccChannelConfig_start(rName string, start bool) string {
	return acctest.ConfigCompose(
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
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

func testAccChannelConfig_update(rName, codec, inputResolution string) string {
	return acctest.ConfigCompose(
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_channel" "test" {
  name          = %[1]q
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.test.arn

  input_specification {
    codec            = %[2]q
    input_resolution = %[3]q
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
`, rName, codec, inputResolution))
}

func testAccChannelConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
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
		testAccChannelBaseConfig(rName),
		testAccChannelBaseS3Config(rName),
		testAccChannelBaseMultiplexConfig(rName),
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
