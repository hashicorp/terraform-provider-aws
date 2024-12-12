// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_directory_service_log_subscription", name="Log Subscription")
func resourceLogSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLogSubscriptionCreate,
		ReadWithoutTimeout:   resourceLogSubscriptionRead,
		DeleteWithoutTimeout: resourceLogSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrLogGroupName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLogSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID := d.Get("directory_id").(string)
	input := &directoryservice.CreateLogSubscriptionInput{
		DirectoryId:  aws.String(directoryID),
		LogGroupName: aws.String(d.Get(names.AttrLogGroupName).(string)),
	}

	_, err := conn.CreateLogSubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Directory Service Log Subscription (%s): %s", directoryID, err)
	}

	d.SetId(directoryID)

	return append(diags, resourceLogSubscriptionRead(ctx, d, meta)...)
}

func resourceLogSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	logSubscription, err := findLogSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Log Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Log Subscription (%s): %s", d.Id(), err)
	}

	d.Set("directory_id", logSubscription.DirectoryId)
	d.Set(names.AttrLogGroupName, logSubscription.LogGroupName)

	return diags
}

func resourceLogSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	log.Printf("[DEBUG] Deleting Directory Service Log Subscription: %s", d.Id())
	_, err := conn.DeleteLogSubscription(ctx, &directoryservice.DeleteLogSubscriptionInput{
		DirectoryId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Log Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func findLogSubscription(ctx context.Context, conn *directoryservice.Client, input *directoryservice.ListLogSubscriptionsInput) (*awstypes.LogSubscription, error) {
	output, err := findLogSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLogSubscriptions(ctx context.Context, conn *directoryservice.Client, input *directoryservice.ListLogSubscriptionsInput) ([]awstypes.LogSubscription, error) {
	var output []awstypes.LogSubscription

	pages := directoryservice.NewListLogSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LogSubscriptions...)
	}

	return output, nil
}

func findLogSubscriptionByID(ctx context.Context, conn *directoryservice.Client, directoryID string) (*awstypes.LogSubscription, error) {
	input := &directoryservice.ListLogSubscriptionsInput{
		DirectoryId: aws.String(directoryID),
	}

	return findLogSubscription(ctx, conn, input)
}
