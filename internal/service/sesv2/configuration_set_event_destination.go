// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_configuration_set_event_destination", name="Configuration Set Event Destination")
func resourceConfigurationSetEventDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationSetEventDestinationCreate,
		ReadWithoutTimeout:   resourceConfigurationSetEventDestinationRead,
		UpdateWithoutTimeout: resourceConfigurationSetEventDestinationUpdate,
		DeleteWithoutTimeout: resourceConfigurationSetEventDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"configuration_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"event_destination": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_watch_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"event_destination.0.cloud_watch_destination",
								"event_destination.0.event_bridge_destination",
								"event_destination.0.kinesis_firehose_destination",
								"event_destination.0.pinpoint_destination",
								"event_destination.0.sns_destination",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dimension_configuration": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_dimension_value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 256),
												},
												"dimension_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 256),
												},
												"dimension_value_source": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.DimensionValueSource](),
												},
											},
										},
									},
								},
							},
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"event_bridge_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"event_destination.0.cloud_watch_destination",
								"event_destination.0.event_bridge_destination",
								"event_destination.0.kinesis_firehose_destination",
								"event_destination.0.pinpoint_destination",
								"event_destination.0.sns_destination",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_bus_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"kinesis_firehose_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"event_destination.0.cloud_watch_destination",
								"event_destination.0.event_bridge_destination",
								"event_destination.0.kinesis_firehose_destination",
								"event_destination.0.pinpoint_destination",
								"event_destination.0.sns_destination",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delivery_stream_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrIAMRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"matching_event_types": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.EventType](),
							},
						},
						"pinpoint_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"event_destination.0.cloud_watch_destination",
								"event_destination.0.event_bridge_destination",
								"event_destination.0.kinesis_firehose_destination",
								"event_destination.0.pinpoint_destination",
								"event_destination.0.sns_destination",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"application_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"sns_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"event_destination.0.cloud_watch_destination",
								"event_destination.0.event_bridge_destination",
								"event_destination.0.kinesis_firehose_destination",
								"event_destination.0.pinpoint_destination",
								"event_destination.0.sns_destination",
							},
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
				},
			},
			"event_destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	resNameConfigurationSetEventDestination = "Configuration Set Event Destination"
)

func resourceConfigurationSetEventDestinationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		EventDestination:     expandEventDestinationDefinition(d.Get("event_destination").([]any)[0].(map[string]any)),
		EventDestinationName: aws.String(d.Get("event_destination_name").(string)),
	}

	configurationSetEventDestinationID := configurationSetEventDestinationCreateResourceID(d.Get("configuration_set_name").(string), d.Get("event_destination_name").(string))

	out, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateConfigurationSetEventDestination(ctx, in)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeBadRequestException, "Could not access Kinesis Firehose Stream") ||
				tfawserr.ErrMessageContains(err, errCodeBadRequestException, "Could not assume IAM role") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameConfigurationSetEventDestination, configurationSetEventDestinationID, err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameConfigurationSetEventDestination, configurationSetEventDestinationID, errors.New("empty output"))
	}

	d.SetId(configurationSetEventDestinationID)

	return append(diags, resourceConfigurationSetEventDestinationRead(ctx, d, meta)...)
}

func resourceConfigurationSetEventDestinationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	configurationSetName, eventDestinationName, err := configurationSetEventDestinationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findConfigurationSetEventDestinationByTwoPartKey(ctx, conn, configurationSetName, eventDestinationName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 ConfigurationSetEventDestination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameConfigurationSetEventDestination, d.Id(), err)
	}

	d.Set("configuration_set_name", configurationSetName)
	if err := d.Set("event_destination", []any{flattenEventDestination(out)}); err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, resNameConfigurationSetEventDestination, d.Id(), err)
	}
	d.Set("event_destination_name", out.Name)

	return diags
}

func resourceConfigurationSetEventDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	configurationSetName, eventDestinationName, err := configurationSetEventDestinationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("event_destination") {
		in := &sesv2.UpdateConfigurationSetEventDestinationInput{
			ConfigurationSetName: aws.String(configurationSetName),
			EventDestination:     expandEventDestinationDefinition(d.Get("event_destination").([]any)[0].(map[string]any)),
			EventDestinationName: aws.String(eventDestinationName),
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSetEventDestination (%s): %#v", d.Id(), in)
		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (any, error) {
				return conn.UpdateConfigurationSetEventDestination(ctx, in)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeBadRequestException, "Could not access Kinesis Firehose Stream") ||
					tfawserr.ErrMessageContains(err, errCodeBadRequestException, "Could not assume IAM role") {
					return true, err
				}

				return false, err
			},
		)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameConfigurationSetEventDestination, d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationSetEventDestinationRead(ctx, d, meta)...)
}

func resourceConfigurationSetEventDestinationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	configurationSetName, eventDestinationName, err := configurationSetEventDestinationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting SESV2 ConfigurationSetEventDestination: %s", d.Id())
	_, err = conn.DeleteConfigurationSetEventDestination(ctx, &sesv2.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(configurationSetName),
		EventDestinationName: aws.String(eventDestinationName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameConfigurationSetEventDestination, d.Id(), err)
	}

	return diags
}

const configurationSetEventDestinationResourceIDSeparator = "|"

func configurationSetEventDestinationCreateResourceID(configurationSetName, eventDestinationName string) string {
	parts := []string{configurationSetName, eventDestinationName}
	id := strings.Join(parts, configurationSetEventDestinationResourceIDSeparator)

	return id
}

func configurationSetEventDestinationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, configurationSetEventDestinationResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected CONFIGURATION_SET_NAME%[2]sEVENT_DESTINATION_NAME", id, configurationSetEventDestinationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findConfigurationSetEventDestinationByTwoPartKey(ctx context.Context, conn *sesv2.Client, configurationSetName, eventDestinationName string) (*types.EventDestination, error) {
	input := &sesv2.GetConfigurationSetEventDestinationsInput{
		ConfigurationSetName: aws.String(configurationSetName),
	}

	return findConfigurationSetEventDestination(ctx, conn, input, func(v *types.EventDestination) bool {
		return aws.ToString(v.Name) == eventDestinationName
	})
}

