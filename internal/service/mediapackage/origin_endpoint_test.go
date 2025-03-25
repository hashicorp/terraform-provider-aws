// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmp "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackage"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfmediapackage "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackage"
)

func TestAccMediaPackageOriginEndpoint_hlsPackage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_package_origin_endpoint.test"
	channelID := fmt.Sprintf("test-%s", time.Now().Format("20060102150405"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx, channelID),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_hlsPackage(channelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
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

func TestAccMediaPackageOriginEndpoint_mssPackage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_package_origin_endpoint.test"
	channelID := fmt.Sprintf("test-%s", time.Now().Format("20060102150405"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx, channelID),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_mssPackage(channelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
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

func TestAccMediaPackageOriginEndpoint_dashPackage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_package_origin_endpoint.test"
	channelID := fmt.Sprintf("test-%s", time.Now().Format("20060102150405"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx, channelID),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_dashPackage(channelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
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

func TestAccMediaPackageOriginEndpoint_cmafPackage(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_package_origin_endpoint.test"
	channelID := fmt.Sprintf("test-%s", time.Now().Format("20060102150405"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx, channelID),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_cmafPackage(channelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
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

func TestAccMediaPackageOriginEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	channelID := fmt.Sprintf("test-%s", time.Now().Format("20060102150405"))
	resourceName := "aws_media_package_origin_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx, channelID),
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_hlsPackage(channelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmp.ResourceOriginEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOriginEndpointConfig_baseResource(channelID string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = "%[1]s"	
}

resource "aws_iam_role" "encryption_test" {
  name = "encryption-role-%[1]s"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "mediapackage.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "encryption_test" {
  name        = "policy-%[1]s"
  role        = aws_iam_role.encryption_test.name
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      },
    ],
  })
}

resource "aws_iam_role" "secrets_manager_test" {
  name = "secret-role-%[1]s"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "mediapackage.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name = "secret-%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({
    MediaPackageCDNIdentifier = "test1234"
  })
}

resource "aws_iam_role_policy" "secrets_manager_test" {
  name   = "test"
  role   = aws_iam_role.secrets_manager_test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      }
    ]
  })
}

`, channelID)
}

func testAccOriginEndpointConfig_hlsPackage(channelID string) string {
	return fmt.Sprintf(`%s
resource "aws_media_package_origin_endpoint" "test" {
  origin_endpoint_id = "endpoint-%s"
  channel_id = aws_media_package_channel.test.id
  hls_package {
	ad_markers = "PASSTHROUGH"
    ad_triggers = ["SPLICE_INSERT"]
    ads_on_delivery_restrictions = "NONE"
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = aws_iam_role.encryption_test.arn
        system_ids  = ["edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"]
        url         = "https://example.com/hoge"
      }
      constant_initialization_vector = "0123456789abcdef0123456789abcdef"
      encryption_method = "AES_128"
      key_rotation_interval_seconds = 600
      repeat_ext_x_key = true
    }
    include_dvb_subtitles = true
	include_iframe_only_stream = true
	playlist_type = "EVENT"
	playlist_window_seconds = 60
	program_date_time_interval_seconds = 600
	segment_duration_seconds = 6
	stream_selection {
	  max_video_bits_per_second = 1000000
	  min_video_bits_per_second = 500000	
	  stream_order = "ORIGINAL"
	}
  }
  authorization {
    cdn_identifier_secret = aws_secretsmanager_secret.test.arn
    secrets_role_arn      = aws_iam_role.secrets_manager_test.arn
  }
  description = "description"
  manifest_name = "example-manifest-name"
  origination = "ALLOW"
  start_over_window_seconds = 3600
  time_delay_seconds = 60
  whitelist = ["192.168.0.1/24"]
}
`, testAccOriginEndpointConfig_baseResource(channelID), channelID)
}

func testAccOriginEndpointConfig_cmafPackage(channelID string) string {
	return fmt.Sprintf(`%s
