// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_securityhub_standards_subscription")
func ResourceStandardsSubscription() *schema.Resource {
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
	}
}

func resourceStandardsSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	standardsARN := d.Get("standards_arn").(string)
	input := &securityhub.BatchEnableStandardsInput{
		StandardsSubscriptionRequests: []*securityhub.StandardsSubscriptionRequest{{
			StandardsArn: aws.String(standardsARN),
		}},
	}

	output, err := conn.BatchEnableStandardsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Standards Subscription (%s): %s", standardsARN, err)
	}

	d.SetId(aws.StringValue(output.StandardsSubscriptions[0].StandardsSubscriptionArn))

	if _, err := waitStandardsSubscriptionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Standards Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStandardsSubscriptionRead(ctx, d, meta)...)
}

func resourceStandardsSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	output, err := FindStandardsSubscriptionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourceStandardsSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Standards Subscription: %s", d.Id())
	_, err := conn.BatchDisableStandardsWithContext(ctx, &securityhub.BatchDisableStandardsInput{
		StandardsSubscriptionArns: aws.StringSlice([]string{d.Id()}),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Standard (%s): %s", d.Id(), err)
	}

	_, err = waitStandardsSubscriptionDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Standards Subscription (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func FindStandardsSubscription(ctx context.Context, conn *securityhub.SecurityHub, input *securityhub.GetEnabledStandardsInput) (*securityhub.StandardsSubscription, error) {
	output, err := FindStandardsSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindStandardsSubscriptions(ctx context.Context, conn *securityhub.SecurityHub, input *securityhub.GetEnabledStandardsInput) ([]*securityhub.StandardsSubscription, error) {
	var output []*securityhub.StandardsSubscription

	err := conn.GetEnabledStandardsPagesWithContext(ctx, input, func(page *securityhub.GetEnabledStandardsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StandardsSubscriptions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindStandardsSubscriptionByARN(ctx context.Context, conn *securityhub.SecurityHub, arn string) (*securityhub.StandardsSubscription, error) {
	input := &securityhub.GetEnabledStandardsInput{
		StandardsSubscriptionArns: aws.StringSlice([]string{arn}),
	}

	output, err := FindStandardsSubscription(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.StandardsStatus); status == securityhub.StandardsStatusFailed {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

const (
	standardsStatusNotFound = "NotFound"

	standardsSubscriptionCreateTimeout = 3 * time.Minute
	standardsSubscriptionDeleteTimeout = 3 * time.Minute
)

func statusStandardsSubscription(ctx context.Context, conn *securityhub.SecurityHub, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStandardsSubscriptionByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			// Return a fake result and status to deal with the INCOMPLETE subscription status
			// being a target for both Create and Delete.
			return "", standardsStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StandardsStatus), nil
	}
}

func waitStandardsSubscriptionCreated(ctx context.Context, conn *securityhub.SecurityHub, arn string) (*securityhub.StandardsSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{securityhub.StandardsStatusPending},
		Target:  []string{securityhub.StandardsStatusReady, securityhub.StandardsStatusIncomplete},
		Refresh: statusStandardsSubscription(ctx, conn, arn),
		Timeout: standardsSubscriptionCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*securityhub.StandardsSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitStandardsSubscriptionDeleted(ctx context.Context, conn *securityhub.SecurityHub, arn string) (*securityhub.StandardsSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{securityhub.StandardsStatusDeleting},
		Target:  []string{standardsStatusNotFound, securityhub.StandardsStatusIncomplete},
		Refresh: statusStandardsSubscription(ctx, conn, arn),
		Timeout: standardsSubscriptionDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*securityhub.StandardsSubscription); ok {
		return output, err
	}

	return nil, err
}
