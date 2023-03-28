package globalaccelerator

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_globalaccelerator_custom_routing_accelerator")
func ResourceCustomRoutingAccelerator() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomRoutingAcceleratorCreate,
		Read:   resourceCustomRoutingAcceleratorRead,
		Update: resourceCustomRoutingAcceleratorUpdate,
		Delete: resourceCustomRoutingAcceleratorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flow_logs_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"flow_logs_s3_bucket": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"flow_logs_s3_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z-]+$`), "only alphanumeric characters and hyphens are allowed"),
					validation.StringDoesNotMatch(regexp.MustCompile(`^-`), "cannot start with a hyphen"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.IpAddressTypeIpv4,
				ValidateFunc: validation.StringInSlice(globalaccelerator.IpAddressType_Values(), false),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomRoutingAcceleratorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(context.TODO(), d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &globalaccelerator.CreateCustomRoutingAcceleratorInput{
		Name:             aws.String(name),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		Tags:             Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Global Accelerator Custom Routing Accelerator: %s", input)
	output, err := conn.CreateCustomRoutingAccelerator(input)

	if err != nil {
		return fmt.Errorf("error creating Global Accelerator Custom Routing Accelerator (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Accelerator.AcceleratorArn))

	if _, err := waitCustomRoutingAcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateAcceleratorAttributesInput(v.([]interface{})[0].(map[string]interface{}))
		input.AcceleratorArn = aws.String(d.Id())

		log.Printf("[DEBUG] Updating Global Accelerator Accelerator attributes: %s", input)
		if _, err := conn.UpdateAcceleratorAttributes(input); err != nil {
			return fmt.Errorf("error updating Global Accelerator Accelerator (%s) attributes: %w", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
		}
	}

	return resourceCustomRoutingAcceleratorRead(d, meta)
}

func resourceCustomRoutingAcceleratorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	accelerator, err := FindCustomRoutingAcceleratorByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Custom Routing Accelerator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator Custom Routing Accelerator (%s): %w", d.Id(), err)
	}

	d.Set("enabled", accelerator.Enabled)
	d.Set("dns_name", accelerator.DnsName)
	d.Set("hosted_zone_id", route53ZoneID)
	d.Set("name", accelerator.Name)
	d.Set("ip_address_type", accelerator.IpAddressType)

	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return fmt.Errorf("error setting ip_sets: %w", err)
	}

	acceleratorAttributes, err := FindCustomRoutingAcceleratorAttributesByARN(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator Custom Routing Accelerator (%s) attributes: %w", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenCustomRoutingAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return fmt.Errorf("error setting attributes: %w", err)
	}

	tags, err := ListTags(context.TODO(), conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Global Accelerator Custom Routing Accelerator (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCustomRoutingAcceleratorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	if d.HasChanges("name", "ip_address_type", "enabled") {
		input := &globalaccelerator.UpdateCustomRoutingAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Name:           aws.String(d.Get("name").(string)),
			Enabled:        aws.Bool(d.Get("enabled").(bool)),
		}

		if v, ok := d.GetOk("ip_address_type"); ok {
			input.IpAddressType = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Global Accelerator Custom Routing Accelerator: %s", input)
		if _, err := conn.UpdateCustomRoutingAccelerator(input); err != nil {
			return fmt.Errorf("error updating Global Accelerator Custom Routing Accelerator (%s): %w", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", d.Id(), err)
		}
	}

	if d.HasChange("attributes") {
		o, n := d.GetChange("attributes")
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
				oInput := expandUpdateCustomRoutingAcceleratorAttributesInput(o.([]interface{})[0].(map[string]interface{}))
				oInput.AcceleratorArn = aws.String(d.Id())
				nInput := expandUpdateCustomRoutingAcceleratorAttributesInput(n.([]interface{})[0].(map[string]interface{}))
				nInput.AcceleratorArn = aws.String(d.Id())

				// To change flow logs bucket and prefix attributes while flows are enabled, first disable flow logs.
				if aws.BoolValue(oInput.FlowLogsEnabled) && aws.BoolValue(nInput.FlowLogsEnabled) {
					oInput.FlowLogsEnabled = aws.Bool(false)

					log.Printf("[DEBUG] Updating Global Accelerator Custom Routing Accelerator attributes: %s", oInput)
					if _, err := conn.UpdateCustomRoutingAcceleratorAttributes(oInput); err != nil {
						return fmt.Errorf("error updating Global Accelerator Custom Routing Accelerator (%s) attributes: %w", d.Id(), err)
					}

					if _, err := waitCustomRoutingAcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
						return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", d.Id(), err)
					}
				}

				log.Printf("[DEBUG] Updating Global Accelerator Accelerator attributes: %s", nInput)
				if _, err := conn.UpdateCustomRoutingAcceleratorAttributes(nInput); err != nil {
					return fmt.Errorf("error updating Global Accelerator Custom Routing Accelerator (%s) attributes: %w", d.Id(), err)
				}

				if _, err := waitCustomRoutingAcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(context.TODO(), conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Global Accelerator Custom Routing Accelerator (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceCustomRoutingAcceleratorRead(d, meta)
}

func resourceCustomRoutingAcceleratorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	{
		input := &globalaccelerator.UpdateCustomRoutingAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Enabled:        aws.Bool(false),
		}

		log.Printf("[DEBUG] Updating Global Accelerator Custom Routing Accelerator: %s", input)
		_, err := conn.UpdateCustomRoutingAccelerator(input)

		if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error disabling Global Accelerator Custom Routing Accelerator (%s): %w", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %w", d.Id(), err)
		}
	}

	{
		input := &globalaccelerator.DeleteCustomRoutingAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Deleting Global Accelerator Custom Routing  Accelerator (%s)", d.Id())
		_, err := conn.DeleteCustomRoutingAccelerator(input)

		if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error deleting Global Accelerator Custom Routing Accelerator (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func FindCustomRoutingAcceleratorByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingAccelerator, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	return findCustomRoutingAccelerator(conn, input)
}

func findCustomRoutingAccelerator(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingAcceleratorInput) (*globalaccelerator.CustomRoutingAccelerator, error) {
	output, err := conn.DescribeCustomRoutingAccelerator(input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Accelerator == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Accelerator, nil
}

func FindCustomRoutingAcceleratorAttributesByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingAcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	return findCustomRoutingAcceleratorAttributes(conn, input)
}

func findCustomRoutingAcceleratorAttributes(conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput) (*globalaccelerator.CustomRoutingAcceleratorAttributes, error) {
	output, err := conn.DescribeCustomRoutingAcceleratorAttributes(input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AcceleratorAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AcceleratorAttributes, nil
}

func statusCustomRoutingAccelerator(conn *globalaccelerator.GlobalAccelerator, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accelerator, err := FindCustomRoutingAcceleratorByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return accelerator, aws.StringValue(accelerator.Status), nil
	}
}

func waitCustomRoutingAcceleratorDeployed(conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.CustomRoutingAccelerator, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: statusCustomRoutingAccelerator(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*globalaccelerator.CustomRoutingAccelerator); ok {
		return v, err
	}

	return nil, err
}

func expandUpdateCustomRoutingAcceleratorAttributesInput(tfMap map[string]interface{}) *globalaccelerator.UpdateCustomRoutingAcceleratorAttributesInput {
	return (*globalaccelerator.UpdateCustomRoutingAcceleratorAttributesInput)(expandUpdateAcceleratorAttributesInput(tfMap))
}

func flattenCustomRoutingAcceleratorAttributes(apiObject *globalaccelerator.CustomRoutingAcceleratorAttributes) map[string]interface{} {
	return flattenAcceleratorAttributes((*globalaccelerator.AcceleratorAttributes)(apiObject))
}
