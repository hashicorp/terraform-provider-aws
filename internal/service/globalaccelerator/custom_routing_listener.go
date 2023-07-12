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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_globalaccelerator_custom_routing_listener")
func ResourceCustomRoutingListener() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	acceleratorARN := d.Get("accelerator_arn").(string)
	input := &globalaccelerator.CreateCustomRoutingListenerInput{
		AcceleratorArn:   aws.String(acceleratorARN),
		IdempotencyToken: aws.String(id.UniqueId()),
		PortRanges:       expandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	output, err := conn.CreateCustomRoutingListenerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Custom Routing Listener: %s", err)
	}

	d.SetId(aws.StringValue(output.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return append(diags, resourceCustomRoutingListenerRead(ctx, d, meta)...)
}

func resourceCustomRoutingListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	listener, err := FindCustomRoutingListenerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Custom Routing Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Listener (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

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
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.UpdateCustomRoutingListenerInput{
		ListenerArn: aws.String(d.Id()),
		PortRanges:  expandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	_, err := conn.UpdateCustomRoutingListenerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Listener (%s): %s", d.Id(), err)
	}

	// Updating a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return append(diags, resourceCustomRoutingListenerRead(ctx, d, meta)...)
}

func resourceCustomRoutingListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	acceleratorARN := d.Get("accelerator_arn").(string)

	log.Printf("[DEBUG] Deleting Global Accelerator Custom Routing Listener (%s)", d.Id())
	_, err := conn.DeleteCustomRoutingListenerWithContext(ctx, &globalaccelerator.DeleteCustomRoutingListenerInput{
		ListenerArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeListenerNotFoundException) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Custom Routing Listener (%s): %s", d.Id(), err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return diags
}

func FindCustomRoutingListenerByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingListener, error) {
	input := &globalaccelerator.DescribeCustomRoutingListenerInput{
		ListenerArn: aws.String(arn),
	}

	return findCustomRoutingListener(ctx, conn, input)
}

func findCustomRoutingListener(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingListenerInput) (*globalaccelerator.CustomRoutingListener, error) {
	output, err := conn.DescribeCustomRoutingListenerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeListenerNotFoundException) {
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
