package aws

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLbListener() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbListenerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"load_balancer_arn", "port"},
			},
			"load_balancer_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"arn"},
			},
			"port": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"arn"},
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
			"last_rule_priority": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"next_rule_priority": {
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
		},
	}
}

func dataSourceAwsLbListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	var listenerArn string

	if v, ok := d.GetOk("arn"); ok {
		listenerArn = v.(string)
		log.Printf("[DEBUG] read listener %s", d.Get("arn").(string))
	} else {
		lbArn, lbOk := d.GetOk("load_balancer_arn")
		port, portOk := d.GetOk("port")
		if !lbOk || !portOk {
			return errors.New("both load_balancer_arn and port must be set")
		}
		resp, err := conn.DescribeListeners(&elbv2.DescribeListenersInput{
			LoadBalancerArn: aws.String(lbArn.(string)),
		})
		if err != nil {
			return err
		}
		if len(resp.Listeners) == 0 {
			return fmt.Errorf("[DEBUG] no listener exists for load balancer: %s", lbArn)
		}
		for _, listener := range resp.Listeners {
			if *listener.Port == int64(port.(int)) {
				listenerArn = *listener.ListenerArn
				log.Printf("[DEBUG] read listener for %s:%s", lbArn, port)
			}
		}
	}

	if listenerArn == "" {
		return errors.New("failed to get listener arn with given arguments")
	}

	lastPriority, nextPriority, err := getListenerRulePriority(conn, listenerArn)
	if err != nil {
		return err
	}

	d.SetId(listenerArn)
	d.Set("last_rule_priority", lastPriority)
	d.Set("next_rule_priority", nextPriority)
	return resourceAwsLbListenerRead(d, meta)
}

func getListenerRulePriority(conn *elbv2.ELBV2, arn string) (last, next int64, err error) {
	var priorities []int
	var nextMarker *string

	for {
		out, aerr := conn.DescribeRules(&elbv2.DescribeRulesInput{
			ListenerArn: aws.String(arn),
			Marker:      nextMarker,
		})
		if aerr != nil {
			err = aerr
			return
		}
		for _, rule := range out.Rules {
			if *rule.Priority != "default" {
				p, _ := strconv.Atoi(*rule.Priority)
				priorities = append(priorities, p)
			}
		}
		if out.NextMarker == nil {
			break
		}
		nextMarker = out.NextMarker
	}

	l := len(priorities)
	if l == 0 || l == priorities[l-1] {
		last = int64(l)
		next = int64(l + 1)
	} else {
		last = int64(priorities[l-1])
		sort.IntSlice(priorities).Sort()
		for i, p := range priorities {
			if i+1 != p {
				next = int64(i + 1)
			}
		}
	}

	return
}
