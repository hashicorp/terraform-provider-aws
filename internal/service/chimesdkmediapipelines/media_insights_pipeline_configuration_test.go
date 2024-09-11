// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkmediapipelines_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfchimesdkmediapipelines "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkmediapipelines"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChimeSDKMediaPipelinesMediaInsightsPipelineConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mipc awstypes.MediaInsightsPipelineConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKMediaPipelinesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKMediaPipelinesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_basic(rName, roleName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &mipc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "resource_access_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "elements.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "AmazonTranscribeCallAnalyticsProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`media-insights-pipeline-configuration/+.`)),
				),
			},
			{
				Config:             testAccMediaInsightsPipelineConfigurationConfig_basic(rName, roleName, streamName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccChimeSDKMediaPipelinesMediaInsightsPipelineConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mipc awstypes.MediaInsightsPipelineConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKMediaPipelinesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKMediaPipelinesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_basic(rName, roleName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &mipc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchimesdkmediapipelines.ResourceMediaInsightsPipelineConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeSDKMediaPipelinesMediaInsightsPipelineConfiguration_updateAllProcessorTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 awstypes.MediaInsightsPipelineConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKMediaPipelinesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKMediaPipelinesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_transcribeCallAnalyticsProcessor(rName, roleName1, streamName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName1)),
					resource.TestCheckResourceAttr(resourceName, "elements.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "AmazonTranscribeCallAnalyticsProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "elements.1.kinesis_data_stream_sink_configuration.0.insights_target", "kinesis", regexache.MustCompile(fmt.Sprintf(`stream/%s`, streamName1))),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.%"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.disabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.0.type", "IssueDetection"),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.rules.0.issue_detection_configuration.0.%"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.1.type", "KeywordMatch"),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.rules.1.keyword_match_configuration.0.%"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.2.type", "Sentiment"),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.rules.2.sentiment_configuration.0.%"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`media-insights-pipeline-configuration/+.`)),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_transcribeProcessor(rName, roleName1, streamName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v2),
					testAccCheckMediaInsightsPipelineConfigurationNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName1)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`media-insights-pipeline-configuration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "elements.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "AmazonTranscribeProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "elements.1.kinesis_data_stream_sink_configuration.0.insights_target", "kinesis", regexache.MustCompile(fmt.Sprintf(`stream/%s`, streamName2))),
					resource.TestCheckNoResourceAttr(resourceName, "real_time_alert_configuration.0.%"),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_s3RecordingSink(rName, roleName2, streamName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v3),
					testAccCheckMediaInsightsPipelineConfigurationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName2)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`media-insights-pipeline-configuration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "elements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "S3RecordingSink"),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_voiceAnalytics(rName, roleName2, streamName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v4),
					testAccCheckMediaInsightsPipelineConfigurationNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName2)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`media-insights-pipeline-configuration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "elements.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "VoiceAnalyticsProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "LambdaFunctionSink"),
					resource.TestCheckResourceAttr(resourceName, "elements.2.type", "SnsTopicSink"),
					resource.TestCheckResourceAttr(resourceName, "elements.3.type", "SqsQueueSink"),
					resource.TestCheckResourceAttr(resourceName, "elements.4.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "elements.4.kinesis_data_stream_sink_configuration.0.insights_target", "kinesis", regexache.MustCompile(fmt.Sprintf(`stream/%s`, streamName2))),
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

func TestAccChimeSDKMediaPipelinesMediaInsightsPipelineConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var mipc awstypes.MediaInsightsPipelineConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKMediaPipelinesEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKMediaPipelinesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_tags1(rName, roleName, streamName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &mipc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_tags2(rName, roleName, streamName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &mipc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_tags1(rName, roleName, streamName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &mipc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" {
				continue
			}

			_, err := tfchimesdkmediapipelines.FindMediaInsightsPipelineConfigurationByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if errs.IsA[*awstypes.NotFoundException](err) {
					return nil
				}
				return err
			}

			return create.Error(names.ChimeSDKMediaPipelines, create.ErrActionCheckingDestroyed,
				tfchimesdkmediapipelines.ResNameMediaInsightsPipelineConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMediaInsightsPipelineConfigurationExists(ctx context.Context, name string, mipc *awstypes.MediaInsightsPipelineConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ChimeSDKMediaPipelines, create.ErrActionCheckingExistence,
				tfchimesdkmediapipelines.ResNameMediaInsightsPipelineConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ChimeSDKMediaPipelines, create.ErrActionCheckingExistence,
				tfchimesdkmediapipelines.ResNameMediaInsightsPipelineConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)
		resp, err := tfchimesdkmediapipelines.FindMediaInsightsPipelineConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ChimeSDKMediaPipelines, create.ErrActionCheckingExistence,
				tfchimesdkmediapipelines.ResNameMediaInsightsPipelineConfiguration, rs.Primary.ID, err)
		}

		*mipc = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)

	input := &chimesdkmediapipelines.ListMediaInsightsPipelineConfigurationsInput{}
	_, err := conn.ListMediaInsightsPipelineConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckMediaInsightsPipelineConfigurationNotRecreated(before, after *awstypes.MediaInsightsPipelineConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.ToString(before.MediaInsightsPipelineConfigurationId), aws.ToString(after.MediaInsightsPipelineConfigurationId); beforeID != afterID {
			return create.Error(names.ChimeSDKMediaPipelines, create.ErrActionCheckingNotRecreated,
				tfchimesdkmediapipelines.ResNameMediaInsightsPipelineConfiguration, beforeID, errors.New("recreated"))
		}

		return nil
	}
}

func testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName string) string {
	return fmt.Sprintf(`
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
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_kinesis_stream" "test" {
  name        = %[2]q
  shard_count = 2
}
`, roleName, streamName)
}

func testAccMediaInsightsPipelineConfigurationConfig_basic(rName, roleName, streamName string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
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
`, rName))
}

