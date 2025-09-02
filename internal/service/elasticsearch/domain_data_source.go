// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elasticsearch_domain", name="Domain")
func dataSourceDomain() *schema.Resource {
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
						"warm_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"warm_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
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
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Computed: true,
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
			"kibana_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	ds, err := findDomainByName(ctx, conn, domainName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain (%s): %s", domainName, err)
	}

	dc, err := findDomainConfigByName(ctx, conn, domainName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain (%s) Config: %s", domainName, err)
	}

	d.SetId(aws.ToString(ds.ARN))

	if ds.AccessPolicies != nil && aws.ToString(ds.AccessPolicies) != "" {
		policies, err := structure.NormalizeJsonString(*ds.AccessPolicies)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("access_policies", policies)
	}
	if err := d.Set("advanced_options", ds.AdvancedOptions); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_options: %s", err)
	}
	if err := d.Set("advanced_security_options", flattenAdvancedSecurityOptions(ds.AdvancedSecurityOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_security_options: %s", err)
	}
	d.Set(names.AttrARN, ds.ARN)
	if dc.AutoTuneOptions != nil {
		if err := d.Set("auto_tune_options", []any{flattenAutoTuneOptions(dc.AutoTuneOptions.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_tune_options: %s", err)
		}
	}
	if err := d.Set("cluster_config", flattenElasticsearchClusterConfig(ds.ElasticsearchClusterConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_config: %s", err)
	}
	if err := d.Set("cognito_options", flattenCognitoOptions(ds.CognitoOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_options: %s", err)
	}
	d.Set("created", ds.Created)
	d.Set("deleted", ds.Deleted)
	d.Set("domain_id", ds.DomainId)
	if err := d.Set("ebs_options", flattenEBSOptions(ds.EBSOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_options: %s", err)
	}
	d.Set("elasticsearch_version", ds.ElasticsearchVersion)
	if err := d.Set("encryption_at_rest", flattenEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_at_rest: %s", err)
	}
	d.Set(names.AttrEndpoint, ds.Endpoint)
	d.Set("kibana_endpoint", getKibanaEndpoint(d))
	if err := d.Set("log_publishing_options", flattenLogPublishingOptions(ds.LogPublishingOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_publishing_options: %s", err)
	}
	if err := d.Set("node_to_node_encryption", flattenNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_to_node_encryption: %s", err)
	}
	d.Set("processing", ds.Processing)
	if err := d.Set("snapshot_options", flattenSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snapshot_options: %s", err)
	}
	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", []any{flattenVPCDerivedInfo(ds.VPCOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_options: %s", err)
		}

		d.Set(names.AttrEndpoint, ds.Endpoints["vpc"])
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
	} else if ds.Endpoint != nil {
		d.Set(names.AttrEndpoint, ds.Endpoint)
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Elasticsearch Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
