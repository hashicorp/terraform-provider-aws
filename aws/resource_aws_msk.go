package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsMsk() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskCreate,
		Read:   resourceAwsMskRead,
		Update: resourceAwsMskUpdate,
		Delete: resourceAwsMskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"broker_node_group_info": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encryption_info": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsMskCreate(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsMskRead(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsMskUpdate(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsMskDelete(d *schema.ResourceData, meta interface{}) error {
}
