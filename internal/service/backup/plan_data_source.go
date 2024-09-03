// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_plan")
func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePlanRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRule: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"completion_window": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"copy_action": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_vault_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"lifecycle": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cold_storage_after": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"delete_after": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"opt_in_to_archive_for_supported_resources": {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"enable_continuous_backup": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"lifecycle": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cold_storage_after": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"delete_after": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"opt_in_to_archive_for_supported_resources": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"recovery_point_tags": tftags.TagsSchema(),
						"rule_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSchedule: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_window": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"target_vault_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: planHash,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("plan_id").(string)

	resp, err := conn.GetBackupPlan(ctx, &backup.GetBackupPlanInput{
		BackupPlanId: aws.String(id),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Backup Plan: %s", err)
	}

	d.SetId(aws.ToString(resp.BackupPlanId))
	d.Set(names.AttrARN, resp.BackupPlanArn)
	d.Set(names.AttrName, resp.BackupPlan.BackupPlanName)
	d.Set(names.AttrVersion, resp.VersionId)
	if err := d.Set(names.AttrRule, flattenPlanRules(ctx, resp.BackupPlan.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	tags, err := listTags(ctx, conn, aws.ToString(resp.BackupPlanArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Plan (%s): %s", id, err)
	}
	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
