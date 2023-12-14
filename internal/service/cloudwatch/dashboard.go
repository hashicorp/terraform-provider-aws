// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_dashboard")
func ResourceDashboard() *schema.Resource {
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
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

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	dashboardName := d.Get("dashboard_name").(string)
	log.Printf("[DEBUG] Reading CloudWatch Dashboard: %s", dashboardName)
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	params := cloudwatch.GetDashboardInput{
		DashboardName: aws.String(d.Id()),
	}

	resp, err := conn.GetDashboardWithContext(ctx, &params)
	if !d.IsNewResource() && IsDashboardNotFoundErr(err) {
		create.LogNotFoundRemoveState(names.CloudWatch, create.ErrActionReading, ResNameDashboard, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CloudWatch, create.ErrActionReading, ResNameDashboard, d.Id(), err)
	}

	d.Set("dashboard_arn", resp.DashboardArn)
	d.Set("dashboard_name", resp.DashboardName)
	d.Set("dashboard_body", resp.DashboardBody)
	return diags
}

func resourceDashboardPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)
	params := cloudwatch.PutDashboardInput{
		DashboardBody: aws.String(d.Get("dashboard_body").(string)),
		DashboardName: aws.String(d.Get("dashboard_name").(string)),
	}

	log.Printf("[DEBUG] Putting CloudWatch Dashboard: %#v", params)

	_, err := conn.PutDashboardWithContext(ctx, &params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Putting dashboard failed: %s", err)
	}
	d.SetId(d.Get("dashboard_name").(string))
	log.Println("[INFO] CloudWatch Dashboard put finished")

	return append(diags, resourceDashboardRead(ctx, d, meta)...)
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[INFO] Deleting CloudWatch Dashboard %s", d.Id())
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)
	params := cloudwatch.DeleteDashboardsInput{
		DashboardNames: []*string{aws.String(d.Id())},
	}

	if _, err := conn.DeleteDashboardsWithContext(ctx, &params); err != nil {
		if IsDashboardNotFoundErr(err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Dashboard: %s", err)
	}
	log.Printf("[INFO] CloudWatch Dashboard %s deleted", d.Id())

	return diags
}

func IsDashboardNotFoundErr(err error) bool {
	return tfawserr.ErrMessageContains(
		err,
		"ResourceNotFound",
		"does not exist")
}
