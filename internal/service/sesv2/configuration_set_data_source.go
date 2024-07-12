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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_configuration_set")
func DataSourceConfigurationSet() *schema.Resource {
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
	DSNameConfigurationSet = "Configuration Set Data Source"
)

func dataSourceConfigurationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	name := d.Get("configuration_set_name").(string)

	out, err := FindConfigurationSetByID(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, DSNameConfigurationSet, name, err)
	}

	d.SetId(aws.ToString(out.ConfigurationSetName))

	d.Set(names.AttrARN, configurationSetNameToARN(meta, aws.ToString(out.ConfigurationSetName)))
	d.Set("configuration_set_name", out.ConfigurationSetName)

	if out.DeliveryOptions != nil {
		if err := d.Set("delivery_options", []interface{}{flattenDeliveryOptions(out.DeliveryOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("delivery_options", nil)
	}

	if out.ReputationOptions != nil {
		if err := d.Set("reputation_options", []interface{}{flattenReputationOptions(out.ReputationOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("reputation_options", nil)
	}

	if out.SendingOptions != nil {
		if err := d.Set("sending_options", []interface{}{flattenSendingOptions(out.SendingOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("sending_options", nil)
	}

	if out.SuppressionOptions != nil {
		if err := d.Set("suppression_options", []interface{}{flattenSuppressionOptions(out.SuppressionOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("suppression_options", nil)
	}

	if out.TrackingOptions != nil {
		if err := d.Set("tracking_options", []interface{}{flattenTrackingOptions(out.TrackingOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("tracking_options", nil)
	}

	if out.VdmOptions != nil {
		if err := d.Set("vdm_options", []interface{}{flattenVDMOptions(out.VdmOptions)}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
		}
	} else {
		d.Set("vdm_options", nil)
	}

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, DSNameConfigurationSet, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameConfigurationSet, d.Id(), err)
	}

	return diags
}
