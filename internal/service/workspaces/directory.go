package workspaces

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceDirectoryCreate,
		Read:   resourceDirectoryRead,
		Update: resourceDirectoryUpdate,
		Delete: resourceDirectoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"alias": {
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
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_chromeos": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_ios": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_linux": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_osx": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_web": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_windows": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
						},
						"device_type_zeroclient": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(workspaces.AccessPropertyValue_Values(), false),
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

func resourceDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	directoryID := d.Get("directory_id").(string)

	input := &workspaces.RegisterWorkspaceDirectoryInput{
		DirectoryId:       aws.String(directoryID),
		EnableSelfService: aws.Bool(false), // this is handled separately below
		EnableWorkDocs:    aws.Bool(false),
		Tenancy:           aws.String(workspaces.TenancyShared),
		Tags:              tags.IgnoreAws().WorkspacesTags(),
	}

	if v, ok := d.GetOk("subnet_ids"); ok {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Registering WorkSpaces Directory: %s", input)
	_, err := tfresource.RetryWhenAwsErrCodeEquals(
		DirectoryRegisterInvalidResourceStateTimeout,
		func() (interface{}, error) {
			return conn.RegisterWorkspaceDirectory(input)
		},
		// "error registering WorkSpaces Directory (d-000000000000): InvalidResourceStateException: The specified directory is not in a valid state. Confirm that the directory has a status of Active, and try again."
		workspaces.ErrCodeInvalidResourceStateException,
	)

	if err != nil {
		return fmt.Errorf("error registering WorkSpaces Directory (%s): %w", directoryID, err)
	}

	d.SetId(directoryID)

	_, err = WaitDirectoryRegistered(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for WorkSpaces Directory (%s) to register: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("self_service_permissions"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) self-service permissions", directoryID)
		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(directoryID),
			SelfservicePermissions: ExpandSelfServicePermissions(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting WorkSpaces Directory (%s) self-service permissions: %w", directoryID, err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) self-service permissions", directoryID)
	}

	if v, ok := d.GetOk("workspace_access_properties"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) access properties", directoryID)
		_, err := conn.ModifyWorkspaceAccessProperties(&workspaces.ModifyWorkspaceAccessPropertiesInput{
			ResourceId:                aws.String(directoryID),
			WorkspaceAccessProperties: ExpandWorkspaceAccessProperties(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting WorkSpaces Directory (%s) access properties: %w", directoryID, err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) access properties", directoryID)
	}

	if v, ok := d.GetOk("workspace_creation_properties"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) creation properties", directoryID)
		_, err := conn.ModifyWorkspaceCreationProperties(&workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(directoryID),
			WorkspaceCreationProperties: ExpandWorkspaceCreationProperties(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting WorkSpaces Directory (%s) creation properties: %w", directoryID, err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) creation properties", directoryID)
	}

	if v, ok := d.GetOk("ip_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		ipGroupIds := v.(*schema.Set)
		log.Printf("[DEBUG] Associating WorkSpaces Directory (%s) with IP Groups %s", directoryID, ipGroupIds.List())
		_, err := conn.AssociateIpGroups(&workspaces.AssociateIpGroupsInput{
			DirectoryId: aws.String(directoryID),
			GroupIds:    flex.ExpandStringSet(ipGroupIds),
		})
		if err != nil {
			return fmt.Errorf("error asassociating WorkSpaces Directory (%s) ip groups: %w", directoryID, err)
		}
		log.Printf("[INFO] Associated WorkSpaces Directory (%s) IP Groups", directoryID)
	}

	return resourceDirectoryRead(d, meta)
}

func resourceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	directory, err := FindDirectoryByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading WorkSpaces Directory (%s): %w", d.Id(), err)
	}

	d.Set("directory_id", directory.DirectoryId)
	if err := d.Set("subnet_ids", flex.FlattenStringSet(directory.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("registration_code", directory.RegistrationCode)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set("alias", directory.Alias)

	if err := d.Set("self_service_permissions", FlattenSelfServicePermissions(directory.SelfservicePermissions)); err != nil {
		return fmt.Errorf("error setting self_service_permissions: %w", err)
	}

	if err := d.Set("workspace_access_properties", FlattenWorkspaceAccessProperties(directory.WorkspaceAccessProperties)); err != nil {
		return fmt.Errorf("error setting workspace_access_properties: %w", err)
	}

	if err := d.Set("workspace_creation_properties", FlattenWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return fmt.Errorf("error setting workspace_creation_properties: %w", err)
	}

	if err := d.Set("ip_group_ids", flex.FlattenStringSet(directory.IpGroupIds)); err != nil {
		return fmt.Errorf("error setting ip_group_ids: %w", err)
	}

	if err := d.Set("dns_ip_addresses", flex.FlattenStringSet(directory.DnsIpAddresses)); err != nil {
		return fmt.Errorf("error setting dns_ip_addresses: %w", err)
	}

	tags, err := tftags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags: %w", err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn

	if d.HasChange("self_service_permissions") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) self-service permissions", d.Id())
		permissions := d.Get("self_service_permissions").([]interface{})

		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: ExpandSelfServicePermissions(permissions),
		})
		if err != nil {
			return fmt.Errorf("error updating WorkSpaces Directory (%s) self service permissions: %w", d.Id(), err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) self-service permissions", d.Id())
	}

	if d.HasChange("workspace_access_properties") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) access properties", d.Id())
		properties := d.Get("workspace_access_properties").([]interface{})

		_, err := conn.ModifyWorkspaceAccessProperties(&workspaces.ModifyWorkspaceAccessPropertiesInput{
			ResourceId:                aws.String(d.Id()),
			WorkspaceAccessProperties: ExpandWorkspaceAccessProperties(properties),
		})
		if err != nil {
			return fmt.Errorf("error updating WorkSpaces Directory (%s) access properties: %w", d.Id(), err)
		}
		log.Printf("[INFO] Modified WorkSpaces Directory (%s) access properties", d.Id())
	}

	if d.HasChange("workspace_creation_properties") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory (%s) creation properties", d.Id())
		properties := d.Get("workspace_creation_properties").([]interface{})

		_, err := conn.ModifyWorkspaceCreationProperties(&workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(d.Id()),
			WorkspaceCreationProperties: ExpandWorkspaceCreationProperties(properties),
		})
		if err != nil {
			return fmt.Errorf("error updating WorkSpaces Directory (%s) creation properties: %w", d.Id(), err)
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
		_, err := conn.AssociateIpGroups(&workspaces.AssociateIpGroupsInput{
			DirectoryId: aws.String(d.Id()),
			GroupIds:    flex.ExpandStringSet(added),
		})
		if err != nil {
			return fmt.Errorf("error asassociating WorkSpaces Directory (%s) IP Groups: %w", d.Id(), err)
		}

		log.Printf("[DEBUG] Disassociating WorkSpaces Directory (%s) with IP Groups %s", d.Id(), removed.GoString())
		_, err = conn.DisassociateIpGroups(&workspaces.DisassociateIpGroupsInput{
			DirectoryId: aws.String(d.Id()),
			GroupIds:    flex.ExpandStringSet(removed),
		})
		if err != nil {
			return fmt.Errorf("error disasassociating WorkSpaces Directory (%s) IP Groups: %w", d.Id(), err)
		}

		log.Printf("[INFO] Updated WorkSpaces Directory (%s) IP Groups", d.Id())
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.WorkspacesUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceDirectoryRead(d, meta)
}

func resourceDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn

	log.Printf("[DEBUG] Deregistering WorkSpaces Directory: %s", d.Id())
	_, err := tfresource.RetryWhenAwsErrCodeEquals(
		DirectoryDeregisterInvalidResourceStateTimeout,
		func() (interface{}, error) {
			return conn.DeregisterWorkspaceDirectory(&workspaces.DeregisterWorkspaceDirectoryInput{
				DirectoryId: aws.String(d.Id()),
			})
		},
		// "error deregistering WorkSpaces Directory (d-000000000000): InvalidResourceStateException: The specified directory is not in a valid state. Confirm that the directory has a status of Active, and try again."
		workspaces.ErrCodeInvalidResourceStateException,
	)

	if tfawserr.ErrCodeEquals(err, workspaces.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deregistering WorkSpaces Directory (%s): %w", d.Id(), err)
	}

	_, err = WaitDirectoryDeregistered(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for WorkSpaces Directory (%s) to deregister: %w", d.Id(), err)
	}

	return nil
}

func ExpandWorkspaceAccessProperties(properties []interface{}) *workspaces.WorkspaceAccessProperties {
	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	result := &workspaces.WorkspaceAccessProperties{}

	p := properties[0].(map[string]interface{})

	if p["device_type_android"].(string) != "" {
		result.DeviceTypeAndroid = aws.String(p["device_type_android"].(string))
	}

	if p["device_type_chromeos"].(string) != "" {
		result.DeviceTypeChromeOs = aws.String(p["device_type_chromeos"].(string))
	}

	if p["device_type_ios"].(string) != "" {
		result.DeviceTypeIos = aws.String(p["device_type_ios"].(string))
	}

	if p["device_type_linux"].(string) != "" {
		result.DeviceTypeLinux = aws.String(p["device_type_linux"].(string))
	}

	if p["device_type_osx"].(string) != "" {
		result.DeviceTypeOsx = aws.String(p["device_type_osx"].(string))
	}

	if p["device_type_web"].(string) != "" {
		result.DeviceTypeWeb = aws.String(p["device_type_web"].(string))
	}

	if p["device_type_windows"].(string) != "" {
		result.DeviceTypeWindows = aws.String(p["device_type_windows"].(string))
	}

	if p["device_type_zeroclient"].(string) != "" {
		result.DeviceTypeZeroClient = aws.String(p["device_type_zeroclient"].(string))
	}

	return result
}

