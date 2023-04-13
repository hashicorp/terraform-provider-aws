package chimesdkmediapipelines_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkmediapipelines"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfchimesdkmediapipelines "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkmediapipelines"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChimeSDKMediaPipelinesMediaInsightsPipelineConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var mipc chimesdkmediapipelines.MediaInsightsPipelineConfiguration
	configName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kinesisStreamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, chimesdkmediapipelines.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkmediapipelines.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_basic(roleName, configName, kinesisStreamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &mipc),
					resource.TestCheckResourceAttr(resourceName, "name", configName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_access_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "AmazonTranscribeCallAnalyticsProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "chime", regexp.MustCompile(`media-insights-pipeline-configuration/+.`)),
				),
			},
			{
				Config:             testAccMediaInsightsPipelineConfigurationConfig_basic(roleName, configName, kinesisStreamName),
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

	var mipc chimesdkmediapipelines.MediaInsightsPipelineConfiguration
	configName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kinesisStreamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, chimesdkmediapipelines.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkmediapipelines.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_basic(roleName, configName, kinesisStreamName),
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

	var v1, v2, v3, v4 chimesdkmediapipelines.MediaInsightsPipelineConfiguration
	configName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kinesisStreamName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kinesisStreamName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_chimesdkmediapipelines_media_insights_pipeline_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, chimesdkmediapipelines.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkmediapipelines.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_transcribeCallAnalyticsProcessor(roleName1, configName, kinesisStreamName1, formatTags("Key1", "Value1")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", configName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName1)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "AmazonTranscribeCallAnalyticsProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "elements.1.kinesis_data_stream_sink_configuration.0.insights_target", "kinesis", regexp.MustCompile(fmt.Sprintf(`stream/%s`, kinesisStreamName1))),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.%"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.0.type", "IssueDetection"),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.rules.0.issue_detection_configuration.0.%"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.1.type", "KeywordMatch"),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.rules.1.keyword_match_configuration.0.%"),
					resource.TestCheckResourceAttr(resourceName, "real_time_alert_configuration.0.rules.2.type", "Sentiment"),
					resource.TestCheckResourceAttrSet(resourceName, "real_time_alert_configuration.0.rules.2.sentiment_configuration.0.%"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "chime", regexp.MustCompile(`media-insights-pipeline-configuration/+.`)),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_transcribeProcessor(roleName1, configName, kinesisStreamName2, formatTags("Key1", "Value1", "Key2", "Value2")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v2),
					testAccCheckMediaInsightsPipelineConfigurationNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", configName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName1)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "chime", regexp.MustCompile(`media-insights-pipeline-configuration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "AmazonTranscribeProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "elements.1.kinesis_data_stream_sink_configuration.0.insights_target", "kinesis", regexp.MustCompile(fmt.Sprintf(`stream/%s`, kinesisStreamName2))),
					resource.TestCheckNoResourceAttr(resourceName, "real_time_alert_configuration.0.%"),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_s3RecordingSink(roleName2, configName, formatTags("Key1", "Value3")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v3),
					testAccCheckMediaInsightsPipelineConfigurationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", configName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName2)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "chime", regexp.MustCompile(`media-insights-pipeline-configuration/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "S3RecordingSink"),
				),
			},
			{
				Config: testAccMediaInsightsPipelineConfigurationConfig_voiceAnalytics(roleName2, configName, kinesisStreamName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInsightsPipelineConfigurationExists(ctx, resourceName, &v4),
					testAccCheckMediaInsightsPipelineConfigurationNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "name", configName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "resource_access_role_arn", "iam", fmt.Sprintf(`role/%s`, roleName2)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "chime", regexp.MustCompile(`media-insights-pipeline-configuration/+.`)),
					resource.TestCheckNoResourceAttr(resourceName, "tags.%"),
					resource.TestCheckResourceAttr(resourceName, "elements.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "elements.0.type", "VoiceAnalyticsProcessor"),
					resource.TestCheckResourceAttr(resourceName, "elements.1.type", "LambdaFunctionSink"),
					resource.TestCheckResourceAttr(resourceName, "elements.2.type", "SnsTopicSink"),
					resource.TestCheckResourceAttr(resourceName, "elements.3.type", "SqsQueueSink"),
					resource.TestCheckResourceAttr(resourceName, "elements.4.type", "KinesisDataStreamSink"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "elements.4.kinesis_data_stream_sink_configuration.0.insights_target", "kinesis", regexp.MustCompile(fmt.Sprintf(`stream/%s`, kinesisStreamName2))),
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

func testAccCheckMediaInsightsPipelineConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKMediaPipelinesConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" {
				continue
			}

			_, err := tfchimesdkmediapipelines.FindMediaInsightsPipelineConfigurationByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, chimesdkmediapipelines.ErrCodeNotFoundException) {
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

func testAccCheckMediaInsightsPipelineConfigurationExists(ctx context.Context, name string, mipc *chimesdkmediapipelines.MediaInsightsPipelineConfiguration) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKMediaPipelinesConn()
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKMediaPipelinesConn()

	input := &chimesdkmediapipelines.ListMediaInsightsPipelineConfigurationsInput{}
	_, err := conn.ListMediaInsightsPipelineConfigurationsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckMediaInsightsPipelineConfigurationNotRecreated(before, after *chimesdkmediapipelines.MediaInsightsPipelineConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.StringValue(before.MediaInsightsPipelineConfigurationId), aws.StringValue(after.MediaInsightsPipelineConfigurationId); beforeID != afterID {
			return create.Error(names.ChimeSDKMediaPipelines, create.ErrActionCheckingNotRecreated,
				tfchimesdkmediapipelines.ResNameMediaInsightsPipelineConfiguration, beforeID, errors.New("recreated"))
		}

		return nil
	}
}

