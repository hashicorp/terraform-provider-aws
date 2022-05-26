package route53recoverycontrolconfig

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceControlPanel() *schema.Resource {
	return &schema.Resource{
		Create: resourceControlPanelCreate,
		Read:   resourceControlPanelRead,
		Update: resourceControlPanelUpdate,
		Delete: resourceControlPanelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_control_panel": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_control_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceControlPanelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.CreateControlPanelInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ClusterArn:       aws.String(d.Get("cluster_arn").(string)),
		ControlPanelName: aws.String(d.Get("name").(string)),
	}

	output, err := conn.CreateControlPanel(input)
	result := output.ControlPanel

	if err != nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Control Panel: %w", err)
	}

	if result == nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Control Panel: empty response")
	}

	d.SetId(aws.StringValue(result.ControlPanelArn))

	if _, err := waitControlPanelCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Control Panel (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceControlPanelRead(d, meta)
}

func resourceControlPanelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.DescribeControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeControlPanel(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Control Panel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Control Panel: %s", err)
	}

	if output == nil || output.ControlPanel == nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Control Panel: %s", "empty response")
	}

	result := output.ControlPanel
	d.Set("arn", result.ControlPanelArn)
	d.Set("cluster_arn", result.ClusterArn)
	d.Set("default_control_panel", result.DefaultControlPanel)
	d.Set("name", result.Name)
	d.Set("routing_control_count", result.RoutingControlCount)
	d.Set("status", result.Status)

	return nil
}

func resourceControlPanelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.UpdateControlPanelInput{
		ControlPanelName: aws.String(d.Get("name").(string)),
		ControlPanelArn:  aws.String(d.Get("arn").(string)),
	}

	_, err := conn.UpdateControlPanel(input)

	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Control Panel: %s", err)
	}

	return resourceControlPanelRead(d, meta)
}

func resourceControlPanelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Control Panel: %s", d.Id())
	_, err := conn.DeleteControlPanel(&r53rcc.DeleteControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 Recovery Control Config Control Panel: %w", err)
	}

	_, err = waitControlPanelDeleted(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Control Panel (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
