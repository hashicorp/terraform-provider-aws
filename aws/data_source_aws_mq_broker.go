package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsMqBroker() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsmQBrokerRead,

		Schema: map[string]*schema.Schema{
			"broker_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"broker_name"},
			},
			"broker_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"broker_id"},
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revision": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"deployment_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"console_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"maintenance_window_start_time": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day_of_week": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time_of_day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"user": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      resourceAwsMqUserHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"console_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"groups": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsmQBrokerRead(d *schema.ResourceData, meta interface{}) error {
	if brokerId, ok := d.GetOk("broker_id"); ok {
		d.SetId(brokerId.(string))
	} else {
		brokerName := d.Get("broker_name").(string)
		brokerId := getBrokerId(meta, brokerName)
		if brokerId == "" {
			return fmt.Errorf("Failed to get broker id with name: %s", brokerName)
		}
		d.SetId(brokerId)
	}
	return resourceAwsMqBrokerRead(d, meta)
}

func getBrokerId(meta interface{}, name string) (id string) {
	conn := meta.(*AWSClient).mqconn
	var nextToken string
	for {
		out, err := conn.ListBrokers(&mq.ListBrokersInput{NextToken: aws.String(nextToken)})
		if err != nil {
			log.Printf("[DEBUG] Failed to list brokers: %s", err)
			return ""
		}
		for _, broker := range out.BrokerSummaries {
			if *broker.BrokerName == name {
				return *broker.BrokerId
			}
		}
		if out.NextToken == nil {
			break
		}
		nextToken = *out.NextToken
	}
	return ""
}
