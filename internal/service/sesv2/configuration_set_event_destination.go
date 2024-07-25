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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sesv2_configuration_set_event_destination")
func ResourceConfigurationSetEventDestination() *schema.Resource {
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
							Type:     schema.TypeList,
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
	ResNameConfigurationSetEventDestination = "Configuration Set Event Destination"
)

func resourceConfigurationSetEventDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(d.Get("configuration_set_name").(string)),
		EventDestination:     expandEventDestination(d.Get("event_destination").([]interface{})[0].(map[string]interface{})),
		EventDestinationName: aws.String(d.Get("event_destination_name").(string)),
	}

	configurationSetEventDestinationID := FormatConfigurationSetEventDestinationID(d.Get("configuration_set_name").(string), d.Get("event_destination_name").(string))

	out, err := conn.CreateConfigurationSetEventDestination(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameConfigurationSetEventDestination, configurationSetEventDestinationID, err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, ResNameConfigurationSetEventDestination, configurationSetEventDestinationID, errors.New("empty output"))
	}

	d.SetId(configurationSetEventDestinationID)

	return append(diags, resourceConfigurationSetEventDestinationRead(ctx, d, meta)...)
}

func resourceConfigurationSetEventDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	configurationSetName, _, err := ParseConfigurationSetEventDestinationID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameConfigurationSetEventDestination, d.Id(), err)
	}

	out, err := FindConfigurationSetEventDestinationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 ConfigurationSetEventDestination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameConfigurationSetEventDestination, d.Id(), err)
	}

	d.Set("configuration_set_name", configurationSetName)
	d.Set("event_destination_name", out.Name)

	if err := d.Set("event_destination", []interface{}{flattenEventDestination(out)}); err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameConfigurationSetEventDestination, d.Id(), err)
	}

	return diags
}

func resourceConfigurationSetEventDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	configurationSetName, eventDestinationName, err := ParseConfigurationSetEventDestinationID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSetEventDestination, d.Id(), err)
	}

	if d.HasChanges("event_destination") {
		in := &sesv2.UpdateConfigurationSetEventDestinationInput{
			ConfigurationSetName: aws.String(configurationSetName),
			EventDestination:     expandEventDestination(d.Get("event_destination").([]interface{})[0].(map[string]interface{})),
			EventDestinationName: aws.String(eventDestinationName),
		}

		log.Printf("[DEBUG] Updating SESV2 ConfigurationSetEventDestination (%s): %#v", d.Id(), in)
		_, err := conn.UpdateConfigurationSetEventDestination(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, ResNameConfigurationSetEventDestination, d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationSetEventDestinationRead(ctx, d, meta)...)
}

func resourceConfigurationSetEventDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 ConfigurationSetEventDestination %s", d.Id())

	configurationSetName, eventDestinationName, err := ParseConfigurationSetEventDestinationID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameConfigurationSetEventDestination, d.Id(), err)
	}

	_, err = conn.DeleteConfigurationSetEventDestination(ctx, &sesv2.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(configurationSetName),
		EventDestinationName: aws.String(eventDestinationName),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, ResNameConfigurationSetEventDestination, d.Id(), err)
	}

	return diags
}

func FindConfigurationSetEventDestinationByID(ctx context.Context, conn *sesv2.Client, id string) (types.EventDestination, error) {
	configurationSetName, eventDestinationName, err := ParseConfigurationSetEventDestinationID(id)
	if err != nil {
		return types.EventDestination{}, err
	}

	in := &sesv2.GetConfigurationSetEventDestinationsInput{
		ConfigurationSetName: aws.String(configurationSetName),
	}
	out, err := conn.GetConfigurationSetEventDestinations(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return types.EventDestination{}, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return types.EventDestination{}, err
	}

	if out == nil {
		return types.EventDestination{}, tfresource.NewEmptyResultError(in)
	}

	for _, eventDestination := range out.EventDestinations {
		if aws.ToString(eventDestination.Name) == eventDestinationName {
			return eventDestination, nil
		}
	}

	return types.EventDestination{}, &retry.NotFoundError{}
}

func flattenEventDestination(apiObject types.EventDestination) map[string]interface{} {
	m := map[string]interface{}{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.CloudWatchDestination; v != nil {
		m["cloud_watch_destination"] = []interface{}{flattenCloudWatchDestination(v)}
	}

	if v := apiObject.EventBridgeDestination; v != nil {
		m["event_bridge_destination"] = []interface{}{flattenEventBridgeDestination(v)}
	}

	if v := apiObject.KinesisFirehoseDestination; v != nil {
		m["kinesis_firehose_destination"] = []interface{}{flattenKinesisFirehoseDestination(v)}
	}

	if v := apiObject.MatchingEventTypes; v != nil {
		m["matching_event_types"] = enum.Slice(apiObject.MatchingEventTypes...)
	}

	if v := apiObject.PinpointDestination; v != nil {
		m["pinpoint_destination"] = []interface{}{flattenPinpointDestination(v)}
	}

	if v := apiObject.SnsDestination; v != nil {
		m["sns_destination"] = []interface{}{flattenSNSDestination(v)}
	}

	return m
}

func flattenCloudWatchDestination(apiObject *types.CloudWatchDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.DimensionConfigurations; v != nil {
		m["dimension_configuration"] = flattenCloudWatchDimensionConfigurations(v)
	}

	return m
}

func flattenEventBridgeDestination(apiObject *types.EventBridgeDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.EventBusArn; v != nil {
		m["event_bus_arn"] = aws.ToString(v)
	}

	return m
}

func flattenKinesisFirehoseDestination(apiObject *types.KinesisFirehoseDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.DeliveryStreamArn; v != nil {
		m["delivery_stream_arn"] = aws.ToString(v)
	}

	if v := apiObject.IamRoleArn; v != nil {
		m[names.AttrIAMRoleARN] = aws.ToString(v)
	}

	return m
}

func flattenPinpointDestination(apiObject *types.PinpointDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.ApplicationArn; v != nil {
		m["application_arn"] = aws.ToString(v)
	}

	return m
}

func flattenSNSDestination(apiObject *types.SnsDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.TopicArn; v != nil {
		m[names.AttrTopicARN] = aws.ToString(v)
	}

	return m
}

func flattenCloudWatchDimensionConfigurations(apiObjects []types.CloudWatchDimensionConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenCloudWatchDimensionConfiguration(apiObject))
	}

	return l
}

