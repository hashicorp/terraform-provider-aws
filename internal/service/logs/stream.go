// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_cloudwatch_log_stream")
func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		DeleteWithoutTimeout: resourceStreamDelete,

		Importer: &schema.ResourceImporter{
			State: resourceStreamImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validStreamName,
			},
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn(ctx)

	name := d.Get("name").(string)
	input := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(d.Get("log_group_name").(string)),
		LogStreamName: aws.String(name),
	}

	_, err := conn.CreateLogStreamWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CloudWatch Logs Log Stream (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindLogStreamByTwoPartKey(ctx, conn, d.Get("log_group_name").(string), d.Id())
	})

	if err != nil {
		return diag.Errorf("waiting for CloudWatch Logs Log Stream (%s) create: %s", d.Id(), err)
	}

	return resourceStreamRead(ctx, d, meta)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn(ctx)

	ls, err := FindLogStreamByTwoPartKey(ctx, conn, d.Get("log_group_name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Log Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Log Stream (%s): %s", d.Id(), err)
	}

	d.Set("arn", ls.Arn)
	d.Set("name", ls.LogStreamName)

	return nil
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn(ctx)

	log.Printf("[INFO] Deleting CloudWatch Logs Log Stream: %s", d.Id())
	_, err := conn.DeleteLogStreamWithContext(ctx, &cloudwatchlogs.DeleteLogStreamInput{
		LogGroupName:  aws.String(d.Get("log_group_name").(string)),
		LogStreamName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Logs Log Stream (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceStreamImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'log-group-name:log-stream-name'", d.Id())
	}

	logGroupName := parts[0]
	logStreamName := parts[1]

	d.SetId(logStreamName)
	d.Set("log_group_name", logGroupName)

	return []*schema.ResourceData{d}, nil
}

func FindLogStreamByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, logGroupName, name string) (*cloudwatchlogs.LogStream, error) { // nosemgrep:ci.logs-in-func-name
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroupName),
		LogStreamNamePrefix: aws.String(name),
	}
	var output *cloudwatchlogs.LogStream

	err := conn.DescribeLogStreamsPagesWithContext(ctx, input, func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LogStreams {
			if aws.StringValue(v.LogStreamName) == name {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
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

func validStreamName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if regexp.MustCompile(`:`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"colons not allowed in %q:", k))
	}
	if len(value) < 1 || len(value) > 512 {
		errors = append(errors, fmt.Errorf(
			"%q must be between 1 and 512 characters: %q", k, value))
	}

	return
}
