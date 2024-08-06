// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/detective"
	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			names.AttrCreatedTime: {
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
	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	input := &detective.CreateGraphInput{
		Tags: getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InternalServerException](ctx, timeout, func() (interface{}, error) {
		return conn.CreateGraph(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Detective Graph: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*detective.CreateGraphOutput).GraphArn))

	return append(diags, resourceGraphRead(ctx, d, meta)...)
}

func resourceGraphRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	graph, err := FindGraphByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Detective Graph (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Graph (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreatedTime, aws.ToTime(graph.CreatedTime).Format(time.RFC3339))
	d.Set("graph_arn", graph.Arn)

	return diags
}

func resourceGraphUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceGraphRead(ctx, d, meta)
}

func resourceGraphDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	log.Printf("[DEBUG] Deleting Detective Graph: %s", d.Id())
	_, err := conn.DeleteGraph(ctx, &detective.DeleteGraphInput{
		GraphArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Detective Graph (%s): %s", d.Id(), err)
	}

	return diags
}

func FindGraphByARN(ctx context.Context, conn *detective.Client, arn string) (*awstypes.Graph, error) {
	input := &detective.ListGraphsInput{}

	return findGraph(ctx, conn, input, func(v awstypes.Graph) bool {
		return aws.ToString(v.Arn) == arn
	})
}

func findGraph(ctx context.Context, conn *detective.Client, input *detective.ListGraphsInput, filter tfslices.Predicate[awstypes.Graph]) (*awstypes.Graph, error) {
	output, err := findGraphs(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGraphs(ctx context.Context, conn *detective.Client, input *detective.ListGraphsInput, filter tfslices.Predicate[awstypes.Graph]) ([]awstypes.Graph, error) {
	var output []awstypes.Graph

	pages := detective.NewListGraphsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.GraphList {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
