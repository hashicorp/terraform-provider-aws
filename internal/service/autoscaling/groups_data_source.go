package autoscaling

import (
	"fmt"
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
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

	input := &autoscaling.DescribeAutoScalingGroupsInput{}

	if v, ok := d.GetOk("names"); ok && len(v.([]interface{})) > 0 {
		input.AutoScalingGroupNames = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildFiltersDataSource(v.(*schema.Set))
	}

	groups, err := findGroups(conn, input)

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Groups: %w", err)
	}

	var arns, names []string

	for _, group := range groups {
		arns = append(arns, aws.StringValue(group.AutoScalingGroupARN))
		names = append(names, aws.StringValue(group.AutoScalingGroupName))
	}

	sort.Strings(arns)
	sort.Strings(names)

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("names", names)

	return nil
}
