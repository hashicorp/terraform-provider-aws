package batch

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSchedulingPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSchedulingPolicyRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fair_share_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_reservation": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"share_decay_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"share_distribution": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"share_identifier": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"weight_factor": {
										Type:     schema.TypeFloat,
										Computed: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSchedulingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).BatchConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	arn := d.Get("arn").(string)

	resp, err := conn.DescribeSchedulingPoliciesWithContext(ctx, &batch.DescribeSchedulingPoliciesInput{
		Arns: []*string{aws.String(arn)},
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error calling DescribeSchedulingPoliciesWithContext API for arn: %s", arn))
	}

	if len(resp.SchedulingPolicies) == 0 {
		return diag.FromErr(fmt.Errorf("no matches found for arn: %s", arn))
	}

	if len(resp.SchedulingPolicies) > 1 {
		return diag.FromErr(fmt.Errorf("multiple matches found for arn: %s", arn))
	}

	schedulingPolicy := resp.SchedulingPolicies[0]
	d.SetId(aws.StringValue(schedulingPolicy.Arn))
	d.Set("name", schedulingPolicy.Name)

	if err := d.Set("fair_share_policy", flattenFairsharePolicy(schedulingPolicy.FairsharePolicy)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting fair share policy: %s", err))
	}

	if err := d.Set("tags", KeyValueTags(schedulingPolicy.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	return nil
}
