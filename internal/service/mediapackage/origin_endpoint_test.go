// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/mediapackage/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfmediapackage "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackage"
)

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccMediaPackageOriginEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// resourceName := "aws_media_package_origin_endpoint.test"
	seed := time.Now().Format("20060102150405")
	channelID := "test_basic" + seed

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
				Config: testAccOriginEndpointConfig_basic(channelID),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
				),
			},
			// {
			// 	ResourceName:            resourceName,
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			// },
		},
	})
}

// func TestAccMediaPackageOriginEndpoint_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}
//
// 	resourceName := "aws_mediapackage_origin_endpoint.test"
// 	channelID := "aws_mediapackage_channel.test_disappears.id"
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckDirectoryService(ctx, t)
// 			testAccPreCheck(ctx, t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx, channelID),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccOriginEndpointConfig_basic(channelID),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					testAccCheckOriginEndpointExists(ctx, resourceName, channelID),
// 					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediapackage.ResourceOriginEndpoint(), resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccOriginEndpointConfig_basic(channelID string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = "%[1]s"	
}

// resource "aws_iam_role" "test" {
//   name = "role-%[1]s"
//   assume_role_policy = jsonencode({
//     Version = "2012-10-17",
//     Statement = [
//       {
//         Action = "sts:AssumeRole",
//         Effect = "Allow",
//         Principal = {
//           Service = "mediapackage.amazonaws.com"
//         },
//       },
//     ],
//   })
// }
// 
// resource "aws_iam_policy" "test" {
//   name        = "SecretManagerAccessPolicy"
//   description = "A policy to allow access to AWS Secret Manager"
//   policy      = jsonencode({
//     Version = "2012-10-17",
//     Statement = [
//       {
//         Action = [
//           "secretsmanager:GetSecretValue",
//           "secretsmanager:DescribeSecret",
//           "secretsmanager:ListSecrets"
//         ],
//         Effect   = "Allow",
//         Resource = "*"
//       },
//     ],
//   })
// }

// resource "aws_iam_role_policy_attachment" "test" {
//   role       = aws_iam_role.test.name
//   policy_arn = aws_iam_policy.test.arn
// }

resource "aws_media_package_origin_endpoint" "test" {
  id = "endpoint-%[1]s"
  channel_id = aws_media_package_channel.test.id
	
  hls_package {
	ad_markers = "PASSTHROUGH"
    ad_triggers = ["SPLICE_INSERT"]
    ads_on_delivery_restrictions = "NONE"
    // encryption {
    //   speke_key_provider {
    //     resource_id = "example-resource-id"
    //     role_arn    = aws_iam_role.test.arn
    //     system_ids  = ["123e4567-e89b-12d3-a456-426614174000"]
    //     url         = "https://example.com/hoge"
    //   }
    // }
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

  // authorization {
  //   cdn_identifier_secret = "example-cdn-identifier-secret"
  //   secrets_role_arn      = aws_iam_role.test.arn
  // }

  // cmaf_package {
  //   encryption {
  //     speke_key_provider {
  //       resource_id = "example-resource-id"
  //       role_arn    = aws_iam_role.test.arn
  //       system_ids  = ["123e4567-e89b-12d3-a456-426614174000"]
  //       url         = "https://example.com/hoge"
  //     }
  //   }
	// hls_manifests {
	// 	ad_markers = "NONE"
	// 	ad_triggers = ["SPLICE_INSERT"]
	// 	ads_on_delivery_restrictions = "NONE"
	// 	include_iframe_only_stream = true
	// 	manifest_name = "example-manifest-name"
	// 	playlist_type = "NONE"
	// 	playlist_window_seconds = 60
	// 	program_date_time_interval_seconds = 600
  //   	segment_duration_seconds = 6
	// }

  // dash_package {
  //   ad_triggers = ["SPLICE_INSERT"]
  //   encryption {
  //     speke_key_provider {
  //       resource_id = "example-resource-id"
  //       role_arn    = aws_iam_role.test.arn
  //       system_ids  = ["123e4567-e89b-12d3-a456-426614174000"]
  //       url         = "https://example.com/hoge"
  //     }
  //   }
  //   period_triggers = ["ADS"]
  //   segment_duration_seconds = 6
  //   segment_template_format  = "NUMBER_WITH_TIMELINE"
  //   stream_selection {
  //     max_video_bits_per_second = 1000000
  //     min_video_bits_per_second = 500000
  //     stream_order              = "ORIGINAL"
  //   }
  //   suggested_presentation_delay_seconds = 60
  //   utc_timing                          = "HTTP-HEAD"
  //   utc_timing_uri                      = "https://time.example.com"
  // }
  // 
  // description = "example description"
  // hls_package {
  //   ad_markers = "PASSTHROUGH"
  //   encryption {
  //     speke_key_provider {
  //       resource_id = "example-resource-id"
  //       role_arn    = aws_iam_role.test.arn
  //       system_ids  = ["123e4567-e89b-12d3-a456-426614174000"]
  //       url         = "https://example.com/hoge"
  //     }
  //   }
  //   include_iframe_only_stream = true
  //   playlist_type              = "EVENT"
  //   playlist_window_seconds    = 60
  //   program_date_time_interval_seconds = 600
  //   segment_duration_seconds   = 6
  //   use_audio_rendition_group = true
  // }
  // 
  // manifest_name = "example-manifest-name"
  // mss_package {
  //   encryption {
  //     speke_key_provider {
  //       resource_id = "example-resource-id"
  //       role_arn    = aws_iam_role.test.arn
  //       system_ids  = ["123e4567-e89b-12d3-a456-426614174000"]
  //       url         = "https://example.com/hoge"
  //     }
  //   }
  //   manifest_window_seconds = 60
  //   segment_duration_seconds = 6
  // }
  // 
  // origination = "ALLOW"
  // start_over_window_seconds = 3600
  // tags = {
  //   Name = "example"
  // }
  // time_delay_seconds = 60
  // whitelist = ["192.168.0.1/24"]
}
`, channelID)
}

func testAccCheckOriginEndpointDestroy(ctx context.Context, channelID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mediapackage_origin_endpoint" && rs.Type != "aws_mediapackage_channel" {
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
