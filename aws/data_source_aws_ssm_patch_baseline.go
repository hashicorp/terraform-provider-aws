package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsSsmPatchBaseline() *schema.Resource {
	return &schema.Resource{
		Read: dataAwsSsmPatchBaselineRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"default_baseline": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"operating_system": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			// Computed values
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataAwsSsmPatchBaselineRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	params := &ssm.DescribePatchBaselinesInput{}
	if v, ok := d.GetOk("filter"); ok {
		params.Filters = buildPatchBaselineFilters(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading DescribePatchBaselines: %s", params)

	resp, err := ssmconn.DescribePatchBaselines(params)

	if err != nil {
		return fmt.Errorf("Error describing SSM PatchBaselines: %s", err)
	}

	var filteredBaselines []*ssm.PatchBaselineIdentity
	if os, ok := d.GetOk("operating_system"); ok {
		for _, baseline := range resp.BaselineIdentities {
			if os.(string) == *baseline.OperatingSystem {
				filteredBaselines = append(filteredBaselines, baseline)
			}
		}
	}

	if db, ok := d.GetOk("default_baseline"); ok {
		var ln int
		for _, baseline := range filteredBaselines {
			if db.(bool) == *baseline.DefaultBaseline {
				filteredBaselines[ln] = baseline
				ln++
			}
		}
		filteredBaselines = filteredBaselines[:ln]
	}

	if len(filteredBaselines) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredBaselines) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria")
	}

	baseline := *filteredBaselines[0]

	d.SetId(*baseline.BaselineId)
	d.Set("name", baseline.BaselineName)
	d.Set("description", baseline.BaselineDescription)
	d.Set("default_baseline", baseline.DefaultBaseline)
	d.Set("operating_system", baseline.OperatingSystem)

	return nil
}

func buildPatchBaselineFilters(set *schema.Set) []*ssm.PatchOrchestratorFilter {
	var filters []*ssm.PatchOrchestratorFilter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		filters = append(filters, &ssm.PatchOrchestratorFilter{
			Key:    aws.String(m["name"].(string)),
			Values: filterValues,
		})
	}
	return filters
}
