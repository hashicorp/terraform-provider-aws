// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_workspaces_directory", name="Directory")
// @Tags(identifierAttribute="id")
func ResourceDirectory() *schema.Resource {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_chromeos": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_ios": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_linux": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_osx": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_web": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_windows": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
						},
						"device_type_zeroclient": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenAccessPropertyEnumValues(types.AccessPropertyValue("").Values()), false),
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	_, err := tfresource.RetryWhenIsA[*types.InvalidResourceStateException](ctx, DirectoryRegisterInvalidResourceStateTimeout,
		func() (interface{}, error) {
			return conn.RegisterWorkspaceDirectory(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering WorkSpaces Directory (%s): %s", directoryID, err)
	}

	d.SetId(directoryID)

	_, err = WaitDirectoryRegistered(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkSpaces Directory (%s) to register: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("self_service_permissions"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) self-service permissions", directoryID)
		_, err := conn.ModifySelfservicePermissions(ctx, &workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(directoryID),
			SelfservicePermissions: ExpandSelfServicePermissions(v.([]interface{})),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) self-service permissions: %s", directoryID, err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) self-service permissions", directoryID)
	}

	if v, ok := d.GetOk("workspace_access_properties"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) access properties", directoryID)
		_, err := conn.ModifyWorkspaceAccessProperties(ctx, &workspaces.ModifyWorkspaceAccessPropertiesInput{
			ResourceId:                aws.String(directoryID),
			WorkspaceAccessProperties: ExpandWorkspaceAccessProperties(v.([]interface{})),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) access properties: %s", directoryID, err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) access properties", directoryID)
	}

	if v, ok := d.GetOk("workspace_creation_properties"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) creation properties", directoryID)
		_, err := conn.ModifyWorkspaceCreationProperties(ctx, &workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(directoryID),
			WorkspaceCreationProperties: ExpandWorkspaceCreationProperties(v.([]interface{})),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WorkSpaces Directory (%s) creation properties: %s", directoryID, err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) creation properties", directoryID)
	}

	if v, ok := d.GetOk("ip_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		ipGroupIds := v.(*schema.Set)
		log.Printf("[DEBUG] Associating WorkSpaces Directory (%s) with IP Groups %s", directoryID, ipGroupIds.List())
		_, err := conn.AssociateIpGroups(ctx, &workspaces.AssociateIpGroupsInput{
			DirectoryId: aws.String(directoryID),
			GroupIds:    flex.ExpandStringValueSet(ipGroupIds),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "asassociating WorkSpaces Directory (%s) ip groups: %s", directoryID, err)
		}
		log.Printf("[INFO] Associated WorkSpaces Directory (%s) IP Groups", directoryID)
	}

	return append(diags, resourceDirectoryRead(ctx, d, meta)...)
}

func resourceDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	directory, err := FindDirectoryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Directory (%s): %s", d.Id(), err)
	}

	d.Set("directory_id", directory.DirectoryId)
	if err := d.Set(names.AttrSubnetIDs, flex.FlattenStringValueSet(directory.SubnetIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("registration_code", directory.RegistrationCode)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set(names.AttrAlias, directory.Alias)

	if err := d.Set("self_service_permissions", FlattenSelfServicePermissions(directory.SelfservicePermissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting self_service_permissions: %s", err)
	}

	if err := d.Set("workspace_access_properties", FlattenWorkspaceAccessProperties(directory.WorkspaceAccessProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_access_properties: %s", err)
	}

	if err := d.Set("workspace_creation_properties", FlattenWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_creation_properties: %s", err)
	}

	if err := d.Set("ip_group_ids", flex.FlattenStringValueSet(directory.IpGroupIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_group_ids: %s", err)
	}

	if err := d.Set("dns_ip_addresses", flex.FlattenStringValueSet(directory.DnsIpAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dns_ip_addresses: %s", err)
	}

	return diags
}

func resourceDirectoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	if d.HasChange("self_service_permissions") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) self-service permissions", d.Id())
		permissions := d.Get("self_service_permissions").([]interface{})

		_, err := conn.ModifySelfservicePermissions(ctx, &workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: ExpandSelfServicePermissions(permissions),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) self service permissions: %s", d.Id(), err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) self-service permissions", d.Id())
	}

	if d.HasChange("workspace_access_properties") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) access properties", d.Id())
		properties := d.Get("workspace_access_properties").([]interface{})

		_, err := conn.ModifyWorkspaceAccessProperties(ctx, &workspaces.ModifyWorkspaceAccessPropertiesInput{
			ResourceId:                aws.String(d.Id()),
			WorkspaceAccessProperties: ExpandWorkspaceAccessProperties(properties),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) access properties: %s", d.Id(), err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) access properties", d.Id())
	}

	if d.HasChange("workspace_creation_properties") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) creation properties", d.Id())
		properties := d.Get("workspace_creation_properties").([]interface{})

		_, err := conn.ModifyWorkspaceCreationProperties(ctx, &workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(d.Id()),
			WorkspaceCreationProperties: ExpandWorkspaceCreationProperties(properties),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces Directory (%s) creation properties: %s", d.Id(), err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) creation properties", d.Id())
	}

	if d.HasChange("ip_group_ids") {
		o, n := d.GetChange("ip_group_ids")
		old := o.(*schema.Set)
		new := n.(*schema.Set)
		added := new.Difference(old)
		removed := old.Difference(new)

		log.Printf("[DEBUG] Associating WorkSpaces Directory (%s) with IP Groups %s", d.Id(), added.GoString())
		_, err := conn.AssociateIpGroups(ctx, &workspaces.AssociateIpGroupsInput{
			DirectoryId: aws.String(d.Id()),
			GroupIds:    flex.ExpandStringValueSet(added),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "asassociating WorkSpaces Directory (%s) IP Groups: %s", d.Id(), err)
		}

		log.Printf("[DEBUG] Disassociating WorkSpaces Directory (%s) with IP Groups %s", d.Id(), removed.GoString())
		_, err = conn.DisassociateIpGroups(ctx, &workspaces.DisassociateIpGroupsInput{
			DirectoryId: aws.String(d.Id()),
			GroupIds:    flex.ExpandStringValueSet(removed),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disasassociating WorkSpaces Directory (%s) IP Groups: %s", d.Id(), err)
		}

		log.Printf("[INFO] Updated WorkSpaces Directory (%s) IP Groups", d.Id())
	}

	return append(diags, resourceDirectoryRead(ctx, d, meta)...)
}

func resourceDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	log.Printf("[DEBUG] Deregistering WorkSpaces Directory: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.InvalidResourceStateException](ctx, DirectoryRegisterInvalidResourceStateTimeout,
		func() (interface{}, error) {
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

	_, err = WaitDirectoryDeregistered(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkSpaces Directory (%s) to deregister: %s", d.Id(), err)
	}

	return diags
}

func ExpandWorkspaceAccessProperties(properties []interface{}) *types.WorkspaceAccessProperties {
	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	result := &types.WorkspaceAccessProperties{}

	p := properties[0].(map[string]interface{})

	if p["device_type_android"].(string) != "" {
		result.DeviceTypeAndroid = types.AccessPropertyValue(p["device_type_android"].(string))
	}

	if p["device_type_chromeos"].(string) != "" {
		result.DeviceTypeChromeOs = types.AccessPropertyValue(p["device_type_chromeos"].(string))
	}

	if p["device_type_ios"].(string) != "" {
		result.DeviceTypeIos = types.AccessPropertyValue(p["device_type_ios"].(string))
	}

	if p["device_type_linux"].(string) != "" {
		result.DeviceTypeLinux = types.AccessPropertyValue(p["device_type_linux"].(string))
	}

	if p["device_type_osx"].(string) != "" {
		result.DeviceTypeOsx = types.AccessPropertyValue(p["device_type_osx"].(string))
	}

	if p["device_type_web"].(string) != "" {
		result.DeviceTypeWeb = types.AccessPropertyValue(p["device_type_web"].(string))
	}

	if p["device_type_windows"].(string) != "" {
		result.DeviceTypeWindows = types.AccessPropertyValue(p["device_type_windows"].(string))
	}

	if p["device_type_zeroclient"].(string) != "" {
		result.DeviceTypeZeroClient = types.AccessPropertyValue(p["device_type_zeroclient"].(string))
	}

	return result
}

func ExpandSelfServicePermissions(permissions []interface{}) *types.SelfservicePermissions {
	if len(permissions) == 0 || permissions[0] == nil {
		return nil
	}

	result := &types.SelfservicePermissions{}

	p := permissions[0].(map[string]interface{})

	if p["change_compute_type"].(bool) {
		result.ChangeComputeType = types.ReconnectEnumEnabled
	} else {
		result.ChangeComputeType = types.ReconnectEnumDisabled
	}

	if p["increase_volume_size"].(bool) {
		result.IncreaseVolumeSize = types.ReconnectEnumEnabled
	} else {
		result.IncreaseVolumeSize = types.ReconnectEnumDisabled
	}

	if p["rebuild_workspace"].(bool) {
		result.RebuildWorkspace = types.ReconnectEnumEnabled
	} else {
		result.RebuildWorkspace = types.ReconnectEnumDisabled
	}

	if p["restart_workspace"].(bool) {
		result.RestartWorkspace = types.ReconnectEnumEnabled
	} else {
		result.RestartWorkspace = types.ReconnectEnumDisabled
	}

	if p["switch_running_mode"].(bool) {
		result.SwitchRunningMode = types.ReconnectEnumEnabled
	} else {
		result.SwitchRunningMode = types.ReconnectEnumDisabled
	}

	return result
}

func ExpandWorkspaceCreationProperties(properties []interface{}) *types.WorkspaceCreationProperties {
	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	p := properties[0].(map[string]interface{})

	result := &types.WorkspaceCreationProperties{
		EnableInternetAccess:            aws.Bool(p["enable_internet_access"].(bool)),
		EnableMaintenanceMode:           aws.Bool(p["enable_maintenance_mode"].(bool)),
		UserEnabledAsLocalAdministrator: aws.Bool(p["user_enabled_as_local_administrator"].(bool)),
	}

	if p["custom_security_group_id"].(string) != "" {
		result.CustomSecurityGroupId = aws.String(p["custom_security_group_id"].(string))
	}

	if p["default_ou"].(string) != "" {
		result.DefaultOu = aws.String(p["default_ou"].(string))
	}

	return result
}

func FlattenWorkspaceAccessProperties(properties *types.WorkspaceAccessProperties) []interface{} {
	if properties == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"device_type_android":    string(properties.DeviceTypeAndroid),
			"device_type_chromeos":   string(properties.DeviceTypeChromeOs),
			"device_type_ios":        string(properties.DeviceTypeIos),
			"device_type_linux":      string(properties.DeviceTypeLinux),
			"device_type_osx":        string(properties.DeviceTypeOsx),
			"device_type_web":        string(properties.DeviceTypeWeb),
			"device_type_windows":    string(properties.DeviceTypeWindows),
			"device_type_zeroclient": string(properties.DeviceTypeZeroClient),
		},
	}
}