resource "aws_media_package_origin_endpoint" "test" {
  origin_endpoint_id = "endpoint-%s"
  channel_id = aws_media_package_channel.test.id
  cmaf_package {
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = "arn:aws:iam::730335654035:role/omote-encryption-role"
        system_ids  = ["edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"]
        url         = "https://example.com/hoge"
		encryption_contract_configuration {
		  preset_speke20_audio = "PRESET-AUDIO-1"
		  preset_speke20_video = "PRESET-VIDEO-1"
		}
      }
	  constant_initialization_vector = "0123456789abcdef0123456789abcdef"
      encryption_method = "SAMPLE_AES"
	  key_rotation_interval_seconds = 0
    }
	hls_manifests {
	  id = "test"
	  ad_markers = "NONE"
	  ad_triggers = ["SPLICE_INSERT"]
	  ads_on_delivery_restrictions = "NONE"
	  include_iframe_only_stream = true
	  manifest_name = "test"
	  playlist_type = "NONE"
	  playlist_window_seconds = 60
	  program_date_time_interval_seconds = 600
	}
	segment_duration_seconds = 6
	segment_prefix = "test"
	stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order = "ORIGINAL"
	}
  }
  authorization {
    cdn_identifier_secret = aws_secretsmanager_secret.test.arn
    secrets_role_arn      = aws_iam_role.secrets_manager_test.arn
  }
  description = "description"
  manifest_name = "example-manifest-name"
  origination = "ALLOW"
  start_over_window_seconds = 3600
  time_delay_seconds = 60
  whitelist = ["192.168.0.1/24"]
}
`, testAccOriginEndpointConfig_baseResource(channelID), channelID)
}

func testAccOriginEndpointConfig_dashPackage(channelID string) string {
	return fmt.Sprintf(`%s
resource "aws_media_package_origin_endpoint" "test" {
  origin_endpoint_id = "endpoint-%s"
  channel_id = aws_media_package_channel.test.id
  dash_package {
    ad_triggers = ["SPLICE_INSERT"]
	ads_on_delivery_restrictions = "NONE"
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = aws_iam_role.encryption_test.arn
        system_ids  = ["edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"]
        url         = "https://example.com/hoge"
		encryption_contract_configuration {
		  preset_speke20_audio = "PRESET-AUDIO-1"
		  preset_speke20_video = "PRESET-VIDEO-1"
		}
      }
	  key_rotation_interval_seconds = 0
    }
    include_iframe_only_stream = true
    manifest_layout            = "FULL"
	manifest_window_seconds    = 60
	min_buffer_time_seconds    = 60
	min_update_period_seconds  = 60
    period_triggers = ["ADS"]
	profile = "NONE"
    segment_duration_seconds = 6
    stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order              = "ORIGINAL"
    }
    suggested_presentation_delay_seconds = 60
    utc_timing = "HTTP-HEAD"
    utc_timing_uri = "https://example.com/hoge"
  }
  authorization {
    cdn_identifier_secret = aws_secretsmanager_secret.test.arn
    secrets_role_arn      = aws_iam_role.secrets_manager_test.arn
  }
  description = "description"
  manifest_name = "example-manifest-name"
  origination = "ALLOW"
  start_over_window_seconds = 3600
  time_delay_seconds = 60
  whitelist = ["192.168.0.1/24"]
}
`, testAccOriginEndpointConfig_baseResource(channelID), channelID)
}

func testAccOriginEndpointConfig_mssPackage(channelID string) string {
	return fmt.Sprintf(`%s
resource "aws_media_package_origin_endpoint" "test" {
  origin_endpoint_id = "endpoint-%s"
  channel_id = aws_media_package_channel.test.id
  mss_package {
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = aws_iam_role.encryption_test.arn
        system_ids  = ["edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"]
        url         = "https://example.com/hoge"
      }
    }
	manifest_window_seconds  = 60
    segment_duration_seconds = 6
    stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order              = "ORIGINAL"
    }
  }
  authorization {
    cdn_identifier_secret = aws_secretsmanager_secret.test.arn
    secrets_role_arn      = aws_iam_role.secrets_manager_test.arn
  }
  description = "description"
  manifest_name = "example-manifest-name"
  origination = "ALLOW"
  start_over_window_seconds = 3600
  time_delay_seconds = 60
  whitelist = ["192.168.0.1/24"]
}
`, testAccOriginEndpointConfig_baseResource(channelID), channelID)
}

func testAccCheckOriginEndpointDestroy(ctx context.Context, channelID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mediapackage_origin_endpoint" {
				continue
			}

			_, err := tfmediapackage.FindOriginEndpointByID(ctx, conn, rs.Primary.ID, channelID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.MediaPackage, create.ErrActionCheckingDestroyed, tfmediapackage.ResNameOriginEndpoint, rs.Primary.ID, err)
			}

			return create.Error(names.MediaPackage, create.ErrActionCheckingDestroyed, tfmediapackage.ResNameOriginEndpoint, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOriginEndpointExists(ctx context.Context, name, channelID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaPackage, create.ErrActionCheckingExistence, tfmediapackage.ResNameOriginEndpoint, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaPackage, create.ErrActionCheckingExistence, tfmediapackage.ResNameOriginEndpoint, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		_, err := tfmediapackage.FindOriginEndpointByID(ctx, conn, rs.Primary.ID, channelID)
		if err != nil {
			return create.Error(names.MediaPackage, create.ErrActionCheckingExistence, tfmediapackage.ResNameOriginEndpoint, rs.Primary.ID, err)
		}

		return nil
	}
}
