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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_globalaccelerator_listener", name="Listener")
func resourceListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCreate,
		ReadWithoutTimeout:   resourceListenerRead,
		UpdateWithoutTimeout: resourceListenerUpdate,
		DeleteWithoutTimeout: resourceListenerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"accelerator_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_affinity": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ClientAffinityNone,
				ValidateDiagFunc: enum.Validate[awstypes.ClientAffinity](),
			},
			"port_range": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			names.AttrProtocol: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Protocol](),
			},
		},
	}
}

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	acceleratorARN := d.Get("accelerator_arn").(string)
	input := &globalaccelerator.CreateListenerInput{
		AcceleratorArn:   aws.String(acceleratorARN),
		ClientAffinity:   awstypes.ClientAffinity(d.Get("client_affinity").(string)),
		IdempotencyToken: aws.String(id.UniqueId()),
		PortRanges:       expandPortRanges(d.Get("port_range").(*schema.Set).List()),
		Protocol:         awstypes.Protocol(d.Get(names.AttrProtocol).(string)),
	}

	output, err := conn.CreateListener(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Listener: %s", err)
	}

	d.SetId(aws.ToString(output.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	listener, err := findListenerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Listener (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := listenerOrEndpointGroupARNToAcceleratorARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("accelerator_arn", acceleratorARN)
	d.Set("client_affinity", listener.ClientAffinity)
	if err := d.Set("port_range", flattenPortRanges(listener.PortRanges)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting port_range: %s", err)
	}
	d.Set(names.AttrProtocol, listener.Protocol)

	return diags
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	acceleratorARN := d.Get("accelerator_arn").(string)
	input := &globalaccelerator.UpdateListenerInput{
		ClientAffinity: awstypes.ClientAffinity(d.Get("client_affinity").(string)),
		ListenerArn:    aws.String(d.Id()),
		PortRanges:     expandPortRanges(d.Get("port_range").(*schema.Set).List()),
		Protocol:       awstypes.Protocol(d.Get(names.AttrProtocol).(string)),
	}

	_, err := conn.UpdateListener(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Listener (%s): %s", d.Id(), err)
	}

	// Updating a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	log.Printf("[DEBUG] Deleting Global Accelerator Listener: %s", d.Id())
	_, err := conn.DeleteListener(ctx, &globalaccelerator.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ListenerNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Listener (%s): %s", d.Id(), err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending.
	acceleratorARN := d.Get("accelerator_arn").(string)
	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return diags
}

func findListenerByARN(ctx context.Context, conn *globalaccelerator.Client, arn string) (*awstypes.Listener, error) {
	input := &globalaccelerator.DescribeListenerInput{
		ListenerArn: aws.String(arn),
	}

	output, err := conn.DescribeListener(ctx, input)

	if errs.IsA[*awstypes.ListenerNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Listener == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Listener, nil
}

func expandPortRange(tfMap map[string]interface{}) *awstypes.PortRange {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PortRange{}

	if v, ok := tfMap["from_port"].(int); ok && v != 0 {
		apiObject.FromPort = aws.Int32(int32(v))
	}

	if v, ok := tfMap["to_port"].(int); ok && v != 0 {
		apiObject.ToPort = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPortRanges(tfList []interface{}) []awstypes.PortRange {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PortRange

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPortRange(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenPortRange(apiObject *awstypes.PortRange) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FromPort; v != nil {
		tfMap["from_port"] = aws.ToInt32(v)
	}

	if v := apiObject.ToPort; v != nil {
		tfMap["to_port"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenPortRanges(apiObjects []awstypes.PortRange) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPortRange(&apiObject))
	}

	return tfList
}
