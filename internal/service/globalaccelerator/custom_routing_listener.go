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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_globalaccelerator_custom_routing_listener", name="Custom Routing Listener")
func resourceCustomRoutingListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomRoutingListenerCreate,
		ReadWithoutTimeout:   resourceCustomRoutingListenerRead,
		UpdateWithoutTimeout: resourceCustomRoutingListenerUpdate,
		DeleteWithoutTimeout: resourceCustomRoutingListenerDelete,

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
		},
	}
}

func resourceCustomRoutingListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	acceleratorARN := d.Get("accelerator_arn").(string)
	input := &globalaccelerator.CreateCustomRoutingListenerInput{
		AcceleratorArn:   aws.String(acceleratorARN),
		IdempotencyToken: aws.String(id.UniqueId()),
		PortRanges:       expandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	output, err := conn.CreateCustomRoutingListener(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Custom Routing Listener: %s", err)
	}

	d.SetId(aws.ToString(output.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return append(diags, resourceCustomRoutingListenerRead(ctx, d, meta)...)
}

func resourceCustomRoutingListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	listener, err := findCustomRoutingListenerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Custom Routing Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Listener (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := listenerOrEndpointGroupARNToAcceleratorARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("accelerator_arn", acceleratorARN)
	if err := d.Set("port_range", flattenPortRanges(listener.PortRanges)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting port_range: %s", err)
	}

	return diags
}

func resourceCustomRoutingListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	acceleratorARN := d.Get("accelerator_arn").(string)
	input := &globalaccelerator.UpdateCustomRoutingListenerInput{
		ListenerArn: aws.String(d.Id()),
		PortRanges:  expandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	_, err := conn.UpdateCustomRoutingListener(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Listener (%s): %s", d.Id(), err)
	}

	// Updating a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return append(diags, resourceCustomRoutingListenerRead(ctx, d, meta)...)
}

func resourceCustomRoutingListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	log.Printf("[DEBUG] Deleting Global Accelerator Custom Routing Listener (%s)", d.Id())
	_, err := conn.DeleteCustomRoutingListener(ctx, &globalaccelerator.DeleteCustomRoutingListenerInput{
		ListenerArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ListenerNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Custom Routing Listener (%s): %s", d.Id(), err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending.
	acceleratorARN := d.Get("accelerator_arn").(string)
	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", acceleratorARN, err)
	}

	return diags
}

func findCustomRoutingListenerByARN(ctx context.Context, conn *globalaccelerator.Client, arn string) (*awstypes.CustomRoutingListener, error) {
	input := &globalaccelerator.DescribeCustomRoutingListenerInput{
		ListenerArn: aws.String(arn),
	}

	output, err := conn.DescribeCustomRoutingListener(ctx, input)

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
