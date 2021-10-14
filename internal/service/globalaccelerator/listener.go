package globalaccelerator

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceListenerCreate,
		Read:   resourceListenerRead,
		Update: resourceListenerUpdate,
		Delete: resourceListenerDelete,

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
			"client_affinity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.ClientAffinityNone,
				ValidateFunc: validation.StringInSlice(globalaccelerator.ClientAffinity_Values(), false),
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(globalaccelerator.Protocol_Values(), false),
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

func resourceListenerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn
	acceleratorARN := d.Get("accelerator_arn").(string)

	opts := &globalaccelerator.CreateListenerInput{
		AcceleratorArn:   aws.String(acceleratorARN),
		ClientAffinity:   aws.String(d.Get("client_affinity").(string)),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Protocol:         aws.String(d.Get("protocol").(string)),
		PortRanges:       resourceAwsGlobalAcceleratorListenerExpandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Create Global Accelerator listener: %s", opts)

	resp, err := conn.CreateListener(opts)
	if err != nil {
		return fmt.Errorf("error creating Global Accelerator listener: %w", err)
	}

	d.SetId(aws.StringValue(resp.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return resourceListenerRead(d, meta)
}

func resourceListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn

	listener, err := FindListenerByARN(conn, d.Id())

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
	d.Set("client_affinity", listener.ClientAffinity)
	d.Set("protocol", listener.Protocol)
	if err := d.Set("port_range", resourceAwsGlobalAcceleratorListenerFlattenPortRanges(listener.PortRanges)); err != nil {
		return fmt.Errorf("error setting port_range: %w", err)
	}

	return nil
}

func resourceListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.UpdateListenerInput{
		ClientAffinity: aws.String(d.Get("client_affinity").(string)),
		ListenerArn:    aws.String(d.Id()),
		Protocol:       aws.String(d.Get("protocol").(string)),
		PortRanges:     resourceAwsGlobalAcceleratorListenerExpandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Updating Global Accelerator listener: %s", input)
	if _, err := conn.UpdateListener(input); err != nil {
		return fmt.Errorf("error updating Global Accelerator listener (%s): %w", d.Id(), err)
	}

	// Updating a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return resourceListenerRead(d, meta)
}

func resourceListenerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Global Accelerator listener (%s)", d.Id())
	_, err := conn.DeleteListener(input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeListenerNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Global Accelerator listener (%s): %w", d.Id(), err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return nil
}

func resourceAwsGlobalAcceleratorListenerExpandPortRanges(portRanges []interface{}) []*globalaccelerator.PortRange {
	out := make([]*globalaccelerator.PortRange, len(portRanges))

	for i, raw := range portRanges {
		portRange := raw.(map[string]interface{})
		m := globalaccelerator.PortRange{}

		m.FromPort = aws.Int64(int64(portRange["from_port"].(int)))
		m.ToPort = aws.Int64(int64(portRange["to_port"].(int)))

		out[i] = &m
	}

	return out
}

func resourceAwsGlobalAcceleratorListenerFlattenPortRanges(portRanges []*globalaccelerator.PortRange) []interface{} {
	out := make([]interface{}, len(portRanges))

	for i, portRange := range portRanges {
		m := make(map[string]interface{})

		m["from_port"] = aws.Int64Value(portRange.FromPort)
		m["to_port"] = aws.Int64Value(portRange.ToPort)

		out[i] = m
	}

	return out
}
