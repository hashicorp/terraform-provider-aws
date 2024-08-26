// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rum

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rum/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rum_metrics_destination", name="Metrics Destination")
func resourceMetricsDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMetricsDestinationPut,
		ReadWithoutTimeout:   resourceMetricsDestinationRead,
		UpdateWithoutTimeout: resourceMetricsDestinationPut,
		DeleteWithoutTimeout: resourceMetricsDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_monitor_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDestination: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MetricDestination](),
			},
			names.AttrDestinationARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrIAMRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceMetricsDestinationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	name := d.Get("app_monitor_name").(string)
	input := &rum.PutRumMetricsDestinationInput{
		AppMonitorName: aws.String(name),
		Destination:    awstypes.MetricDestination(d.Get(names.AttrDestination).(string)),
	}

	if v, ok := d.GetOk(names.AttrDestinationARN); ok {
		input.DestinationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIAMRoleARN); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	_, err := conn.PutRumMetricsDestination(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch RUM Metrics Destination (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceMetricsDestinationRead(ctx, d, meta)...)
}

func resourceMetricsDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	dest, err := findMetricsDestinationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch RUM Metrics Destination %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch RUM Metrics Destination (%s): %s", d.Id(), err)
	}

	d.Set("app_monitor_name", d.Id())
	d.Set(names.AttrDestination, dest.Destination)
	d.Set(names.AttrDestinationARN, dest.DestinationArn)
	d.Set(names.AttrIAMRoleARN, dest.IamRoleArn)

	return diags
}

func resourceMetricsDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	input := &rum.DeleteRumMetricsDestinationInput{
		AppMonitorName: aws.String(d.Id()),
		Destination:    awstypes.MetricDestination(d.Get(names.AttrDestination).(string)),
	}

	if v, ok := d.GetOk(names.AttrDestinationARN); ok {
		input.DestinationArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting CloudWatch RUM Metrics Destination: %s", d.Id())
	_, err := conn.DeleteRumMetricsDestination(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch RUM Metrics Destination (%s): %s", d.Id(), err)
	}

	return diags
}

func findMetricsDestination(ctx context.Context, conn *rum.Client, input *rum.ListRumMetricsDestinationsInput) (*awstypes.MetricDestinationSummary, error) {
	output, err := findMetricsDestinations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMetricsDestinations(ctx context.Context, conn *rum.Client, input *rum.ListRumMetricsDestinationsInput) ([]awstypes.MetricDestinationSummary, error) {
	var output []awstypes.MetricDestinationSummary

	pages := rum.NewListRumMetricsDestinationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Destinations...)
	}

	return output, nil
}

func findMetricsDestinationByName(ctx context.Context, conn *rum.Client, name string) (*awstypes.MetricDestinationSummary, error) {
	input := &rum.ListRumMetricsDestinationsInput{
		AppMonitorName: aws.String(name),
	}

	return findMetricsDestination(ctx, conn, input)
}
