package ssm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourcePatchBaseline() *schema.Resource {
	return &schema.Resource{
		Read: dataPatchBaselineRead,
		Schema: map[string]*schema.Schema{
			"approved_patches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"approved_patches_compliance_level": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"approved_patches_enable_non_security": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"approval_rule": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"approve_after_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"approve_until_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"compliance_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_non_security": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"patch_filter": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"values": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"default_baseline": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_filter": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"values": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"operating_system": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ssm.OperatingSystem_Values(), false),
			},
			"owner": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"rejected_patches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rejected_patches_action": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"configuration": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"products": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataPatchBaselineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	filters := []*ssm.PatchOrchestratorFilter{
		{
			Key: aws.String("OWNER"),
			Values: []*string{
				aws.String(d.Get("owner").(string)),
			},
		},
	}

	if v, ok := d.GetOk("name_prefix"); ok {
		filters = append(filters, &ssm.PatchOrchestratorFilter{
			Key: aws.String("NAME_PREFIX"),
			Values: []*string{
				aws.String(v.(string)),
			},
		})
	}

	params := &ssm.DescribePatchBaselinesInput{
		Filters: filters,
	}

	log.Printf("[DEBUG] Reading DescribePatchBaselines: %s", params)

	resp, err := conn.DescribePatchBaselines(params)

	if err != nil {
		return fmt.Errorf("Error describing SSM PatchBaselines: %w", err)
	}

	var filteredBaselines []*ssm.PatchBaselineIdentity
	if v, ok := d.GetOk("operating_system"); ok {
		for _, baseline := range resp.BaselineIdentities {
			if v.(string) == aws.StringValue(baseline.OperatingSystem) {
				filteredBaselines = append(filteredBaselines, baseline)
			}
		}
	}

	if v, ok := d.GetOk("default_baseline"); ok {
		for _, baseline := range filteredBaselines {
			if v.(bool) == aws.BoolValue(baseline.DefaultBaseline) {
				filteredBaselines = []*ssm.PatchBaselineIdentity{baseline}
				break
			}
		}
	}

	if len(filteredBaselines) < 1 || filteredBaselines[0] == nil {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredBaselines) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria")
	}

	baseline := filteredBaselines[0]

	input := &ssm.GetPatchBaselineInput{
		BaselineId: baseline.BaselineId,
	}

	output, err := conn.GetPatchBaseline(input)

	if err != nil {
		return fmt.Errorf("Error getting SSM PatchBaseline: %w", err)
	}

	d.SetId(aws.StringValue(baseline.BaselineId))
	d.Set("approved_patches", aws.StringValueSlice(output.ApprovedPatches))
	d.Set("approved_patches_compliance_level", output.ApprovedPatchesComplianceLevel)
	d.Set("approved_patches_enable_non_security", output.ApprovedPatchesEnableNonSecurity)
	d.Set("approval_rule", flattenPatchRuleGroup(output.ApprovalRules))
	d.Set("default_baseline", baseline.DefaultBaseline)
	d.Set("description", baseline.BaselineDescription)
	d.Set("global_filter", flattenPatchFilterGroup(output.GlobalFilters))
	d.Set("name", baseline.BaselineName)
	d.Set("operating_system", baseline.OperatingSystem)
	d.Set("rejected_patches", aws.StringValueSlice(output.RejectedPatches))
	d.Set("rejected_patches_action", output.RejectedPatchesAction)
	d.Set("source", flattenPatchSource(output.Sources))

	return nil
}
