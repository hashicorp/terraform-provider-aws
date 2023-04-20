package datapipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
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
	conn := meta.(*conns.AWSClient).DataPipelineConn()

	uniqueID := id.UniqueId()
	input := datapipeline.CreatePipelineInput{
		Name:     aws.String(d.Get("name").(string)),
		UniqueId: aws.String(uniqueID),
		Tags:     GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	resp, err := conn.CreatePipelineWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating datapipeline: %s", err)
	}

	d.SetId(aws.StringValue(resp.PipelineId))

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataPipelineConn()

	v, err := PipelineRetrieve(ctx, d.Id(), conn)
	if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) || tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) || v == nil {
		log.Printf("[WARN] DataPipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error describing DataPipeline (%s): %s", d.Id(), err)
	}

	d.Set("name", v.Name)
	d.Set("description", v.Description)

	SetTagsOut(ctx, v.Tags)

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataPipelineConn()

	opts := datapipeline.DeletePipelineInput{
		PipelineId: aws.String(d.Id()),
	}

	_, err := conn.DeletePipelineWithContext(ctx, &opts)
	if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) || tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Data Pipeline %s: %s", d.Id(), err)
	}

	if err := WaitForDeletion(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Data Pipeline %s: %s", d.Id(), err)
	}
	return nil
}

func PipelineRetrieve(ctx context.Context, id string, conn *datapipeline.DataPipeline) (*datapipeline.PipelineDescription, error) {
	opts := datapipeline.DescribePipelinesInput{
		PipelineIds: []*string{aws.String(id)},
	}

	resp, err := conn.DescribePipelinesWithContext(ctx, &opts)
	if err != nil {
		return nil, err
	}

	var pipeline *datapipeline.PipelineDescription

	for _, p := range resp.PipelineDescriptionList {
		if p == nil {
			continue
		}

		if aws.StringValue(p.PipelineId) == id {
			pipeline = p
			break
		}
	}

	return pipeline, nil
}

func WaitForDeletion(ctx context.Context, conn *datapipeline.DataPipeline, pipelineID string) error {
	params := &datapipeline.DescribePipelinesInput{
		PipelineIds: []*string{aws.String(pipelineID)},
	}
	return retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribePipelinesWithContext(ctx, params)
		if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) || tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
			return nil
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("DataPipeline (%s) still exists", pipelineID))
	})
}
