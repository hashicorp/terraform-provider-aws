package globalaccelerator

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEndpointGroup() *schema.Resource {
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
			"arn": {
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
						"weight": {
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.HealthCheckProtocolTcp,
				ValidateFunc: validation.StringInSlice(globalaccelerator.HealthCheckProtocol_Values(), false),
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
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	input := &globalaccelerator.CreateEndpointGroupInput{
		EndpointGroupRegion: aws.String(meta.(*conns.AWSClient).Region),
		IdempotencyToken:    aws.String(resource.UniqueId()),
		ListenerArn:         aws.String(d.Get("listener_arn").(string)),
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.EndpointConfigurations = expandEndpointConfigurations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("endpoint_group_region"); ok {
		input.EndpointGroupRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_interval_seconds"); ok {
		input.HealthCheckIntervalSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_path"); ok {
		input.HealthCheckPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_port"); ok {
		input.HealthCheckPort = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_protocol"); ok {
		input.HealthCheckProtocol = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port_override"); ok && v.(*schema.Set).Len() > 0 {
		input.PortOverrides = expandPortOverrides(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("threshold_count"); ok {
		input.ThresholdCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		input.TrafficDialPercentage = aws.Float64(v)
	}

	resp, err := conn.CreateEndpointGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Global Accelerator Endpoint Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.EndpointGroup.EndpointGroupArn))

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return resourceEndpointGroupRead(ctx, d, meta)
}

func resourceEndpointGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	endpointGroup, err := FindEndpointGroupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator endpoint group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Global Accelerator Endpoint Group (%s): %s", d.Id(), err)
	}

	listenerARN, err := EndpointGroupARNToListenerARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("arn", endpointGroup.EndpointGroupArn)
	if err := d.Set("endpoint_configuration", flattenEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return diag.Errorf("setting endpoint_configuration: %s", err)
	}
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	d.Set("health_check_interval_seconds", endpointGroup.HealthCheckIntervalSeconds)
	d.Set("health_check_path", endpointGroup.HealthCheckPath)
	d.Set("health_check_port", endpointGroup.HealthCheckPort)
	d.Set("health_check_protocol", endpointGroup.HealthCheckProtocol)
	d.Set("listener_arn", listenerARN)
	if err := d.Set("port_override", flattenPortOverrides(endpointGroup.PortOverrides)); err != nil {
		return diag.Errorf("setting port_override: %s", err)
	}
	d.Set("threshold_count", endpointGroup.ThresholdCount)
	d.Set("traffic_dial_percentage", endpointGroup.TrafficDialPercentage)

	return nil
}

func resourceEndpointGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	input := &globalaccelerator.UpdateEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.EndpointConfigurations = expandEndpointConfigurations(v.(*schema.Set).List())
	} else {
		input.EndpointConfigurations = []*globalaccelerator.EndpointConfiguration{}
	}

	if v, ok := d.GetOk("health_check_interval_seconds"); ok {
		input.HealthCheckIntervalSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_path"); ok {
		input.HealthCheckPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_port"); ok {
		input.HealthCheckPort = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_protocol"); ok {
		input.HealthCheckProtocol = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port_override"); ok && v.(*schema.Set).Len() > 0 {
		input.PortOverrides = expandPortOverrides(v.(*schema.Set).List())
	} else {
		input.PortOverrides = []*globalaccelerator.PortOverride{}
	}

	if v, ok := d.GetOk("threshold_count"); ok {
		input.ThresholdCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		input.TrafficDialPercentage = aws.Float64(v)
	}

	_, err := conn.UpdateEndpointGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Global Accelerator Endpoint Group (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return resourceEndpointGroupRead(ctx, d, meta)
}

func resourceEndpointGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	log.Printf("[DEBUG] Deleting Global Accelerator Endpoint Group: %s", d.Id())
	_, err := conn.DeleteEndpointGroupWithContext(ctx, &globalaccelerator.DeleteEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Global Accelerator Endpoint Group (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return nil
}

func expandEndpointConfiguration(tfMap map[string]interface{}) *globalaccelerator.EndpointConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.EndpointConfiguration{}

	if v, ok := tfMap["client_ip_preservation_enabled"].(bool); ok {
		apiObject.ClientIPPreservationEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["endpoint_id"].(string); ok && v != "" {
		apiObject.EndpointId = aws.String(v)
	}

	if v, ok := tfMap["weight"].(int); ok && v != 0 {
		apiObject.Weight = aws.Int64(int64(v))
	}

	return apiObject
}

func expandEndpointConfigurations(tfList []interface{}) []*globalaccelerator.EndpointConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*globalaccelerator.EndpointConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEndpointConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPortOverride(tfMap map[string]interface{}) *globalaccelerator.PortOverride {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.PortOverride{}

	if v, ok := tfMap["endpoint_port"].(int); ok && v != 0 {
		apiObject.EndpointPort = aws.Int64(int64(v))
	}

	if v, ok := tfMap["listener_port"].(int); ok && v != 0 {
		apiObject.ListenerPort = aws.Int64(int64(v))
	}

	return apiObject
}

func expandPortOverrides(tfList []interface{}) []*globalaccelerator.PortOverride {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*globalaccelerator.PortOverride

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPortOverride(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEndpointDescription(apiObject *globalaccelerator.EndpointDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ClientIPPreservationEnabled; v != nil {
		tfMap["client_ip_preservation_enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.EndpointId; v != nil {
		tfMap["endpoint_id"] = aws.StringValue(v)
	}

	if v := apiObject.Weight; v != nil {
		tfMap["weight"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenEndpointDescriptions(apiObjects []*globalaccelerator.EndpointDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEndpointDescription(apiObject))
	}

	return tfList
}

func flattenPortOverride(apiObject *globalaccelerator.PortOverride) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EndpointPort; v != nil {
		tfMap["endpoint_port"] = aws.Int64Value(v)
	}

	if v := apiObject.ListenerPort; v != nil {
		tfMap["listener_port"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenPortOverrides(apiObjects []*globalaccelerator.PortOverride) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPortOverride(apiObject))
	}

	return tfList
}