func ExpandSelfServicePermissions(permissions []interface{}) *workspaces.SelfservicePermissions {
	if len(permissions) == 0 || permissions[0] == nil {
		return nil
	}

	result := &workspaces.SelfservicePermissions{}

	p := permissions[0].(map[string]interface{})

	if p["change_compute_type"].(bool) {
		result.ChangeComputeType = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.ChangeComputeType = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["increase_volume_size"].(bool) {
		result.IncreaseVolumeSize = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.IncreaseVolumeSize = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["rebuild_workspace"].(bool) {
		result.RebuildWorkspace = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.RebuildWorkspace = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["restart_workspace"].(bool) {
		result.RestartWorkspace = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.RestartWorkspace = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["switch_running_mode"].(bool) {
		result.SwitchRunningMode = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.SwitchRunningMode = aws.String(workspaces.ReconnectEnumDisabled)
	}

	return result
}

func ExpandWorkspaceCreationProperties(properties []interface{}) *workspaces.WorkspaceCreationProperties {
	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	p := properties[0].(map[string]interface{})

	result := &workspaces.WorkspaceCreationProperties{
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

func FlattenWorkspaceAccessProperties(properties *workspaces.WorkspaceAccessProperties) []interface{} {
	if properties == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"device_type_android":    aws.StringValue(properties.DeviceTypeAndroid),
			"device_type_chromeos":   aws.StringValue(properties.DeviceTypeChromeOs),
			"device_type_ios":        aws.StringValue(properties.DeviceTypeIos),
			"device_type_linux":      aws.StringValue(properties.DeviceTypeLinux),
			"device_type_osx":        aws.StringValue(properties.DeviceTypeOsx),
			"device_type_web":        aws.StringValue(properties.DeviceTypeWeb),
			"device_type_windows":    aws.StringValue(properties.DeviceTypeWindows),
			"device_type_zeroclient": aws.StringValue(properties.DeviceTypeZeroClient),
		},
	}
}

func FlattenSelfServicePermissions(permissions *workspaces.SelfservicePermissions) []interface{} {
	if permissions == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}

	switch *permissions.ChangeComputeType {
	case workspaces.ReconnectEnumEnabled:
		result["change_compute_type"] = true
	case workspaces.ReconnectEnumDisabled:
		result["change_compute_type"] = false
	default:
		result["change_compute_type"] = nil
	}

	switch *permissions.IncreaseVolumeSize {
	case workspaces.ReconnectEnumEnabled:
		result["increase_volume_size"] = true
	case workspaces.ReconnectEnumDisabled:
		result["increase_volume_size"] = false
	default:
		result["increase_volume_size"] = nil
	}

	switch *permissions.RebuildWorkspace {
	case workspaces.ReconnectEnumEnabled:
		result["rebuild_workspace"] = true
	case workspaces.ReconnectEnumDisabled:
		result["rebuild_workspace"] = false
	default:
		result["rebuild_workspace"] = nil
	}

	switch *permissions.RestartWorkspace {
	case workspaces.ReconnectEnumEnabled:
		result["restart_workspace"] = true
	case workspaces.ReconnectEnumDisabled:
		result["restart_workspace"] = false
	default:
		result["restart_workspace"] = nil
	}

	switch *permissions.SwitchRunningMode {
	case workspaces.ReconnectEnumEnabled:
		result["switch_running_mode"] = true
	case workspaces.ReconnectEnumDisabled:
		result["switch_running_mode"] = false
	default:
		result["switch_running_mode"] = nil
	}

	return []interface{}{result}
}

func FlattenWorkspaceCreationProperties(properties *workspaces.DefaultWorkspaceCreationProperties) []interface{} {
	if properties == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"custom_security_group_id":            aws.StringValue(properties.CustomSecurityGroupId),
			"default_ou":                          aws.StringValue(properties.DefaultOu),
			"enable_internet_access":              aws.BoolValue(properties.EnableInternetAccess),
			"enable_maintenance_mode":             aws.BoolValue(properties.EnableMaintenanceMode),
			"user_enabled_as_local_administrator": aws.BoolValue(properties.UserEnabledAsLocalAdministrator),
		},
	}
}
