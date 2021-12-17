package autoscaling

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func buildFiltersDataSource(set *schema.Set) []*autoscaling.Filter {
	var filters []*autoscaling.Filter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}

		// In previous iterations, users were expected to provide "key" and "value" tag names.
		// With the addition of asgs filters, the signature is "tag-key" and "tag-value", so these conditions prevent breaking changes.
		// https://docs.aws.amazon.com/sdk-for-go/api/service/autoscaling/#Filter
		name := m["name"].(string)
		if name == "key" {
			name = "tag-key"
		}
		if name == "value" {
			name = "tag-value"
		}
		filters = append(filters, &autoscaling.Filter{
			Name:   aws.String(name),
			Values: filterValues,
		})
	}
	return filters
}

func dataSourceGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	log.Printf("[DEBUG] Reading Autoscaling Groups.")

	var rawName []string
	var rawArn []string
	var err error

	params := autoscaling.DescribeAutoScalingGroupsInput{}

	if v, ok := d.GetOk("filter"); ok {
		params.Filters = buildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("names"); ok {
		params.AutoScalingGroupNames = flex.ExpandStringList(v.([]interface{}))
	}

	err = conn.DescribeAutoScalingGroupsPages(&params, func(resp *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
		for _, group := range resp.AutoScalingGroups {
			rawName = append(rawName, aws.StringValue(group.AutoScalingGroupName))
			rawArn = append(rawArn, aws.StringValue(group.AutoScalingGroupARN))
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error fetching Autoscaling Groups: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	sort.Strings(rawName)
	sort.Strings(rawArn)

	if err := d.Set("names", rawName); err != nil {
		return fmt.Errorf("[WARN] Error setting Autoscaling Group Names: %w", err)
	}

	if err := d.Set("arns", rawArn); err != nil {
		return fmt.Errorf("[WARN] Error setting Autoscaling Group ARNs: %w", err)
	}

	return nil
}
