package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsDbProxyDefaultTargetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbProxyDefaultTargetGroupUpdate,
		Read:   resourceAwsDbProxyDefaultTargetGroupRead,
		Update: resourceAwsDbProxyDefaultTargetGroupUpdate,
		Delete: resourceAwsDbProxyDefaultTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"connection_pool_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_borrow_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"init_query": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max_connections_percent": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_idle_connections_percent": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"session_pinning_filters": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceAwsDbProxyDefaultTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	tg, err := resourceAwsDbProxyDefaultTargetGroupGet(conn, d.Id())

	if err != nil {
		if isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			log.Printf("[WARN] DB Proxy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if tg == nil {
		log.Printf("[WARN] DB Proxy default target group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", tg.TargetGroupArn)
	d.Set("db_proxy_name", tg.DBProxyName)
	d.Set("name", tg.TargetGroupName)

	cpc := tg.ConnectionPoolConfig

	d.Set("connection_borrow_timeout", cpc.ConnectionBorrowTimeout)
	d.Set("init_query", cpc.InitQuery)
	d.Set("max_connections_percent", cpc.MaxConnectionsPercent)
	d.Set("max_idle_connections_percent", cpc.MaxIdleConnectionsPercent)
	d.Set("session_pinning_filters", flattenStringSet(cpc.SessionPinningFilters))

	return nil
}

func resourceAwsDbProxyDefaultTargetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	oName, nName := d.GetChange("name")

	params := rds.ModifyDBProxyTargetGroupInput{
		DBProxyName:     aws.String(d.Get("db_proxy_name").(string)),
		TargetGroupName: aws.String(oName.(string)),
		NewName:         aws.String(nName.(string)),
	}

	if v, ok := d.GetOk("connection_pool_config"); ok {
		params.ConnectionPoolConfig = expandDbProxyConnectionPoolConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Update DB Proxy default target group: %#v", params)
	_, err := conn.ModifyDBProxyTargetGroup(&params)
	if err != nil {
		return fmt.Errorf("Error updating DB Proxy default target group: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusModifying},
		Target:  []string{rds.DBProxyStatusAvailable},
		Refresh: resourceAwsDbProxyDefaultTargetGroupRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateChangeConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for DB Proxy default target group update: %s", err)
	}

	return resourceAwsDbProxyDefaultTargetGroupRead(d, meta)
}

func expandDbProxyConnectionPoolConfig(configs []interface{}) *rds.ConnectionPoolConfiguration {
	if len(configs) < 1 {
		return nil
	}

	config := configs[0].(map[string]interface{})

	result := &rds.ConnectionPoolConfiguration{
		ConnectionBorrowTimeout:   aws.Int64(int64(config["connection_borrow_timeout"].(int))),
		InitQuery:                 aws.String(config["init_query"].(string)),
		MaxConnectionsPercent:     aws.Int64(int64(config["max_connections_percent"].(int))),
		MaxIdleConnectionsPercent: aws.Int64(int64(config["max_idle_connections_percent"].(int))),
		SessionPinningFilters:     expandStringSet(config["session_pinning_filters"].(*schema.Set)),
	}

	return result
}

func resourceAwsDbProxyDefaultTargetGroupGet(conn *rds.RDS, proxyName string) (*rds.DBProxyTargetGroup, error) {
	params := &rds.DescribeDBProxyTargetGroupsInput{
		DBProxyName: aws.String(proxyName),
	}

	resp, err := conn.DescribeDBProxyTargetGroups(params)

	if err != nil {
		return nil, err
	}

	// Return default target group
	for _, tg := range resp.TargetGroups {
		if *tg.IsDefault {
			return tg, nil
		}
	}

	return nil, nil
}

func resourceAwsDbProxyDefaultTargetGroupRefreshFunc(conn *rds.RDS, proxyName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		tg, err := resourceAwsDbProxyDefaultTargetGroupGet(conn, proxyName)

		if err != nil {
			if isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
				return 42, "", nil
			}
			return 42, "", err
		}

		return tg, *tg.Status, nil
	}
}

func resourceAwsDbProxyDefaultTargetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy DB Proxy default target group. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