func findConfigurationSetEventDestination(ctx context.Context, conn *sesv2.Client, input *sesv2.GetConfigurationSetEventDestinationsInput, filter tfslices.Predicate[*types.EventDestination]) (*types.EventDestination, error) {
	output, err := findConfigurationSetEventDestinations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurationSetEventDestinations(ctx context.Context, conn *sesv2.Client, input *sesv2.GetConfigurationSetEventDestinationsInput, filter tfslices.Predicate[*types.EventDestination]) ([]types.EventDestination, error) {
	output, err := conn.GetConfigurationSetEventDestinations(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.EventDestinations, tfslices.PredicateValue(filter)), nil
}

func flattenEventDestination(apiObject *types.EventDestination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.CloudWatchDestination; v != nil {
		tfMap["cloud_watch_destination"] = []any{flattenCloudWatchDestination(v)}
	}

	if v := apiObject.EventBridgeDestination; v != nil {
		tfMap["event_bridge_destination"] = []any{flattenEventBridgeDestination(v)}
	}

	if v := apiObject.KinesisFirehoseDestination; v != nil {
		tfMap["kinesis_firehose_destination"] = []any{flattenKinesisFirehoseDestination(v)}
	}

	if v := apiObject.MatchingEventTypes; v != nil {
		tfMap["matching_event_types"] = enum.Slice(apiObject.MatchingEventTypes...)
	}

	if v := apiObject.PinpointDestination; v != nil {
		tfMap["pinpoint_destination"] = []any{flattenPinpointDestination(v)}
	}

	if v := apiObject.SnsDestination; v != nil {
		tfMap["sns_destination"] = []any{flattenSNSDestination(v)}
	}

	return tfMap
}

func flattenCloudWatchDestination(apiObject *types.CloudWatchDestination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DimensionConfigurations; v != nil {
		tfMap["dimension_configuration"] = flattenCloudWatchDimensionConfigurations(v)
	}

	return tfMap
}

func flattenEventBridgeDestination(apiObject *types.EventBridgeDestination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.EventBusArn; v != nil {
		tfMap["event_bus_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenKinesisFirehoseDestination(apiObject *types.KinesisFirehoseDestination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DeliveryStreamArn; v != nil {
		tfMap["delivery_stream_arn"] = aws.ToString(v)
	}

	if v := apiObject.IamRoleArn; v != nil {
		tfMap[names.AttrIAMRoleARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenPinpointDestination(apiObject *types.PinpointDestination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ApplicationArn; v != nil {
		tfMap["application_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenSNSDestination(apiObject *types.SnsDestination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.TopicArn; v != nil {
		tfMap[names.AttrTopicARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenCloudWatchDimensionConfigurations(apiObjects []types.CloudWatchDimensionConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCloudWatchDimensionConfiguration(&apiObject))
	}

	return tfList
}

func flattenCloudWatchDimensionConfiguration(apiObject *types.CloudWatchDimensionConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"dimension_value_source": string(apiObject.DimensionValueSource),
	}

	if v := apiObject.DefaultDimensionValue; v != nil {
		tfMap["default_dimension_value"] = aws.ToString(v)
	}

	if v := apiObject.DimensionName; v != nil {
		tfMap["dimension_name"] = aws.ToString(v)
	}

	return tfMap
}

func expandEventDestinationDefinition(tfMap map[string]any) *types.EventDestinationDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EventDestinationDefinition{}

	if v, ok := tfMap["cloud_watch_destination"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchDestination = expandCloudWatchDestination(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}

	if v, ok := tfMap["event_bridge_destination"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EventBridgeDestination = expandEventBridgeDestination(v[0].(map[string]any))
	}

	if v, ok := tfMap["kinesis_firehose_destination"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.KinesisFirehoseDestination = expandKinesisFirehoseDestination(v[0].(map[string]any))
	}

	if v, ok := tfMap["matching_event_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchingEventTypes = flex.ExpandStringyValueSet[types.EventType](v)
	}

	if v, ok := tfMap["pinpoint_destination"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.PinpointDestination = expandPinpointDestinaton(v[0].(map[string]any))
	}

	if v, ok := tfMap["sns_destination"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SnsDestination = expandSNSDestination(v[0].(map[string]any))
	}

	return apiObject
}

func expandCloudWatchDestination(tfMap map[string]any) *types.CloudWatchDestination {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CloudWatchDestination{}

	if v, ok := tfMap["dimension_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DimensionConfigurations = expandCloudWatchDimensionConfigurations(v)
	}

	return apiObject
}

func expandEventBridgeDestination(tfMap map[string]any) *types.EventBridgeDestination {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EventBridgeDestination{}

	if v, ok := tfMap["event_bus_arn"].(string); ok && v != "" {
		apiObject.EventBusArn = aws.String(v)
	}

	return apiObject
}

func expandKinesisFirehoseDestination(tfMap map[string]any) *types.KinesisFirehoseDestination {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.KinesisFirehoseDestination{}

	if v, ok := tfMap["delivery_stream_arn"].(string); ok && v != "" {
		apiObject.DeliveryStreamArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrIAMRoleARN].(string); ok && v != "" {
		apiObject.IamRoleArn = aws.String(v)
	}

	return apiObject
}

func expandPinpointDestinaton(tfMap map[string]any) *types.PinpointDestination {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PinpointDestination{}

	if v, ok := tfMap["application_arn"].(string); ok && v != "" {
		apiObject.ApplicationArn = aws.String(v)
	}

	return apiObject
}

func expandSNSDestination(tfMap map[string]any) *types.SnsDestination {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SnsDestination{}

	if v, ok := tfMap[names.AttrTopicARN].(string); ok && v != "" {
		apiObject.TopicArn = aws.String(v)
	}

	return apiObject
}

func expandCloudWatchDimensionConfigurations(tfList []any) []types.CloudWatchDimensionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.CloudWatchDimensionConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandCloudWatchDimensionConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCloudWatchDimensionConfiguration(tfMap map[string]any) *types.CloudWatchDimensionConfiguration {
	apiObject := &types.CloudWatchDimensionConfiguration{}

	if v, ok := tfMap["default_dimension_value"].(string); ok && v != "" {
		apiObject.DefaultDimensionValue = aws.String(v)
	}

	if v, ok := tfMap["dimension_name"].(string); ok && v != "" {
		apiObject.DimensionName = aws.String(v)
	}

	if v, ok := tfMap["dimension_value_source"].(string); ok && v != "" {
		apiObject.DimensionValueSource = types.DimensionValueSource(v)
	}

	return apiObject
}
