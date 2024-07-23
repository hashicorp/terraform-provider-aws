// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_scheduling_policy", name="Scheduling Policy")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/batch;batch.SchedulingPolicyDetail")
func ResourceSchedulingPolicy() *schema.Resource {
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
							Set: func(v interface{}) int {
								var buf bytes.Buffer
								m := v.(map[string]interface{})
								buf.WriteString(m["share_identifier"].(string))
								if v, ok := m["weight_factor"]; ok {
									buf.WriteString(fmt.Sprintf("%s-", v))
								}
								return create.StringHashcode(buf.String())
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSchedulingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &batch.CreateSchedulingPolicyInput{
		FairsharePolicy: expandFairsharePolicy(d.Get("fair_share_policy").([]interface{})),
		Name:            aws.String(name),
		Tags:            getTagsIn(ctx),
	}

	output, err := conn.CreateSchedulingPolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Scheduling Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Arn))

	return append(diags, resourceSchedulingPolicyRead(ctx, d, meta)...)
}

func resourceSchedulingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	sp, err := FindSchedulingPolicyByARN(ctx, conn, d.Id())

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

func resourceSchedulingPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	if d.HasChange("fair_share_policy") {
		input := &batch.UpdateSchedulingPolicyInput{
			Arn:             aws.String(d.Id()),
			FairsharePolicy: expandFairsharePolicy(d.Get("fair_share_policy").([]interface{})),
		}

		_, err := conn.UpdateSchedulingPolicyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Scheduling Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSchedulingPolicyRead(ctx, d, meta)...)
}

func resourceSchedulingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	log.Printf("[DEBUG] Deleting Batch Scheduling Policy: %s", d.Id())
	_, err := conn.DeleteSchedulingPolicyWithContext(ctx, &batch.DeleteSchedulingPolicyInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Batch Scheduling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindSchedulingPolicyByARN(ctx context.Context, conn *batch.Batch, arn string) (*batch.SchedulingPolicyDetail, error) {
	input := &batch.DescribeSchedulingPoliciesInput{
		Arns: aws.StringSlice([]string{arn}),
	}

	output, err := findSchedulingPolicy(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findSchedulingPolicy(ctx context.Context, conn *batch.Batch, input *batch.DescribeSchedulingPoliciesInput) (*batch.SchedulingPolicyDetail, error) {
	output, err := conn.DescribeSchedulingPoliciesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SchedulingPolicies) == 0 || output.SchedulingPolicies[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SchedulingPolicies); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.SchedulingPolicies[0], nil
}

func expandFairsharePolicy(fairsharePolicy []interface{}) *batch.FairsharePolicy {
	if len(fairsharePolicy) == 0 || fairsharePolicy[0] == nil {
		return nil
	}

	tfMap, ok := fairsharePolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &batch.FairsharePolicy{
		ComputeReservation: aws.Int64(int64(tfMap["compute_reservation"].(int))),
		ShareDecaySeconds:  aws.Int64(int64(tfMap["share_decay_seconds"].(int))),
	}

	shareDistributions := tfMap["share_distribution"].(*schema.Set).List()

	fairsharePolicyShareDistributions := []*batch.ShareAttributes{}

	for _, shareDistribution := range shareDistributions {
		data := shareDistribution.(map[string]interface{})

		schedulingPolicyConfig := &batch.ShareAttributes{
			ShareIdentifier: aws.String(data["share_identifier"].(string)),
			WeightFactor:    aws.Float64(data["weight_factor"].(float64)),
		}
		fairsharePolicyShareDistributions = append(fairsharePolicyShareDistributions, schedulingPolicyConfig)
	}

	result.ShareDistribution = fairsharePolicyShareDistributions

	return result
}

func flattenFairsharePolicy(fairsharePolicy *batch.FairsharePolicy) []interface{} {
	if fairsharePolicy == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"compute_reservation": aws.Int64Value(fairsharePolicy.ComputeReservation),
		"share_decay_seconds": aws.Int64Value(fairsharePolicy.ShareDecaySeconds),
	}

	shareDistributions := fairsharePolicy.ShareDistribution

	fairsharePolicyShareDistributions := []interface{}{}
	for _, shareDistribution := range shareDistributions {
		sdValues := map[string]interface{}{
			"share_identifier": aws.StringValue(shareDistribution.ShareIdentifier),
			"weight_factor":    aws.Float64Value(shareDistribution.WeightFactor),
		}
		fairsharePolicyShareDistributions = append(fairsharePolicyShareDistributions, sdValues)
	}
	values["share_distribution"] = fairsharePolicyShareDistributions

	return []interface{}{values}
}
