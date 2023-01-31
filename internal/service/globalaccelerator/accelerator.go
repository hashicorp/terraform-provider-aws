package globalaccelerator

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// Global Route53 Zone ID for Global Accelerators, exported as a
// convenience attribute for Route53 aliases (see
// https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html).
const route53ZoneID = "Z2BJ6XQ5FK7U4H"

func ResourceAccelerator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAcceleratorCreate,
		ReadWithoutTimeout:   resourceAcceleratorRead,
		UpdateWithoutTimeout: resourceAcceleratorUpdate,
		DeleteWithoutTimeout: resourceAcceleratorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.IpAddressTypeIpv4,
				ValidateFunc: validation.StringInSlice(globalaccelerator.IpAddressType_Values(), false),
			},
			"ip_addresses": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAcceleratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &globalaccelerator.CreateAcceleratorInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Name:             aws.String(name),
		Tags:             Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ip_addresses"); ok && len(v.([]interface{})) > 0 {
		input.IpAddresses = flex.ExpandStringList(v.([]interface{}))
	}

	output, err := conn.CreateAcceleratorWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Global Accelerator Accelerator (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Accelerator.AcceleratorArn))

	if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateAcceleratorAttributesInput(v.([]interface{})[0].(map[string]interface{}))
		input.AcceleratorArn = aws.String(d.Id())

		_, err := conn.UpdateAcceleratorAttributesWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
		}

		if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
		}
	}

	return resourceAcceleratorRead(ctx, d, meta)
}

func resourceAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	accelerator, err := FindAcceleratorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Accelerator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	d.Set("dns_name", accelerator.DnsName)
	d.Set("enabled", accelerator.Enabled)
	d.Set("hosted_zone_id", route53ZoneID)
	d.Set("ip_address_type", accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return diag.Errorf("setting ip_sets: %s", err)
	}
	d.Set("name", accelerator.Name)

	acceleratorAttributes, err := FindAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("reading Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return diag.Errorf("setting attributes: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("listing tags for Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceAcceleratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	if d.HasChanges("name", "ip_address_type", "enabled") {
		input := &globalaccelerator.UpdateAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Enabled:        aws.Bool(d.Get("enabled").(bool)),
			Name:           aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("ip_address_type"); ok {
			input.IpAddressType = aws.String(v.(string))
		}

		_, err := conn.UpdateAcceleratorWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Global Accelerator Accelerator (%s): %s", d.Id(), err)
		}

		if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
		}
	}

	if d.HasChange("attributes") {
		o, n := d.GetChange("attributes")
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
				oInput := expandUpdateAcceleratorAttributesInput(o.([]interface{})[0].(map[string]interface{}))
				oInput.AcceleratorArn = aws.String(d.Id())
				nInput := expandUpdateAcceleratorAttributesInput(n.([]interface{})[0].(map[string]interface{}))
				nInput.AcceleratorArn = aws.String(d.Id())

				// To change flow logs bucket and prefix attributes while flows are enabled, first disable flow logs.
				if aws.BoolValue(oInput.FlowLogsEnabled) && aws.BoolValue(nInput.FlowLogsEnabled) {
					oInput.FlowLogsEnabled = aws.Bool(false)

					_, err := conn.UpdateAcceleratorAttributesWithContext(ctx, oInput)

					if err != nil {
						return diag.Errorf("updating Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
					}

					if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
						return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
					}
				}

				_, err := conn.UpdateAcceleratorAttributesWithContext(ctx, nInput)

				if err != nil {
					return diag.Errorf("updating Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
				}

				if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating Global Accelerator Accelerator (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAcceleratorRead(ctx, d, meta)
}

func resourceAcceleratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	input := &globalaccelerator.UpdateAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
		Enabled:        aws.Bool(false),
	}

	_, err := conn.UpdateAcceleratorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("disabling Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Global Accelerator Accelerator: %s", d.Id())
	_, err = conn.DeleteAcceleratorWithContext(ctx, &globalaccelerator.DeleteAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	return nil
}

func expandUpdateAcceleratorAttributesInput(tfMap map[string]interface{}) *globalaccelerator.UpdateAcceleratorAttributesInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.UpdateAcceleratorAttributesInput{}

	if v, ok := tfMap["flow_logs_enabled"].(bool); ok {
		apiObject.FlowLogsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["flow_logs_s3_bucket"].(string); ok && v != "" {
		apiObject.FlowLogsS3Bucket = aws.String(v)
	}

	if v, ok := tfMap["flow_logs_s3_prefix"].(string); ok && v != "" {
		apiObject.FlowLogsS3Prefix = aws.String(v)
	}

	return apiObject
}

func flattenIPSet(apiObject *globalaccelerator.IpSet) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IpAddresses; v != nil {
		tfMap["ip_addresses"] = aws.StringValueSlice(v)
	}

	if v := apiObject.IpFamily; v != nil {
		tfMap["ip_family"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenIPSets(apiObjects []*globalaccelerator.IpSet) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenIPSet(apiObject))
	}

	return tfList
}

func flattenAcceleratorAttributes(apiObject *globalaccelerator.AcceleratorAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FlowLogsEnabled; v != nil {
		tfMap["flow_logs_enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.FlowLogsS3Bucket; v != nil {
		tfMap["flow_logs_s3_bucket"] = aws.StringValue(v)
	}

	if v := apiObject.FlowLogsS3Prefix; v != nil {
		tfMap["flow_logs_s3_prefix"] = aws.StringValue(v)
	}

	return tfMap
}
