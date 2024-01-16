// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_globalaccelerator_accelerator", name="Accelerator")
// @Tags(identifierAttribute="id")
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
			"dual_stack_dns_name": {
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
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "only alphanumeric characters and hyphens are allowed"),
					validation.StringDoesNotMatch(regexache.MustCompile(`^-`), "cannot start with a hyphen"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAcceleratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	name := d.Get("name").(string)
	input := &globalaccelerator.CreateAcceleratorInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		IdempotencyToken: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ip_addresses"); ok && len(v.([]interface{})) > 0 {
		input.IpAddresses = flex.ExpandStringList(v.([]interface{}))
	}

	output, err := conn.CreateAcceleratorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Accelerator (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Accelerator.AcceleratorArn))

	if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateAcceleratorAttributesInput(v.([]interface{})[0].(map[string]interface{}))
		input.AcceleratorArn = aws.String(d.Id())

		_, err := conn.UpdateAcceleratorAttributesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
		}

		if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAcceleratorRead(ctx, d, meta)...)
}

func resourceAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	accelerator, err := FindAcceleratorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Accelerator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	d.Set("dns_name", accelerator.DnsName)
	d.Set("dual_stack_dns_name", accelerator.DualStackDnsName)
	d.Set("enabled", accelerator.Enabled)
	d.Set("hosted_zone_id", meta.(*conns.AWSClient).GlobalAcceleratorHostedZoneID())
	d.Set("ip_address_type", accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_sets: %s", err)
	}
	d.Set("name", accelerator.Name)

	acceleratorAttributes, err := FindAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}

	return diags
}

func resourceAcceleratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

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
			return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Accelerator (%s): %s", d.Id(), err)
		}

		if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
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
						return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
					}

					if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
					}
				}

				_, err := conn.UpdateAcceleratorAttributesWithContext(ctx, nInput)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
				}

				if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceAcceleratorRead(ctx, d, meta)...)
}

func resourceAcceleratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	input := &globalaccelerator.UpdateAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
		Enabled:        aws.Bool(false),
	}

	_, err := conn.UpdateAcceleratorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	if _, err := waitAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Accelerator (%s) deployment: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Global Accelerator Accelerator: %s", d.Id())
	_, err = conn.DeleteAcceleratorWithContext(ctx, &globalaccelerator.DeleteAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	return diags
}

func FindAcceleratorByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.Accelerator, error) {
	input := &globalaccelerator.DescribeAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	return findAccelerator(ctx, conn, input)
}

func findAccelerator(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeAcceleratorInput) (*globalaccelerator.Accelerator, error) {
	output, err := conn.DescribeAcceleratorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil, &retry.NotFoundError{
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

func FindAcceleratorAttributesByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.AcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	return findAcceleratorAttributes(ctx, conn, input)
}

func findAcceleratorAttributes(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeAcceleratorAttributesInput) (*globalaccelerator.AcceleratorAttributes, error) {
	output, err := conn.DescribeAcceleratorAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil, &retry.NotFoundError{
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

func statusAccelerator(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accelerator, err := FindAcceleratorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return accelerator, aws.StringValue(accelerator.Status), nil
	}
}

func waitAcceleratorDeployed(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.Accelerator, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: statusAccelerator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*globalaccelerator.Accelerator); ok {
		return output, err
	}

	return nil, err
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
