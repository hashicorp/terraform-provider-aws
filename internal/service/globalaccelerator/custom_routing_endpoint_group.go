// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_globalaccelerator_custom_routing_endpoint_group")
func ResourceCustomRoutingEndpointGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomRoutingEndpointGroupCreate,
		ReadWithoutTimeout:   resourceCustomRoutingEndpointGroupRead,
		DeleteWithoutTimeout: resourceCustomRoutingEndpointGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_configuration": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"protocols": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(globalaccelerator.CustomRoutingProtocol_Values(), false),
							},
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"endpoint_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
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
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceCustomRoutingEndpointGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	input := &globalaccelerator.CreateCustomRoutingEndpointGroupInput{
		DestinationConfigurations: expandCustomRoutingDestinationConfigurations(d.Get("destination_configuration").(*schema.Set).List()),
		EndpointGroupRegion:       aws.String(meta.(*conns.AWSClient).Region),
		IdempotencyToken:          aws.String(id.UniqueId()),
		ListenerArn:               aws.String(d.Get("listener_arn").(string)),
	}

	if v, ok := d.GetOk("endpoint_group_region"); ok {
		input.EndpointGroupRegion = aws.String(v.(string))
	}

	output, err := conn.CreateCustomRoutingEndpointGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Custom Routing Endpoint Group: %s", err)
	}

	d.SetId(aws.StringValue(output.EndpointGroup.EndpointGroupArn))

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		input := &globalaccelerator.AddCustomRoutingEndpointsInput{
			EndpointGroupArn:       aws.String(d.Id()),
			EndpointConfigurations: expandCustomRoutingEndpointConfigurations(v.(*schema.Set).List()),
		}

		_, err := conn.AddCustomRoutingEndpointsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding Global Accelerator Custom Routing Endpoint Group (%s) endpoints: %s", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", acceleratorARN, err)
		}
	}

	return append(diags, resourceCustomRoutingEndpointGroupRead(ctx, d, meta)...)
}

func resourceCustomRoutingEndpointGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	endpointGroup, err := FindCustomRoutingEndpointGroupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Custom Routing Endpoint Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Endpoint Group (%s): %s", d.Id(), err)
	}

	listenerARN, err := EndpointGroupARNToListenerARN(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("arn", endpointGroup.EndpointGroupArn)
	if err := d.Set("destination_configuration", flattenCustomRoutingDestinationDescriptions(endpointGroup.DestinationDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting destination_configuration: %s", err)
	}
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	if err := d.Set("endpoint_configuration", flattenCustomRoutingEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	d.Set("listener_arn", listenerARN)

	return diags
}

func resourceCustomRoutingEndpointGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	log.Printf("[DEBUG] Deleting Global Accelerator Custom Routing Endpoint Group (%s)", d.Id())
	_, err := conn.DeleteCustomRoutingEndpointGroupWithContext(ctx, &globalaccelerator.DeleteCustomRoutingEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Custom Routing Endpoint Group (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return diags
}

func FindCustomRoutingEndpointGroupByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingEndpointGroup, error) {
	input := &globalaccelerator.DescribeCustomRoutingEndpointGroupInput{
		EndpointGroupArn: aws.String(arn),
	}

	return findCustomRoutingEndpointGroup(ctx, conn, input)
}

func findCustomRoutingEndpointGroup(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingEndpointGroupInput) (*globalaccelerator.CustomRoutingEndpointGroup, error) {
	output, err := conn.DescribeCustomRoutingEndpointGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException) {
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

func expandCustomRoutingEndpointDestinationConfiguration(tfMap map[string]interface{}) *globalaccelerator.CustomRoutingDestinationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.CustomRoutingDestinationConfiguration{}

	if v, ok := tfMap["from_port"].(int); ok && v != 0 {
		apiObject.FromPort = aws.Int64(int64(v))
	}

	if v, ok := tfMap["protocols"].(*schema.Set); ok {
		apiObject.Protocols = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["to_port"].(int); ok && v != 0 {
		apiObject.ToPort = aws.Int64(int64(v))
	}

	return apiObject
}

func expandCustomRoutingDestinationConfigurations(tfList []interface{}) []*globalaccelerator.CustomRoutingDestinationConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*globalaccelerator.CustomRoutingDestinationConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCustomRoutingEndpointDestinationConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCustomRoutingEndpointConfiguration(tfMap map[string]interface{}) *globalaccelerator.CustomRoutingEndpointConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.CustomRoutingEndpointConfiguration{}

	if v, ok := tfMap["endpoint_id"].(string); ok && v != "" {
		apiObject.EndpointId = aws.String(v)
	}

	return apiObject
}

func expandCustomRoutingEndpointConfigurations(tfList []interface{}) []*globalaccelerator.CustomRoutingEndpointConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*globalaccelerator.CustomRoutingEndpointConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCustomRoutingEndpointConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCustomRoutingDestinationDescription(apiObject *globalaccelerator.CustomRoutingDestinationDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FromPort; v != nil {
		tfMap["from_port"] = aws.Int64Value(v)
	}

	if v := apiObject.Protocols; v != nil {
		tfMap["protocols"] = aws.StringValueSlice(v)
	}

	if v := apiObject.ToPort; v != nil {
		tfMap["to_port"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenCustomRoutingDestinationDescriptions(apiObjects []*globalaccelerator.CustomRoutingDestinationDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCustomRoutingDestinationDescription(apiObject))
	}

	return tfList
}

func flattenCustomRoutingEndpointDescription(apiObject *globalaccelerator.CustomRoutingEndpointDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EndpointId; v != nil {
		tfMap["endpoint_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenCustomRoutingEndpointDescriptions(apiObjects []*globalaccelerator.CustomRoutingEndpointDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCustomRoutingEndpointDescription(apiObject))
	}

	return tfList
}
