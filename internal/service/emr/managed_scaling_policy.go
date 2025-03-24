// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
						"minimum_capacity_units": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"unit_type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ComputeLimitsUnitType](),
						},
					},
				},
			},
		},
	}
}

func resourceManagedScalingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	clusterID := d.Get("cluster_id").(string)
	if v := d.Get("compute_limits").(*schema.Set).List(); len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)
		computeLimits := &awstypes.ComputeLimits{
			UnitType:             awstypes.ComputeLimitsUnitType(tfMap["unit_type"].(string)),
			MinimumCapacityUnits: aws.Int32(int32(tfMap["minimum_capacity_units"].(int))),
			MaximumCapacityUnits: aws.Int32(int32(tfMap["maximum_capacity_units"].(int))),
		}
		if v, ok := tfMap["maximum_core_capacity_units"].(int); ok && v > 0 {
			computeLimits.MaximumCoreCapacityUnits = aws.Int32(int32(v))

			if v, ok := tfMap["maximum_ondemand_capacity_units"].(int); ok && v > 0 {
				computeLimits.MaximumOnDemandCapacityUnits = aws.Int32(int32(v))
			}
		} else if v, ok := tfMap["maximum_ondemand_capacity_units"].(int); ok && v >= 0 {
			computeLimits.MaximumOnDemandCapacityUnits = aws.Int32(int32(v))
		}
		input := &emr.PutManagedScalingPolicyInput{
			ClusterId: aws.String(clusterID),
			ManagedScalingPolicy: &awstypes.ManagedScalingPolicy{
				ComputeLimits: computeLimits,
			},
		}

		_, err := conn.PutManagedScalingPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting EMR Managed Scaling Policy: %s", err)
		}
	}

	d.SetId(clusterID)

	return diags
}

func resourceManagedScalingPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	managedScalingPolicy, err := findManagedScalingPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Managed Scaling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Managed Scaling Policy (%s): %s", d.Id(), err)
	}

	d.Set("cluster_id", d.Id())
	if err := d.Set("compute_limits", flattenComputeLimits(managedScalingPolicy.ComputeLimits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_limits: %s", err)
	}

	return diags
}

func resourceManagedScalingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	log.Printf("[INFO] Deleting EMR Managed Scaling Policy: %s", d.Id())
	_, err := conn.RemoveManagedScalingPolicy(ctx, &emr.RemoveManagedScalingPolicyInput{
		ClusterId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeValidationException, "A job flow that is shutting down, terminated, or finished may not be modified") ||
		tfawserr.ErrMessageContains(err, errCodeValidationException, "is not valid") ||
		errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Managed Scaling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findManagedScalingPolicyByID(ctx context.Context, conn *emr.Client, id string) (*awstypes.ManagedScalingPolicy, error) {
	input := &emr.GetManagedScalingPolicyInput{
		ClusterId: aws.String(id),
	}

	return findManagedScalingPolicy(ctx, conn, input)
}

func findManagedScalingPolicy(ctx context.Context, conn *emr.Client, input *emr.GetManagedScalingPolicyInput) (*awstypes.ManagedScalingPolicy, error) {
	output, err := conn.GetManagedScalingPolicy(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeValidationException, "A job flow that is shutting down, terminated, or finished may not be modified") ||
		tfawserr.ErrMessageContains(err, errCodeValidationException, "is not valid") ||
		errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ManagedScalingPolicy == nil {
		return nil, tfresource.NewEmptyResultError((input))
	}

	return output.ManagedScalingPolicy, nil
}

func flattenComputeLimits(apiObject *awstypes.ComputeLimits) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}
