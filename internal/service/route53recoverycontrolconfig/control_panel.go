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

// @SDKResource("aws_route53recoverycontrolconfig_control_panel")
func ResourceControlPanel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceControlPanelCreate,
		ReadWithoutTimeout:   resourceControlPanelRead,
		UpdateWithoutTimeout: resourceControlPanelUpdate,
		DeleteWithoutTimeout: resourceControlPanelDelete,
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
			"default_control_panel": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_control_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceControlPanelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.CreateControlPanelInput{
		ClientToken:      aws.String(id.UniqueId()),
		ClusterArn:       aws.String(d.Get("cluster_arn").(string)),
		ControlPanelName: aws.String(d.Get(names.AttrName).(string)),
	}

	output, err := conn.CreateControlPanelWithContext(ctx, input)
	result := output.ControlPanel

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Control Panel: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Control Panel: empty response")
	}

	d.SetId(aws.StringValue(result.ControlPanelArn))

	if _, err := waitControlPanelCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Control Panel (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceControlPanelRead(ctx, d, meta)...)
}

func resourceControlPanelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.DescribeControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeControlPanelWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Control Panel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Control Panel: %s", err)
	}

	if output == nil || output.ControlPanel == nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Control Panel: %s", "empty response")
	}

	result := output.ControlPanel
	d.Set(names.AttrARN, result.ControlPanelArn)
	d.Set("cluster_arn", result.ClusterArn)
	d.Set("default_control_panel", result.DefaultControlPanel)
	d.Set(names.AttrName, result.Name)
	d.Set("routing_control_count", result.RoutingControlCount)
	d.Set(names.AttrStatus, result.Status)

	return diags
}

func resourceControlPanelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.UpdateControlPanelInput{
		ControlPanelName: aws.String(d.Get(names.AttrName).(string)),
		ControlPanelArn:  aws.String(d.Get(names.AttrARN).(string)),
	}

	_, err := conn.UpdateControlPanelWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Control Panel: %s", err)
	}

	return append(diags, resourceControlPanelRead(ctx, d, meta)...)
}

func resourceControlPanelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Control Panel: %s", d.Id())
	_, err := conn.DeleteControlPanelWithContext(ctx, &r53rcc.DeleteControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Control Panel: %s", err)
	}

	_, err = waitControlPanelDeleted(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Control Panel (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
