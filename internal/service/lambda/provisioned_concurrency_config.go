// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lambda_provisioned_concurrency_config", name="Provisioned Concurrency Config")
func resourceProvisionedConcurrencyConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProvisionedConcurrencyConfigCreate,
		ReadWithoutTimeout:   resourceProvisionedConcurrencyConfigRead,
		UpdateWithoutTimeout: resourceProvisionedConcurrencyConfigUpdate,
		DeleteWithoutTimeout: resourceProvisionedConcurrencyConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceProvisionedConcurrencyConfigV0().CoreConfigSchema().ImpliedType(),
				Upgrade: provisionedConcurrencyConfigStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"provisioned_concurrent_executions": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"qualifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

const (
	provisionedConcurrencyConfigResourceIDPartCount = 2
)

func resourceProvisionedConcurrencyConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	id := errs.Must(flex.FlattenResourceId([]string{functionName, qualifier}, provisionedConcurrencyConfigResourceIDPartCount, true))
	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName:                    aws.String(functionName),
		ProvisionedConcurrentExecutions: aws.Int32(int32(d.Get("provisioned_concurrent_executions").(int))),
		Qualifier:                       aws.String(qualifier),
	}

	_, err := conn.PutProvisionedConcurrencyConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Provisioned Concurrency Config (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitProvisionedConcurrencyConfigReady(ctx, conn, functionName, qualifier, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Provisioned Concurrency Config (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProvisionedConcurrencyConfigRead(ctx, d, meta)...)
}

func resourceProvisionedConcurrencyConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), provisionedConcurrencyConfigResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	functionName, qualifier := parts[0], parts[1]

	output, err := findProvisionedConcurrencyConfigByTwoPartKey(ctx, conn, functionName, qualifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Provisioned Concurrency Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Provisioned Concurrency Config (%s): %s", d.Id(), err)
	}

	d.Set("function_name", functionName)
	d.Set("provisioned_concurrent_executions", output.AllocatedProvisionedConcurrentExecutions)
	d.Set("qualifier", qualifier)

	return diags
}

func resourceProvisionedConcurrencyConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), provisionedConcurrencyConfigResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	functionName, qualifier := parts[0], parts[1]

	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName:                    aws.String(functionName),
		ProvisionedConcurrentExecutions: aws.Int32(int32(d.Get("provisioned_concurrent_executions").(int))),
		Qualifier:                       aws.String(qualifier),
	}

	_, err = conn.PutProvisionedConcurrencyConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda Provisioned Concurrency Config (%s): %s", d.Id(), err)
	}

	if _, err := waitProvisionedConcurrencyConfigReady(ctx, conn, functionName, qualifier, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Provisioned Concurrency Config (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceProvisionedConcurrencyConfigRead(ctx, d, meta)...)
}

func resourceProvisionedConcurrencyConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), provisionedConcurrencyConfigResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	functionName, qualifier := parts[0], parts[1]

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Lambda Provisioned Concurrency Config %q", d.Id())
		return diags
	}

	log.Printf("[INFO] Deleting Lambda Provisioned Concurrency Config: %s", d.Id())
	_, err = conn.DeleteProvisionedConcurrencyConfig(ctx, &lambda.DeleteProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	})

	if errs.IsA[*awstypes.ProvisionedConcurrencyConfigNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Provisioned Concurrency Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findProvisionedConcurrencyConfig(ctx context.Context, conn *lambda.Client, input *lambda.GetProvisionedConcurrencyConfigInput) (*lambda.GetProvisionedConcurrencyConfigOutput, error) {
	output, err := conn.GetProvisionedConcurrencyConfig(ctx, input)

	if errs.IsA[*awstypes.ProvisionedConcurrencyConfigNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findProvisionedConcurrencyConfigByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, qualifier string) (*lambda.GetProvisionedConcurrencyConfigOutput, error) {
	input := &lambda.GetProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	return findProvisionedConcurrencyConfig(ctx, conn, input)
}

func statusProvisionedConcurrencyConfig(ctx context.Context, conn *lambda.Client, functionName, qualifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findProvisionedConcurrencyConfigByTwoPartKey(ctx, conn, functionName, qualifier)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitProvisionedConcurrencyConfigReady(ctx context.Context, conn *lambda.Client, functionName, qualifier string, timeout time.Duration) (*lambda.GetProvisionedConcurrencyConfigOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProvisionedConcurrencyStatusEnumInProgress),
		Target:  enum.Slice(awstypes.ProvisionedConcurrencyStatusEnumReady),
		Refresh: statusProvisionedConcurrencyConfig(ctx, conn, functionName, qualifier),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lambda.GetProvisionedConcurrencyConfigOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}
