// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_workspaces_directory", name="Directory")
// @Tags(identifierAttribute="id")
func resourceDirectory() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDirectoryCreate,
		ReadWithoutTimeout:   resourceDirectoryRead,
		UpdateWithoutTimeout: resourceDirectoryUpdate,
		DeleteWithoutTimeout: resourceDirectoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"directory_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"iam_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
			},
			"registration_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"relay_state_parameter_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "RelayState",
						},
						names.AttrStatus: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.SamlStatusEnumDisabled,
							ValidateDiagFunc: enum.Validate[types.SamlStatusEnum](),
						},
						"user_access_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"self_service_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"change_compute_type": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"increase_volume_size": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"rebuild_workspace": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"restart_workspace": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"switch_running_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"workspace_access_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_type_android": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_chromeos": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_ios": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_linux": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_osx": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_web": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_windows": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
						"device_type_zeroclient": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AccessPropertyValue](),
						},
					},
				},
			},
			"workspace_creation_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_security_group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_ou": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_internet_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"enable_maintenance_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"user_enabled_as_local_administrator": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"workspace_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	directoryID := d.Get("directory_id").(string)
	input := &workspaces.RegisterWorkspaceDirectoryInput{
		DirectoryId:       aws.String(directoryID),
		EnableSelfService: aws.Bool(false), // this is handled separately below
		EnableWorkDocs:    aws.Bool(false),
		Tenancy:           types.TenancyShared,
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrSubnetIDs); ok {
		input.SubnetIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*types.InvalidResourceStateException](ctx, timeout,
		func() (any, error) {
			return conn.RegisterWorkspaceDirectory(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering WorkSpaces Directory (%s): %s", directoryID, err)
	}

	d.SetId(directoryID)

	if _, err := waitDirectoryRegistered(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkSpaces Directory (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("saml_properties"); ok {
		input := &workspaces.ModifySamlPropertiesInput{
			ResourceId:     aws.String(d.Id()),
			SamlProperties: expandSAMLProperties(v.([]any)),
		}

		_, err := conn.ModifySamlProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) SAML properties: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("self_service_permissions"); ok {
		input := &workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: expandSelfservicePermissions(v.([]any)),
		}

		_, err := conn.ModifySelfservicePermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) self-service permissions: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("workspace_access_properties"); ok {
		input := &workspaces.ModifyWorkspaceAccessPropertiesInput{
			ResourceId:                aws.String(d.Id()),
			WorkspaceAccessProperties: expandWorkspaceAccessProperties(v.([]any)),
		}

		_, err := conn.ModifyWorkspaceAccessProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) access properties: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("workspace_creation_properties"); ok {
		input := &workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(d.Id()),
			WorkspaceCreationProperties: expandWorkspaceCreationProperties(v.([]any)),
		}

		_, err := conn.ModifyWorkspaceCreationProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) creation properties: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("ip_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input := &workspaces.AssociateIpGroupsInput{
			DirectoryId: aws.String(d.Id()),
			GroupIds:    flex.ExpandStringValueSet(v.(*schema.Set)),
		}

		_, err := conn.AssociateIpGroups(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "asassociating WorkSpaces Directory (%s) IP Groups: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDirectoryRead(ctx, d, meta)...)
}

func resourceDirectoryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	directory, err := findDirectoryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Directory (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAlias, directory.Alias)
	d.Set("directory_id", directory.DirectoryId)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set("dns_ip_addresses", directory.DnsIpAddresses)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("ip_group_ids", directory.IpGroupIds)
	d.Set("registration_code", directory.RegistrationCode)
	if err := d.Set("self_service_permissions", flattenSelfservicePermissions(directory.SelfservicePermissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting self_service_permissions: %s", err)
	}
	if err := d.Set("saml_properties", flattenSAMLProperties(directory.SamlProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_properties: %s", err)
	}
	d.Set(names.AttrSubnetIDs, directory.SubnetIds)
	if err := d.Set("workspace_access_properties", flattenWorkspaceAccessProperties(directory.WorkspaceAccessProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_access_properties: %s", err)
	}
	if err := d.Set("workspace_creation_properties", flattenDefaultWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_creation_properties: %s", err)
	}
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)

	return diags
}

func resourceDirectoryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	if d.HasChange("saml_properties") {
		tfListSAMLProperties := d.Get("saml_properties").([]any)
		tfMap := tfListSAMLProperties[0].(map[string]any)

		var dels []types.DeletableSamlProperty
		if tfMap["relay_state_parameter_name"].(string) == "" {
			dels = append(dels, types.DeletableSamlPropertySamlPropertiesRelayStateParameterName)
		}
		if tfMap["user_access_url"].(string) == "" {
			dels = append(dels, types.DeletableSamlPropertySamlPropertiesUserAccessUrl)
		}

		input := &workspaces.ModifySamlPropertiesInput{
			PropertiesToDelete: dels,
			ResourceId:         aws.String(d.Id()),
			SamlProperties:     expandSAMLProperties(tfListSAMLProperties),
		}

		_, err := conn.ModifySamlProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) SAML properties: %s", d.Id(), err)
		}
	}

	if d.HasChange("self_service_permissions") {
		input := &workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: expandSelfservicePermissions(d.Get("self_service_permissions").([]any)),
		}

		_, err := conn.ModifySelfservicePermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) self-service permissions: %s", d.Id(), err)
		}
	}

	if d.HasChange("workspace_access_properties") {
		input := &workspaces.ModifyWorkspaceAccessPropertiesInput{
			ResourceId:                aws.String(d.Id()),
			WorkspaceAccessProperties: expandWorkspaceAccessProperties(d.Get("workspace_access_properties").([]any)),
		}

		_, err := conn.ModifyWorkspaceAccessProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) access properties: %s", d.Id(), err)
		}
	}

	if d.HasChange("workspace_creation_properties") {
		input := &workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(d.Id()),
			WorkspaceCreationProperties: expandWorkspaceCreationProperties(d.Get("workspace_creation_properties").([]any)),
		}

		_, err := conn.ModifyWorkspaceCreationProperties(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) creation properties: %s", d.Id(), err)
		}
	}

	if d.HasChange("ip_group_ids") {
		o, n := d.GetChange("ip_group_ids")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os), os.Difference(ns)

		if add.Len() > 0 {
			input := &workspaces.AssociateIpGroupsInput{
				DirectoryId: aws.String(d.Id()),
				GroupIds:    flex.ExpandStringValueSet(add),
			}

			_, err := conn.AssociateIpGroups(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating WorkSpaces Directory (%s) IP Groups: %s", d.Id(), err)
			}
		}

		if del.Len() > 0 {
			input := &workspaces.DisassociateIpGroupsInput{
				DirectoryId: aws.String(d.Id()),
				GroupIds:    flex.ExpandStringValueSet(del),
			}

			_, err := conn.DisassociateIpGroups(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating WorkSpaces Directory (%s) IP Groups: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDirectoryRead(ctx, d, meta)...)
}

func resourceDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	log.Printf("[DEBUG] Deleting WorkSpaces Directory: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*types.InvalidResourceStateException](ctx, timeout,
		func() (any, error) {
			return conn.DeregisterWorkspaceDirectory(ctx, &workspaces.DeregisterWorkspaceDirectoryInput{
				DirectoryId: aws.String(d.Id()),
			})
		})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering WorkSpaces Directory (%s): %s", d.Id(), err)
	}

	if _, err := waitDirectoryDeregistered(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkSpaces Directory (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDirectoryByID(ctx context.Context, conn *workspaces.Client, id string) (*types.WorkspaceDirectory, error) {
	input := &workspaces.DescribeWorkspaceDirectoriesInput{
		DirectoryIds: []string{id},
	}

	output, err := findDirectory(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if itypes.IsZero(output) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if state := output.State; state == types.WorkspaceDirectoryStateDeregistered {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findDirectory(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspaceDirectoriesInput) (*types.WorkspaceDirectory, error) {
	output, err := findDirectories(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDirectories(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspaceDirectoriesInput) ([]types.WorkspaceDirectory, error) {
	var output []types.WorkspaceDirectory

	pages := workspaces.NewDescribeWorkspaceDirectoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Directories...)
	}

	return output, nil
}

func statusDirectory(ctx context.Context, conn *workspaces.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDirectoryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitDirectoryRegistered(ctx context.Context, conn *workspaces.Client, directoryID string) (*types.WorkspaceDirectory, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceDirectoryStateRegistering),
		Target:  enum.Slice(types.WorkspaceDirectoryStateRegistered),
		Refresh: statusDirectory(ctx, conn, directoryID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.WorkspaceDirectory); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func waitDirectoryDeregistered(ctx context.Context, conn *workspaces.Client, directoryID string) (*types.WorkspaceDirectory, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.WorkspaceDirectoryStateRegistering,
			types.WorkspaceDirectoryStateRegistered,
			types.WorkspaceDirectoryStateDeregistering,
		),
		Target:  []string{},
		Refresh: statusDirectory(ctx, conn, directoryID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.WorkspaceDirectory); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func expandWorkspaceAccessProperties(tfList []any) *types.WorkspaceAccessProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &types.WorkspaceAccessProperties{}
	tfMap := tfList[0].(map[string]any)

	if tfMap["device_type_android"].(string) != "" {
		apiObject.DeviceTypeAndroid = types.AccessPropertyValue(tfMap["device_type_android"].(string))
	}

	if tfMap["device_type_chromeos"].(string) != "" {
		apiObject.DeviceTypeChromeOs = types.AccessPropertyValue(tfMap["device_type_chromeos"].(string))
	}

	if tfMap["device_type_ios"].(string) != "" {
		apiObject.DeviceTypeIos = types.AccessPropertyValue(tfMap["device_type_ios"].(string))
	}

	if tfMap["device_type_linux"].(string) != "" {
		apiObject.DeviceTypeLinux = types.AccessPropertyValue(tfMap["device_type_linux"].(string))
	}

	if tfMap["device_type_osx"].(string) != "" {
		apiObject.DeviceTypeOsx = types.AccessPropertyValue(tfMap["device_type_osx"].(string))
	}

	if tfMap["device_type_web"].(string) != "" {
		apiObject.DeviceTypeWeb = types.AccessPropertyValue(tfMap["device_type_web"].(string))
	}

	if tfMap["device_type_windows"].(string) != "" {
		apiObject.DeviceTypeWindows = types.AccessPropertyValue(tfMap["device_type_windows"].(string))
	}

	if tfMap["device_type_zeroclient"].(string) != "" {
		apiObject.DeviceTypeZeroClient = types.AccessPropertyValue(tfMap["device_type_zeroclient"].(string))
	}

	return apiObject
}

func expandSAMLProperties(tfList []any) *types.SamlProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &types.SamlProperties{}

	if tfMap["relay_state_parameter_name"].(string) != "" {
		apiObject.RelayStateParameterName = aws.String(tfMap["relay_state_parameter_name"].(string))
	}

	if tfMap[names.AttrStatus].(string) != "" {
		apiObject.Status = types.SamlStatusEnum(tfMap[names.AttrStatus].(string))
	}

	if tfMap["user_access_url"].(string) != "" {
		apiObject.UserAccessUrl = aws.String(tfMap["user_access_url"].(string))
	}

	return apiObject
}

func expandSelfservicePermissions(tfList []any) *types.SelfservicePermissions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &types.SelfservicePermissions{}
	tfMap := tfList[0].(map[string]any)

	if tfMap["change_compute_type"].(bool) {
		apiObject.ChangeComputeType = types.ReconnectEnumEnabled
	} else {
		apiObject.ChangeComputeType = types.ReconnectEnumDisabled
	}

	if tfMap["increase_volume_size"].(bool) {
		apiObject.IncreaseVolumeSize = types.ReconnectEnumEnabled
	} else {
		apiObject.IncreaseVolumeSize = types.ReconnectEnumDisabled
	}

	if tfMap["rebuild_workspace"].(bool) {
		apiObject.RebuildWorkspace = types.ReconnectEnumEnabled
	} else {
		apiObject.RebuildWorkspace = types.ReconnectEnumDisabled
	}

	if tfMap["restart_workspace"].(bool) {
		apiObject.RestartWorkspace = types.ReconnectEnumEnabled
	} else {
		apiObject.RestartWorkspace = types.ReconnectEnumDisabled
	}

	if tfMap["switch_running_mode"].(bool) {
		apiObject.SwitchRunningMode = types.ReconnectEnumEnabled
	} else {
		apiObject.SwitchRunningMode = types.ReconnectEnumDisabled
	}

	return apiObject
}

func expandWorkspaceCreationProperties(tfList []any) *types.WorkspaceCreationProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &types.WorkspaceCreationProperties{
		EnableInternetAccess:            aws.Bool(tfMap["enable_internet_access"].(bool)),
		EnableMaintenanceMode:           aws.Bool(tfMap["enable_maintenance_mode"].(bool)),
		UserEnabledAsLocalAdministrator: aws.Bool(tfMap["user_enabled_as_local_administrator"].(bool)),
	}

	if tfMap["custom_security_group_id"].(string) != "" {
		apiObject.CustomSecurityGroupId = aws.String(tfMap["custom_security_group_id"].(string))
	}

	if tfMap["default_ou"].(string) != "" {
		apiObject.DefaultOu = aws.String(tfMap["default_ou"].(string))
	}

	return apiObject
}

func flattenWorkspaceAccessProperties(apiObject *types.WorkspaceAccessProperties) []any {
	if apiObject == nil {
		return []any{}
	}

	return []any{
		map[string]any{
			"device_type_android":    apiObject.DeviceTypeAndroid,
			"device_type_chromeos":   apiObject.DeviceTypeChromeOs,
			"device_type_ios":        apiObject.DeviceTypeIos,
			"device_type_linux":      apiObject.DeviceTypeLinux,
			"device_type_osx":        apiObject.DeviceTypeOsx,
			"device_type_web":        apiObject.DeviceTypeWeb,
			"device_type_windows":    apiObject.DeviceTypeWindows,
			"device_type_zeroclient": apiObject.DeviceTypeZeroClient,
		},
	}
}

func flattenSAMLProperties(apiObject *types.SamlProperties) []any {
	if apiObject == nil {
		return []any{}
	}

	return []any{
		map[string]any{
			"relay_state_parameter_name": aws.ToString(apiObject.RelayStateParameterName),
			names.AttrStatus:             apiObject.Status,
			"user_access_url":            aws.ToString(apiObject.UserAccessUrl),
		},
	}
}

func flattenSelfservicePermissions(apiObject *types.SelfservicePermissions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	switch apiObject.ChangeComputeType {
	case types.ReconnectEnumEnabled:
		tfMap["change_compute_type"] = true
	case types.ReconnectEnumDisabled:
		tfMap["change_compute_type"] = false
	default:
		tfMap["change_compute_type"] = nil
	}

	switch apiObject.IncreaseVolumeSize {
	case types.ReconnectEnumEnabled:
		tfMap["increase_volume_size"] = true
	case types.ReconnectEnumDisabled:
		tfMap["increase_volume_size"] = false
	default:
		tfMap["increase_volume_size"] = nil
	}

	switch apiObject.RebuildWorkspace {
	case types.ReconnectEnumEnabled:
		tfMap["rebuild_workspace"] = true
	case types.ReconnectEnumDisabled:
		tfMap["rebuild_workspace"] = false
	default:
		tfMap["rebuild_workspace"] = nil
	}

	switch apiObject.RestartWorkspace {
	case types.ReconnectEnumEnabled:
		tfMap["restart_workspace"] = true
	case types.ReconnectEnumDisabled:
		tfMap["restart_workspace"] = false
	default:
		tfMap["restart_workspace"] = nil
	}

	switch apiObject.SwitchRunningMode {
	case types.ReconnectEnumEnabled:
		tfMap["switch_running_mode"] = true
	case types.ReconnectEnumDisabled:
		tfMap["switch_running_mode"] = false
	default:
		tfMap["switch_running_mode"] = nil
	}

	return []any{tfMap}
}

func flattenDefaultWorkspaceCreationProperties(apiObject *types.DefaultWorkspaceCreationProperties) []any {
	if apiObject == nil {
		return []any{}
	}

	return []any{
		map[string]any{
			"custom_security_group_id":            aws.ToString(apiObject.CustomSecurityGroupId),
			"default_ou":                          aws.ToString(apiObject.DefaultOu),
			"enable_internet_access":              aws.ToBool(apiObject.EnableInternetAccess),
			"enable_maintenance_mode":             aws.ToBool(apiObject.EnableMaintenanceMode),
			"user_enabled_as_local_administrator": aws.ToBool(apiObject.UserEnabledAsLocalAdministrator),
		},
	}
}
