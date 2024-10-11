// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datapipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datapipeline"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datapipeline/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datapipeline_pipeline", name="Pipeline")
// @Tags(identifierAttribute="id")
func ResourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineCreate,
		ReadWithoutTimeout:   resourcePipelineRead,
		UpdateWithoutTimeout: resourcePipelineUpdate,
		DeleteWithoutTimeout: resourcePipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)

	uniqueID := id.UniqueId()
	input := datapipeline.CreatePipelineInput{
		Name:     aws.String(d.Get(names.AttrName).(string)),
		UniqueId: aws.String(uniqueID),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	resp, err := conn.CreatePipeline(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating datapipeline: %s", err)
	}

	d.SetId(aws.ToString(resp.PipelineId))

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)

	v, err := PipelineRetrieve(ctx, d.Id(), conn)
	if errs.IsA[*awstypes.PipelineNotFoundException](err) || errs.IsA[*awstypes.PipelineDeletedException](err) || v == nil {
		log.Printf("[WARN] DataPipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing DataPipeline (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, v.Name)
	d.Set(names.AttrDescription, v.Description)

	setTagsOut(ctx, v.Tags)

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)

	opts := datapipeline.DeletePipelineInput{
		PipelineId: aws.String(d.Id()),
	}

	_, err := conn.DeletePipeline(ctx, &opts)
	if errs.IsA[*awstypes.PipelineNotFoundException](err) || errs.IsA[*awstypes.PipelineDeletedException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Data Pipeline %s: %s", d.Id(), err)
	}

	if err := WaitForDeletion(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Data Pipeline %s: %s", d.Id(), err)
	}
	return diags
}

func PipelineRetrieve(ctx context.Context, id string, conn *datapipeline.Client) (*awstypes.PipelineDescription, error) {
	opts := datapipeline.DescribePipelinesInput{
		PipelineIds: []string{id},
	}

	resp, err := conn.DescribePipelines(ctx, &opts)
	if err != nil {
		return nil, err
	}

	var pipeline awstypes.PipelineDescription

	for _, p := range resp.PipelineDescriptionList {
		if aws.ToString(p.PipelineId) == id {
			pipeline = p
			break
		}
	}

	return &pipeline, nil
}

func WaitForDeletion(ctx context.Context, conn *datapipeline.Client, pipelineID string) error {
	params := &datapipeline.DescribePipelinesInput{
		PipelineIds: []string{pipelineID},
	}
	return retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribePipelines(ctx, params)
		if errs.IsA[*awstypes.PipelineNotFoundException](err) || errs.IsA[*awstypes.PipelineDeletedException](err) {
			return nil
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("DataPipeline (%s) still exists", pipelineID))
	})
}
