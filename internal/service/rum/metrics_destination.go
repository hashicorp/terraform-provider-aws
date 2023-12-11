// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rum

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_rum_metrics_destination")
func ResourceMetricsDestination() *schema.Resource {
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
			"destination": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudwatchrum.MetricDestination_Values(), false),
			},
			"destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceMetricsDestinationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RUMConn(ctx)

	name := d.Get("app_monitor_name").(string)
	input := &cloudwatchrum.PutRumMetricsDestinationInput{
		AppMonitorName: aws.String(name),
		Destination:    aws.String(d.Get("destination").(string)),
	}

	if v, ok := d.GetOk("destination_arn"); ok {
		input.DestinationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	_, err := conn.PutRumMetricsDestinationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("putting CloudWatch RUM Metrics Destination (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return resourceMetricsDestinationRead(ctx, d, meta)
}

func resourceMetricsDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RUMConn(ctx)

	dest, err := FindMetricsDestinationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch RUM Metrics Destination %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch RUM Metrics Destination (%s): %s", d.Id(), err)
	}

	d.Set("app_monitor_name", d.Id())
	d.Set("destination", dest.Destination)
	d.Set("destination_arn", dest.DestinationArn)
	d.Set("iam_role_arn", dest.IamRoleArn)

	return nil
}

func resourceMetricsDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RUMConn(ctx)

	input := &cloudwatchrum.DeleteRumMetricsDestinationInput{
		AppMonitorName: aws.String(d.Id()),
		Destination:    aws.String(d.Get("destination").(string)),
	}

	if v, ok := d.GetOk("destination_arn"); ok {
		input.DestinationArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting CloudWatch RUM Metrics Destination: %s", d.Id())
	_, err := conn.DeleteRumMetricsDestinationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch RUM Metrics Destination (%s): %s", d.Id(), err)
	}

	return nil
}

func FindMetricsDestinationByName(ctx context.Context, conn *cloudwatchrum.CloudWatchRUM, name string) (*cloudwatchrum.MetricDestinationSummary, error) {
	input := &cloudwatchrum.ListRumMetricsDestinationsInput{
		AppMonitorName: aws.String(name),
	}
	var output []*cloudwatchrum.MetricDestinationSummary

	err := conn.ListRumMetricsDestinationsPagesWithContext(ctx, input, func(page *cloudwatchrum.ListRumMetricsDestinationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Destinations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}
