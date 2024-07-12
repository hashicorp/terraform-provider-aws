// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_grafana_workspace", name="Workspace")
// @Tags
func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"account_access_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"notification_destinations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"organization_role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organizational_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_configuration_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	workspaceID := d.Get("workspace_id").(string)
	workspace, err := findWorkspaceByID(ctx, conn, workspaceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace (%s): %s", workspaceID, err)
	}

	d.SetId(workspaceID)
	d.Set("account_access_type", workspace.AccountAccessType)
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonmanagedgrafana.html#amazonmanagedgrafana-resources-for-iam-policies.
	workspaceARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "grafana",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("/workspaces/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, workspaceARN)
	d.Set("authentication_providers", workspace.Authentication.Providers)
	d.Set(names.AttrCreatedDate, workspace.Created.Format(time.RFC3339))
	d.Set("data_sources", workspace.DataSources)
	d.Set(names.AttrDescription, workspace.Description)
	d.Set(names.AttrEndpoint, workspace.Endpoint)
	d.Set("grafana_version", workspace.GrafanaVersion)
	d.Set(names.AttrLastUpdatedDate, workspace.Modified.Format(time.RFC3339))
	d.Set(names.AttrName, workspace.Name)
	d.Set("notification_destinations", workspace.NotificationDestinations)
	d.Set("organization_role_name", workspace.OrganizationRoleName)
	d.Set("organizational_units", workspace.OrganizationalUnits)
	d.Set("permission_type", workspace.PermissionType)
	d.Set(names.AttrRoleARN, workspace.WorkspaceRoleArn)
	d.Set("saml_configuration_status", workspace.Authentication.SamlConfigurationStatus)
	d.Set("stack_set_name", workspace.StackSetName)
	d.Set(names.AttrStatus, workspace.Status)

	setTagsOut(ctx, workspace.Tags)

	return diags
}
