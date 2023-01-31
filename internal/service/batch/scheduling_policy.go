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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

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
			"arn": {
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSchedulingPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BatchConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	fairsharePolicy := expandFairsharePolicy(d.Get("fair_share_policy").([]interface{}))

	input := &batch.CreateSchedulingPolicyInput{
		FairsharePolicy: fairsharePolicy,
		Name:            aws.String(name),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateSchedulingPolicyWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Batch Scheduling Policy (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Batch Scheduling Policy (%s): empty output", name))
	}

	arn := aws.StringValue(output.Arn)
	log.Printf("[DEBUG] Scheduling Policy created: %s", arn)
	d.SetId(arn)

	return resourceSchedulingPolicyRead(ctx, d, meta)
}

func resourceSchedulingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BatchConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	sp, err := GetSchedulingPolicy(ctx, conn, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if sp == nil {
		log.Printf("[WARN] Batch Scheduling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", sp.Arn)
	d.Set("name", sp.Name)

	if err := d.Set("fair_share_policy", flattenFairsharePolicy(sp.FairsharePolicy)); err != nil {
		return diag.FromErr(err)
	}

	tags := KeyValueTags(sp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceSchedulingPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BatchConn()

	input := &batch.UpdateSchedulingPolicyInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("fair_share_policy") {
		fairsharePolicy := expandFairsharePolicy(d.Get("fair_share_policy").([]interface{}))
		input.FairsharePolicy = fairsharePolicy
	}

	_, err := conn.UpdateSchedulingPolicyWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("updating SchedulingPolicy (%s): %w", d.Id(), err))
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceSchedulingPolicyRead(ctx, d, meta)
}

func resourceSchedulingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BatchConn()

	log.Printf("[DEBUG] Deleting Batch Scheduling Policy: %s", d.Id())
	_, err := conn.DeleteSchedulingPolicyWithContext(ctx, &batch.DeleteSchedulingPolicyInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SchedulingPolicy (%s): %w", d.Id(), err))
	}

	return nil
}

func GetSchedulingPolicy(ctx context.Context, conn *batch.Batch, arn string) (*batch.SchedulingPolicyDetail, error) {
	resp, err := conn.DescribeSchedulingPoliciesWithContext(ctx, &batch.DescribeSchedulingPoliciesInput{
		Arns: []*string{aws.String(arn)},
	})
	if err != nil {
		return nil, err
	}

	numSchedulingPolicies := len(resp.SchedulingPolicies)
	switch {
	case numSchedulingPolicies == 0:
		log.Printf("[DEBUG] Scheduling Policy %q is already gone", arn)
		return nil, nil
	case numSchedulingPolicies == 1:
		return resp.SchedulingPolicies[0], nil
	case numSchedulingPolicies > 1:
		return nil, fmt.Errorf("Multiple Scheduling Policy with arn %s", arn)
	}
	return nil, nil
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
