// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_stream", name="Log Stream")
func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		DeleteWithoutTimeout: resourceStreamDelete,

		Importer: &schema.ResourceImporter{
			State: resourceStreamImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLogGroupName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogStreamName,
			},
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(d.Get(names.AttrLogGroupName).(string)),
		LogStreamName: aws.String(name),
	}

	_, err := conn.CreateLogStream(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Logs Log Stream (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (any, error) {
		return findLogStreamByTwoPartKey(ctx, conn, d.Get(names.AttrLogGroupName).(string), d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Logs Log Stream (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	ls, err := findLogStreamByTwoPartKey(ctx, conn, d.Get(names.AttrLogGroupName).(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Log Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Stream (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, ls.Arn)
	d.Set(names.AttrName, ls.LogStreamName)

	return diags
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[INFO] Deleting CloudWatch Logs Log Stream: %s", d.Id())
	_, err := conn.DeleteLogStream(ctx, &cloudwatchlogs.DeleteLogStreamInput{
		LogGroupName:  aws.String(d.Get(names.AttrLogGroupName).(string)),
		LogStreamName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Stream (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceStreamImport(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'log-group-name:log-stream-name'", d.Id())
	}

	logGroupName := parts[0]
	logStreamName := parts[1]

	d.SetId(logStreamName)
	d.Set(names.AttrLogGroupName, logGroupName)

	return []*schema.ResourceData{d}, nil
}

func findLogStreamByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName, name string) (*awstypes.LogStream, error) { // nosemgrep:ci.logs-in-func-name
	input := cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroupName),
		LogStreamNamePrefix: aws.String(name),
	}

	return findLogStream(ctx, conn, &input, func(v *awstypes.LogStream) bool {
		return aws.ToString(v.LogStreamName) == name
	})
}

func findLogStream(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogStreamsInput, filter tfslices.Predicate[*awstypes.LogStream]) (*awstypes.LogStream, error) { // nosemgrep:ci.logs-in-func-name
	output, err := findLogStreams(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLogStreams(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogStreamsInput, filter tfslices.Predicate[*awstypes.LogStream]) ([]awstypes.LogStream, error) { // nosemgrep:ci.logs-in-func-name
	var output []awstypes.LogStream

	pages := cloudwatchlogs.NewDescribeLogStreamsPaginator(conn, input)
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

		for _, v := range page.LogStreams {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