func FlattenSelfServicePermissions(permissions *types.SelfservicePermissions) []interface{} {
	if permissions == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}

	switch permissions.ChangeComputeType {
	case types.ReconnectEnumEnabled:
		result["change_compute_type"] = true
	case types.ReconnectEnumDisabled:
		result["change_compute_type"] = false
	default:
		result["change_compute_type"] = nil
	}

	switch permissions.IncreaseVolumeSize {
	case types.ReconnectEnumEnabled:
		result["increase_volume_size"] = true
	case types.ReconnectEnumDisabled:
		result["increase_volume_size"] = false
	default:
		result["increase_volume_size"] = nil
	}

	switch permissions.RebuildWorkspace {
	case types.ReconnectEnumEnabled:
		result["rebuild_workspace"] = true
	case types.ReconnectEnumDisabled:
		result["rebuild_workspace"] = false
	default:
		result["rebuild_workspace"] = nil
	}

	switch permissions.RestartWorkspace {
	case types.ReconnectEnumEnabled:
		result["restart_workspace"] = true
	case types.ReconnectEnumDisabled:
		result["restart_workspace"] = false
	default:
		result["restart_workspace"] = nil
	}

	switch permissions.SwitchRunningMode {
	case types.ReconnectEnumEnabled:
		result["switch_running_mode"] = true
	case types.ReconnectEnumDisabled:
		result["switch_running_mode"] = false
	default:
		result["switch_running_mode"] = nil
	}

	return []interface{}{result}
}

func FlattenWorkspaceCreationProperties(properties *types.DefaultWorkspaceCreationProperties) []interface{} {
	if properties == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"custom_security_group_id":            aws.ToString(properties.CustomSecurityGroupId),
			"default_ou":                          aws.ToString(properties.DefaultOu),
			"enable_internet_access":              aws.ToBool(properties.EnableInternetAccess),
			"enable_maintenance_mode":             aws.ToBool(properties.EnableMaintenanceMode),
			"user_enabled_as_local_administrator": aws.ToBool(properties.UserEnabledAsLocalAdministrator),
		},
	}
}

func flattenAccessPropertyEnumValues(t []types.AccessPropertyValue) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
