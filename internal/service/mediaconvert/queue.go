// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_media_convert_queue", name="Queue")
// @Tags
func ResourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueueCreate,
		ReadWithoutTimeout:   resourceQueueRead,
		UpdateWithoutTimeout: resourceQueueUpdate,
		DeleteWithoutTimeout: resourceQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pricing_plan": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      mediaconvert.PricingPlanOnDemand,
				ValidateFunc: validation.StringInSlice(mediaconvert.PricingPlan_Values(), false),
			},
			"reservation_plan_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"commitment": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(mediaconvert.Commitment_Values(), false),
						},
						"renewal_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(mediaconvert.RenewalType_Values(), false),
						},
						"reserved_slots": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      mediaconvert.QueueStatusActive,
				ValidateFunc: validation.StringInSlice(mediaconvert.QueueStatus_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertConn(ctx)

	name := d.Get("name").(string)
	input := &mediaconvert.CreateQueueInput{
		Name:        aws.String(name),
		PricingPlan: aws.String(d.Get("pricing_plan").(string)),
		Status:      aws.String(d.Get("status").(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.Get("reservation_plan_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		input.ReservationPlanSettings = expandReservationPlanSettings(v[0].(map[string]interface{}))
	}

	output, err := conn.CreateQueueWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Media Convert Queue (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Queue.Name))

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertConn(ctx)

	queue, err := FindQueueByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Media Convert Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert Queue (%s): %s", d.Id(), err)
	}

	d.Set("arn", queue.Arn)
	d.Set("description", queue.Description)
	d.Set("name", queue.Name)
	d.Set("pricing_plan", queue.PricingPlan)
	if err := d.Set("reservation_plan_settings", flattenReservationPlan(queue.ReservationPlan)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting reservation_plan_settings: %s", err)
	}
	d.Set("status", queue.Status)

	tags, err := listTags(ctx, conn, aws.StringValue(queue.Arn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Media Convert Queue (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertConn(ctx)

	if d.HasChanges("description", "reservation_plan_settings", "status") {
		input := &mediaconvert.UpdateQueueInput{
			Name:   aws.String(d.Id()),
			Status: aws.String(d.Get("status").(string)),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.Get("reservation_plan_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
			input.ReservationPlanSettings = expandReservationPlanSettings(v[0].(map[string]interface{}))
		}

		_, err := conn.UpdateQueueWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Media Convert Queue (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := updateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertConn(ctx)

	log.Printf("[DEBUG] Deleting Media Convert Queue: %s", d.Id())
	_, err := conn.DeleteQueueWithContext(ctx, &mediaconvert.DeleteQueueInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, mediaconvert.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Media Convert Queue (%s): %s", d.Id(), err)
	}

	return diags
}

func FindQueueByName(ctx context.Context, conn *mediaconvert.MediaConvert, name string) (*mediaconvert.Queue, error) {
	input := &mediaconvert.GetQueueInput{
		Name: aws.String(name),
	}

	output, err := conn.GetQueueWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, mediaconvert.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Queue == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Queue, nil
}
