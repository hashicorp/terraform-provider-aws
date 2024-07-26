// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_globalaccelerator_endpoint_group", name="Endpoint Group")
func resourceEndpointGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointGroupCreate,
		ReadWithoutTimeout:   resourceEndpointGroupRead,
		UpdateWithoutTimeout: resourceEndpointGroupUpdate,
		DeleteWithoutTimeout: resourceEndpointGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_ip_preservation_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"endpoint_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						names.AttrWeight: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 255),
						},
					},
				},
			},
			"endpoint_group_region": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidRegionName,
			},
			"health_check_interval_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(10, 30),
			},
			"health_check_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"health_check_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"health_check_protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.HealthCheckProtocolTcp,
				ValidateDiagFunc: enum.Validate[awstypes.HealthCheckProtocol](),
			},
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"port_override": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"listener_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"threshold_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(1, 10),
			},
			"traffic_dial_percentage": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      100.0,
				ValidateFunc: validation.FloatBetween(0.0, 100.0),
			},
		},
	}
}

func resourceEndpointGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	input := &globalaccelerator.CreateEndpointGroupInput{
		EndpointGroupRegion: aws.String(meta.(*conns.AWSClient).Region),
		IdempotencyToken:    aws.String(id.UniqueId()),
		ListenerArn:         aws.String(d.Get("listener_arn").(string)),
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.EndpointConfigurations = expandEndpointConfigurations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("endpoint_group_region"); ok {
		input.EndpointGroupRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_interval_seconds"); ok {
		input.HealthCheckIntervalSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("health_check_path"); ok {
		input.HealthCheckPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_port"); ok {
		input.HealthCheckPort = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("health_check_protocol"); ok {
		input.HealthCheckProtocol = awstypes.HealthCheckProtocol(v.(string))
	}

	if v, ok := d.GetOk("port_override"); ok && v.(*schema.Set).Len() > 0 {
		input.PortOverrides = expandPortOverrides(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("threshold_count"); ok {
		input.ThresholdCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		input.TrafficDialPercentage = aws.Float32(float32(v))
	}

	output, err := conn.CreateEndpointGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Endpoint Group: %s", err)
	}

	d.SetId(aws.ToString(output.EndpointGroup.EndpointGroupArn))

	acceleratorARN, err := listenerOrEndpointGroupARNToAcceleratorARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return append(diags, resourceEndpointGroupRead(ctx, d, meta)...)
}

func resourceEndpointGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	endpointGroup, err := findEndpointGroupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator endpoint group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Endpoint Group (%s): %s", d.Id(), err)
	}

	listenerARN, err := endpointGroupARNToListenerARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, endpointGroup.EndpointGroupArn)
	if err := d.Set("endpoint_configuration", flattenEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	d.Set("health_check_interval_seconds", endpointGroup.HealthCheckIntervalSeconds)
	d.Set("health_check_path", endpointGroup.HealthCheckPath)
	d.Set("health_check_port", endpointGroup.HealthCheckPort)
	d.Set("health_check_protocol", endpointGroup.HealthCheckProtocol)
	d.Set("listener_arn", listenerARN)
	if err := d.Set("port_override", flattenPortOverrides(endpointGroup.PortOverrides)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting port_override: %s", err)
	}
	d.Set("threshold_count", endpointGroup.ThresholdCount)
	d.Set("traffic_dial_percentage", endpointGroup.TrafficDialPercentage)

	return diags
}

func resourceEndpointGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	input := &globalaccelerator.UpdateEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.EndpointConfigurations = expandEndpointConfigurations(v.(*schema.Set).List())
	} else {
		input.EndpointConfigurations = []awstypes.EndpointConfiguration{}
	}

	if v, ok := d.GetOk("health_check_interval_seconds"); ok {
		input.HealthCheckIntervalSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("health_check_path"); ok {
		input.HealthCheckPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_port"); ok {
		input.HealthCheckPort = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("health_check_protocol"); ok {
		input.HealthCheckProtocol = awstypes.HealthCheckProtocol(v.(string))
	}

	if v, ok := d.GetOk("port_override"); ok && v.(*schema.Set).Len() > 0 {
		input.PortOverrides = expandPortOverrides(v.(*schema.Set).List())
	} else {
		input.PortOverrides = []awstypes.PortOverride{}
	}

	if v, ok := d.GetOk("threshold_count"); ok {
		input.ThresholdCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		input.TrafficDialPercentage = aws.Float32(float32(v))
	}

	_, err := conn.UpdateEndpointGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Endpoint Group (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := listenerOrEndpointGroupARNToAcceleratorARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return append(diags, resourceEndpointGroupRead(ctx, d, meta)...)
}

func resourceEndpointGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	log.Printf("[DEBUG] Deleting Global Accelerator Endpoint Group: %s", d.Id())
	_, err := conn.DeleteEndpointGroup(ctx, &globalaccelerator.DeleteEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.EndpointGroupNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Endpoint Group (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := listenerOrEndpointGroupARNToAcceleratorARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return diags
}

func findEndpointGroupByARN(ctx context.Context, conn *globalaccelerator.Client, arn string) (*awstypes.EndpointGroup, error) {
	input := &globalaccelerator.DescribeEndpointGroupInput{
		EndpointGroupArn: aws.String(arn),
	}

	output, err := conn.DescribeEndpointGroup(ctx, input)

	if errs.IsA[*awstypes.EndpointGroupNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.EndpointGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EndpointGroup, nil
}

func expandEndpointConfiguration(tfMap map[string]interface{}) *awstypes.EndpointConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EndpointConfiguration{}

	if v, ok := tfMap["client_ip_preservation_enabled"].(bool); ok {
		apiObject.ClientIPPreservationEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["endpoint_id"].(string); ok && v != "" {
		apiObject.EndpointId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrWeight].(int); ok {
		apiObject.Weight = aws.Int32(int32(v))
	}

	return apiObject
}

func expandEndpointConfigurations(tfList []interface{}) []awstypes.EndpointConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.EndpointConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEndpointConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPortOverride(tfMap map[string]interface{}) *awstypes.PortOverride {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PortOverride{}

	if v, ok := tfMap["endpoint_port"].(int); ok && v != 0 {
		apiObject.EndpointPort = aws.Int32(int32(v))
	}

	if v, ok := tfMap["listener_port"].(int); ok && v != 0 {
		apiObject.ListenerPort = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPortOverrides(tfList []interface{}) []awstypes.PortOverride {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PortOverride

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPortOverride(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenEndpointDescription(apiObject *awstypes.EndpointDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ClientIPPreservationEnabled; v != nil {
		tfMap["client_ip_preservation_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.EndpointId; v != nil {
		tfMap["endpoint_id"] = aws.ToString(v)
	}

	if v := apiObject.Weight; v != nil {
		tfMap[names.AttrWeight] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenEndpointDescriptions(apiObjects []awstypes.EndpointDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEndpointDescription(&apiObject))
	}

	return tfList
}

func flattenPortOverride(apiObject *awstypes.PortOverride) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EndpointPort; v != nil {
		tfMap["endpoint_port"] = aws.ToInt32(v)
	}

	if v := apiObject.ListenerPort; v != nil {
		tfMap["listener_port"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenPortOverrides(apiObjects []awstypes.PortOverride) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPortOverride(&apiObject))
	}

	return tfList
}
