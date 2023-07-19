// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_detective_graph", name="Graph")
// @Tags(identifierAttribute="id")
func ResourceGraph() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGraphCreate,
		ReadWithoutTimeout:   resourceGraphRead,
		UpdateWithoutTimeout: resourceGraphUpdate,
		DeleteWithoutTimeout: resourceGraphDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"graph_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGraphCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	input := &detective.CreateGraphInput{
		Tags: getTagsIn(ctx),
	}

	var output *detective.CreateGraphOutput
	var err error
	err = retry.RetryContext(ctx, GraphOperationTimeout, func() *retry.RetryError {
		output, err = conn.CreateGraphWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, detective.ErrCodeInternalServerException) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateGraphWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("creating detective Graph: %s", err)
	}

	d.SetId(aws.StringValue(output.GraphArn))

	return resourceGraphRead(ctx, d, meta)
}

func resourceGraphRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	resp, err := FindGraphByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) || resp == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("reading detective Graph (%s): %s", d.Id(), err)
	}

	d.Set("created_time", aws.TimeValue(resp.CreatedTime).Format(time.RFC3339))
	d.Set("graph_arn", resp.Arn)

	return nil
}

func resourceGraphUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceGraphRead(ctx, d, meta)
}

func resourceGraphDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	input := &detective.DeleteGraphInput{
		GraphArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteGraphWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("deleting detective Graph (%s): %s", d.Id(), err)
	}

	return nil
}