func flattenCloudWatchDimensionConfiguration(apiObject types.CloudWatchDimensionConfiguration) map[string]interface{} {
	m := map[string]interface{}{
		"dimension_value_source": string(apiObject.DimensionValueSource),
	}

	if v := apiObject.DefaultDimensionValue; v != nil {
		m["default_dimension_value"] = aws.ToString(v)
	}

	if v := apiObject.DimensionName; v != nil {
		m["dimension_name"] = aws.ToString(v)
	}

	return m
}

func expandEventDestination(tfMap map[string]interface{}) *types.EventDestinationDefinition {
	if tfMap == nil {
		return nil
	}

	a := &types.EventDestinationDefinition{}

	if v, ok := tfMap["cloud_watch_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.CloudWatchDestination = expandCloudWatchDestination(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		a.Enabled = v
	}

	if v, ok := tfMap["event_bridge_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.EventBridgeDestination = expandEventBridgeDestinaton(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["kinesis_firehose_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.KinesisFirehoseDestination = expandKinesisFirehoseDestination(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["matching_event_types"].([]interface{}); ok && len(v) > 0 {
		a.MatchingEventTypes = stringsToEventTypes(flex.ExpandStringList(v))
	}

	if v, ok := tfMap["pinpoint_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.PinpointDestination = expandPinpointDestinaton(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sns_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SnsDestination = expandSNSDestination(v[0].(map[string]interface{}))
	}

	return a
}

func expandCloudWatchDestination(tfMap map[string]interface{}) *types.CloudWatchDestination {
	if tfMap == nil {
		return nil
	}

	a := &types.CloudWatchDestination{}

	if v, ok := tfMap["dimension_configuration"].([]interface{}); ok && len(v) > 0 {
		a.DimensionConfigurations = expandCloudWatchDimensionConfigurations(v)
	}

	return a
}

func expandEventBridgeDestinaton(tfMap map[string]interface{}) *types.EventBridgeDestination {
	if tfMap == nil {
		return nil
	}

	a := &types.EventBridgeDestination{}

	if v, ok := tfMap["event_bus_arn"].(string); ok && v != "" {
		a.EventBusArn = aws.String(v)
	}

	return a
}

func expandKinesisFirehoseDestination(tfMap map[string]interface{}) *types.KinesisFirehoseDestination {
	if tfMap == nil {
		return nil
	}

	a := &types.KinesisFirehoseDestination{}

	if v, ok := tfMap["delivery_stream_arn"].(string); ok && v != "" {
		a.DeliveryStreamArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrIAMRoleARN].(string); ok && v != "" {
		a.IamRoleArn = aws.String(v)
	}

	return a
}

func expandPinpointDestinaton(tfMap map[string]interface{}) *types.PinpointDestination {
	if tfMap == nil {
		return nil
	}

	a := &types.PinpointDestination{}

	if v, ok := tfMap["application_arn"].(string); ok && v != "" {
		a.ApplicationArn = aws.String(v)
	}

	return a
}

func expandSNSDestination(tfMap map[string]interface{}) *types.SnsDestination {
	if tfMap == nil {
		return nil
	}

	a := &types.SnsDestination{}

	if v, ok := tfMap[names.AttrTopicARN].(string); ok && v != "" {
		a.TopicArn = aws.String(v)
	}

	return a
}

func expandCloudWatchDimensionConfigurations(tfList []interface{}) []types.CloudWatchDimensionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.CloudWatchDimensionConfiguration

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		s = append(s, expandCloudWatchDimensionConfiguration(m))
	}

	return s
}

func expandCloudWatchDimensionConfiguration(tfMap map[string]interface{}) types.CloudWatchDimensionConfiguration {
	a := types.CloudWatchDimensionConfiguration{}

	if v, ok := tfMap["default_dimension_value"].(string); ok && v != "" {
		a.DefaultDimensionValue = aws.String(v)
	}

	if v, ok := tfMap["dimension_name"].(string); ok && v != "" {
		a.DimensionName = aws.String(v)
	}

	if v, ok := tfMap["dimension_value_source"].(string); ok && v != "" {
		a.DimensionValueSource = types.DimensionValueSource(v)
	}

	return a
}

func FormatConfigurationSetEventDestinationID(configurationSetName, eventDestinationName string) string {
	return fmt.Sprintf("%s|%s", configurationSetName, eventDestinationName)
}

func ParseConfigurationSetEventDestinationID(id string) (string, string, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return "", "", errors.New("please make sure the ID is in the form CONFIGURATION_SET_NAME|EVENT_DESTINATION_NAME")
	}

	return idParts[0], idParts[1], nil
}

func stringsToEventTypes(values []*string) []types.EventType {
	var eventTypes []types.EventType

	for _, eventType := range values {
		eventTypes = append(eventTypes, types.EventType(aws.ToString(eventType)))
	}

	return eventTypes
}
