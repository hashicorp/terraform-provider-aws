// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudwatch

import (
	"context"
	"log"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_dashboard", name="Dashboard")
func resourceDashboard() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDashboardPut,
		ReadWithoutTimeout:   resourceDashboardRead,
		UpdateWithoutTimeout: resourceDashboardPut,
		DeleteWithoutTimeout: resourceDashboardDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// Note that we specify both the `dashboard_body` and
		// the `dashboard_name` as being required, even though
		// according to the REST API documentation both are
		// optional: http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutDashboard.html#API_PutDashboard_RequestParameters
		Schema: map[string]*schema.Schema{
			"dashboard_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard_body": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"dashboard_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validDashboardName,
			},
		},
	}
}

func resourceDashboardPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	name := d.Get("dashboard_name").(string)
	input := &cloudwatch.PutDashboardInput{
		DashboardBody: aws.String(d.Get("dashboard_body").(string)),
		DashboardName: aws.String(name),
	}

	_, err := conn.PutDashboard(ctx, input)

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, name)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return smerr.AppendEnrich(ctx, diags, resourceDashboardRead(ctx, d, meta))
}

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	output, err := findDashboardByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		smerr.AppendOne(ctx, diags, sdkdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	d.Set("dashboard_arn", output.DashboardArn)
	d.Set("dashboard_body", output.DashboardBody)
	d.Set("dashboard_name", output.DashboardName)

	return diags
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Dashboard: %s", d.Id())
	input := cloudwatch.DeleteDashboardsInput{
		DashboardNames: []string{d.Id()},
	}
	_, err := conn.DeleteDashboards(ctx, &input)

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return diags
}

func findDashboardByName(ctx context.Context, conn *cloudwatch.Client, name string) (*cloudwatch.GetDashboardOutput, error) {
	input := &cloudwatch.GetDashboardInput{
		DashboardName: aws.String(name),
	}

	output, err := conn.GetDashboard(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFound) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}
