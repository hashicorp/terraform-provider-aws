// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_opensearch_domain")
func DataSourceDomain() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDomainRead,

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"advanced_options": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"advanced_security_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"anonymous_auth_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"internal_user_database_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_tune_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"maintenance_schedule": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cron_expression_for_recurrence": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrDuration: {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrUnit: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrValue: {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"start_at": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"rollback_on_disable": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_off_peak_window": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cold_storage_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"dedicated_master_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"dedicated_master_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrInstanceCount: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"multi_az_with_standby_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"warm_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"warm_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"warm_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_awareness_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"zone_awareness_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"cognito_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRoleARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrUserPoolID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dashboard_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deleted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"ebs_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ebs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrThroughput: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrVolumeType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"encryption_at_rest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kibana_endpoint": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use 'dashboard_endpoint' attribute instead",
			},
			"log_publishing_options": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCloudWatchLogGroupARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"log_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"node_to_node_encryption": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"off_peak_window_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"off_peak_window": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"window_start_time": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hours": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"minutes": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"processing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"snapshot_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automated_snapshot_start_hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"software_update_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_software_update_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vpc_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ds, err := FindDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "your query returned no results")
	}

	reqDescribeDomainConfig := &opensearchservice.DescribeDomainConfigInput{
		DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
	}

	respDescribeDomainConfig, err := conn.DescribeDomainConfigWithContext(ctx, reqDescribeDomainConfig)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "querying config for opensearch_domain: %s", err)
	}

	if respDescribeDomainConfig.DomainConfig == nil {
		return sdkdiag.AppendErrorf(diags, "your query returned no results")
	}

	dc := respDescribeDomainConfig.DomainConfig

	d.SetId(aws.StringValue(ds.ARN))

	if ds.AccessPolicies != nil && aws.StringValue(ds.AccessPolicies) != "" {
		policies, err := structure.NormalizeJsonString(aws.StringValue(ds.AccessPolicies))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "access policies contain an invalid JSON: %s", err)
		}
		d.Set("access_policies", policies)
	}

	if err := d.Set("advanced_options", flex.FlattenStringMap(ds.AdvancedOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_options: %s", err)
	}

	d.Set(names.AttrARN, ds.ARN)
	d.Set("domain_id", ds.DomainId)
	d.Set(names.AttrEndpoint, ds.Endpoint)
	d.Set("dashboard_endpoint", getDashboardEndpoint(d))
	d.Set("kibana_endpoint", getKibanaEndpoint(d))

	if err := d.Set("advanced_security_options", flattenAdvancedSecurityOptions(ds.AdvancedSecurityOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_security_options: %s", err)
	}

	if dc.AutoTuneOptions != nil {
		if err := d.Set("auto_tune_options", []interface{}{flattenAutoTuneOptions(dc.AutoTuneOptions.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_tune_options: %s", err)
		}
	}

	if err := d.Set("ebs_options", flattenEBSOptions(ds.EBSOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_options: %s", err)
	}

	if err := d.Set("encryption_at_rest", flattenEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_at_rest: %s", err)
	}

	if err := d.Set("node_to_node_encryption", flattenNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_to_node_encryption: %s", err)
	}

	if err := d.Set("cluster_config", flattenClusterConfig(ds.ClusterConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_config: %s", err)
	}

	if err := d.Set("snapshot_options", flattenSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snapshot_options: %s", err)
	}

	if err := d.Set("software_update_options", flattenSoftwareUpdateOptions(ds.SoftwareUpdateOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting software_update_options: %s", err)
	}

	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", []interface{}{flattenVPCDerivedInfo(ds.VPCOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_options: %s", err)
		}

		endpoints := flex.FlattenStringMap(ds.Endpoints)
		if err := d.Set(names.AttrEndpoint, endpoints["vpc"]); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
		}
		d.Set("dashboard_endpoint", getDashboardEndpoint(d))
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set(names.AttrEndpoint, ds.Endpoint)
			d.Set("dashboard_endpoint", getDashboardEndpoint(d))
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	if err := d.Set("log_publishing_options", flattenLogPublishingOptions(ds.LogPublishingOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_publishing_options: %s", err)
	}

	d.Set(names.AttrEngineVersion, ds.EngineVersion)
	d.Set(names.AttrIPAddressType, ds.IPAddressType)

	if err := d.Set("cognito_options", flattenCognitoOptions(ds.CognitoOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_options: %s", err)
	}

	if ds.OffPeakWindowOptions != nil {
		if err := d.Set("off_peak_window_options", []interface{}{flattenOffPeakWindowOptions(ds.OffPeakWindowOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting off_peak_window_options: %s", err)
		}
	} else {
		d.Set("off_peak_window_options", nil)
	}

	d.Set("created", ds.Created)
	d.Set("deleted", ds.Deleted)
	d.Set("processing", ds.Processing)

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for OpenSearch Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
