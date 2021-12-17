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
		CreateContext: resourceSchedulingPolicyCreate,
		ReadContext:   resourceSchedulingPolicyRead,
		UpdateContext: resourceSchedulingPolicyUpdate,
		DeleteContext: resourceSchedulingPolicyDelete,
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
										Type:     schema.TypeString,
										Required: true,
										// limited to 255 alphanumeric characters, optionally followed by an asterisk (*).
										ValidateFunc: validation.StringLenBetween(1, 256),
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
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
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
			WeightFactor:    aws.Float64(float64(data["weight_factor"].(float64))),
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
