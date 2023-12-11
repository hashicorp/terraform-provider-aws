// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_apprunner_start_deployment", name="Start Deployment")
func resourceStartDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStartDeploymentCreate,
		ReadWithoutTimeout:   resourceStartDeploymentRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"service_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"operation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"started_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ended_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStartDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	service_arn := d.Get("service_arn").(string)
	input := &apprunner.StartDeploymentInput{
		ServiceArn: aws.String(service_arn),
	}

	output, err := conn.StartDeployment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "initiating App Runner Start Deployment Operation (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(output.OperationId))

	if _, err := waitStartDeploymentSucceeded(ctx, conn, service_arn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Start Deployment Operation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStartDeploymentRead(ctx, d, meta)...)
}

func resourceStartDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	output, err := findStartDeploymentOperationByServiceARN(ctx, meta.(*conns.AWSClient).AppRunnerClient(ctx), d.Get("service_arn").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner Start Deployment Operation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding App Runner Start Deployment Operation (%s): %s", d.Id(), err)
	}

	d.Set("operation_id", output.Id)
	d.Set("started_at", aws.ToTime(output.StartedAt).Format(time.RFC3339))
	d.Set("ended_at", aws.ToTime(output.EndedAt).Format(time.RFC3339))
	d.Set("status", output.Status)

	return diags
}

func waitStartDeploymentSucceeded(ctx context.Context, conn *apprunner.Client, arn string) (*types.OperationSummary, error) {
	const (
		timeout = 15 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(types.OperationStatusSucceeded),
		Refresh: statusStartDeployment(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OperationSummary); ok {
		return output, err
	}

	return nil, err
}

func statusStartDeployment(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStartDeploymentOperationByServiceARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findStartDeploymentOperationByServiceARN(ctx context.Context, conn *apprunner.Client, arn string) (*types.OperationSummary, error) {
	input := &apprunner.ListOperationsInput{
		ServiceArn: aws.String(arn),
	}

	output, err := conn.ListOperations(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(output.OperationSummaryList) == 0 {
		return nil, &retry.NotFoundError{
			Message:     "start deployment operation not found",
			LastRequest: input,
		}
	}

	var operation types.OperationSummary
	var found bool
	for _, op := range output.OperationSummaryList {
		if aws.ToString(op.TargetArn) == arn {
			operation = op
			found = true
			break
		}
	}

	if !found {
		return nil, &retry.NotFoundError{
			Message:     "start deployment operation not found",
			LastRequest: input,
		}
	}

	return &operation, nil
}
