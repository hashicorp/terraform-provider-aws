// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_scheduling_policy", name="Scheduling Policy")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/batch/types;types.SchedulingPolicyDetail")
func resourceSchedulingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchedulingPolicyCreate,
		ReadWithoutTimeout:   resourceSchedulingPolicyRead,
		UpdateWithoutTimeout: resourceSchedulingPolicyUpdate,
		DeleteWithoutTimeout: resourceSchedulingPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fair_share_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_reservation": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 99),
						},
						"share_decay_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 604800),
						},
						"share_distribution": {
							Type: schema.TypeSet,
							// There can be no more than 500 fair share identifiers active in a job queue.
							MaxItems: 500,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"share_identifier": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validShareIdentifier,
									},
									"weight_factor": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatBetween(0.0001, 999.9999),
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSchedulingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &batch.CreateSchedulingPolicyInput{
		FairsharePolicy: expandFairsharePolicy(d.Get("fair_share_policy").([]any)),
		Name:            aws.String(name),
		Tags:            getTagsIn(ctx),
	}

	output, err := conn.CreateSchedulingPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Scheduling Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Arn))

	return append(diags, resourceSchedulingPolicyRead(ctx, d, meta)...)
}

func resourceSchedulingPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	sp, err := findSchedulingPolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Scheduling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Scheduling Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, sp.Arn)
	if err := d.Set("fair_share_policy", flattenFairsharePolicy(sp.FairsharePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting fair_share_policy: %s", err)
	}
	d.Set(names.AttrName, sp.Name)

	setTagsOut(ctx, sp.Tags)

	return diags
}

func resourceSchedulingPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	if d.HasChange("fair_share_policy") {
		input := &batch.UpdateSchedulingPolicyInput{
			Arn:             aws.String(d.Id()),
			FairsharePolicy: expandFairsharePolicy(d.Get("fair_share_policy").([]any)),
		}

		_, err := conn.UpdateSchedulingPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Scheduling Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSchedulingPolicyRead(ctx, d, meta)...)
}

func resourceSchedulingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	log.Printf("[DEBUG] Deleting Batch Scheduling Policy: %s", d.Id())
	input := batch.DeleteSchedulingPolicyInput{
		Arn: aws.String(d.Id()),
	}
	_, err := conn.DeleteSchedulingPolicy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Batch Scheduling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findSchedulingPolicyByARN(ctx context.Context, conn *batch.Client, arn string) (*awstypes.SchedulingPolicyDetail, error) {
	input := &batch.DescribeSchedulingPoliciesInput{
		Arns: []string{arn},
	}

	return findSchedulingPolicy(ctx, conn, input)
}

func findSchedulingPolicy(ctx context.Context, conn *batch.Client, input *batch.DescribeSchedulingPoliciesInput) (*awstypes.SchedulingPolicyDetail, error) {
	output, err := findSchedulingPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSchedulingPolicies(ctx context.Context, conn *batch.Client, input *batch.DescribeSchedulingPoliciesInput) ([]awstypes.SchedulingPolicyDetail, error) {
	output, err := conn.DescribeSchedulingPolicies(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SchedulingPolicies, nil
}

func expandFairsharePolicy(tfList []any) *awstypes.FairsharePolicy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FairsharePolicy{
		ComputeReservation: aws.Int32(int32(tfMap["compute_reservation"].(int))),
		ShareDecaySeconds:  aws.Int32(int32(tfMap["share_decay_seconds"].(int))),
	}

	for _, tfMapRaw := range tfMap["share_distribution"].(*schema.Set).List() {
		tfMap := tfMapRaw.(map[string]any)
		apiObject.ShareDistribution = append(apiObject.ShareDistribution, awstypes.ShareAttributes{
			ShareIdentifier: aws.String(tfMap["share_identifier"].(string)),
			WeightFactor:    flex.Float64ValueToFloat32(tfMap["weight_factor"].(float64)),
		})
	}

	return apiObject
}

func flattenFairsharePolicy(apiObject *awstypes.FairsharePolicy) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"compute_reservation": aws.ToInt32(apiObject.ComputeReservation),
		"share_decay_seconds": aws.ToInt32(apiObject.ShareDecaySeconds),
	}

	tfList := []any{}
	for _, apiObject := range apiObject.ShareDistribution {
		tfMap := map[string]any{
			"share_identifier": aws.ToString(apiObject.ShareIdentifier),
			"weight_factor":    flex.Float32ToFloat64Value(apiObject.WeightFactor),
		}
		tfList = append(tfList, tfMap)
	}
	tfMap["share_distribution"] = tfList

	return []any{tfMap}
}
