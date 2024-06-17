// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_capacity_provider", name="Capacity Provider")
// @Tags(identifierAttribute="id")
func ResourceCapacityProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapacityProviderCreate,
		ReadWithoutTimeout:   resourceCapacityProviderRead,
		UpdateWithoutTimeout: resourceCapacityProviderUpdate,
		DeleteWithoutTimeout: resourceCapacityProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceCapacityProviderImport,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_group_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_scaling_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"managed_draining": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ManagedDraining](),
						},
						"managed_scaling": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_warmup_period": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(0, 10000),
									},
									"maximum_scaling_step_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									"minimum_scaling_step_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									names.AttrStatus: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ManagedScalingStatus]()},
									"target_capacity": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
						"managed_termination_protection": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ManagedTerminationProtection](),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCapacityProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	name := d.Get(names.AttrName).(string)
	input := ecs.CreateCapacityProviderInput{
		Name:                     aws.String(name),
		AutoScalingGroupProvider: expandAutoScalingGroupProviderCreate(d.Get("auto_scaling_group_provider")),
		Tags:                     getTagsIn(ctx),
	}

	output, err := conn.CreateCapacityProvider(ctx, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateCapacityProvider(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Capacity Provider (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CapacityProvider.CapacityProviderArn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceCapacityProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Capacity Provider (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapacityProviderRead(ctx, d, meta)...)
}

func resourceCapacityProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	output, err := FindCapacityProviderByARN(ctx, conn, partition, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Capacity Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Capacity Provider (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.CapacityProviderArn)

	if err := d.Set("auto_scaling_group_provider", flattenAutoScalingGroupProvider(output.AutoScalingGroupProvider)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_scaling_group_provider: %s", err)
	}

	d.Set(names.AttrName, output.Name)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceCapacityProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ecs.UpdateCapacityProviderInput{
			AutoScalingGroupProvider: expandAutoScalingGroupProviderUpdate(d.Get("auto_scaling_group_provider")),
			Name:                     aws.String(d.Get(names.AttrName).(string)),
		}

		log.Printf("[DEBUG] Updating ECS Capacity Provider: %+v", input)
		err := retry.RetryContext(ctx, capacityProviderUpdateTimeout, func() *retry.RetryError {
			_, err := conn.UpdateCapacityProvider(ctx, input)

			if errs.IsA[*awstypes.UpdateInProgressException](err) {
				return retry.RetryableError(err)
			}

			if err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateCapacityProvider(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Capacity Provider (%s): %s", d.Id(), err)
		}

		if _, err = waitCapacityProviderUpdated(ctx, conn, partition, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Capacity Provider (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapacityProviderRead(ctx, d, meta)...)
}

func resourceCapacityProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	log.Printf("[DEBUG] Deleting ECS Capacity Provider (%s)", d.Id())
	_, err := conn.DeleteCapacityProvider(ctx, &ecs.DeleteCapacityProviderInput{
		CapacityProvider: aws.String(d.Id()),
	})

	// "An error occurred (ClientException) when calling the DeleteCapacityProvider operation: The specified capacity provider does not exist. Specify a valid name or ARN and try again."
	if errs.IsAErrorMessageContains[*awstypes.ClientException](err, "capacity provider does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Capacity Provider (%s): %s", d.Id(), err)
	}

	if _, err := waitCapacityProviderDeleted(ctx, conn, partition, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Capacity Provider (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func resourceCapacityProviderImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set(names.AttrName, d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   "ecs",
		Resource:  fmt.Sprintf("capacity-provider/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func expandAutoScalingGroupProviderCreate(configured interface{}) *awstypes.AutoScalingGroupProvider {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	prov := awstypes.AutoScalingGroupProvider{}
	p := configured.([]interface{})[0].(map[string]interface{})
	arn := p["auto_scaling_group_arn"].(string)
	prov.AutoScalingGroupArn = aws.String(arn)

	if mtp := p["managed_draining"].(string); len(mtp) > 0 {
		prov.ManagedDraining = awstypes.ManagedDraining(mtp)
	}

	prov.ManagedScaling = expandManagedScaling(p["managed_scaling"])

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = awstypes.ManagedTerminationProtection(mtp)
	}

	return &prov
}

func expandAutoScalingGroupProviderUpdate(configured interface{}) *awstypes.AutoScalingGroupProviderUpdate {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	prov := awstypes.AutoScalingGroupProviderUpdate{}
	p := configured.([]interface{})[0].(map[string]interface{})

	if mtp := p["managed_draining"].(string); len(mtp) > 0 {
		prov.ManagedDraining = awstypes.ManagedDraining(mtp)
	}

	prov.ManagedScaling = expandManagedScaling(p["managed_scaling"])

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = awstypes.ManagedTerminationProtection(mtp)
	}

	return &prov
}

func expandManagedScaling(configured interface{}) *awstypes.ManagedScaling {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	tfMap := configured.([]interface{})[0].(map[string]interface{})

	managedScaling := awstypes.ManagedScaling{}

	if v, ok := tfMap["instance_warmup_period"].(int); ok {
		managedScaling.InstanceWarmupPeriod = aws.Int32(int32(v))
	}
	if v, ok := tfMap["maximum_scaling_step_size"].(int); ok && v != 0 {
		managedScaling.MaximumScalingStepSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["minimum_scaling_step_size"].(int); ok && v != 0 {
		managedScaling.MinimumScalingStepSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && len(v) > 0 {
		managedScaling.Status = awstypes.ManagedScalingStatus(v)
	}
	if v, ok := tfMap["target_capacity"].(int); ok && v != 0 {
		managedScaling.TargetCapacity = aws.Int32(int32(v))
	}

	return &managedScaling
}

func flattenAutoScalingGroupProvider(provider *awstypes.AutoScalingGroupProvider) []map[string]interface{} {
	if provider == nil {
		return nil
	}

	p := map[string]interface{}{
		"auto_scaling_group_arn":         aws.ToString(provider.AutoScalingGroupArn),
		"managed_draining":               string(provider.ManagedDraining),
		"managed_scaling":                []map[string]interface{}{},
		"managed_termination_protection": string(provider.ManagedTerminationProtection),
	}

	if provider.ManagedScaling != nil {
		m := map[string]interface{}{
			"instance_warmup_period":    aws.ToInt32(provider.ManagedScaling.InstanceWarmupPeriod),
			"maximum_scaling_step_size": aws.ToInt32(provider.ManagedScaling.MaximumScalingStepSize),
			"minimum_scaling_step_size": aws.ToInt32(provider.ManagedScaling.MinimumScalingStepSize),
			names.AttrStatus:            string(provider.ManagedScaling.Status),
			"target_capacity":           aws.ToInt32(provider.ManagedScaling.TargetCapacity),
		}

		p["managed_scaling"] = []map[string]interface{}{m}
	}

	result := []map[string]interface{}{p}
	return result
}
