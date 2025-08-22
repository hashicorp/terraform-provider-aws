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

// @SDKResource("aws_route53recoverycontrolconfig_control_panel", name="Control Panel")
func resourceControlPanel() *schema.Resource {
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

func resourceControlPanelCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	input := &r53rcc.CreateControlPanelInput{
		ClientToken:      aws.String(id.UniqueId()),
		ClusterArn:       aws.String(d.Get("cluster_arn").(string)),
		ControlPanelName: aws.String(d.Get(names.AttrName).(string)),
	}

	output, err := conn.CreateControlPanel(ctx, input)
	result := output.ControlPanel

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Control Panel: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Control Panel: empty response")
	}

	d.SetId(aws.ToString(result.ControlPanelArn))

	if _, err := waitControlPanelCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Control Panel (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceControlPanelRead(ctx, d, meta)...)
}

func resourceControlPanelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	output, err := findControlPanelByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Control Config Control Panel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Control Panel: %s", err)
	}

	d.Set(names.AttrARN, output.ControlPanelArn)
	d.Set("cluster_arn", output.ClusterArn)
	d.Set("default_control_panel", output.DefaultControlPanel)
	d.Set(names.AttrName, output.Name)
	d.Set("routing_control_count", output.RoutingControlCount)
	d.Set(names.AttrStatus, output.Status)

	return diags
}

func resourceControlPanelUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	input := &r53rcc.UpdateControlPanelInput{
		ControlPanelName: aws.String(d.Get(names.AttrName).(string)),
		ControlPanelArn:  aws.String(d.Get(names.AttrARN).(string)),
	}

	_, err := conn.UpdateControlPanel(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Control Panel: %s", err)
	}

	return append(diags, resourceControlPanelRead(ctx, d, meta)...)
}

func resourceControlPanelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Control Panel: %s", d.Id())
	_, err := conn.DeleteControlPanel(ctx, &r53rcc.DeleteControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Control Panel: %s", err)
	}

	_, err = waitControlPanelDeleted(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Control Panel (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findControlPanelByARN(ctx context.Context, conn *r53rcc.Client, arn string) (*awstypes.ControlPanel, error) {
	input := &r53rcc.DescribeControlPanelInput{
		ControlPanelArn: aws.String(arn),
	}

	output, err := conn.DescribeControlPanel(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.ControlPanel == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ControlPanel, nil
}
