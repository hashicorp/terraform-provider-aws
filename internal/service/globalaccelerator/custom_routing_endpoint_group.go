package globalaccelerator

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCustomRoutingEndpointGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomRoutingEndpointGroupCreate,
		Read:   resourceCustomRoutingEndpointGroupRead,
		Delete: resourceCustomRoutingEndpointGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint_group_region": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"destination_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"protocols": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(ec2.TransportProtocol_Values(), false),
							},
						},
					},
				},
			},

			"endpoint_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},

			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceCustomRoutingEndpointGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn
	region := meta.(*conns.AWSClient).Region

	opts := &globalaccelerator.CreateCustomRoutingEndpointGroupInput{
		EndpointGroupRegion: aws.String(region),
		IdempotencyToken:    aws.String(resource.UniqueId()),
		ListenerArn:         aws.String(d.Get("listener_arn").(string)),
	}

	if v, ok := d.GetOk("destination_configuration"); ok {
		opts.DestinationConfigurations = expandGlobalAcceleratorCustomRoutingDestinationConfigurations(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Create Global Accelerator custom routing endpoint group: %s", opts)

	resp, err := conn.CreateCustomRoutingEndpointGroup(opts)

	if err != nil {
		return fmt.Errorf("error creating Global Accelerator custom routing endpoint group: %w", err)
	}

	d.SetId(aws.StringValue(resp.EndpointGroup.EndpointGroupArn))

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return err
	}

	if _, err := waitCustomRoutingAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		optsEndpoints := &globalaccelerator.AddCustomRoutingEndpointsInput{
			EndpointGroupArn:       resp.EndpointGroup.EndpointGroupArn,
			EndpointConfigurations: expandGlobalAcceleratorCustomRoutingEndpointConfigurations(v.(*schema.Set).List()),
		}

		_, err := conn.AddCustomRoutingEndpoints(optsEndpoints)
		if err != nil {
			return err
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", acceleratorARN, err)
		}
	}

	return resourceCustomRoutingEndpointGroupRead(d, meta)
}

func resourceCustomRoutingEndpointGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn

	endpointGroup, err := FindCustomRoutingEndpointGroupByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator custom routing endpoint group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator custom routing endpoint group (%s): %w", d.Id(), err)
	}

	listenerARN, err := EndpointGroupARNToListenerARN(d.Id())

	if err != nil {
		return err
	}

	d.Set("arn", endpointGroup.EndpointGroupArn)
	if err := d.Set("destination_configuration", flattenGlobalAcceleratorCustomRoutingDestinationDescriptions(endpointGroup.DestinationDescriptions)); err != nil {
		return fmt.Errorf("error setting destination_configuration: %w", err)
	}
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	if err := d.Set("endpoint_configuration", flattenGlobalAcceleratorCustomRoutingEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %w", err)
	}
	d.Set("listener_arn", listenerARN)

	return nil
}

func resourceCustomRoutingEndpointGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn

	input := &globalaccelerator.DeleteCustomRoutingEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Global Accelerator custom routing endpoint group (%s)", d.Id())
	_, err := conn.DeleteCustomRoutingEndpointGroup(input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Global Accelerator custom routing endpoint group (%s): %w", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return err
	}

	if _, err := waitCustomRoutingAcceleratorDeployed(conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", acceleratorARN, err)
	}

	return nil
}

func expandGlobalAcceleratorCustomRoutingDestinationConfigurations(configurations []interface{}) []*globalaccelerator.CustomRoutingDestinationConfiguration {
	out := make([]*globalaccelerator.CustomRoutingDestinationConfiguration, len(configurations))

	for i, raw := range configurations {
		configuration := raw.(map[string]interface{})
		m := globalaccelerator.CustomRoutingDestinationConfiguration{}

		m.FromPort = aws.Int64(int64(configuration["from_port"].(int)))
		m.ToPort = aws.Int64(int64(configuration["to_port"].(int)))
		m.Protocols = aws.StringSlice(configuration["protocols"].([]string))

		out[i] = &m
	}

	return out
}

func expandGlobalAcceleratorCustomRoutingEndpointConfigurations(configurations []interface{}) []*globalaccelerator.CustomRoutingEndpointConfiguration {
	out := make([]*globalaccelerator.CustomRoutingEndpointConfiguration, len(configurations))

	for i, raw := range configurations {
		configuration := raw.(map[string]interface{})
		m := globalaccelerator.CustomRoutingEndpointConfiguration{}

		m.EndpointId = aws.String(configuration["endpoint_id"].(string))

		out[i] = &m
	}

	return out
}

func flattenGlobalAcceleratorCustomRoutingEndpointDescriptions(configurations []*globalaccelerator.CustomRoutingEndpointDescription) []interface{} {
	out := make([]interface{}, len(configurations))

	for i, configuration := range configurations {
		m := make(map[string]interface{})

		m["endpoint_id"] = aws.StringValue(configuration.EndpointId)

		out[i] = m
	}

	return out
}

func flattenGlobalAcceleratorCustomRoutingDestinationDescriptions(configurations []*globalaccelerator.CustomRoutingDestinationDescription) []interface{} {
	out := make([]interface{}, len(configurations))

	for i, configuration := range configurations {
		m := make(map[string]interface{})

		m["from_port"] = int(aws.Int64Value(configuration.FromPort))
		m["to_port"] = int(aws.Int64Value(configuration.ToPort))
		m["protocols"] = aws.StringValueSlice(configuration.Protocols)

		out[i] = m
	}

	return out
}
