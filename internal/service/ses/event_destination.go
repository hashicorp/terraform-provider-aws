// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_event_destination", name="Configuration Set Event Destination")
func resourceEventDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventDestinationCreate,
		ReadWithoutTimeout:   resourceEventDestinationRead,
		DeleteWithoutTimeout: resourceEventDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceEventDestinationImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_destination": {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"kinesis_destination", "sns_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDefaultValue: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.@-]+$`), "must contain only alphanumeric, underscore, hyphen, period, and at signs characters"),
							),
						},
						"dimension_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_:-]+$`), "must contain only alphanumeric, underscore, and hyphen characters"),
							),
						},
						"value_source": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DimensionValueSource](),
						},
					},
				},
			},
			"configuration_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"kinesis_destination": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				MaxItems:      1,
				ConflictsWith: []string{"cloudwatch_destination", "sns_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrStreamARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"matching_types": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.EventType](),
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric, underscore, and hyphen characters"),
				),
			},
			"sns_destination": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cloudwatch_destination", "kinesis_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
		},
	}
}

func resourceEventDestinationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	eventDestinationName := d.Get(names.AttrName).(string)
	input := &ses.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		EventDestination: &awstypes.EventDestination{
			Enabled:            d.Get(names.AttrEnabled).(bool),
			MatchingEventTypes: flex.ExpandStringyValueSet[awstypes.EventType](d.Get("matching_types").(*schema.Set)),
			Name:               aws.String(eventDestinationName),
		},
	}

	if v, ok := d.GetOk("cloudwatch_destination"); ok && v.(*schema.Set).Len() > 0 {
		input.EventDestination.CloudWatchDestination = &awstypes.CloudWatchDestination{
			DimensionConfigurations: expandCloudWatchDimensionConfigurations(v.(*schema.Set).List()),
		}
	}

	if v, ok := d.GetOk("kinesis_destination"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)
		input.EventDestination.KinesisFirehoseDestination = &awstypes.KinesisFirehoseDestination{
			DeliveryStreamARN: aws.String(tfMap[names.AttrStreamARN].(string)),
			IAMRoleARN:        aws.String(tfMap[names.AttrRoleARN].(string)),
		}
	}

	if v, ok := d.GetOk("sns_destination"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)
		input.EventDestination.SNSDestination = &awstypes.SNSDestination{
			TopicARN: aws.String(tfMap[names.AttrTopicARN].(string)),
		}
	}

	_, err := conn.CreateConfigurationSetEventDestination(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Configuration Set Event Destination (%s): %s", eventDestinationName, err)
	}

	d.SetId(eventDestinationName)

	return append(diags, resourceEventDestinationRead(ctx, d, meta)...)
}

func resourceEventDestinationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	configurationSetName := d.Get("configuration_set_name").(string)
	eventDestination, err := findEventDestinationByTwoPartKey(ctx, conn, configurationSetName, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Configuration Set Event Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Configuration Set Event Destination (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("configuration-set/%s:event-destination/%s", configurationSetName, d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if err := d.Set("cloudwatch_destination", flattenCloudWatchDestination(eventDestination.CloudWatchDestination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_destination: %s", err)
	}
	d.Set("configuration_set_name", configurationSetName)
	d.Set(names.AttrEnabled, eventDestination.Enabled)
	if err := d.Set("kinesis_destination", flattenKinesisFirehoseDestination(eventDestination.KinesisFirehoseDestination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kinesis_destination: %s", err)
	}
	if err := d.Set("matching_types", eventDestination.MatchingEventTypes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting matching_types: %s", err)
	}
	d.Set(names.AttrName, eventDestination.Name)
	if err := d.Set("sns_destination", flattenSNSDestination(eventDestination.SNSDestination)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sns_destination: %s", err)
	}

	return diags
}

func resourceEventDestinationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting Configuration Set Event Destination: %s", d.Id())
	_, err := conn.DeleteConfigurationSetEventDestination(ctx, &ses.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		EventDestinationName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ConfigurationSetDoesNotExistException](err) || errs.IsA[*awstypes.EventDestinationDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Configuration Set Event Destination (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEventDestinationImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'configuration-set-name/event-destination-name'", d.Id())
	}

	configurationSetName := parts[0]
	eventDestinationName := parts[1]

	d.SetId(eventDestinationName)
	d.Set("configuration_set_name", configurationSetName)

	return []*schema.ResourceData{d}, nil
}

func findEventDestinationByTwoPartKey(ctx context.Context, conn *ses.Client, configurationSetName, eventDestinationName string) (*awstypes.EventDestination, error) {
	input := &ses.DescribeConfigurationSetInput{
		ConfigurationSetAttributeNames: []awstypes.ConfigurationSetAttribute{awstypes.ConfigurationSetAttributeEventDestinations},
		ConfigurationSetName:           aws.String(configurationSetName),
	}

	return findEventDestination(ctx, conn, input, func(v *awstypes.EventDestination) bool {
		return aws.ToString(v.Name) == eventDestinationName
	})
}

func findEventDestination(ctx context.Context, conn *ses.Client, input *ses.DescribeConfigurationSetInput, filter tfslices.Predicate[*awstypes.EventDestination]) (*awstypes.EventDestination, error) {
	output, err := findEventDestinations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEventDestinations(ctx context.Context, conn *ses.Client, input *ses.DescribeConfigurationSetInput, filter tfslices.Predicate[*awstypes.EventDestination]) ([]awstypes.EventDestination, error) {
	output, err := findConfigurationSet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfslices.Filter(output.EventDestinations, tfslices.PredicateValue(filter)), nil
}

func expandCloudWatchDimensionConfigurations(tfList []any) []awstypes.CloudWatchDimensionConfiguration {
	apiObjects := make([]awstypes.CloudWatchDimensionConfiguration, 0)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.CloudWatchDimensionConfiguration{
			DefaultDimensionValue: aws.String(tfMap[names.AttrDefaultValue].(string)),
			DimensionName:         aws.String(tfMap["dimension_name"].(string)),
			DimensionValueSource:  awstypes.DimensionValueSource(tfMap["value_source"].(string)),
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCloudWatchDestination(apiObject *awstypes.CloudWatchDestination) []any {
	if apiObject == nil {
		return []any{}
	}

	tfList := []any{}

	for _, apiObject := range apiObject.DimensionConfigurations {
		tfMap := map[string]any{
			names.AttrDefaultValue: aws.ToString(apiObject.DefaultDimensionValue),
			"dimension_name":       aws.ToString(apiObject.DimensionName),
			"value_source":         apiObject.DimensionValueSource,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenKinesisFirehoseDestination(apiObject *awstypes.KinesisFirehoseDestination) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrRoleARN:   aws.ToString(apiObject.IAMRoleARN),
		names.AttrStreamARN: aws.ToString(apiObject.DeliveryStreamARN),
	}

	return []any{tfMap}
}

func flattenSNSDestination(apiObject *awstypes.SNSDestination) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrTopicARN: aws.ToString(apiObject.TopicARN),
	}

	return []any{tfMap}
}
