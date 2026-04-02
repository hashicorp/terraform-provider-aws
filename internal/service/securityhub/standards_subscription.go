// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package securityhub

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_securityhub_standards_subscription", name="Standards Subscription")
func resourceStandardsSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStandardsSubscriptionCreate,
		ReadWithoutTimeout:   resourceStandardsSubscriptionRead,
		DeleteWithoutTimeout: resourceStandardsSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"standards_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
	}
}

func resourceStandardsSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	standardsARN := d.Get("standards_arn").(string)
	input := &securityhub.BatchEnableStandardsInput{
		StandardsSubscriptionRequests: []types.StandardsSubscriptionRequest{{
			StandardsArn: aws.String(standardsARN),
		}},
	}

	output, err := conn.BatchEnableStandards(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Standards Subscription (%s): %s", standardsARN, err)
	}

	d.SetId(aws.ToString(output.StandardsSubscriptions[0].StandardsSubscriptionArn))

	if _, err := waitStandardsSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Standards Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStandardsSubscriptionRead(ctx, d, meta)...)
}

func resourceStandardsSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := findStandardsSubscriptionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Security Hub Standards Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Standards Subscription (%s): %s", d.Id(), err)
	}

	d.Set("standards_arn", output.StandardsArn)

	return diags
}

func resourceStandardsSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Standards Subscription: %s", d.Id())
	_, err := conn.BatchDisableStandards(ctx, &securityhub.BatchDisableStandardsInput{
		StandardsSubscriptionArns: []string{d.Id()},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Standard (%s): %s", d.Id(), err)
	}

	if _, err := waitStandardsSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Standards Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStandardsSubscriptionByARN(ctx context.Context, conn *securityhub.Client, arn string) (*types.StandardsSubscription, error) {
	input := &securityhub.GetEnabledStandardsInput{
		StandardsSubscriptionArns: []string{arn},
	}

	output, err := findStandardsSubscription(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.StandardsStatus; status == types.StandardsStatusFailed {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findStandardsSubscription(ctx context.Context, conn *securityhub.Client, input *securityhub.GetEnabledStandardsInput) (*types.StandardsSubscription, error) {
	output, err := findStandardsSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStandardsSubscriptions(ctx context.Context, conn *securityhub.Client, input *securityhub.GetEnabledStandardsInput) ([]types.StandardsSubscription, error) {
	var output []types.StandardsSubscription

	pages := securityhub.NewGetEnabledStandardsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.StandardsSubscriptions...)
	}

	return output, nil
}

func statusStandardsSubscriptionCreate(conn *securityhub.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findStandardsSubscriptionByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.StandardsStatus), nil
	}
}

func statusStandardsSubscriptionDelete(conn *securityhub.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findStandardsSubscriptionByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.StandardsStatus == types.StandardsStatusIncomplete {
			return nil, "", nil
		}

		return output, string(output.StandardsStatus), nil
	}
}

func waitStandardsSubscriptionCreated(ctx context.Context, conn *securityhub.Client, arn string, timeout time.Duration) (*types.StandardsSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.StandardsStatusPending),
		Target:  enum.Slice(types.StandardsStatusReady, types.StandardsStatusIncomplete),
		Refresh: statusStandardsSubscriptionCreate(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.StandardsSubscription); ok {
		if reason := output.StandardsStatusReason; reason != nil {
			retry.SetLastError(err, errors.New(string(reason.StatusReasonCode)))
		}

		return output, err
	}

	return nil, err
}

func waitStandardsSubscriptionDeleted(ctx context.Context, conn *securityhub.Client, arn string, timeout time.Duration) (*types.StandardsSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.StandardsStatusDeleting),
		Target:  []string{},
		Refresh: statusStandardsSubscriptionDelete(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.StandardsSubscription); ok {
		if reason := output.StandardsStatusReason; reason != nil {
			retry.SetLastError(err, errors.New(string(reason.StatusReasonCode)))
		}

		return output, err
	}

	return nil, err
}
