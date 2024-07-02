// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_emr_managed_scaling_policy", name="Managed Scaling Policy")
func resourceManagedScalingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceManagedScalingPolicyCreate,
		ReadWithoutTimeout:   resourceManagedScalingPolicyRead,
		DeleteWithoutTimeout: resourceManagedScalingPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"compute_limits": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unit_type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ComputeLimitsUnitType](),
						},
						"minimum_capacity_units": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"maximum_capacity_units": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"maximum_core_capacity_units": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"maximum_ondemand_capacity_units": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceManagedScalingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	if l := d.Get("compute_limits").(*schema.Set).List(); len(l) > 0 && l[0] != nil {
		cl := l[0].(map[string]interface{})
		computeLimits := &awstypes.ComputeLimits{
			UnitType:             awstypes.ComputeLimitsUnitType(cl["unit_type"].(string)),
			MinimumCapacityUnits: aws.Int32(int32(cl["minimum_capacity_units"].(int))),
			MaximumCapacityUnits: aws.Int32(int32(cl["maximum_capacity_units"].(int))),
		}
		if v, ok := cl["maximum_core_capacity_units"].(int); ok && v > 0 {
			computeLimits.MaximumCoreCapacityUnits = aws.Int32(int32(v))

			if v, ok := cl["maximum_ondemand_capacity_units"].(int); ok && v > 0 {
				computeLimits.MaximumOnDemandCapacityUnits = aws.Int32(int32(v))
			}
		} else if v, ok := cl["maximum_ondemand_capacity_units"].(int); ok && v >= 0 {
			computeLimits.MaximumOnDemandCapacityUnits = aws.Int32(int32(v))
		}
		managedScalingPolicy := &awstypes.ManagedScalingPolicy{
			ComputeLimits: computeLimits,
		}

		_, err := conn.PutManagedScalingPolicy(ctx, &emr.PutManagedScalingPolicyInput{
			ClusterId:            aws.String(d.Get("cluster_id").(string)),
			ManagedScalingPolicy: managedScalingPolicy,
		})

		if err != nil {
			log.Printf("[ERROR] EMR.PutManagedScalingPolicy %s", err)
			return sdkdiag.AppendErrorf(diags, "putting EMR Managed Scaling Policy: %s", err)
		}
	}

	d.SetId(d.Get("cluster_id").(string))
	return diags
}

func resourceManagedScalingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	input := &emr.GetManagedScalingPolicyInput{
		ClusterId: aws.String(d.Id()),
	}

	resp, err := conn.GetManagedScalingPolicy(ctx, input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "A job flow that is shutting down, terminated, or finished may not be modified") {
		log.Printf("[WARN] EMR Managed Scaling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "does not exist") {
		log.Printf("[WARN] EMR Managed Scaling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting EMR Managed Scaling Policy (%s): %s", d.Id(), err)
	}

	// Previously after RemoveManagedScalingPolicy the API returned an error, but now it
	// returns an empty response. We keep the original error handling above though just in case.
	if resp == nil || resp.ManagedScalingPolicy == nil {
		log.Printf("[WARN] EMR Managed Scaling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("cluster_id", d.Id())
	d.Set("compute_limits", flattenComputeLimits(resp.ManagedScalingPolicy.ComputeLimits))

	return diags
}

func resourceManagedScalingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	input := &emr.RemoveManagedScalingPolicyInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	}

	_, err := conn.RemoveManagedScalingPolicy(ctx, input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "A job flow that is shutting down, terminated, or finished may not be modified") {
		return diags
	}

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing EMR Managed Scaling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenComputeLimits(apiObject *awstypes.ComputeLimits) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["unit_type"] = string(apiObject.UnitType)

	if v := apiObject.MaximumCapacityUnits; v != nil {
		tfMap["maximum_capacity_units"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumCoreCapacityUnits; v != nil {
		tfMap["maximum_core_capacity_units"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumOnDemandCapacityUnits; v != nil {
		tfMap["maximum_ondemand_capacity_units"] = aws.ToInt32(v)
	}

	if v := apiObject.MinimumCapacityUnits; v != nil {
		tfMap["minimum_capacity_units"] = aws.ToInt32(v)
	}

	return []interface{}{tfMap}
}
