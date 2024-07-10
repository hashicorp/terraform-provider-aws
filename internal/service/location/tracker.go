// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_location_tracker", name="Route Calculator")
// @Tags(identifierAttribute="tracker_arn")
func ResourceTracker() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrackerCreate,
		ReadWithoutTimeout:   resourceTrackerRead,
		UpdateWithoutTimeout: resourceTrackerUpdate,
		DeleteWithoutTimeout: resourceTrackerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"position_filtering": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      locationservice.PositionFilteringTimeBased,
				ValidateFunc: validation.StringInSlice(locationservice.PositionFiltering_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tracker_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tracker_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTrackerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.CreateTrackerInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("position_filtering"); ok {
		input.PositionFiltering = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tracker_name"); ok {
		input.TrackerName = aws.String(v.(string))
	}

	output, err := conn.CreateTrackerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Tracker: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Tracker: empty result")
	}

	d.SetId(aws.StringValue(output.TrackerName))

	return append(diags, resourceTrackerRead(ctx, d, meta)...)
}

func resourceTrackerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DescribeTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	output, err := conn.DescribeTrackerWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Location Service Tracker (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Tracker (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Map (%s): empty response", d.Id())
	}

	d.Set(names.AttrCreateTime, aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrKMSKeyID, output.KmsKeyId)
	d.Set("position_filtering", output.PositionFiltering)

	setTagsOut(ctx, output.Tags)

	d.Set("tracker_arn", output.TrackerArn)
	d.Set("tracker_name", output.TrackerName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourceTrackerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	if d.HasChanges(names.AttrDescription, "position_filtering") {
		input := &locationservice.UpdateTrackerInput{
			TrackerName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("position_filtering"); ok {
			input.PositionFiltering = aws.String(v.(string))
		}

		_, err := conn.UpdateTrackerWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Location Service Tracker (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrackerRead(ctx, d, meta)...)
}

func resourceTrackerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DeleteTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteTrackerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Tracker (%s): %s", d.Id(), err)
	}

	return diags
}
