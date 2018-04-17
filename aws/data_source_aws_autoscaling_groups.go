package aws

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsAutoscalingGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAutoscalingGroupsRead,

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
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsAutoscalingGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	log.Printf("[DEBUG] Reading Autoscaling Groups.")
	d.SetId(time.Now().UTC().String())

	var rawNames []string
	var rawArns []string

	tf := d.Get("filter").(*schema.Set)
	if tf.Len() > 0 {
		out, err := conn.DescribeTags(&autoscaling.DescribeTagsInput{
			Filters: expandAsgTagFilters(tf.List()),
		})
		if err != nil {
			return err
		}

		rawNames = make([]string, len(out.Tags))
		namesSlicePointer := make([]*string, len(out.Tags))
		for i, v := range out.Tags {
			rawNames[i] = *v.ResourceId
			namesSlicePointer[i] = v.ResourceId
		}

		// Fetch the group ARNs.
		rawArns = make([]string, len(namesSlicePointer))
		resp, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: namesSlicePointer,
		})
		if err != nil {
			return fmt.Errorf("Error fetching Autoscaling Groups: %s", err)
		}

		for i, v := range resp.AutoScalingGroups {
			rawArns[i] = *v.AutoScalingGroupARN
		}
	} else {

		resp, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
		if err != nil {
			return fmt.Errorf("Error fetching Autoscaling Groups: %s", err)
		}

		rawNames = make([]string, len(resp.AutoScalingGroups))
		rawArns = make([]string, len(resp.AutoScalingGroups))
		for i, v := range resp.AutoScalingGroups {
			rawNames[i] = *v.AutoScalingGroupName
			rawArns[i] = *v.AutoScalingGroupARN
		}
	}

	sort.Strings(rawNames)

	if err := d.Set("names", rawNames); err != nil {
		return fmt.Errorf("[WARN] Error setting Autoscaling Group Names: %s", err)
	}

	if err := d.Set("arns", rawArns); err != nil {
		return fmt.Errorf("[WARN] Error setting Autoscaling Group ARNs: %s", err)
	}

	return nil

}

func expandAsgTagFilters(in []interface{}) []*autoscaling.Filter {
	out := make([]*autoscaling.Filter, len(in), len(in))
	for i, filter := range in {
		m := filter.(map[string]interface{})
		values := expandStringList(m["values"].(*schema.Set).List())

		out[i] = &autoscaling.Filter{
			Name:   aws.String(m["name"].(string)),
			Values: values,
		}
	}
	return out
}
