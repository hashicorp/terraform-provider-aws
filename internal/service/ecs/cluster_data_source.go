// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_cluster", name="Cluster")
// @Tags
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"active_services_count": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"attachments": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"details": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrValue: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							names.AttrID: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrStatus: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrType: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"attachments_status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"capacity_providers": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				names.AttrClusterName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrConfiguration: {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"execute_command_configuration": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrKMSKeyID: {
											Type:     schema.TypeString,
											Computed: true,
										},
										"log_configuration": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"cloud_watch_encryption_enabled": {
														Type:     schema.TypeBool,
														Computed: true,
													},
													"cloud_watch_log_group_name": {
														Type:     schema.TypeString,
														Computed: true,
													},
													"s3_bucket_encryption_enabled": {
														Type:     schema.TypeBool,
														Computed: true,
													},
													names.AttrS3BucketName: {
														Type:     schema.TypeString,
														Computed: true,
													},
													names.AttrS3KeyPrefix: {
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},
										"logging": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							"managed_storage_configuration": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"fargate_ephemeral_storage_kms_key_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrKMSKeyID: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				"default_capacity_provider_strategy": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"base": {
								Type:     schema.TypeInt,
								Computed: true,
							},
							"capacity_provider": {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrWeight: {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
				},
				"pending_tasks_count": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"registered_container_instances_count": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"running_tasks_count": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"service_connect_defaults": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrNamespace: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"setting": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrValue: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"statistics": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrValue: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	cluster, err := findClusterByNameOrARN(ctx, conn, clusterName, awstypes.ClusterFieldAttachments, awstypes.ClusterFieldStatistics)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Cluster (%s): %s", clusterName, err)
	}

	arn := aws.ToString(cluster.ClusterArn)
	d.SetId(arn)
	d.Set("active_services_count", cluster.ActiveServicesCount)
	d.Set(names.AttrARN, arn)
	if err := d.Set("attachments", flattenClusterAttachments(cluster.Attachments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attachments: %s", err)
	}
	d.Set("attachments_status", cluster.AttachmentsStatus)
	d.Set("capacity_providers", cluster.CapacityProviders)
	if err := d.Set(names.AttrConfiguration, flattenClusterConfiguration(cluster.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	if err := d.Set("default_capacity_provider_strategy", flattenCapacityProviderStrategyItems(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_capacity_provider_strategy: %s", err)
	}
	d.Set("pending_tasks_count", cluster.PendingTasksCount)
	d.Set("registered_container_instances_count", cluster.RegisteredContainerInstancesCount)
	d.Set("running_tasks_count", cluster.RunningTasksCount)
	if cluster.ServiceConnectDefaults != nil {
		if err := d.Set("service_connect_defaults", []any{flattenClusterServiceConnectDefaults(cluster.ServiceConnectDefaults)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting service_connect_defaults: %s", err)
		}
	} else {
		d.Set("service_connect_defaults", nil)
	}
	if err := d.Set("setting", flattenClusterSettings(cluster.Settings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting setting: %s", err)
	}
	if err := d.Set("statistics", flattenKeyValuePairs(cluster.Statistics)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting statistics: %s", err)
	}
	d.Set(names.AttrStatus, cluster.Status)

	setTagsOut(ctx, cluster.Tags)

	return diags
}

func flattenClusterAttachments(apiObjects []awstypes.Attachment) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]any{
			"details":        flattenKeyValuePairs(apiObject.Details),
			names.AttrID:     aws.ToString(apiObject.Id),
			names.AttrStatus: aws.ToString(apiObject.Status),
			names.AttrType:   aws.ToString(apiObject.Type),
		})
	}

	return tfList
}

func flattenKeyValuePairs(apiObjects []awstypes.KeyValuePair) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]any{
			names.AttrName:  aws.ToString(apiObject.Name),
			names.AttrValue: aws.ToString(apiObject.Value),
		})
	}

	return tfList
}