func resourceAccessRole(roleName string) string {
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
  name               = "%s"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}


`, roleName)
}

func kinesisDataStream(streamName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %q
  shard_count = 2
}`, streamName)
}

func testAccMediaInsightsPipelineConfigurationConfig_basic(roleName string, configName string, kinesisStreamName string) string {
	return fmt.Sprintf(`
%s

%s

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %q
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
`, resourceAccessRole(roleName), kinesisDataStream(kinesisStreamName), configName)
}

func testAccMediaInsightsPipelineConfigurationConfig_transcribeCallAnalyticsProcessor(roleName, configName, kinesisStreamName, tags string) string {
	return fmt.Sprintf(`
%s

%s

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %q
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
  %s
}
`, resourceAccessRole(roleName), kinesisDataStream(kinesisStreamName), configName, tags)
}

func testAccMediaInsightsPipelineConfigurationConfig_transcribeProcessor(roleName, configName, kinesisStreamName, tags string) string {
	return fmt.Sprintf(`
%s

%s

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %q
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

  %s
}
`, resourceAccessRole(roleName), kinesisDataStream(kinesisStreamName), configName, tags)
}

func testAccMediaInsightsPipelineConfigurationConfig_s3RecordingSink(roleName, configName, tags string) string {
	return fmt.Sprintf(`
%s

data "aws_partition" "current" {}

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %q
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "S3RecordingSink"
    s3_recording_sink_configuration {
      destination = "arn:${data.aws_partition.current.partition}:s3:::MyBucket"
    }
  }

  %s
}
`, resourceAccessRole(roleName), configName, tags)
}

func testAccMediaInsightsPipelineConfigurationConfig_voiceAnalytics(roleName string, configName string, kinesisStreamName string) string {
	return fmt.Sprintf(`
%s

%s

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}


resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
  name                     = %q
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
`, resourceAccessRole(roleName), kinesisDataStream(kinesisStreamName), configName)
}

func formatTags(tags ...string) string {
	if len(tags) == 0 || len(tags)%2 == 1 {
		return ""
	}
	tagsBlock := "tags = {\n"
	for i := 0; i < len(tags)-1; i += 2 {
		tagsBlock += fmt.Sprintf("  %s = %q\n", tags[i], tags[i+1])
	}
	tagsBlock += "}"
	return tagsBlock
}
