// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoverycontrolconfig_routing_control", name="Routing Control")
func resourceRoutingControl() *schema.Resource {
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

func resourceRoutingControlCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	input := &r53rcc.CreateRoutingControlInput{
		ClientToken:        aws.String(id.UniqueId()),
		ClusterArn:         aws.String(d.Get("cluster_arn").(string)),
		RoutingControlName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("control_panel_arn"); ok {
		input.ControlPanelArn = aws.String(v.(string))
	}

	output, err := conn.CreateRoutingControl(ctx, input)
	result := output.RoutingControl

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Routing Control: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Routing Control: empty response")
	}

	d.SetId(aws.ToString(result.RoutingControlArn))

	if _, err := waitRoutingControlCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Routing Control (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceRoutingControlRead(ctx, d, meta)...)
}

func resourceRoutingControlRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	output, err := findRoutingControlByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Control Config Routing Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Routing Control: %s", err)
	}

	d.Set(names.AttrARN, output.RoutingControlArn)
	d.Set("control_panel_arn", output.ControlPanelArn)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrStatus, output.Status)

	return diags
}

func resourceRoutingControlUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	input := &r53rcc.UpdateRoutingControlInput{
		RoutingControlName: aws.String(d.Get(names.AttrName).(string)),
		RoutingControlArn:  aws.String(d.Get(names.AttrARN).(string)),
	}

	_, err := conn.UpdateRoutingControl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Routing Control: %s", err)
	}

	return append(diags, resourceRoutingControlRead(ctx, d, meta)...)
}

func resourceRoutingControlDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Routing Control: %s", d.Id())
	_, err := conn.DeleteRoutingControl(ctx, &r53rcc.DeleteRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Routing Control: %s", err)
	}

	_, err = waitRoutingControlDeleted(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Routing Control (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findRoutingControlByARN(ctx context.Context, conn *r53rcc.Client, arn string) (*awstypes.RoutingControl, error) {
	input := &r53rcc.DescribeRoutingControlInput{
		RoutingControlArn: aws.String(arn),
	}

	output, err := conn.DescribeRoutingControl(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.RoutingControl == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RoutingControl, nil
}
