package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsMskCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskClusterCreate,
		Read:   resourceAwsMskClusterRead,
		Update: resourceAwsMskClusterUpdate,
		Delete: resourceAwsMskClusterDelete,
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
			},
			"encryption_info": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"broker_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}
func resourceAwsMskClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
}
func resourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
}
func resourceAwsMskClusterUpdate(d *schema.ResourceData, meta interface{}) error {
}
func resourceAwsMskClusterDelete(d *schema.ResourceData, meta interface{}) error {
}
