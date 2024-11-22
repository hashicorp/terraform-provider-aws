// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_globalaccelerator_custom_routing_accelerator", name="Custom Routing Accelerator")
// @Tags(identifierAttribute="id")
func resourceCustomRoutingAccelerator() *schema.Resource {
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
			names.AttrAttributes: {
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
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.IpAddressTypeIpv4,
				ValidateDiagFunc: enum.Validate[awstypes.IpAddressType](),
			},
			names.AttrIPAddresses: {
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
						names.AttrIPAddresses: {
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
			names.AttrName: {
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

func resourceCustomRoutingAcceleratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &globalaccelerator.CreateCustomRoutingAcceleratorInput{
		Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
		Name:             aws.String(name),
		IdempotencyToken: aws.String(id.UniqueId()),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrIPAddressType); ok {
		input.IpAddressType = awstypes.IpAddressType(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIPAddresses); ok && len(v.([]interface{})) > 0 {
		input.IpAddresses = flex.ExpandStringValueList(v.([]interface{}))
	}

	output, err := conn.CreateCustomRoutingAccelerator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Global Accelerator Custom Routing Accelerator (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Accelerator.AcceleratorArn))

	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrAttributes); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateAcceleratorAttributesInput(v.([]interface{})[0].(map[string]interface{}))
		input.AcceleratorArn = aws.String(d.Id())

		if _, err := conn.UpdateAcceleratorAttributes(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCustomRoutingAcceleratorRead(ctx, d, meta)...)
}

func resourceCustomRoutingAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	accelerator, err := findCustomRoutingAcceleratorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Custom Routing Accelerator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDNSName, accelerator.DnsName)
	d.Set(names.AttrEnabled, accelerator.Enabled)
	d.Set(names.AttrHostedZoneID, meta.(*conns.AWSClient).GlobalAcceleratorHostedZoneID(ctx))
	d.Set(names.AttrIPAddressType, accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_sets: %s", err)
	}
	d.Set(names.AttrName, accelerator.Name)

	acceleratorAttributes, err := findCustomRoutingAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrAttributes, []interface{}{flattenCustomRoutingAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}

	return diags
}

func resourceCustomRoutingAcceleratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	if d.HasChanges(names.AttrName, names.AttrIPAddressType, names.AttrEnabled) {
		input := &globalaccelerator.UpdateCustomRoutingAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Name:           aws.String(d.Get(names.AttrName).(string)),
			Enabled:        aws.Bool(d.Get(names.AttrEnabled).(bool)),
		}

		if v, ok := d.GetOk(names.AttrIPAddressType); ok {
			input.IpAddressType = awstypes.IpAddressType(v.(string))
		}

		_, err := conn.UpdateCustomRoutingAccelerator(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
		}

		if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrAttributes) {
		o, n := d.GetChange(names.AttrAttributes)
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
				oInput := expandUpdateCustomRoutingAcceleratorAttributesInput(o.([]interface{})[0].(map[string]interface{}))
				oInput.AcceleratorArn = aws.String(d.Id())
				nInput := expandUpdateCustomRoutingAcceleratorAttributesInput(n.([]interface{})[0].(map[string]interface{}))
				nInput.AcceleratorArn = aws.String(d.Id())

				// To change flow logs bucket and prefix attributes while flows are enabled, first disable flow logs.
				if aws.ToBool(oInput.FlowLogsEnabled) && aws.ToBool(nInput.FlowLogsEnabled) {
					oInput.FlowLogsEnabled = aws.Bool(false)

					_, err := conn.UpdateCustomRoutingAcceleratorAttributes(ctx, oInput)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
					}

					if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", d.Id(), err)
					}
				}

				_, err := conn.UpdateCustomRoutingAcceleratorAttributes(ctx, nInput)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
				}

				if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceCustomRoutingAcceleratorRead(ctx, d, meta)...)
}

func resourceCustomRoutingAcceleratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	input := &globalaccelerator.UpdateCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
		Enabled:        aws.Bool(false),
	}

	_, err := conn.UpdateCustomRoutingAccelerator(ctx, input)

	if errs.IsA[*awstypes.AcceleratorNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	if _, err := waitCustomRoutingAcceleratorDeployed(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Global Accelerator Custom Routing Accelerator (%s) deploy: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Global Accelerator Custom Routing  Accelerator (%s)", d.Id())
	_, err = conn.DeleteCustomRoutingAccelerator(ctx, &globalaccelerator.DeleteCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.AcceleratorNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	return diags
}

func findCustomRoutingAcceleratorByARN(ctx context.Context, conn *globalaccelerator.Client, arn string) (*awstypes.CustomRoutingAccelerator, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorInput{
		AcceleratorArn: aws.String(arn),
	}

	output, err := conn.DescribeCustomRoutingAccelerator(ctx, input)

	if errs.IsA[*awstypes.AcceleratorNotFoundException](err) {
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

func findCustomRoutingAcceleratorAttributesByARN(ctx context.Context, conn *globalaccelerator.Client, arn string) (*awstypes.CustomRoutingAcceleratorAttributes, error) {
	input := &globalaccelerator.DescribeCustomRoutingAcceleratorAttributesInput{
		AcceleratorArn: aws.String(arn),
	}

	output, err := conn.DescribeCustomRoutingAcceleratorAttributes(ctx, input)

	if errs.IsA[*awstypes.AcceleratorNotFoundException](err) {
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

func statusCustomRoutingAccelerator(ctx context.Context, conn *globalaccelerator.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accelerator, err := findCustomRoutingAcceleratorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return accelerator, string(accelerator.Status), nil
	}
}

func waitCustomRoutingAcceleratorDeployed(ctx context.Context, conn *globalaccelerator.Client, arn string, timeout time.Duration) (*awstypes.CustomRoutingAccelerator, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AcceleratorStatusInProgress),
		Target:  enum.Slice(awstypes.AcceleratorStatusDeployed),
		Refresh: statusCustomRoutingAccelerator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CustomRoutingAccelerator); ok {
		return output, err
	}

	return nil, err
}

func expandUpdateCustomRoutingAcceleratorAttributesInput(tfMap map[string]interface{}) *globalaccelerator.UpdateCustomRoutingAcceleratorAttributesInput {
	return (*globalaccelerator.UpdateCustomRoutingAcceleratorAttributesInput)(expandUpdateAcceleratorAttributesInput(tfMap))
}

func flattenCustomRoutingAcceleratorAttributes(apiObject *awstypes.CustomRoutingAcceleratorAttributes) map[string]interface{} {
	return flattenAcceleratorAttributes((*awstypes.AcceleratorAttributes)(apiObject))
}
