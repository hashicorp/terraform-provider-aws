// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_configuration_set", name="Configuration Set")
// @Tags(identifierAttribute="arn")
func dataSourceConfigurationSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationSetRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"delivery_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_delivery_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"sending_pool_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tls_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"reputation_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"last_fresh_start": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reputation_metrics_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"sending_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sending_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"suppression_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"suppressed_reasons": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"tracking_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_redirect_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"https_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"vdm_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dashboard_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"engagement_metrics": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"guardian_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"optimized_shared_delivery": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	dsNameConfigurationSet = "Configuration Set Data Source"
)

func dataSourceConfigurationSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	name := d.Get("configuration_set_name").(string)

	output, err := findConfigurationSetByID(ctx, conn, name)

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, dsNameConfigurationSet, name, err)
	}

	d.SetId(aws.ToString(output.ConfigurationSetName))

	d.Set(names.AttrARN, configurationSetARN(ctx, meta.(*conns.AWSClient), aws.ToString(output.ConfigurationSetName)))
	d.Set("configuration_set_name", output.ConfigurationSetName)
	if output.DeliveryOptions != nil {
		if err := d.Set("delivery_options", []any{flattenDeliveryOptions(output.DeliveryOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting delivery_options: %s", err)
		}
	} else {
		d.Set("delivery_options", nil)
	}
	if output.ReputationOptions != nil {
		if err := d.Set("reputation_options", []any{flattenReputationOptions(output.ReputationOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting reputation_options: %s", err)
		}
	} else {
		d.Set("reputation_options", nil)
	}
	if output.SendingOptions != nil {
		if err := d.Set("sending_options", []any{flattenSendingOptions(output.SendingOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting sending_options: %s", err)
		}
	} else {
		d.Set("sending_options", nil)
	}
	if output.SuppressionOptions != nil {
		if err := d.Set("suppression_options", []any{flattenSuppressionOptions(output.SuppressionOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting suppression_options: %s", err)
		}
	} else {
		d.Set("suppression_options", nil)
	}
	if output.TrackingOptions != nil {
		if err := d.Set("tracking_options", []any{flattenTrackingOptions(output.TrackingOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tracking_options: %s", err)
		}
	} else {
		d.Set("tracking_options", nil)
	}
	if output.VdmOptions != nil {
		if err := d.Set("vdm_options", []any{flattenVDMOptions(output.VdmOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vdm_options: %s", err)
		}
	} else {
		d.Set("vdm_options", nil)
	}

	return diags
}
