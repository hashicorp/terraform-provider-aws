// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_plan", name="Plan")
// @Tags(identifierAttribute="arn")
func dataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePlanRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
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
						"schedule_expression_timezone": {
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
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	id := d.Get("plan_id").(string)
	output, err := findPlanByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Plan (%s): %s", id, err)
	}

	d.SetId(aws.ToString(output.BackupPlanId))
	d.Set(names.AttrARN, output.BackupPlanArn)
	d.Set(names.AttrName, output.BackupPlan.BackupPlanName)
	if err := d.Set(names.AttrRule, flattenBackupRules(ctx, output.BackupPlan.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set(names.AttrVersion, output.VersionId)

	return diags
}
