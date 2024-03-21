package resiliencehub

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/service/resiliencehub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceResiliencyPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceResiliencyPolicyCreate,
		ReadContext:   ResourceResiliencyPolicyRead,
		UpdateContext: ResourceResiliencyPolicyUpdate,
		DeleteContext: ResourceResiliencyPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z\d][A-Za-z\d_\\-]{1,59}$`), "Up to 60 alphanumeric characters, or hyphens, without spaces. The first character must be a letter or a number."),
			},
			"policy_description": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			"policy": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"AZ":       failurePolicy(true),
						"Hardware": failurePolicy(true),
						"Software": failurePolicy(true),
						"Region":   failurePolicy(false),
					},
				},
			},
			"data_location_constraint": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(resiliencehub.DataLocationConstraint_Values(), false),
			},
			"tier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resiliencehub.ResiliencyPolicyTier_Values(), false),
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func ResourceResiliencyPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResilienceHubConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	policy := expandPolicy(d.Get("policy").([]interface{}))

	input := resiliencehub.CreateResiliencyPolicyInput{
		PolicyName: aws.String(d.Get("name").(string)),
		Tier:       aws.String(d.Get("tier").(string)),
		Tags:       Tags(tags.IgnoreAWS()),
		Policy:     policy,
	}

	if v, ok := d.GetOk("policy_description"); ok {
		input.PolicyDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_location_constraint"); ok {
		input.DataLocationConstraint = aws.String(v.(string))
	}

	res, err := conn.CreateResiliencyPolicy(&input)
	if err != nil {
		return diag.Errorf("error creating resiliency policy: %s", err)
	}

	d.SetId(*res.Policy.PolicyArn)
	d.Set("arn", *res.Policy.PolicyArn)
	return ResourceResiliencyPolicyRead(ctx, d, meta)
}

func ResourceResiliencyPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResilienceHubConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading ResilienceHub Policy %s", d.Id())
	input := resiliencehub.DescribeResiliencyPolicyInput{
		PolicyArn: aws.String(d.Get("arn").(string)),
	}

	resp, err := conn.DescribeResiliencyPolicy(&input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, resiliencehub.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] Resilience Hub Resiliency Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.Errorf("error reading resiliency policy (%s): %s", d.Id(), err)
	}

	p := *resp.Policy

	d.Set("arn", p.PolicyArn)
	d.Set("name", p.PolicyName)
	d.Set("tier", p.Tier)
	d.Set("data_location_constraint", p.DataLocationConstraint)
	d.Set("policy_description", p.PolicyDescription)
	d.Set("policy", flattenPolicy(p.Policy))

	tags := KeyValueTags(p.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}

func ResourceResiliencyPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func ResourceResiliencyPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func failurePolicy(required bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Required: required,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rpoInSecs": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(0),
				},
				"rtoInSecs": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(0),
				},
			},
		},
	}
}

func expandPolicy(d []interface{}) map[string]*resiliencehub.FailurePolicy {
	if len(d) == 0 || d[0] == nil {
		return nil
	}

	policyAsMap := d[0].(map[string]interface{})

	failurePolicy := map[string]*resiliencehub.FailurePolicy{
		"AZ":       expandFailurePolicy(policyAsMap, "AZ"),
		"Software": expandFailurePolicy(policyAsMap, "Software"),
		"Hardware": expandFailurePolicy(policyAsMap, "Hardware"),
	}

	if _, ok := policyAsMap["Region"]; ok {
		failurePolicy["Region"] = expandFailurePolicy(policyAsMap, "Region")
	}

	return failurePolicy
}

func expandFailurePolicy(policyMap map[string]interface{}, key string) *resiliencehub.FailurePolicy {
	fPolicy := policyMap[key].(map[string]interface{})

	rtoInSecs := fPolicy["rtoInSecs"].(int64)
	rpoInSecs := fPolicy["rtoInSecs"].(int64)

	return &resiliencehub.FailurePolicy{
		RpoInSecs: &rpoInSecs,
		RtoInSecs: &rtoInSecs,
	}
}

func flattenPolicy(policy map[string]*resiliencehub.FailurePolicy) map[string]interface{} {
	var tfMap map[string]interface{}

	azFailurePolicy := policy["AZ"]
	hardwareFailurePolicy := policy["Hardware"]
	softwareFailurePolicy := policy["Software"]

	tfMap["AZ"] = flattenFailurePolicy(azFailurePolicy)
	tfMap["Hardware"] = flattenFailurePolicy(hardwareFailurePolicy)
	tfMap["Software"] = flattenFailurePolicy(softwareFailurePolicy)

	if regionFailurePolicy, ok := policy["Region"]; ok {
		tfMap["Region"] = flattenFailurePolicy(regionFailurePolicy)
	}

	return tfMap
}

func flattenFailurePolicy(failurePolicy *resiliencehub.FailurePolicy) map[string]interface{} {
	var tfMap map[string]interface{}

	rtoInSecs := failurePolicy.RtoInSecs
	rpoInSecs := failurePolicy.RpoInSecs

	tfMap["rtoInSecs"] = rtoInSecs
	tfMap["rpoInSecs"] = rpoInSecs

	return tfMap
}
