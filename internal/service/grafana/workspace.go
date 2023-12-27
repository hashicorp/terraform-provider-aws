// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_grafana_workspace", name="Workspace")
// @Tags(identifierAttribute="arn")
func ResourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceCreate,
		ReadWithoutTimeout:   resourceWorkspaceRead,
		UpdateWithoutTimeout: resourceWorkspaceUpdate,
		DeleteWithoutTimeout: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_access_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.AccountAccessType_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(managedgrafana.AuthenticationProviderTypes_Values(), false),
				},
			},
			"configuration": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"data_sources": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(managedgrafana.DataSourceType_Values(), false),
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_access_control": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix_list_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 100,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpce_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 100,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"notification_destinations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(managedgrafana.NotificationDestinationType_Values(), false),
				},
			},
			"organization_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"organizational_units": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.PermissionType_Values(), false),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"saml_configuration_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 100,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 2,
							MaxItems: 100,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	input := &managedgrafana.CreateWorkspaceInput{
		AccountAccessType:       aws.String(d.Get("account_access_type").(string)),
		AuthenticationProviders: flex.ExpandStringList(d.Get("authentication_providers").([]interface{})),
		ClientToken:             aws.String(id.UniqueId()),
		PermissionType:          aws.String(d.Get("permission_type").(string)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("configuration"); ok {
		input.Configuration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_sources"); ok {
		input.WorkspaceDataSources = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.WorkspaceDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("grafana_version"); ok {
		input.GrafanaVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.WorkspaceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_access_control"); ok {
		input.NetworkAccessControl = expandNetworkAccessControl(v.([]interface{}))
	}

	if v, ok := d.GetOk("notification_destinations"); ok {
		input.WorkspaceNotificationDestinations = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("organization_role_name"); ok {
		input.OrganizationRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("organizational_units"); ok {
		input.WorkspaceOrganizationalUnits = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.WorkspaceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stack_set_name"); ok {
		input.StackSetName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_configuration"); ok {
		input.VpcConfiguration = expandVPCConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateWorkspaceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace: %s", err)
	}

	d.SetId(aws.StringValue(output.Workspace.Id))

	if _, err := waitWorkspaceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	workspace, err := FindWorkspaceByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace (%s): %s", d.Id(), err)
	}

	d.Set("account_access_type", workspace.AccountAccessType)
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonmanagedgrafana.html#amazonmanagedgrafana-resources-for-iam-policies.
	workspaceARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   managedgrafana.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("/workspaces/%s", d.Id()),
	}.String()
	d.Set("arn", workspaceARN)
	d.Set("authentication_providers", workspace.Authentication.Providers)
	d.Set("data_sources", workspace.DataSources)
	d.Set("description", workspace.Description)
	d.Set("endpoint", workspace.Endpoint)
	d.Set("grafana_version", workspace.GrafanaVersion)
	d.Set("name", workspace.Name)
	if err := d.Set("network_access_control", flattenNetworkAccessControl(workspace.NetworkAccessControl)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_access_control: %s", err)
	}
	d.Set("notification_destinations", workspace.NotificationDestinations)
	d.Set("organization_role_name", workspace.OrganizationRoleName)
	d.Set("organizational_units", workspace.OrganizationalUnits)
	d.Set("permission_type", workspace.PermissionType)
	d.Set("role_arn", workspace.WorkspaceRoleArn)
	d.Set("saml_configuration_status", workspace.Authentication.SamlConfigurationStatus)
	d.Set("stack_set_name", workspace.StackSetName)
	if err := d.Set("vpc_configuration", flattenVPCConfiguration(workspace.VpcConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_configuration: %s", err)
	}

	setTagsOut(ctx, workspace.Tags)

	input := &managedgrafana.DescribeWorkspaceConfigurationInput{
		WorkspaceId: aws.String(d.Id()),
	}

	output, err := conn.DescribeWorkspaceConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace (%s) configuration: %s", d.Id(), err)
	}

	d.Set("configuration", output.Configuration)

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	if d.HasChangesExcept("configuration", "grafana_version", "tags", "tags_all") {
		input := &managedgrafana.UpdateWorkspaceInput{
			WorkspaceId: aws.String(d.Id()),
		}

		if d.HasChange("account_access_type") {
			input.AccountAccessType = aws.String(d.Get("account_access_type").(string))
		}

		if d.HasChange("data_sources") {
			input.WorkspaceDataSources = flex.ExpandStringList(d.Get("data_sources").([]interface{}))
		}

		if d.HasChange("description") {
			input.WorkspaceDescription = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.WorkspaceName = aws.String(d.Get("name").(string))
		}

		if d.HasChange("network_access_control") {
			if v, ok := d.Get("network_access_control").([]interface{}); ok {
				if len(v) > 0 {
					input.NetworkAccessControl = expandNetworkAccessControl(v)
				} else {
					input.RemoveNetworkAccessConfiguration = aws.Bool(true)
				}
			}
		}

		if d.HasChange("notification_destinations") {
			input.WorkspaceNotificationDestinations = flex.ExpandStringList(d.Get("notification_destinations").([]interface{}))
		}

		if d.HasChange("organization_role_name") {
			input.OrganizationRoleName = aws.String(d.Get("organization_role_name").(string))
		}

		if d.HasChange("organizational_units") {
			input.WorkspaceOrganizationalUnits = flex.ExpandStringList(d.Get("organizational_units").([]interface{}))
		}

		if d.HasChange("permission_type") {
			input.PermissionType = aws.String(d.Get("permission_type").(string))
		}

		if d.HasChange("role_arn") {
			input.WorkspaceRoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("stack_set_name") {
			input.StackSetName = aws.String(d.Get("stack_set_name").(string))
		}

		if d.HasChange("vpc_configuration") {
			if v, ok := d.Get("vpc_configuration").([]interface{}); ok {
				if len(v) > 0 {
					input.VpcConfiguration = expandVPCConfiguration(v)
				} else {
					input.RemoveVpcConfiguration = aws.Bool(true)
				}
			}
		}

		_, err := conn.UpdateWorkspaceWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Grafana Workspace (%s): %s", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChanges("configuration", "grafana_version") {
		input := &managedgrafana.UpdateWorkspaceConfigurationInput{
			Configuration: aws.String(d.Get("configuration").(string)),
			WorkspaceId:   aws.String(d.Id()),
		}

		if d.HasChange("grafana_version") {
			input.GrafanaVersion = aws.String(d.Get("grafana_version").(string))
		}

		_, err := conn.UpdateWorkspaceConfigurationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Grafana Workspace (%s) configuration: %s", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace (%s) configuration update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn(ctx)

	log.Printf("[DEBUG] Deleting Grafana Workspace: %s", d.Id())
	_, err := conn.DeleteWorkspaceWithContext(ctx, &managedgrafana.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana Workspace (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkspaceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandVPCConfiguration(cfg []interface{}) *managedgrafana.VpcConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := managedgrafana.VpcConfiguration{}

	if v, ok := conf["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := conf["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.SubnetIds = flex.ExpandStringSet(v)
	}

	return &out
}

func flattenVPCConfiguration(rs *managedgrafana.VpcConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.SecurityGroupIds != nil {
		m["security_group_ids"] = flex.FlattenStringSet(rs.SecurityGroupIds)
	}
	if rs.SubnetIds != nil {
		m["subnet_ids"] = flex.FlattenStringSet(rs.SubnetIds)
	}

	return []interface{}{m}
}

func expandNetworkAccessControl(cfg []interface{}) *managedgrafana.NetworkAccessConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := managedgrafana.NetworkAccessConfiguration{}

	if v, ok := conf["prefix_list_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.PrefixListIds = flex.ExpandStringSet(v)
	}

	if v, ok := conf["vpce_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.VpceIds = flex.ExpandStringSet(v)
	}

	return &out
}

func flattenNetworkAccessControl(rs *managedgrafana.NetworkAccessConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.PrefixListIds != nil {
		m["prefix_list_ids"] = flex.FlattenStringSet(rs.PrefixListIds)
	}
	if rs.VpceIds != nil {
		m["vpce_ids"] = flex.FlattenStringSet(rs.VpceIds)
	}

	return []interface{}{m}
}
