// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_media_convert_queue", name="Queue")
// @Tags(identifierAttribute="arn")
func resourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueueCreate,
		ReadWithoutTimeout:   resourceQueueRead,
		UpdateWithoutTimeout: resourceQueueUpdate,
		DeleteWithoutTimeout: resourceQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pricing_plan": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.PricingPlanOnDemand,
				ValidateDiagFunc: enum.Validate[types.PricingPlan](),
			},
			"reservation_plan_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"commitment": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.Commitment](),
						},
						"renewal_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.RenewalType](),
						},
						"reserved_slots": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.QueueStatusActive,
				ValidateDiagFunc: enum.Validate[types.QueueStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &mediaconvert.CreateQueueInput{
		Name:        aws.String(name),
		PricingPlan: types.PricingPlan(d.Get("pricing_plan").(string)),
		Status:      types.QueueStatus(d.Get(names.AttrStatus).(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.Get("reservation_plan_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		input.ReservationPlanSettings = expandReservationPlanSettings(v[0].(map[string]interface{}))
	}

	output, err := conn.CreateQueue(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Media Convert Queue (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Queue.Name))

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	queue, err := findQueueByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Media Convert Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert Queue (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, queue.Arn)
	d.Set(names.AttrDescription, queue.Description)
	d.Set(names.AttrName, queue.Name)
	d.Set("pricing_plan", queue.PricingPlan)
	if queue.ReservationPlan != nil {
		if err := d.Set("reservation_plan_settings", []interface{}{flattenReservationPlan(queue.ReservationPlan)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting reservation_plan_settings: %s", err)
		}
	} else {
		d.Set("reservation_plan_settings", nil)
	}
	d.Set(names.AttrStatus, queue.Status)

	return diags
}

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &mediaconvert.UpdateQueueInput{
			Name:   aws.String(d.Id()),
			Status: types.QueueStatus(d.Get(names.AttrStatus).(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.Get("reservation_plan_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
			input.ReservationPlanSettings = expandReservationPlanSettings(v[0].(map[string]interface{}))
		}

		_, err := conn.UpdateQueue(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Media Convert Queue (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	log.Printf("[DEBUG] Deleting Media Convert Queue: %s", d.Id())
	_, err := conn.DeleteQueue(ctx, &mediaconvert.DeleteQueueInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Media Convert Queue (%s): %s", d.Id(), err)
	}

	return diags
}

func findQueueByName(ctx context.Context, conn *mediaconvert.Client, name string) (*types.Queue, error) {
	input := &mediaconvert.GetQueueInput{
		Name: aws.String(name),
	}

	output, err := conn.GetQueue(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

func expandReservationPlanSettings(tfMap map[string]interface{}) *types.ReservationPlanSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ReservationPlanSettings{}

	if v, ok := tfMap["commitment"]; ok {
		apiObject.Commitment = types.Commitment(v.(string))
	}

	if v, ok := tfMap["renewal_type"]; ok {
		apiObject.RenewalType = types.RenewalType(v.(string))
	}

	if v, ok := tfMap["reserved_slots"]; ok {
		apiObject.ReservedSlots = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenReservationPlan(apiObject *types.ReservationPlan) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"commitment":     apiObject.Commitment,
		"renewal_type":   apiObject.RenewalType,
		"reserved_slots": aws.ToInt32(apiObject.ReservedSlots),
	}

	return tfMap
}
