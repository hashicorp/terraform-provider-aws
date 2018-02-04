package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsWorkspacesWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesWorkspaceCreate,
		Read:   resourceAwsWorkspacesWorkspaceRead,
		Update: resourceAwsWorkspacesWorkspaceUpdate,
		Delete: resourceAwsWorkspacesWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"root_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"volume_encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"workspace_properties": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"running_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"root_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"running_mode_auto_stop_timeout_in_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"user_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsWorkspacesWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wsconn

	input := &workspaces.CreateWorkspacesInput{
		Workspaces: buildWorkspaceRequests(d, meta),
	}

	resp, err := conn.CreateWorkspaces(input)
	if err != nil {
		return err
	}

	if len(resp.FailedRequests) > 0 {
		failReq := resp.FailedRequests[0]
		return fmt.Errorf("Creating Workspace failed due to %s", *failReq.ErrorMessage)
	}

	pendingReq := resp.PendingRequests[0]

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceStatePending,
			workspaces.WorkspaceStateStarting,
		},
		Target:  []string{workspaces.WorkspaceStateAvailable},
		Refresh: workspaceRefreshStatusFunc(conn, *pendingReq.WorkspaceId),
		Timeout: 25 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[WARN] Error waiting for Workspace state to be \"%s\": %s", workspaces.WorkspaceStateAvailable, err)
	}

	d.SetId(*pendingReq.WorkspaceId)

	return resourceAwsWorkspacesWorkspaceRead(d, meta)
}

func resourceAwsWorkspacesWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wsconn

	input := &workspaces.DescribeWorkspacesInput{
		WorkspaceIds: []*string{aws.String(d.Id())},
	}

	resp, err := conn.DescribeWorkspaces(input)
	if err != nil {
		if isAWSErr(err, workspaces.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Workspace (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	workspace := resp.Workspaces[0]
	d.Set("bundle_id", workspace.BundleId)
	d.Set("bundle_id", workspace.DirectoryId)
	d.Set("root_volume_encryption_enabled", workspace.RootVolumeEncryptionEnabled)
	d.Set("user_name", workspace.UserName)
	d.Set("user_volume_encryption_enabled", workspace.UserVolumeEncryptionEnabled)
	d.Set("volume_encryption_key", workspace.VolumeEncryptionKey)
	d.Set("workspace_properties", flattenWorkspaceProperties(workspace.WorkspaceProperties))

	return nil
}

func resourceAwsWorkspacesWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wsconn

	input := &workspaces.ModifyWorkspacePropertiesInput{
		WorkspaceId: aws.String(d.Id()),
	}

	if d.HasChange("workspace_properties") {
		properties := d.Get("workspace_properties").(*schema.Set).List()
		if len(properties) > 0 {
			prop := properties[0].(map[string]interface{})
			input.WorkspaceProperties = expandWorkspaceProperties(prop)
		}
	}

	_, err := conn.ModifyWorkspaceProperties(input)
	if err != nil {
		if isAWSErr(err, workspaces.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Workspace (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	return nil
}

func resourceAwsWorkspacesWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wsconn

	input := &workspaces.TerminateWorkspacesInput{
		TerminateWorkspaceRequests: []*workspaces.TerminateRequest{
			&workspaces.TerminateRequest{
				WorkspaceId: aws.String(d.Id()),
			},
		},
	}

	_, err := conn.TerminateWorkspaces(input)
	if err != nil {
		if isAWSErr(err, workspaces.ErrCodeResourceNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceStatePending,
			workspaces.WorkspaceStateAvailable,
			workspaces.WorkspaceStateImpaired,
			workspaces.WorkspaceStateUnhealthy,
			workspaces.WorkspaceStateRebooting,
			workspaces.WorkspaceStateStarting,
			workspaces.WorkspaceStateRebuilding,
			workspaces.WorkspaceStateMaintenance,
			workspaces.WorkspaceStateSuspended,
			workspaces.WorkspaceStateUpdating,
			workspaces.WorkspaceStateStopping,
			workspaces.WorkspaceStateStopped,
		},
		Target: []string{
			workspaces.WorkspaceStateTerminating,
			workspaces.WorkspaceStateTerminated,
		},
		Refresh: workspaceRefreshStatusFunc(conn, d.Id()),
		Timeout: 25 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[WARN] Error waiting for Workspace state to be \"%s\", \"%s\": %s",
			workspaces.WorkspaceStateTerminating,
			workspaces.WorkspaceStateTerminated,
			err)
	}

	return nil
}

func buildWorkspaceRequests(d *schema.ResourceData, meta interface{}) []*workspaces.WorkspaceRequest {
	req := &workspaces.WorkspaceRequest{
		BundleId:                    aws.String(d.Get("bundle_id").(string)),
		DirectoryId:                 aws.String(d.Get("directory_id").(string)),
		UserName:                    aws.String(d.Get("user_name").(string)),
		RootVolumeEncryptionEnabled: aws.Bool(d.Get("root_volume_encryption_enabled").(bool)),
		UserVolumeEncryptionEnabled: aws.Bool(d.Get("user_volume_encryption_enabled").(bool)),
	}

	if v, ok := d.GetOk("volume_encryption_key"); ok {
		req.VolumeEncryptionKey = aws.String(v.(string))
	}
	properties := d.Get("workspace_properties").(*schema.Set).List()
	if len(properties) > 0 {
		prop := properties[0].(map[string]interface{})
		req.WorkspaceProperties = expandWorkspaceProperties(prop)
	}
	return []*workspaces.WorkspaceRequest{req}
}

func expandWorkspaceProperties(configured map[string]interface{}) *workspaces.WorkspaceProperties {
	if len(configured) > 0 {
		return nil
	}

	wp := &workspaces.WorkspaceProperties{}
	if v, ok := configured["compute_type_name"]; ok && v.(string) != "" {
		wp.SetComputeTypeName(v.(string))
	}
	if v, ok := configured["running_mode"]; ok && v.(string) != "" {
		wp.SetRunningMode(v.(string))
	}
	if v, ok := configured["root_volume_size_gib"]; ok && v.(int64) != 0 {
		wp.SetRootVolumeSizeGib(v.(int64))
	}
	if v, ok := configured["running_mode_auto_stop_timeout_in_minutes"]; ok && v.(int64) != 0 {
		wp.SetRunningModeAutoStopTimeoutInMinutes(v.(int64))
	}
	if v, ok := configured["user_volume_size_gib"]; ok && v.(int64) != 0 {
		wp.SetUserVolumeSizeGib(v.(int64))
	}

	return wp
}

func flattenWorkspaceProperties(wp *workspaces.WorkspaceProperties) []map[string]interface{} {
	if wp == nil {
		return nil
	}

	result := make(map[string]interface{}, 1)
	if wp.ComputeTypeName != nil {
		result["compute_type_name"] = *wp.ComputeTypeName
	}
	if wp.RunningMode != nil {
		result["running_mode"] = *wp.RunningMode
	}
	if wp.RootVolumeSizeGib != nil {
		result["root_volume_size_gib"] = *wp.RootVolumeSizeGib
	}
	if wp.RunningModeAutoStopTimeoutInMinutes != nil {
		result["running_mode_auto_stop_timeout_in_minutes"] = *wp.RunningModeAutoStopTimeoutInMinutes
	}
	if wp.UserVolumeSizeGib != nil {
		result["user_volume_size_gib"] = *wp.UserVolumeSizeGib
	}

	return []map[string]interface{}{result}
}

func workspaceRefreshStatusFunc(conn *workspaces.WorkSpaces, workspaceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &workspaces.DescribeWorkspacesInput{
			WorkspaceIds: []*string{aws.String(workspaceID)},
		}
		resp, err := conn.DescribeWorkspaces(input)
		if err != nil {
			return nil, "failed", err
		}
		workspace := resp.Workspaces[0]
		return workspace, *workspace.State, nil
	}
}
