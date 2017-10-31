package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func dataSourceAwsLbListener() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbListenerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"load_balancer_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ssl_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"default_action": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsLbListenerRead(d *schema.ResourceData, meta interface{}) error {
	listenerArn := d.Get("arn").(string)
	lbArn := d.Get("load_balancer_arn").(string)

	switch {
	case listenerArn != "":
		d.SetId(d.Get("arn").(string))
	case lbArn != "":
		elbconn := meta.(*AWSClient).elbv2conn

		resp, err := elbconn.DescribeListeners(&elbv2.DescribeListenersInput{
			LoadBalancerArn: aws.String(lbArn),
		})

		if err != nil {
			if isListenerNotFound(err) {
				log.Printf("[WARN] DescribeListeners - removing %s from state", d.Id())
				d.SetId("")
				return nil
			}
			return errwrap.Wrapf("Error retrieving Listener: {{err}}", err)
		}

		if len(resp.Listeners) != 1 {
			return fmt.Errorf("Multiple listeners found for load balancer %s. This data source only supports single listener load balancers when searching by load balancer.", lbArn)
		}

		d.SetId(*resp.Listeners[0].ListenerArn)
	}

	return resourceAwsLbListenerRead(d, meta)

}
