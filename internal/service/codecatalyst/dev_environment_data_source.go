// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecatalyst

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_codecatalyst_dev_environment", name="Dev Environment")
func DataSourceDevEnvironment() *schema.Resource {
	return &schema.Resource{

		ReadWithoutTimeout: dataSourceDevEnvironmentRead,

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"creator_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"env_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ides": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"runtime": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"inactivity_timeout_minutes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"persistent_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repositories": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRepositoryName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameDevEnvironment = "Dev Environment Data Source"
)

func dataSourceDevEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	spaceName := aws.String(d.Get("space_name").(string))
	projectName := aws.String(d.Get("project_name").(string))
	env_id := d.Get("env_id").(string)

	out, err := findDevEnvironmentByID(ctx, conn, env_id, spaceName, projectName)
	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionReading, DSNameDevEnvironment, d.Id(), err)
	}

	d.SetId(aws.ToString(out.Id))

	d.Set(names.AttrAlias, out.Alias)
	d.Set("creator_id", out.CreatorId)
	d.Set("project_name", out.ProjectName)
	d.Set("space_name", out.SpaceName)
	d.Set(names.AttrInstanceType, out.InstanceType)
	d.Set(names.AttrLastUpdatedTime, out.LastUpdatedTime.String())
	d.Set("inactivity_timeout_minutes", out.InactivityTimeoutMinutes)
	d.Set("persistent_storage", flattenPersistentStorage(out.PersistentStorage))
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrStatusReason, out.StatusReason)

	if err := d.Set("ides", flattenIdes(out.Ides)); err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionSetting, ResNameDevEnvironment, d.Id(), err)
	}

	if err := d.Set("repositories", flattenRepositories(out.Repositories)); err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionSetting, ResNameDevEnvironment, d.Id(), err)
	}

	return diags
}
