// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
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

	name := d.Get("service_arn").(string)
	input := &apprunner.StartDeploymentInput{
		ServiceArn: aws.String(name),
	}

	output, err := conn.StartDeployment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting App Runner Deployment (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.OperationId))

	return append(diags, resourceStartDeploymentRead(ctx, d, meta)...)
}

func resourceStartDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	input := &apprunner.ListOperationsInput{
		ServiceArn: aws.String(d.Id()),
	}

	output, err := conn.ListOperations(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner Deployment (%s): %s", d.Id(), err)
	}

	if len(output.OperationSummaryList) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading App Runner Deployment (%s): not found", d.Id())
	}

	var operation types.OperationSummary
	for _, op := range output.OperationSummaryList {
		if aws.ToString(op.Id) == d.Id() {
			operation = op
			break
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Summary: %s", err)
	}

	d.Set("operation_id", operation.Id)
	d.Set("valid_until", aws.ToTime(operation.StartedAt).Format(time.RFC3339))
	d.Set("status", operation.Status)

	return diags
}

func waitStartDeploymentSucceeded(ctx context.Context, conn *apprunner.Client, arn string) (*types.OperationSummary, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(types.OperationStatusSucceeded),
		Refresh: statusObservabilityConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OperationSummary); ok {
		return output, err
	}

	return nil, err
}
