// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_evidently_segment", name="Segment")
// @Tags(identifierAttribute="arn")
func ResourceSegment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSegmentCreate,
		ReadWithoutTimeout:   resourceSegmentRead,
		UpdateWithoutTimeout: resourceSegmentUpdate,
		DeleteWithoutTimeout: resourceSegmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 160),
			},
			"experiment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			"pattern": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSegmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &evidently.CreateSegmentInput{
		Name:    aws.String(name),
		Pattern: aws.String(d.Get("pattern").(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateSegment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Evidently Segment (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Segment.Arn))

	return append(diags, resourceSegmentRead(ctx, d, meta)...)
}

func resourceSegmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	segment, err := FindSegmentByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Segment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Evidently Segment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, segment.Arn)
	d.Set(names.AttrCreatedTime, aws.ToTime(segment.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, segment.Description)
	d.Set("experiment_count", segment.ExperimentCount)
	d.Set(names.AttrLastUpdatedTime, aws.ToTime(segment.LastUpdatedTime).Format(time.RFC3339))
	d.Set("launch_count", segment.LaunchCount)
	d.Set(names.AttrName, segment.Name)
	d.Set("pattern", segment.Pattern)

	setTagsOut(ctx, segment.Tags)

	return diags
}

func resourceSegmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceSegmentRead(ctx, d, meta)
}

func resourceSegmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Evidently Segment: %s", d.Id())
	_, err := conn.DeleteSegment(ctx, &evidently.DeleteSegmentInput{
		Segment: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Evidently Segment (%s): %s", d.Id(), err)
	}

	return diags
}
