package rds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceEventCategories() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEventCategoriesRead,

		Schema: map[string]*schema.Schema{
			"source_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_categories": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceEventCategoriesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	req := &rds.DescribeEventCategoriesInput{}

	if sourceType := d.Get("source_type").(string); sourceType != "" {
		req.SourceType = aws.String(sourceType)
	}

	log.Printf("[DEBUG] Describe Event Categories %s\n", req)
	resp, err := conn.DescribeEventCategories(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.EventCategoriesMapList) == 0 {
		return fmt.Errorf("Event Categories not found")
	}

	eventCategories := make([]string, 0)

	for _, eventMap := range resp.EventCategoriesMapList {
		for _, v := range eventMap.EventCategories {
			eventCategories = append(eventCategories, aws.StringValue(v))
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	if err := d.Set("event_categories", eventCategories); err != nil {
		return fmt.Errorf("Error setting Event Categories: %w", err)
	}

	return nil

}
