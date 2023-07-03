// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// @SDKResource("aws_globalaccelerator_custom_routing_accelerator", name="Custom Routing Accelerator")
// @Tags(identifierAttribute="id")
func ResourceCustomRoutingAccelerator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomRoutingAcceleratorCreate,
		ReadWithoutTimeout:   resourceCustomRoutingAcceleratorRead,
		UpdateWithoutTimeout: resourceCustomRoutingAcceleratorUpdate,
		DeleteWithoutTimeout: resourceCustomRoutingAcceleratorDelete,

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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomRoutingAcceleratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	name := d.Get("name").(string)
	input := &globalaccelerator.CreateCustomRoutingAcceleratorInput{
		Name:             aws.String(name),
		IdempotencyToken: aws.String(id.UniqueId()),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ip_addresses"); ok && len(v.([]interface{})) > 0 {
		input.IpAddresses = flex.ExpandStringList(v.([]interface{}))
	}

	output, err := conn.CreateCustomRoutingAcceleratorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Custom Routing Accelerator (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Accelerator.AcceleratorArn))

	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateAcceleratorAttributesInput(v.([]interface{})[0].(map[string]interface{}))
		input.AcceleratorArn = aws.String(d.Id())

		if _, err := conn.UpdateAcceleratorAttributesWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCustomRoutingAcceleratorRead(ctx, d, meta)...)
}

func resourceCustomRoutingAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	accelerator, err := FindCustomRoutingAcceleratorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Custom Routing Accelerator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	d.Set("dns_name", accelerator.DnsName)
	d.Set("enabled", accelerator.Enabled)
	d.Set("hosted_zone_id", meta.(*conns.AWSClient).GlobalAcceleratorHostedZoneID())
	d.Set("ip_address_type", accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_sets: %s", err)
	}
	d.Set("name", accelerator.Name)

	acceleratorAttributes, err := FindCustomRoutingAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenCustomRoutingAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}

	return diags
}

func resourceCustomRoutingAcceleratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	if d.HasChanges("name", "ip_address_type", "enabled") {
		input := &globalaccelerator.UpdateCustomRoutingAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Name:           aws.String(d.Get("name").(string)),
			Enabled:        aws.Bool(d.Get("enabled").(bool)),
		}

		if v, ok := d.GetOk("ip_address_type"); ok {
			input.IpAddressType = aws.String(v.(string))
		}

		_, err := conn.UpdateCustomRoutingAcceleratorWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", d.Id(), err)
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

					_, err := conn.UpdateCustomRoutingAcceleratorAttributesWithContext(ctx, oInput)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
					}

					if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", d.Id(), err)
					}
				}

				_, err := conn.UpdateCustomRoutingAcceleratorAttributesWithContext(ctx, nInput)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
				}

				if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceCustomRoutingAcceleratorRead(ctx, d, meta)...)
}

func resourceCustomRoutingAcceleratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)

	input := &globalaccelerator.UpdateCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
		Enabled:        aws.Bool(false),
	}

	_, err := conn.UpdateCustomRoutingAcceleratorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deployment: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Global Accelerator Custom Routing  Accelerator (%s)", d.Id())
	_, err = conn.DeleteCustomRoutingAcceleratorWithContext(ctx, &globalaccelerator.DeleteCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	return diags
}

func FindCustomRoutingAcceleratorByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingAccelerator, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	return findCustomRoutingAccelerator(ctx, conn, input)
}

func findCustomRoutingAccelerator(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingAcceleratorInput) (*globalaccelerator.CustomRoutingAccelerator, error) {
	output, err := conn.DescribeCustomRoutingAcceleratorWithContext(ctx, input)

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

func FindCustomRoutingAcceleratorAttributesByARN(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.CustomRoutingAcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	return findCustomRoutingAcceleratorAttributes(ctx, conn, input)
}

func findCustomRoutingAcceleratorAttributes(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, input *globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput) (*globalaccelerator.CustomRoutingAcceleratorAttributes, error) {
	output, err := conn.DescribeCustomRoutingAcceleratorAttributesWithContext(ctx, input)

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

func statusCustomRoutingAccelerator(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accelerator, err := FindCustomRoutingAcceleratorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return accelerator, aws.StringValue(accelerator.Status), nil
	}
}

func waitCustomRoutingAcceleratorDeployed(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.CustomRoutingAccelerator, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: statusCustomRoutingAccelerator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*globalaccelerator.CustomRoutingAccelerator); ok {
		return output, err
	}

	return nil, err
}

func expandUpdateCustomRoutingAcceleratorAttributesInput(tfMap map[string]interface{}) *globalaccelerator.UpdateCustomRoutingAcceleratorAttributesInput {
	return (*globalaccelerator.UpdateCustomRoutingAcceleratorAttributesInput)(expandUpdateAcceleratorAttributesInput(tfMap))
}

func flattenCustomRoutingAcceleratorAttributes(apiObject *globalaccelerator.CustomRoutingAcceleratorAttributes) map[string]interface{} {
	return flattenAcceleratorAttributes((*globalaccelerator.AcceleratorAttributes)(apiObject))
}
