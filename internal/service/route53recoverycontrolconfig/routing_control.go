// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoverycontrolconfig_routing_control")
func ResourceRoutingControl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoutingControlCreate,
		ReadWithoutTimeout:   resourceRoutingControlRead,
		UpdateWithoutTimeout: resourceRoutingControlUpdate,
		DeleteWithoutTimeout: resourceRoutingControlDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"control_panel_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRoutingControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.CreateRoutingControlInput{
		ClientToken:        aws.String(id.UniqueId()),
		ClusterArn:         aws.String(d.Get("cluster_arn").(string)),
		RoutingControlName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("control_panel_arn"); ok {
		input.ControlPanelArn = aws.String(v.(string))
	}

	output, err := conn.CreateRoutingControlWithContext(ctx, input)
	result := output.RoutingControl

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Routing Control: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Routing Control: empty response")
	}

	d.SetId(aws.StringValue(result.RoutingControlArn))

	if _, err := waitRoutingControlCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Routing Control (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceRoutingControlRead(ctx, d, meta)...)
}

func resourceRoutingControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.DescribeRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeRoutingControlWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Routing Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Routing Control: %s", err)
	}

	if output == nil || output.RoutingControl == nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Routing Control: %s", "empty response")
	}

	result := output.RoutingControl
	d.Set(names.AttrARN, result.RoutingControlArn)
	d.Set("control_panel_arn", result.ControlPanelArn)
	d.Set(names.AttrName, result.Name)
	d.Set(names.AttrStatus, result.Status)

	return diags
}

func resourceRoutingControlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.UpdateRoutingControlInput{
		RoutingControlName: aws.String(d.Get(names.AttrName).(string)),
		RoutingControlArn:  aws.String(d.Get(names.AttrARN).(string)),
	}

	_, err := conn.UpdateRoutingControlWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Routing Control: %s", err)
	}

	return append(diags, resourceRoutingControlRead(ctx, d, meta)...)
}

func resourceRoutingControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Routing Control: %s", d.Id())
	_, err := conn.DeleteRoutingControlWithContext(ctx, &r53rcc.DeleteRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Routing Control: %s", err)
	}

	_, err = waitRoutingControlDeleted(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Routing Control (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
