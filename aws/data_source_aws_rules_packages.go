package aws

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsRulesPackages() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRulesPackagesRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsRulesPackagesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).inspectorconn

	log.Printf("[DEBUG] Reading Rules Packages.")

	var results int64 = 300
	request := &inspector.ListRulesPackagesInput{
		MaxResults: &results,
	}

	log.Printf("[DEBUG] Reading Rules Packages: %s", request)

	resp, err := conn.ListRulesPackages(request)
	if err != nil {
		return fmt.Errorf("Error fetching Rules Packages: %s", err)
	}

	raw := make([]string, len(resp.RulesPackageArns))
	for i, v := range resp.RulesPackageArns {
		raw[i] = *v
	}

	sort.Strings(raw)

	log.Printf("[DEBUG] Output is: %s", raw)

	if err := d.Set("arns", raw); err != nil {
		return fmt.Errorf("[WARN] Error setting Rules Packages: %s", err)
	}

	return nil
}
