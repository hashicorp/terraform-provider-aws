// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkmediapipelines

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const iamPropagationTimeout = 2 * time.Minute

const (
	ResNameMediaInsightsPipelineConfiguration = "Media Insights Pipeline Configuration"
)

var (
	errConvertingElement           = errors.New("unable to convert element")
	errConvertingRuleConfiguration = errors.New("unable to convert rule configuration")
)

// @SDKResource("aws_chimesdkmediapipelines_media_insights_pipeline_configuration", name="Media Insights Pipeline Configuration")
// @Tags(identifierAttribute="arn")
func ResourceMediaInsightsPipelineConfiguration() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceMediaInsightsPipelineConfigurationCreate,
		ReadWithoutTimeout:   resourceMediaInsightsPipelineConfigurationRead,
		UpdateWithoutTimeout: resourceMediaInsightsPipelineConfigurationUpdate,
		DeleteWithoutTimeout: resourceMediaInsightsPipelineConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// Resource creation/update/deletion is atomic and synchronous with the API calls. The timeouts for
		// create and update are dominated by timeout waiting for IAM role changes to propagate.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elements": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MediaInsightsPipelineConfigurationElementType](),
						},
						"amazon_transcribe_call_analytics_processor_configuration": AmazonTranscribeCallAnalyticsProcessorConfigurationSchema(),
						"amazon_transcribe_processor_configuration":                AmazonTranscribeProcessorConfigurationSchema(),
						"kinesis_data_stream_sink_configuration":                   BasicSinkConfigurationSchema(),
						"lambda_function_sink_configuration":                       BasicSinkConfigurationSchema(),
						"sns_topic_sink_configuration":                             BasicSinkConfigurationSchema(),
						"sqs_queue_sink_configuration":                             BasicSinkConfigurationSchema(),
						"s3_recording_sink_configuration":                          S3RecordingSinkConfigurationSchema(),
						"voice_analytics_processor_configuration":                  VoiceAnalyticsProcessorConfigurationSchema(),
					},
				},
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_access_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"real_time_alert_configuration": RealTimeAlertConfigurationSchema(),
			names.AttrTags:                  tftags.TagsSchema(),
			names.AttrTagsAll:               tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func AmazonTranscribeCallAnalyticsProcessorConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"call_analytics_stream_categories": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 20,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 200),
							validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid Category Name matching expression: ^[0-9A-Za-z_.-]+"),
						),
					},
				},
				"content_identification_type": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ContentType](),
				},
				"content_redaction_type": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ContentType](),
				},
				"enable_partial_results_stabilization": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"filter_partial_results": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				names.AttrLanguageCode: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.CallAnalyticsLanguageCode](),
				},
				"language_model_name": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 200),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid language model name matching expression: ^[0-9A-Za-z_.-]+"),
					),
				},
				"partial_results_stability": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PartialResultsStability](),
				},
				"pii_entity_types": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 300),
						validation.StringMatch(regexache.MustCompile(`^[A-Z_, ]+`),
							"Must be a valid comma-separated list of entity types. For valid types see https://docs.aws.amazon.com/chime-sdk/latest/APIReference/API_media-pipelines-chime_AmazonTranscribeCallAnalyticsProcessorConfiguration.html#chimesdk-Type-media-pipelines-chime_AmazonTranscribeCallAnalyticsProcessorConfiguration-CallAnalyticsStreamCategories"),
					),
				},
				"post_call_analytics_settings": PostCallAnalyticsSettingsSchema(),
				"vocabulary_filter_method": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.VocabularyFilterMethod](),
				},
				"vocabulary_filter_name": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 200),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid vocabulary filter name matching expression: ^[0-9A-Za-z_.-]+"),
					),
				},
				"vocabulary_name": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 200),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid vocabulary name matching expression: ^[0-9A-Za-z_.-]+"),
					),
				},
			},
		},
	}
}

func PostCallAnalyticsSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content_redaction_output": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ContentRedactionOutput](),
				},
				"data_access_role_arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"output_encryption_kms_key_id": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 4096),
				},
				"output_location": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexache.MustCompile(`s3://+`), "Must begin with the prefix 's3://'"),
				},
			},
		},
	}
}

func AmazonTranscribeProcessorConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content_identification_type": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ContentType](),
				},
				"content_redaction_type": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ContentType](),
				},
				"enable_partial_results_stabilization": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"filter_partial_results": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				names.AttrLanguageCode: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.CallAnalyticsLanguageCode](),
				},
				"language_model_name": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 200),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid language model name matching expression: ^[0-9A-Za-z_.-]+"),
					),
				},
				"partial_results_stability": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PartialResultsStability](),
				},
				"pii_entity_types": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 300),
						validation.StringMatch(regexache.MustCompile(`^[A-Z_, ]+`),
							"Must be a valid comma-separated list of entity types. For valid types see https://docs.aws.amazon.com/chime-sdk/latest/APIReference/API_media-pipelines-chime_AmazonTranscribeCallAnalyticsProcessorConfiguration.html#chimesdk-Type-media-pipelines-chime_AmazonTranscribeCallAnalyticsProcessorConfiguration-CallAnalyticsStreamCategories"),
					),
				},
				"show_speaker_label": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"vocabulary_filter_method": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.VocabularyFilterMethod](),
				},
				"vocabulary_filter_name": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 200),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid vocabulary filter name matching expression: ^[0-9A-Za-z_.-]+"),
					),
				},
				"vocabulary_name": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 200),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must be a valid vocabulary name matching expression: ^[0-9A-Za-z_.-]+"),
					),
				},
			},
		},
	}
}

func BasicSinkConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"insights_target": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func S3RecordingSinkConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrDestination: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func VoiceAnalyticsProcessorConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"speaker_search_status": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.VoiceAnalyticsConfigurationStatus](),
				},
				"voice_tone_analysis_status": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.VoiceAnalyticsConfigurationStatus](),
				},
			},
		},
	}
}

func RealTimeAlertConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
				"rules": {
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 3,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"issue_detection_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rule_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must match the expression: ^[0-9A-Za-z_.-]+"),
										},
									},
								},
							},
							"keyword_match_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"keywords": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z\s'-]+`), "Must match the expression: ^[0-9A-Za-z\\s'-]+"),
											},
										},
										"negate": {
											Type:     schema.TypeBool,
											Optional: true,
											Computed: true,
										},
										"rule_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must match the expression: ^[0-9A-Za-z_.-]+"),
										},
									},
								},
							},
							"sentiment_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rule_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "Must match the expression: ^[0-9A-Za-z_.-]+"),
										},
										"sentiment_type": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.SentimentType](),
										},
										"time_period": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(60, 1800),
										},
									},
								},
							},
							names.AttrType: {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.RealTimeAlertRuleType](),
							},
						},
					},
				},
			},
		},
	}
}

func resourceMediaInsightsPipelineConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)

	elements, err := expandElements(d.Get("elements").([]interface{}))
	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionCreating,
			ResNameMediaInsightsPipelineConfiguration, d.Get(names.AttrName).(string), err)
	}

	in := &chimesdkmediapipelines.CreateMediaInsightsPipelineConfigurationInput{
		MediaInsightsPipelineConfigurationName: aws.String(d.Get(names.AttrName).(string)),
		ResourceAccessRoleArn:                  aws.String(d.Get("resource_access_role_arn").(string)),
		Elements:                               elements,
		Tags:                                   getTagsIn(ctx),
	}

	if realTimeAlertConfiguration, ok := d.GetOk("real_time_alert_configuration"); ok && len(realTimeAlertConfiguration.([]interface{})) > 0 {
		rtac, err := expandRealTimeAlertConfiguration(realTimeAlertConfiguration.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionCreating,
				ResNameMediaInsightsPipelineConfiguration, d.Get(names.AttrName).(string), err)
		}
		in.RealTimeAlertConfiguration = rtac
	}

	// Retry when forbidden exception is received; iam role propagation is eventually consistent
	var out *chimesdkmediapipelines.CreateMediaInsightsPipelineConfigurationOutput
	createError := tfresource.Retry(ctx, iamPropagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreateMediaInsightsPipelineConfiguration(ctx, in)
		if err != nil {
			var forbiddenException *awstypes.ForbiddenException
			if errors.As(err, &forbiddenException) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if createError != nil {
		return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionCreating, ResNameMediaInsightsPipelineConfiguration, d.Get(names.AttrName).(string), createError)
	}

	if out == nil || out.MediaInsightsPipelineConfiguration == nil {
		return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionCreating, ResNameMediaInsightsPipelineConfiguration, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.MediaInsightsPipelineConfiguration.MediaInsightsPipelineConfigurationArn))

	return append(diags, resourceMediaInsightsPipelineConfigurationRead(ctx, d, meta)...)
}

func resourceMediaInsightsPipelineConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)

	out, err := FindMediaInsightsPipelineConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ChimeSDKMediaPipelines MediaInsightsPipelineConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionReading, ResNameMediaInsightsPipelineConfiguration, d.Id(), err)
	}

	d.Set(names.AttrARN, out.MediaInsightsPipelineConfigurationArn)
	d.Set(names.AttrName, out.MediaInsightsPipelineConfigurationName)
	d.Set(names.AttrID, out.MediaInsightsPipelineConfigurationId)
	d.Set("resource_access_role_arn", out.ResourceAccessRoleArn)
	if err := d.Set("elements", flattenElements(out.Elements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting elements: %s", err)
	}
	if out.RealTimeAlertConfiguration != nil {
		if err := d.Set("real_time_alert_configuration", flattenRealTimeAlertConfiguration(out.RealTimeAlertConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting real time alert configuration: %s", err)
		}
	}

	return diags
}

func resourceMediaInsightsPipelineConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		elements, err := expandElements(d.Get("elements").([]interface{}))
		if err != nil {
			return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionUpdating, ResNameMediaInsightsPipelineConfiguration, d.Id(), err)
		}

		in := &chimesdkmediapipelines.UpdateMediaInsightsPipelineConfigurationInput{
			Identifier:            aws.String(d.Id()),
			ResourceAccessRoleArn: aws.String(d.Get("resource_access_role_arn").(string)),
			Elements:              elements,
		}
		if realTimeAlertConfiguration, ok := d.GetOk("real_time_alert_configuration"); ok && len(realTimeAlertConfiguration.([]interface{})) > 0 {
			rtac, err := expandRealTimeAlertConfiguration(realTimeAlertConfiguration.([]interface{})[0].(map[string]interface{}))
			if err != nil {
				return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionUpdating, ResNameMediaInsightsPipelineConfiguration, d.Id(), err)
			}
			in.RealTimeAlertConfiguration = rtac
		}

		// Retry when forbidden exception is received; iam role propagation is eventually consistent
		updateError := tfresource.Retry(ctx, iamPropagationTimeout, func() *retry.RetryError {
			var err error
			_, err = conn.UpdateMediaInsightsPipelineConfiguration(ctx, in)
			if err != nil {
				var forbiddenException *awstypes.ForbiddenException
				if errors.As(err, &forbiddenException) {
					return retry.RetryableError(err)
				}
				return retry.NonRetryableError(err)
			}

			return nil
		})
		if updateError != nil {
			return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionUpdating, ResNameMediaInsightsPipelineConfiguration, d.Id(), updateError)
		}
	}

	return append(diags, resourceMediaInsightsPipelineConfigurationRead(ctx, d, meta)...)
}

func resourceMediaInsightsPipelineConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ChimeSDKMediaPipelinesClient(ctx)

	log.Printf("[INFO] Deleting ChimeSDKMediaPipelines MediaInsightsPipelineConfiguration %s", d.Id())

	// eventual consistency may cause an initial failure
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 15*time.Second, func() (interface{}, error) {
		return conn.DeleteMediaInsightsPipelineConfiguration(ctx, &chimesdkmediapipelines.DeleteMediaInsightsPipelineConfigurationInput{
			Identifier: aws.String(d.Id()),
		})
	}, "ConflictException", "Cannot delete a Media Insights Pipeline Configuration while it is in use by 1 or more Voice Connectors")

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ChimeSDKMediaPipelines, create.ErrActionDeleting, ResNameMediaInsightsPipelineConfiguration, d.Id(), err)
	}

	return diags
}

func FindMediaInsightsPipelineConfigurationByID(ctx context.Context, conn *chimesdkmediapipelines.Client, id string) (*awstypes.MediaInsightsPipelineConfiguration, error) {
	in := &chimesdkmediapipelines.GetMediaInsightsPipelineConfigurationInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetMediaInsightsPipelineConfiguration(ctx, in)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.MediaInsightsPipelineConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.MediaInsightsPipelineConfiguration, nil
}

