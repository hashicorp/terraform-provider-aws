package globalaccelerator

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_globalaccelerator_custom_routing_endpoint_group")
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
								ValidateFunc: validation.StringInSlice(globalaccelerator.CustomRoutingProtocol_Values(), false),
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
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	region := meta.(*conns.AWSClient).Region

	opts := &globalaccelerator.CreateCustomRoutingEndpointGroupInput{
		EndpointGroupRegion: aws.String(region),
		IdempotencyToken:    aws.String(resource.UniqueId()),
		ListenerArn:         aws.String(d.Get("listener_arn").(string)),
	}

	if v, ok := d.GetOk("destination_configuration"); ok {
		opts.DestinationConfigurations = expandCustomRoutingDestinationConfigurations(v.(*schema.Set).List())
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
			EndpointConfigurations: expandCustomRoutingEndpointConfigurations(v.(*schema.Set).List()),
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
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

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
	if err := d.Set("destination_configuration", flattenCustomRoutingDestinationDescriptions(endpointGroup.DestinationDescriptions)); err != nil {
		return fmt.Errorf("error setting destination_configuration: %w", err)
	}
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	if err := d.Set("endpoint_configuration", flattenCustomRoutingEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %w", err)
	}
	d.Set("listener_arn", listenerARN)

	return nil
}

func resourceCustomRoutingEndpointGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

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

func expandCustomRoutingDestinationConfigurations(configurations []interface{}) []*globalaccelerator.CustomRoutingDestinationConfiguration {
	if len(configurations) == 0 {
		return nil
	}

	var apiObjects []*globalaccelerator.CustomRoutingDestinationConfiguration

	for _, tfMapRaw := range configurations {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCustomRoutingEndpointDestinationConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCustomRoutingEndpointDestinationConfiguration(tfMap map[string]interface{}) *globalaccelerator.CustomRoutingDestinationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.CustomRoutingDestinationConfiguration{}

	if v, ok := tfMap["from_port"].(int64); ok {
		apiObject.FromPort = aws.Int64(v)
	}

	if v, ok := tfMap["to_port"].(int64); ok {
		apiObject.ToPort = aws.Int64(v)
	}

	if v, ok := tfMap["protocols"].(*schema.Set); ok {
		apiObject.Protocols = flex.ExpandStringSet(v)
	}

	return apiObject
}

func expandCustomRoutingEndpointConfigurations(configurations []interface{}) []*globalaccelerator.CustomRoutingEndpointConfiguration {
	out := make([]*globalaccelerator.CustomRoutingEndpointConfiguration, len(configurations))

	for i, raw := range configurations {
		configuration := raw.(map[string]interface{})
		m := globalaccelerator.CustomRoutingEndpointConfiguration{}

		m.EndpointId = aws.String(configuration["endpoint_id"].(string))

		out[i] = &m
	}

	return out
}

func flattenCustomRoutingEndpointDescriptions(configurations []*globalaccelerator.CustomRoutingEndpointDescription) []interface{} {
	out := make([]interface{}, len(configurations))

	for i, configuration := range configurations {
		m := make(map[string]interface{})

		m["endpoint_id"] = aws.StringValue(configuration.EndpointId)

		out[i] = m
	}

	return out
}

func flattenCustomRoutingDestinationDescriptions(configurations []*globalaccelerator.CustomRoutingDestinationDescription) []interface{} {
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
