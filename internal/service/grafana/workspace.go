// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_grafana_workspace", name="Workspace")
// @Tags(identifierAttribute="arn")
func resourceWorkspace() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AccountAccessType](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.AuthenticationProviderTypes](),
				},
			},
			names.AttrConfiguration: {
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.DataSourceType](),
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.NotificationDestinationType](),
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PermissionType](),
			},
			names.AttrRoleARN: {
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
			names.AttrVPCConfiguration: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 100,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
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
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	input := &grafana.CreateWorkspaceInput{
		AccountAccessType:       awstypes.AccountAccessType(d.Get("account_access_type").(string)),
		AuthenticationProviders: flex.ExpandStringyValueList[awstypes.AuthenticationProviderTypes](d.Get("authentication_providers").([]interface{})),
		ClientToken:             aws.String(id.UniqueId()),
		PermissionType:          awstypes.PermissionType(d.Get("permission_type").(string)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok {
		input.Configuration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_sources"); ok {
		input.WorkspaceDataSources = flex.ExpandStringyValueList[awstypes.DataSourceType](v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.WorkspaceDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("grafana_version"); ok {
		input.GrafanaVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.WorkspaceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_access_control"); ok {
		input.NetworkAccessControl = expandNetworkAccessControl(v.([]interface{}))
	}

	if v, ok := d.GetOk("notification_destinations"); ok {
		input.WorkspaceNotificationDestinations = flex.ExpandStringyValueList[awstypes.NotificationDestinationType](v.([]interface{}))
	}

	if v, ok := d.GetOk("organization_role_name"); ok {
		input.OrganizationRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("organizational_units"); ok {
		input.WorkspaceOrganizationalUnits = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.WorkspaceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stack_set_name"); ok {
		input.StackSetName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCConfiguration); ok {
		input.VpcConfiguration = expandVPCConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateWorkspace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace: %s", err)
	}

	d.SetId(aws.ToString(output.Workspace.Id))

	if _, err := waitWorkspaceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	workspace, err := findWorkspaceByID(ctx, conn, d.Id())

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
		Service:   "grafana",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("/workspaces/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, workspaceARN)
	d.Set("authentication_providers", workspace.Authentication.Providers)
	d.Set("data_sources", workspace.DataSources)
	d.Set(names.AttrDescription, workspace.Description)
	d.Set(names.AttrEndpoint, workspace.Endpoint)
	d.Set("grafana_version", workspace.GrafanaVersion)
	d.Set(names.AttrName, workspace.Name)
	if err := d.Set("network_access_control", flattenNetworkAccessControl(workspace.NetworkAccessControl)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_access_control: %s", err)
	}
	d.Set("notification_destinations", workspace.NotificationDestinations)
	d.Set("organization_role_name", workspace.OrganizationRoleName)
	d.Set("organizational_units", workspace.OrganizationalUnits)
	d.Set("permission_type", workspace.PermissionType)
	d.Set(names.AttrRoleARN, workspace.WorkspaceRoleArn)
	d.Set("saml_configuration_status", workspace.Authentication.SamlConfigurationStatus)
	d.Set("stack_set_name", workspace.StackSetName)
	if err := d.Set(names.AttrVPCConfiguration, flattenVPCConfiguration(workspace.VpcConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_configuration: %s", err)
	}

	setTagsOut(ctx, workspace.Tags)

	input := &grafana.DescribeWorkspaceConfigurationInput{
		WorkspaceId: aws.String(d.Id()),
	}

	output, err := conn.DescribeWorkspaceConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace (%s) configuration: %s", d.Id(), err)
	}

	d.Set(names.AttrConfiguration, output.Configuration)

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	if d.HasChangesExcept(names.AttrConfiguration, "grafana_version", names.AttrTags, names.AttrTagsAll) {
		input := &grafana.UpdateWorkspaceInput{
			WorkspaceId: aws.String(d.Id()),
		}

		if d.HasChange("account_access_type") {
			input.AccountAccessType = awstypes.AccountAccessType(d.Get("account_access_type").(string))
		}

		if d.HasChange("data_sources") {
			input.WorkspaceDataSources = flex.ExpandStringyValueList[awstypes.DataSourceType](d.Get("data_sources").([]interface{}))
		}

		if d.HasChange(names.AttrDescription) {
			input.WorkspaceDescription = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrName) {
			input.WorkspaceName = aws.String(d.Get(names.AttrName).(string))
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
			input.WorkspaceNotificationDestinations = flex.ExpandStringyValueList[awstypes.NotificationDestinationType](d.Get("notification_destinations").([]interface{}))
		}

		if d.HasChange("organization_role_name") {
			input.OrganizationRoleName = aws.String(d.Get("organization_role_name").(string))
		}

		if d.HasChange("organizational_units") {
			input.WorkspaceOrganizationalUnits = flex.ExpandStringValueList(d.Get("organizational_units").([]interface{}))
		}

		if d.HasChange("permission_type") {
			input.PermissionType = awstypes.PermissionType(d.Get("permission_type").(string))
		}

		if d.HasChange(names.AttrRoleARN) {
			input.WorkspaceRoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
		}

		if d.HasChange("stack_set_name") {
			input.StackSetName = aws.String(d.Get("stack_set_name").(string))
		}

		if d.HasChange(names.AttrVPCConfiguration) {
			if v, ok := d.Get(names.AttrVPCConfiguration).([]interface{}); ok {
				if len(v) > 0 {
					input.VpcConfiguration = expandVPCConfiguration(v)
				} else {
					input.RemoveVpcConfiguration = aws.Bool(true)
				}
			}
		}

		_, err := conn.UpdateWorkspace(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Grafana Workspace (%s): %s", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChanges(names.AttrConfiguration, "grafana_version") {
		input := &grafana.UpdateWorkspaceConfigurationInput{
			Configuration: aws.String(d.Get(names.AttrConfiguration).(string)),
			WorkspaceId:   aws.String(d.Id()),
		}

		if d.HasChange("grafana_version") {
			input.GrafanaVersion = aws.String(d.Get("grafana_version").(string))
		}

		_, err := conn.UpdateWorkspaceConfiguration(ctx, input)

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
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	log.Printf("[DEBUG] Deleting Grafana Workspace: %s", d.Id())
	_, err := conn.DeleteWorkspace(ctx, &grafana.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findWorkspaceByID(ctx context.Context, conn *grafana.Client, id string) (*awstypes.WorkspaceDescription, error) {
	input := &grafana.DescribeWorkspaceInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspace(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workspace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workspace, nil
}

func statusWorkspace(ctx context.Context, conn *grafana.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWorkspaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitWorkspaceCreated(ctx context.Context, conn *grafana.Client, id string, timeout time.Duration) (*awstypes.WorkspaceDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspaceStatusCreating),
		Target:  enum.Slice(awstypes.WorkspaceStatusActive),
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceUpdated(ctx context.Context, conn *grafana.Client, id string, timeout time.Duration) (*awstypes.WorkspaceDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspaceStatusUpdating, awstypes.WorkspaceStatusVersionUpdating),
		Target:  enum.Slice(awstypes.WorkspaceStatusActive),
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceDeleted(ctx context.Context, conn *grafana.Client, id string, timeout time.Duration) (*awstypes.WorkspaceDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspaceStatusDeleting),
		Target:  []string{},
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}

func expandVPCConfiguration(tfList []interface{}) *awstypes.VpcConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := awstypes.VpcConfiguration{}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return &apiObject
}

func flattenVPCConfiguration(apiObject *awstypes.VpcConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := make(map[string]interface{})
	tfMap[names.AttrSecurityGroupIDs] = apiObject.SecurityGroupIds
	tfMap[names.AttrSubnetIDs] = apiObject.SubnetIds

	return []interface{}{tfMap}
}

func expandNetworkAccessControl(tfList []interface{}) *awstypes.NetworkAccessConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := awstypes.NetworkAccessConfiguration{}

	if v, ok := tfMap["prefix_list_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.PrefixListIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["vpce_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.VpceIds = flex.ExpandStringValueSet(v)
	}

	return &apiObject
}

func flattenNetworkAccessControl(apiObject *awstypes.NetworkAccessConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := make(map[string]interface{})
	tfMap["prefix_list_ids"] = apiObject.PrefixListIds
	tfMap["vpce_ids"] = apiObject.VpceIds

	return []interface{}{tfMap}
}