func expandElements(inputElements []interface{}) ([]awstypes.MediaInsightsPipelineConfigurationElement, error) {
	if len(inputElements) == 0 || inputElements[0] == nil {
		return nil, nil
	}
	elements := make([]awstypes.MediaInsightsPipelineConfigurationElement, 0, len(inputElements))
	for _, inputElement := range inputElements {
		apiElement, err := expandElement(inputElement)
		if err != nil {
			return nil, err
		}

		if apiElement == (awstypes.MediaInsightsPipelineConfigurationElement{}) {
			continue
		}
		elements = append(elements, apiElement)
	}
	return elements, nil
}

func expandElement(inputElement interface{}) (awstypes.MediaInsightsPipelineConfigurationElement, error) {
	inputMapRaw, ok := inputElement.(map[string]interface{})
	if !ok {
		return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
	}

	element := awstypes.MediaInsightsPipelineConfigurationElement{
		Type: awstypes.MediaInsightsPipelineConfigurationElementType(inputMapRaw[names.AttrType].(string)),
	}

	switch {
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeAmazonTranscribeCallAnalyticsProcessor:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["amazon_transcribe_call_analytics_processor_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.AmazonTranscribeCallAnalyticsProcessorConfiguration = &awstypes.AmazonTranscribeCallAnalyticsProcessorConfiguration{
			LanguageCode: awstypes.CallAnalyticsLanguageCode(rawConfiguration[names.AttrLanguageCode].(string)),
		}

		if callAnalyticsStreamCategories, ok := rawConfiguration["call_analytics_stream_categories"].([]interface{}); ok && len(callAnalyticsStreamCategories) > 0 {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.CallAnalyticsStreamCategories = flex.ExpandStringValueList(callAnalyticsStreamCategories)
		}

		if contentIdentificationType, ok := rawConfiguration["content_identification_type"].(string); ok && contentIdentificationType != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.ContentIdentificationType = awstypes.ContentType(contentIdentificationType)
		}

		if contentRedactionType, ok := rawConfiguration["content_redaction_type"].(string); ok && contentRedactionType != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.ContentRedactionType = awstypes.ContentType(contentRedactionType)
		}

		if enablePartialResultsStabilization, ok := rawConfiguration["enable_partial_results_stabilization"].(bool); ok {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.EnablePartialResultsStabilization = enablePartialResultsStabilization
		}

		if filterPartialResults, ok := rawConfiguration["filter_partial_results"].(bool); ok {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.FilterPartialResults = filterPartialResults
		}

		if languageModelName, ok := rawConfiguration["language_model_name"].(string); ok && languageModelName != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.LanguageModelName = aws.String(languageModelName)
		}

		if partialResultsStability, ok := rawConfiguration["partial_results_stability"].(string); ok && partialResultsStability != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.PartialResultsStability = awstypes.PartialResultsStability(partialResultsStability)
		}

		if piiEntityTypes, ok := rawConfiguration["pii_entity_types"].(string); ok && piiEntityTypes != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.PiiEntityTypes = aws.String(piiEntityTypes)
		}

		if postCallAnalyticsSettings, ok := rawConfiguration["post_call_analytics_settings"].([]interface{}); ok && len(postCallAnalyticsSettings) == 1 {
			rawPostCallSettings := postCallAnalyticsSettings[0].(map[string]interface{})

			postCallSettingsApi := &awstypes.PostCallAnalyticsSettings{
				DataAccessRoleArn: aws.String(rawPostCallSettings["data_access_role_arn"].(string)),
				OutputLocation:    aws.String(rawPostCallSettings["output_location"].(string)),
			}
			if contentRedactionOutput, ok := rawPostCallSettings["content_redaction_output"].(string); ok && len(contentRedactionOutput) > 0 {
				postCallSettingsApi.ContentRedactionOutput = awstypes.ContentRedactionOutput(contentRedactionOutput)
			}
			if outputEncryptionKMSKeyId, ok := rawPostCallSettings["output_encryption_kms_key_id"].(string); ok && len(outputEncryptionKMSKeyId) > 0 {
				postCallSettingsApi.OutputEncryptionKMSKeyId = aws.String(outputEncryptionKMSKeyId)
			}
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.PostCallAnalyticsSettings = postCallSettingsApi
		}

		if vocabularyFilterMethod, ok := rawConfiguration["vocabulary_filter_method"].(string); ok && vocabularyFilterMethod != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.VocabularyFilterMethod = awstypes.VocabularyFilterMethod(vocabularyFilterMethod)
		}

		if vocabularyFilterName, ok := rawConfiguration["vocabulary_filter_name"].(string); ok && vocabularyFilterName != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.VocabularyFilterName = aws.String(vocabularyFilterName)
		}

		if vocabularyName, ok := rawConfiguration["vocabulary_name"].(string); ok && vocabularyName != "" {
			element.AmazonTranscribeCallAnalyticsProcessorConfiguration.VocabularyName = aws.String(vocabularyName)
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeAmazonTranscribeProcessor:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["amazon_transcribe_processor_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.AmazonTranscribeProcessorConfiguration = &awstypes.AmazonTranscribeProcessorConfiguration{
			LanguageCode: awstypes.CallAnalyticsLanguageCode(rawConfiguration[names.AttrLanguageCode].(string)),
		}

		if contentIdentificationType, ok := rawConfiguration["content_identification_type"].(string); ok && contentIdentificationType != "" {
			element.AmazonTranscribeProcessorConfiguration.ContentIdentificationType = awstypes.ContentType(contentIdentificationType)
		}

		if contentRedactionType, ok := rawConfiguration["content_redaction_type"].(string); ok && contentRedactionType != "" {
			element.AmazonTranscribeProcessorConfiguration.ContentRedactionType = awstypes.ContentType(contentRedactionType)
		}

		if enablePartialResultsStabilization, ok := rawConfiguration["enable_partial_results_stabilization"].(bool); ok {
			element.AmazonTranscribeProcessorConfiguration.EnablePartialResultsStabilization = enablePartialResultsStabilization
		}

		if filterPartialResults, ok := rawConfiguration["filter_partial_results"].(bool); ok {
			element.AmazonTranscribeProcessorConfiguration.FilterPartialResults = filterPartialResults
		}

		if languageModelName, ok := rawConfiguration["language_model_name"].(string); ok && languageModelName != "" {
			element.AmazonTranscribeProcessorConfiguration.LanguageModelName = aws.String(languageModelName)
		}

		if partialResultsStability, ok := rawConfiguration["partial_results_stability"].(string); ok && partialResultsStability != "" {
			element.AmazonTranscribeProcessorConfiguration.PartialResultsStability = awstypes.PartialResultsStability(partialResultsStability)
		}

		if piiEntityTypes, ok := rawConfiguration["pii_entity_types"].(string); ok && piiEntityTypes != "" {
			element.AmazonTranscribeProcessorConfiguration.PiiEntityTypes = aws.String(piiEntityTypes)
		}

		if showSpeakerLabel, ok := rawConfiguration["show_speaker_label"].(bool); ok {
			element.AmazonTranscribeProcessorConfiguration.ShowSpeakerLabel = showSpeakerLabel
		}

		if vocabularyFilterMethod, ok := rawConfiguration["vocabulary_filter_method"].(string); ok && vocabularyFilterMethod != "" {
			element.AmazonTranscribeProcessorConfiguration.VocabularyFilterMethod = awstypes.VocabularyFilterMethod(vocabularyFilterMethod)
		}

		if vocabularyFilterName, ok := rawConfiguration["vocabulary_filter_name"].(string); ok && vocabularyFilterName != "" {
			element.AmazonTranscribeProcessorConfiguration.VocabularyFilterName = aws.String(vocabularyFilterName)
		}

		if vocabularyName, ok := rawConfiguration["vocabulary_name"].(string); ok && vocabularyName != "" {
			element.AmazonTranscribeProcessorConfiguration.VocabularyName = aws.String(vocabularyName)
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeKinesisDataStreamSink:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["kinesis_data_stream_sink_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.KinesisDataStreamSinkConfiguration = &awstypes.KinesisDataStreamSinkConfiguration{
			InsightsTarget: aws.String(rawConfiguration["insights_target"].(string)),
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeSnsTopicSink:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["sns_topic_sink_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.SnsTopicSinkConfiguration = &awstypes.SnsTopicSinkConfiguration{
			InsightsTarget: aws.String(rawConfiguration["insights_target"].(string)),
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeSqsQueueSink:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["sqs_queue_sink_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.SqsQueueSinkConfiguration = &awstypes.SqsQueueSinkConfiguration{
			InsightsTarget: aws.String(rawConfiguration["insights_target"].(string)),
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeLambdaFunctionSink:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["lambda_function_sink_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.LambdaFunctionSinkConfiguration = &awstypes.LambdaFunctionSinkConfiguration{
			InsightsTarget: aws.String(rawConfiguration["insights_target"].(string)),
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeS3RecordingSink:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["s3_recording_sink_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.S3RecordingSinkConfiguration = &awstypes.S3RecordingSinkConfiguration{
			Destination: aws.String(rawConfiguration[names.AttrDestination].(string)),
		}
	case element.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeVoiceAnalyticsProcessor:
		var configuration []interface{}
		if configuration, ok = inputMapRaw["voice_analytics_processor_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.MediaInsightsPipelineConfigurationElement{}, errConvertingElement
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		element.VoiceAnalyticsProcessorConfiguration = &awstypes.VoiceAnalyticsProcessorConfiguration{
			SpeakerSearchStatus:     awstypes.VoiceAnalyticsConfigurationStatus(rawConfiguration["speaker_search_status"].(string)),
			VoiceToneAnalysisStatus: awstypes.VoiceAnalyticsConfigurationStatus(rawConfiguration["voice_tone_analysis_status"].(string)),
		}
	}
	return element, nil
}

func expandRealTimeAlertConfiguration(inputConfiguration map[string]interface{}) (*awstypes.RealTimeAlertConfiguration, error) {
	apiConfiguration := &awstypes.RealTimeAlertConfiguration{
		Disabled: inputConfiguration["disabled"].(bool),
	}
	if inputRules, ok := inputConfiguration["rules"].([]interface{}); ok && len(inputRules) > 0 {
		rules := make([]awstypes.RealTimeAlertRule, 0, len(inputRules))
		for _, inputRule := range inputRules {
			rule, err := expandRealTimeAlertRule(inputRule)
			if err != nil {
				return apiConfiguration, err
			}
			rules = append(rules, rule)
		}
		apiConfiguration.Rules = rules
	}
	return apiConfiguration, nil
}

func expandRealTimeAlertRule(inputRule interface{}) (awstypes.RealTimeAlertRule, error) {
	inputRuleRaw, ok := inputRule.(map[string]interface{})
	if !ok {
		return awstypes.RealTimeAlertRule{}, nil
	}
	ruleType := awstypes.RealTimeAlertRuleType(inputRuleRaw[names.AttrType].(string))
	apiRule := awstypes.RealTimeAlertRule{
		Type: ruleType,
	}

	switch {
	case ruleType == awstypes.RealTimeAlertRuleTypeIssueDetection:
		var configuration []interface{}
		if configuration, ok = inputRuleRaw["issue_detection_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.RealTimeAlertRule{}, errConvertingRuleConfiguration
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		apiConfiguration := &awstypes.IssueDetectionConfiguration{
			RuleName: aws.String(rawConfiguration["rule_name"].(string)),
		}

		apiRule.IssueDetectionConfiguration = apiConfiguration
	case ruleType == awstypes.RealTimeAlertRuleTypeKeywordMatch:
		var configuration []interface{}
		if configuration, ok = inputRuleRaw["keyword_match_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.RealTimeAlertRule{}, errConvertingRuleConfiguration
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		apiConfiguration := &awstypes.KeywordMatchConfiguration{
			Keywords: flex.ExpandStringValueList((rawConfiguration["keywords"].([]interface{}))),
			RuleName: aws.String(rawConfiguration["rule_name"].(string)),
		}

		if negate, ok := rawConfiguration["negate"]; ok {
			apiConfiguration.Negate = negate.(bool)
		}

		apiRule.KeywordMatchConfiguration = apiConfiguration
	case ruleType == awstypes.RealTimeAlertRuleTypeSentiment:
		var configuration []interface{}
		if configuration, ok = inputRuleRaw["sentiment_configuration"].([]interface{}); !ok || len(configuration) != 1 {
			return awstypes.RealTimeAlertRule{}, errConvertingRuleConfiguration
		}

		rawConfiguration := configuration[0].(map[string]interface{})
		apiConfiguration := &awstypes.SentimentConfiguration{
			RuleName:      aws.String(rawConfiguration["rule_name"].(string)),
			SentimentType: awstypes.SentimentType(rawConfiguration["sentiment_type"].(string)),
			TimePeriod:    aws.Int32(int32(rawConfiguration["time_period"].(int))),
		}

		apiRule.SentimentConfiguration = apiConfiguration
	}
	return apiRule, nil
}

func flattenElements(apiElements []awstypes.MediaInsightsPipelineConfigurationElement) []interface{} {
	if len(apiElements) == 0 {
		return nil
	}
	var tfElements []interface{}
	for _, apiElement := range apiElements {
		if apiElement == (awstypes.MediaInsightsPipelineConfigurationElement{}) {
			continue
		}
		tfElements = append(tfElements, flattenElement(apiElement))
	}
	return tfElements
}

func flattenElement(apiElement awstypes.MediaInsightsPipelineConfigurationElement) map[string]interface{} {
	if apiElement == (awstypes.MediaInsightsPipelineConfigurationElement{}) {
		return nil
	}
	tfMap := map[string]interface{}{}
	tfMap[names.AttrType] = string(apiElement.Type)

	configuration := map[string]interface{}{}

	switch {
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeAmazonTranscribeCallAnalyticsProcessor:
		processorConfiguration := apiElement.AmazonTranscribeCallAnalyticsProcessorConfiguration
		configuration["call_analytics_stream_categories"] = processorConfiguration.CallAnalyticsStreamCategories
		configuration["content_identification_type"] = processorConfiguration.ContentIdentificationType
		configuration["content_redaction_type"] = processorConfiguration.ContentRedactionType
		configuration["enable_partial_results_stabilization"] = processorConfiguration.EnablePartialResultsStabilization
		configuration["filter_partial_results"] = processorConfiguration.FilterPartialResults
		configuration[names.AttrLanguageCode] = processorConfiguration.LanguageCode
		configuration["language_model_name"] = processorConfiguration.LanguageModelName
		configuration["partial_results_stability"] = processorConfiguration.PartialResultsStability
		configuration["pii_entity_types"] = processorConfiguration.PiiEntityTypes

		if processorConfiguration.PostCallAnalyticsSettings != nil {
			postCallSettings := map[string]interface{}{}
			postCallSettings["content_redaction_output"] = processorConfiguration.PostCallAnalyticsSettings.ContentRedactionOutput
			postCallSettings["data_access_role_arn"] = processorConfiguration.PostCallAnalyticsSettings.DataAccessRoleArn
			postCallSettings["output_encryption_kms_key_id"] = processorConfiguration.PostCallAnalyticsSettings.OutputEncryptionKMSKeyId
			postCallSettings["output_location"] = processorConfiguration.PostCallAnalyticsSettings.OutputLocation
			configuration["post_call_analytics_settings"] = []interface{}{postCallSettings}
		}

		configuration["vocabulary_filter_method"] = processorConfiguration.VocabularyFilterMethod
		configuration["vocabulary_filter_name"] = processorConfiguration.VocabularyFilterName
		configuration["vocabulary_name"] = processorConfiguration.VocabularyName
		tfMap["amazon_transcribe_call_analytics_processor_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeAmazonTranscribeProcessor:
		processorConfiguration := apiElement.AmazonTranscribeProcessorConfiguration
		configuration["content_identification_type"] = processorConfiguration.ContentIdentificationType
		configuration["content_redaction_type"] = processorConfiguration.ContentRedactionType
		configuration["enable_partial_results_stabilization"] = processorConfiguration.EnablePartialResultsStabilization
		configuration["filter_partial_results"] = processorConfiguration.FilterPartialResults
		configuration[names.AttrLanguageCode] = processorConfiguration.LanguageCode
		configuration["language_model_name"] = processorConfiguration.LanguageModelName
		configuration["partial_results_stability"] = processorConfiguration.PartialResultsStability
		configuration["pii_entity_types"] = processorConfiguration.PiiEntityTypes
		configuration["show_speaker_label"] = processorConfiguration.ShowSpeakerLabel
		configuration["vocabulary_filter_method"] = processorConfiguration.VocabularyFilterMethod
		configuration["vocabulary_filter_name"] = processorConfiguration.VocabularyFilterName
		configuration["vocabulary_name"] = processorConfiguration.VocabularyName
		tfMap["amazon_transcribe_processor_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeKinesisDataStreamSink:
		processorConfiguration := apiElement.KinesisDataStreamSinkConfiguration
		configuration["insights_target"] = processorConfiguration.InsightsTarget
		tfMap["kinesis_data_stream_sink_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeLambdaFunctionSink:
		processorConfiguration := apiElement.LambdaFunctionSinkConfiguration
		configuration["insights_target"] = processorConfiguration.InsightsTarget
		tfMap["lambda_function_sink_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeSnsTopicSink:
		processorConfiguration := apiElement.SnsTopicSinkConfiguration
		configuration["insights_target"] = processorConfiguration.InsightsTarget
		tfMap["sns_topic_sink_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeSqsQueueSink:
		processorConfiguration := apiElement.SqsQueueSinkConfiguration
		configuration["insights_target"] = processorConfiguration.InsightsTarget
		tfMap["sqs_queue_sink_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeS3RecordingSink:
		processorConfiguration := apiElement.S3RecordingSinkConfiguration
		configuration[names.AttrDestination] = processorConfiguration.Destination
		tfMap["s3_recording_sink_configuration"] = []interface{}{configuration}
	case apiElement.Type == awstypes.MediaInsightsPipelineConfigurationElementTypeVoiceAnalyticsProcessor:
		processorConfiguration := apiElement.VoiceAnalyticsProcessorConfiguration
		configuration["speaker_search_status"] = processorConfiguration.SpeakerSearchStatus
		configuration["voice_tone_analysis_status"] = processorConfiguration.VoiceToneAnalysisStatus
		tfMap["voice_analytics_processor_configuration"] = []interface{}{configuration}
	}
	return tfMap
}

func flattenRealTimeAlertConfiguration(apiConfiguration *awstypes.RealTimeAlertConfiguration) []interface{} {
	if apiConfiguration == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	tfMap["disabled"] = apiConfiguration.Disabled
	if apiConfiguration.Rules == nil {
		return []interface{}{}
	}
	var tfRules []interface{}

	for _, apiRule := range apiConfiguration.Rules {
		if apiRule == (awstypes.RealTimeAlertRule{}) {
			continue
		}
		tfRules = append(tfRules, flattenRealTimeAlertRule(apiRule))
	}
	tfMap["rules"] = tfRules
	return []interface{}{tfMap}
}

func flattenRealTimeAlertRule(apiRule awstypes.RealTimeAlertRule) interface{} {
	if apiRule == (awstypes.RealTimeAlertRule{}) {
		return nil
	}
	tfMap := map[string]interface{}{}
	tfMap[names.AttrType] = string(apiRule.Type)

	configuration := map[string]interface{}{}

	switch {
	case apiRule.Type == awstypes.RealTimeAlertRuleTypeIssueDetection:
		issueDetectionConfiguration := apiRule.IssueDetectionConfiguration
		configuration["rule_name"] = issueDetectionConfiguration.RuleName
		tfMap["issue_detection_configuration"] = []interface{}{configuration}
	case apiRule.Type == awstypes.RealTimeAlertRuleTypeKeywordMatch:
		keywordMatchConfiguration := apiRule.KeywordMatchConfiguration
		configuration["rule_name"] = keywordMatchConfiguration.RuleName
		configuration["keywords"] = keywordMatchConfiguration.Keywords
		configuration["negate"] = keywordMatchConfiguration.Negate
		tfMap["keyword_match_configuration"] = []interface{}{configuration}
	case apiRule.Type == awstypes.RealTimeAlertRuleTypeSentiment:
		sentimentConfiguration := apiRule.SentimentConfiguration
		configuration["rule_name"] = sentimentConfiguration.RuleName
		configuration["sentiment_type"] = sentimentConfiguration.SentimentType
		configuration["time_period"] = sentimentConfiguration.TimePeriod
		tfMap["sentiment_configuration"] = []interface{}{configuration}
	}
	return tfMap
}
