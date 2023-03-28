package globalaccelerator

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_globalaccelerator_custom_routing_listener")
func ResourceCustomRoutingListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomRoutingListenerCreate,
		Read:   resourceCustomRoutingListenerRead,
		Update: resourceCustomRoutingListenerUpdate,
		Delete: resourceCustomRoutingListenerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"accelerator_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_range": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
		},
	}
}

func resourceCustomRoutingListenerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	acceleratorARN := d.Get("accelerator_arn").(string)

	opts := &globalaccelerator.CreateCustomRoutingListenerInput{
		AcceleratorArn:   aws.String(acceleratorARN),
		IdempotencyToken: aws.String(resource.UniqueId()),
		PortRanges:       expandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Create Global Accelerator custom routing listener: %s", opts)

	resp, err := conn.CreateCustomRoutingListener(opts)
	if err != nil {
		return fmt.Errorf("error creating Global Accelerator custom routing listener: %w", err)
	}

	d.SetId(aws.StringValue(resp.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return resourceCustomRoutingListenerRead(d, meta)
}

func resourceCustomRoutingListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	listener, err := FindCustomRoutingListenerByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator listener (%s): %w", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return err
	}

	d.Set("accelerator_arn", acceleratorARN)
	if err := d.Set("port_range", flattenPortRanges(listener.PortRanges)); err != nil {
		return fmt.Errorf("error setting port_range: %w", err)
	}

	return nil
}

func resourceCustomRoutingListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.UpdateCustomRoutingListenerInput{
		ListenerArn: aws.String(d.Id()),
		PortRanges:  expandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Updating Global Accelerator custom routing listener: %s", input)
	if _, err := conn.UpdateCustomRoutingListener(input); err != nil {
		return fmt.Errorf("error updating Global Accelerator custom routing listener (%s): %w", d.Id(), err)
	}

	// Updating a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return resourceCustomRoutingListenerRead(d, meta)
}

func resourceCustomRoutingListenerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.DeleteCustomRoutingListenerInput{
		ListenerArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Global Accelerator custom routing listener (%s)", d.Id())
	_, err := conn.DeleteCustomRoutingListener(input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeListenerNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Global Accelerator custom routing listener (%s): %w", d.Id(), err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending.
	if _, err := waitCustomRoutingAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return nil
}
