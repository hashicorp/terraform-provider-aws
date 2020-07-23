package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/workspaces/waiter"
)

func resourceAwsWorkspacesDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesDirectoryCreate,
		Read:   resourceAwsWorkspacesDirectoryRead,
		Update: resourceAwsWorkspacesDirectoryUpdate,
		Delete: resourceAwsWorkspacesDirectoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"creation_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  nil,
						},
						"default_ou": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"internet_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"maintenance_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"local_admin": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
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
			"workspace_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registration_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWorkspacesDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	directoryId := d.Get("directory_id").(string)

	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WorkspacesTags()

	input := &workspaces.RegisterWorkspaceDirectoryInput{
		DirectoryId:       aws.String(directoryId),
		EnableSelfService: aws.Bool(false), // this is handled separately below
		EnableWorkDocs:    aws.Bool(false),
		Tenancy:           aws.String(workspaces.TenancyShared),
		Tags:              tags,
	}

	if v, ok := d.GetOk("subnet_ids"); ok {
		input.SubnetIds = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Regestering WorkSpaces Directory...\n%#v\n", *input)
	_, err := conn.RegisterWorkspaceDirectory(input)
	if err != nil {
		return err
	}
	d.SetId(directoryId)

	log.Printf("[DEBUG] Waiting for WorkSpaces Directory %q to become registered...", directoryId)
	_, err = waiter.DirectoryRegistered(conn, directoryId)
	if err != nil {
		return fmt.Errorf("error registering directory: %s", err)
	}
	log.Printf("[DEBUG] WorkSpaces Directory %q is registered", directoryId)

	log.Printf("[DEBUG] Modifying WorkSpaces Directory %q self-service permissions...", directoryId)
	if v, ok := d.GetOk("self_service_permissions"); ok {
		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(directoryId),
			SelfservicePermissions: expandSelfServicePermissions(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting self service permissions: %s", err)
		}
	}
	log.Printf("[DEBUG] WorkSpaces Directory %q self-service permissions are set", directoryId)

	log.Printf("[DEBUG] Modifying workspaces directory %q creation properties...", d.Id())
	if v, ok := d.GetOk("creation_properties"); ok {
		_, err := conn.ModifyWorkspaceCreationProperties(&workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(directoryId),
			WorkspaceCreationProperties: expandCreationProperties(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting workspace creation properties: %s", err)
		}
	}
	log.Printf("[DEBUG] Workspaces directory %q creation properties are set", d.Id())

	return resourceAwsWorkspacesDirectoryRead(d, meta)
}

func resourceAwsWorkspacesDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	rawOutput, state, err := waiter.DirectoryState(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting WorkSpaces Directory (%s): %s", d.Id(), err)
	}
	if state == workspaces.WorkspaceDirectoryStateDeregistered {
		log.Printf("[WARN] WorkSpaces Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	directory := rawOutput.(*workspaces.WorkspaceDirectory)
	d.Set("directory_id", directory.DirectoryId)
	if err := d.Set("subnet_ids", flattenStringSet(directory.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("registration_code", directory.RegistrationCode)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set("alias", directory.Alias)
	if err := d.Set("self_service_permissions", flattenSelfServicePermissions(directory.SelfservicePermissions)); err != nil {
		return fmt.Errorf("error setting self_service_permissions: %s", err)
	}

	if err := d.Set("ip_group_ids", flattenStringSet(directory.IpGroupIds)); err != nil {
		return fmt.Errorf("error setting ip_group_ids: %s", err)
	}

	if err := d.Set("dns_ip_addresses", flattenStringSet(directory.DnsIpAddresses)); err != nil {
		return fmt.Errorf("error setting dns_ip_addresses: %s", err)
	}

	if err := d.Set("creation_properties", flattenCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return fmt.Errorf("error setting creation_properties: %s", err)
	}

	tags, err := keyvaluetags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags: %s", err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWorkspacesDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	if d.HasChange("self_service_permissions") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory %q self-service permissions...", d.Id())
		permissions := d.Get("self_service_permissions").([]interface{})

		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: expandSelfServicePermissions(permissions),
		})
		if err != nil {
			return fmt.Errorf("error updating self service permissions: %s", err)
		}
		log.Printf("[DEBUG] WorkSpaces Directory %q self-service permissions are set", d.Id())
	}

	if d.HasChange("creation_properties") {
		log.Printf("[DEBUG] Modifying workspaces directory %q creation properties...", d.Id())
		permissions := d.Get("creation_properties").([]interface{})

		_, err := conn.ModifyWorkspaceCreationProperties(&workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(d.Id()),
			WorkspaceCreationProperties: expandCreationProperties(permissions),
		})
		if err != nil {
			return fmt.Errorf("error updating workspace creation properties: %s", err)
		}
		log.Printf("[DEBUG] Workspaces directory %q workspace creation properties are set", d.Id())
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.WorkspacesUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWorkspacesDirectoryRead(d, meta)
}

func resourceAwsWorkspacesDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	err := workspacesDirectoryDelete(d.Id(), conn)
	if err != nil {
		return fmt.Errorf("error deleting WorkSpaces Directory (%s): %s", d.Id(), err)
	}

	return nil
}

func workspacesDirectoryDelete(id string, conn *workspaces.WorkSpaces) error {
	log.Printf("[DEBUG] Deregistering WorkSpaces Directory %q", id)
	_, err := conn.DeregisterWorkspaceDirectory(&workspaces.DeregisterWorkspaceDirectoryInput{
		DirectoryId: aws.String(id),
	})
	if err != nil {
		return fmt.Errorf("error deregistering WorkSpaces Directory %q: %w", id, err)
	}

	log.Printf("[DEBUG] Waiting for WorkSpaces Directory %q to be deregistered", id)
	_, err = waiter.DirectoryDeregistered(conn, id)
	if err != nil {
		return fmt.Errorf("error waiting for WorkSpaces Directory %q to be deregistered: %w", id, err)
	}
	log.Printf("[DEBUG] WorkSpaces Directory %q is deregistered", id)

	return nil
}

func expandSelfServicePermissions(permissions []interface{}) *workspaces.SelfservicePermissions {
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

func flattenSelfServicePermissions(permissions *workspaces.SelfservicePermissions) []interface{} {
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

func expandCreationProperties(permissions []interface{}) *workspaces.WorkspaceCreationProperties {
	if len(permissions) == 0 || permissions[0] == nil {
		return nil
	}

	result := &workspaces.WorkspaceCreationProperties{}

	p := permissions[0].(map[string]interface{})

	if len(p["security_group"].(string)) > 0 {
		result.CustomSecurityGroupId = aws.String(p["security_group"].(string))
	}
	if len(p["default_ou"].(string)) > 0 {
		result.DefaultOu = aws.String(p["default_ou"].(string))
	}
	result.EnableInternetAccess = aws.Bool(p["internet_access"].(bool))
	result.EnableMaintenanceMode = aws.Bool(p["maintenance_mode"].(bool))
	result.UserEnabledAsLocalAdministrator = aws.Bool(p["local_admin"].(bool))

	return result
}

func flattenCreationProperties(permissions *workspaces.DefaultWorkspaceCreationProperties) []interface{} {
	if permissions == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}

	result["security_group"] = permissions.CustomSecurityGroupId
	result["default_ou"] = permissions.DefaultOu
	result["internet_access"] = permissions.EnableInternetAccess
	result["maintenance_mode"] = permissions.EnableMaintenanceMode
	result["local_admin"] = permissions.UserEnabledAsLocalAdministrator

	return []interface{}{result}
}
