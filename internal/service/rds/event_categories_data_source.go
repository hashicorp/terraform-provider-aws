package rds

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceEventCategories() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEventCategoriesRead,

		Schema: map[string]*schema.Schema{
			"event_categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"source_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(rds.SourceType_Values(), false),
			},
		},
	}
}

func dataSourceEventCategoriesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.DescribeEventCategoriesInput{}

	if v, ok := d.GetOk("source_type"); ok {
		input.SourceType = aws.String(v.(string))
	}

	output, err := findEventCategoriesMaps(conn, input)

	if err != nil {
		return fmt.Errorf("error reading RDS Event Categories: %w", err)
	}

	var eventCategories []string

	for _, v := range output {
		eventCategories = append(eventCategories, aws.StringValueSlice(v.EventCategories)...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("event_categories", eventCategories)

	return nil
}

func findEventCategoriesMaps(conn *rds.RDS, input *rds.DescribeEventCategoriesInput) ([]*rds.EventCategoriesMap, error) {
	var output []*rds.EventCategoriesMap

	page, err := conn.DescribeEventCategories(input)

	if err != nil {
		return nil, err
	}

	for _, v := range page.EventCategoriesMapList {
		if v != nil {
			output = append(output, v)
		}
	}

	return output, nil
}
