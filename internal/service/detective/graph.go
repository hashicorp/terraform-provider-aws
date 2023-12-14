// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
	var diags diag.Diagnostics

	const (
		timeout = 4 * time.Minute
	)
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	input := &detective.CreateGraphInput{
		Tags: getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.CreateGraphWithContext(ctx, input)
	}, detective.ErrCodeInternalServerException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Detective Graph: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*detective.CreateGraphOutput).GraphArn))

	return append(diags, resourceGraphRead(ctx, d, meta)...)
}

func resourceGraphRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	graph, err := FindGraphByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Detective Graph (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Graph (%s): %s", d.Id(), err)
	}

	d.Set("created_time", aws.TimeValue(graph.CreatedTime).Format(time.RFC3339))
	d.Set("graph_arn", graph.Arn)

	return diags
}

func resourceGraphUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceGraphRead(ctx, d, meta)
}

func resourceGraphDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	log.Printf("[DEBUG] Deleting Detective Graph: %s", d.Id())
	_, err := conn.DeleteGraphWithContext(ctx, &detective.DeleteGraphInput{
		GraphArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Detective Graph (%s): %s", d.Id(), err)
	}

	return diags
}

func FindGraphByARN(ctx context.Context, conn *detective.Detective, arn string) (*detective.Graph, error) {
	input := &detective.ListGraphsInput{}

	return findGraph(ctx, conn, input, func(v *detective.Graph) bool {
		return aws.StringValue(v.Arn) == arn
	})
}

func findGraph(ctx context.Context, conn *detective.Detective, input *detective.ListGraphsInput, filter tfslices.Predicate[*detective.Graph]) (*detective.Graph, error) {
	output, err := findGraphs(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findGraphs(ctx context.Context, conn *detective.Detective, input *detective.ListGraphsInput, filter tfslices.Predicate[*detective.Graph]) ([]*detective.Graph, error) {
	var output []*detective.Graph

	err := conn.ListGraphsPagesWithContext(ctx, input, func(page *detective.ListGraphsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GraphList {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
