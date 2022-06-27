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

func ResourceRoutingControl() *schema.Resource {
	return &schema.Resource{
		Create: resourceRoutingControlCreate,
		Read:   resourceRoutingControlRead,
		Update: resourceRoutingControlUpdate,
		Delete: resourceRoutingControlDelete,
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
			"control_panel_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRoutingControlCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.CreateRoutingControlInput{
		ClientToken:        aws.String(resource.UniqueId()),
		ClusterArn:         aws.String(d.Get("cluster_arn").(string)),
		RoutingControlName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("control_panel_arn"); ok {
		input.ControlPanelArn = aws.String(v.(string))
	}

	output, err := conn.CreateRoutingControl(input)
	result := output.RoutingControl

	if err != nil {
		return fmt.Errorf("error creating Route53 Recovery Control Config Routing Control: %w", err)
	}

	if result == nil {
		return fmt.Errorf("error creating Route53 Recovery Control Config Routing Control: empty response")
	}

	d.SetId(aws.StringValue(result.RoutingControlArn))

	if _, err := waitRoutingControlCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route53 Recovery Control Config Routing Control (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceRoutingControlRead(d, meta)
}

func resourceRoutingControlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.DescribeRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeRoutingControl(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Routing Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Route53 Recovery Control Config Routing Control: %w", err)
	}

	if output == nil || output.RoutingControl == nil {
		return fmt.Errorf("error describing Route53 Recovery Control Config Routing Control: %s", "empty response")
	}

	result := output.RoutingControl
	d.Set("arn", result.RoutingControlArn)
	d.Set("control_panel_arn", result.ControlPanelArn)
	d.Set("name", result.Name)
	d.Set("status", result.Status)

	return nil
}

func resourceRoutingControlUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.UpdateRoutingControlInput{
		RoutingControlName: aws.String(d.Get("name").(string)),
		RoutingControlArn:  aws.String(d.Get("arn").(string)),
	}

	_, err := conn.UpdateRoutingControl(input)

	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Routing Control: %s", err)
	}

	return resourceRoutingControlRead(d, meta)
}

func resourceRoutingControlDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Routing Control: %s", d.Id())
	_, err := conn.DeleteRoutingControl(&r53rcc.DeleteRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 Recovery Control Config Routing Control: %w", err)
	}

	_, err = waitRoutingControlDeleted(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Recovery Control Config Routing Control (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
