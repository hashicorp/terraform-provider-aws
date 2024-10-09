// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVoiceConnectorStreaming_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_streaming.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorStreamingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorStreamingConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorStreamingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_retention", "5"),
					resource.TestCheckResourceAttr(resourceName, "disabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "streaming_notification_targets.#", acctest.Ct1),
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

func testAccVoiceConnectorStreaming_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_streaming.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorStreamingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorStreamingConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorStreamingExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchime.ResourceVoiceConnectorStreaming(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVoiceConnectorStreaming_update(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_streaming.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorStreamingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorStreamingConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorStreamingExists(ctx, resourceName),
				),
			},
			{
				Config: testAccVoiceConnectorStreamingConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorStreamingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_retention", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "disabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "streaming_notification_targets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "media_insights_configuration.0.disabled", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(resourceName,
						"media_insights_configuration.0.configuration_arn",
						"chime",
						regexache.MustCompile(fmt.Sprintf(`media-insights-pipeline-configuration/test-config-%s`, name)),
					),
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

func testAccVoiceConnectorStreamingConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  disabled                       = false
  data_retention                 = 5
  streaming_notification_targets = ["SQS"]
}
`, name)
}

func testAccVoiceConnectorStreamingConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  disabled                       = false
  data_retention                 = 2
  streaming_notification_targets = ["SQS", "SNS"]
  media_insights_configuration {
    disabled          = false
    configuration_arn = aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test.arn
  }
}

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = "test-config-%[1]s"
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      language_code = "en-US"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.test.arn
    }
  }
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["mediapipelines.chime.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "test" {
  name               = "resource_access_role-%[1]s"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_kinesis_stream" "test" {
  name        = "kvs-%[1]s"
  shard_count = 2
}
`, name)
}

func testAccCheckVoiceConnectorStreamingExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector streaming configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)
		input := &chimesdkvoice.GetVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorStreamingConfiguration(ctx, input)
		if err != nil {
			return err
		}

		if resp == nil || resp.StreamingConfiguration == nil {
			return fmt.Errorf("no Chime Voice Connector Streaming configuration (%s) found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVoiceConnectorStreamingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chime_voice_connector_termination" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)
			input := &chimesdkvoice.GetVoiceConnectorStreamingConfigurationInput{
				VoiceConnectorId: aws.String(rs.Primary.ID),
			}
			resp, err := conn.GetVoiceConnectorStreamingConfiguration(ctx, input)

			if errs.IsA[*awstypes.NotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil && resp.StreamingConfiguration != nil {
				return fmt.Errorf("error Chime Voice Connector streaming configuration still exists")
			}
		}

		return nil
	}
}
