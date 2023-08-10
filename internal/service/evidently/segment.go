// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 160),
			},
			"experiment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
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
	conn := meta.(*conns.AWSClient).EvidentlyConn(ctx)

	name := d.Get("name").(string)
	input := &cloudwatchevidently.CreateSegmentInput{
		Name:    aws.String(name),
		Pattern: aws.String(d.Get("pattern").(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateSegmentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CloudWatch Evidently Segment (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Segment.Arn))

	return resourceSegmentRead(ctx, d, meta)
}

func resourceSegmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn(ctx)

	segment, err := FindSegmentByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Segment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Evidently Segment (%s): %s", d.Id(), err)
	}

	d.Set("arn", segment.Arn)
	d.Set("created_time", aws.TimeValue(segment.CreatedTime).Format(time.RFC3339))
	d.Set("description", segment.Description)
	d.Set("experiment_count", segment.ExperimentCount)
	d.Set("last_updated_time", aws.TimeValue(segment.LastUpdatedTime).Format(time.RFC3339))
	d.Set("launch_count", segment.LaunchCount)
	d.Set("name", segment.Name)
	d.Set("pattern", segment.Pattern)

	setTagsOut(ctx, segment.Tags)

	return nil
}

func resourceSegmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceSegmentRead(ctx, d, meta)
}

func resourceSegmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Evidently Segment: %s", d.Id())
	_, err := conn.DeleteSegmentWithContext(ctx, &cloudwatchevidently.DeleteSegmentInput{
		Segment: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Evidently Segment (%s): %s", d.Id(), err)
	}

	return nil
}