func testAccMediaInsightsPipelineConfigurationConfig_transcribeCallAnalyticsProcessor(rName, roleName, streamName string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      call_analytics_stream_categories = [
        "category_1",
        "category_2"
      ]
      content_redaction_type               = "PII"
      enable_partial_results_stabilization = true
      filter_partial_results               = true
      language_code                        = "en-US"
      language_model_name                  = "MyLanguageModel"
      partial_results_stability            = "high"
      pii_entity_types                     = "ADDRESS,BANK_ACCOUNT_NUMBER"
      post_call_analytics_settings {
        content_redaction_output     = "redacted"
        data_access_role_arn         = aws_iam_role.test.arn
        output_encryption_kms_key_id = "MyKmsKeyId "
        output_location              = "s3://MyBucket"
      }
      vocabulary_filter_method = "mask"
      vocabulary_filter_name   = "MyVocabularyFilter"
      vocabulary_name          = "MyVocabulary"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.test.arn
    }
  }

  real_time_alert_configuration {
    disabled = false

    rules {
      type = "IssueDetection"
      issue_detection_configuration {
        rule_name = "MyIssueDetectionRule"
      }
    }

    rules {
      type = "KeywordMatch"
      keyword_match_configuration {
        keywords  = ["keyword1", "keyword2"]
        negate    = false
        rule_name = "MyKeywordMatchRule"
      }
    }

    rules {
      type = "Sentiment"
      sentiment_configuration {
        rule_name      = "MySentimentRule"
        sentiment_type = "NEGATIVE"
        time_period    = 60
      }
    }
  }
}
`, rName))
}

func testAccMediaInsightsPipelineConfigurationConfig_transcribeProcessor(rName, roleName, streamName string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "AmazonTranscribeProcessor"
    amazon_transcribe_processor_configuration {
      content_identification_type          = "PII"
      enable_partial_results_stabilization = true
      filter_partial_results               = true
      language_code                        = "en-US"
      language_model_name                  = "MyLanguageModel"
      partial_results_stability            = "high"
      pii_entity_types                     = "ADDRESS,BANK_ACCOUNT_NUMBER"
      show_speaker_label                   = true
      vocabulary_filter_method             = "mask"
      vocabulary_filter_name               = "MyVocabularyFilter"
      vocabulary_name                      = "MyVocabulary"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.test.arn
    }
  }
}
`, rName))
}

func testAccMediaInsightsPipelineConfigurationConfig_s3RecordingSink(rName, roleName, streamName string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "S3RecordingSink"
    s3_recording_sink_configuration {
      destination = "arn:${data.aws_partition.current.partition}:s3:::MyBucket"
    }
  }
}
`, rName))
}

func testAccMediaInsightsPipelineConfigurationConfig_voiceAnalytics(rName, roleName, streamName string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "VoiceAnalyticsProcessor"
    voice_analytics_processor_configuration {
      speaker_search_status      = "Enabled"
      voice_tone_analysis_status = "Enabled"
    }
  }

  elements {
    type = "LambdaFunctionSink"
    lambda_function_sink_configuration {
      insights_target = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:MyFunction"
    }
  }

  elements {
    type = "SnsTopicSink"
    sns_topic_sink_configuration {
      insights_target = "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:topic/MyTopic"
    }
  }

  elements {
    type = "SqsQueueSink"
    sqs_queue_sink_configuration {
      insights_target = "arn:${data.aws_partition.current.partition}:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:queue/MyQueue"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.test.arn
    }
  }
}
`, rName))
}

func testAccMediaInsightsPipelineConfigurationConfig_tags1(rName, roleName, streamName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
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

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccMediaInsightsPipelineConfigurationConfig_tags2(rName, roleName, streamName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccMediaInsightsPipelineConfigurationConfigBase(roleName, streamName),
		fmt.Sprintf(`
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %[1]q
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

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
