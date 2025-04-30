// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_workspaces_pool", name="Pool")
// @Tags(identifierAttribute="id")
func ResourcePool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePoolCreate,
		ReadWithoutTimeout:   resourcePoolRead,
		UpdateWithoutTimeout: resourcePoolUpdate,
		DeleteWithoutTimeout: resourcePoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"application_settings": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrS3BucketName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      false,
							ValidateFunc: validation.StringInSlice(enum.Slice(types.ApplicationSettingsStatusEnumEnabled, types.ApplicationSettingsStatusEnumDisabled), false),
						},
						"settings_group": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"capacity": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"available_user_sessions": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"actual_user_sessions": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"active_user_sessions": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"desired_user_sessions": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Required: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"timeout_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_user_duration_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      54000,
							ValidateFunc: validation.IntBetween(1, 43200),
						},
						"disconnect_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      900,
							ValidateFunc: validation.IntBetween(1, 36000),
						},
						"idle_disconnect_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      900,
							ValidateFunc: validation.IntBetween(1, 36000),
						},
					},
				},
			},
		},
	}
}

const (
	ResNamePool = "Pool"
)

func resourcePoolCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	in := &workspaces.CreateWorkspacesPoolInput{
		BundleId:    aws.String(d.Get("bundle_id").(string)),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		DirectoryId: aws.String(d.Get("directory_id").(string)),
		PoolName:    aws.String(d.Get("pool_name").(string)),
		Tags:        getTagsIn(ctx),
	}
	if v, ok := d.GetOk("application_settings"); ok {
		in.ApplicationSettings = expandApplicationSettings(v.([]any))
	}
	if v, ok := d.GetOk("capacity"); ok {
		in.Capacity.DesiredUserSessions = expandCapacity(v.([]any)).DesiredUserSessions
	}
	if v, ok := d.GetOk("timeout_settings"); ok {
		in.TimeoutSettings = expandTimeoutSettings(v.([]any))
	}

	out, err := conn.CreateWorkspacesPool(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionCreating, ResNamePool, d.Get("pool_name").(string), err)
	}

	d.SetId(aws.ToString(out.WorkspacesPool.PoolId))

	if _, err := waitPoolCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionCreating, ResNamePool, d.Get("pool_name").(string), err)
	}

	if err := startPool(ctx, conn, d.Id()); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionCreating, ResNamePool, d.Get("pool_name").(string), err)
	}

	return append(diags, resourcePoolRead(ctx, d, meta)...)
}

func resourcePoolRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	out, err := findPoolByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionReading, ResNamePool, d.Id(), err)
	}

	if err := d.Set("application_settings", flattenApplicationSettings(out.ApplicationSettings)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionSetting, ResNamePool, d.Id(), err)
	}
	d.Set(names.AttrARN, out.PoolArn)
	d.Set("bundle_id", out.BundleId)
	if err := d.Set("capacity", flattenCapacity(out.CapacityStatus)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionSetting, ResNamePool, d.Id(), err)
	}
	d.Set(names.AttrDescription, out.Description)
	d.Set("directory_id", out.DirectoryId)
	d.Set(names.AttrID, out.PoolId)
	d.Set("pool_name", out.PoolName)
	d.Set(names.AttrState, out.State)
	if err := d.Set("timeout_settings", flattenTimeoutSettings(out.TimeoutSettings)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionSetting, ResNamePool, d.Id(), err)
	}

	return diags
}

func resourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	update := false
	shouldStop := false
	currentState := d.Get(names.AttrState).(string)

	in := &workspaces.UpdateWorkspacesPoolInput{
		PoolId: aws.String(d.Id()),
	}

	if d.HasChange("bundle_id") {
		shouldStop = true
		in.BundleId = aws.String(d.Get("bundle_id").(string))
		update = true
	}

	if d.HasChange("directory_id") {
		shouldStop = true
		in.DirectoryId = aws.String(d.Get("directory_id").(string))
		update = true
	}

	if d.HasChange("application_settings") {
		in.ApplicationSettings = expandApplicationSettings(d.Get("application_settings").([]any))
		update = true
	}

	if d.HasChange("capacity") {
		in.Capacity.DesiredUserSessions = expandCapacity(d.Get("capacity").([]any)).DesiredUserSessions
		update = true
	}

	if d.HasChange("timeout_settings") {
		timeoutSettings := expandTimeoutSettings(d.Get("timeout_settings").([]any))
		timeoutSettingsRequest := &types.TimeoutSettings{}

		old, new := d.GetChange("timeout_settings")
		oldSettings := old.([]any)
		newSettings := new.([]any)

		if len(oldSettings) > 0 && len(newSettings) > 0 {
			oldMaxUserDuration := oldSettings[0].(map[string]any)["max_user_duration_in_seconds"].(int)
			newMaxUserDuration := newSettings[0].(map[string]any)["max_user_duration_in_seconds"].(int)

			if oldMaxUserDuration != newMaxUserDuration {
				log.Printf("[DEBUG] max_user_duration_in_seconds changed from %d to %d", oldMaxUserDuration, newMaxUserDuration)
				shouldStop = true
				timeoutSettingsRequest.MaxUserDurationInSeconds = timeoutSettings.MaxUserDurationInSeconds
			}
		}

		timeoutSettingsRequest.DisconnectTimeoutInSeconds = timeoutSettings.DisconnectTimeoutInSeconds
		timeoutSettingsRequest.IdleDisconnectTimeoutInSeconds = timeoutSettings.IdleDisconnectTimeoutInSeconds
		in.TimeoutSettings = timeoutSettingsRequest
		update = true
	}

	if shouldStop && currentState != string(types.WorkspacesPoolStateStopped) {
		if err := stopPool(ctx, conn, d.Id()); err != nil {
			return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionUpdating, ResNamePool, d.Id(), err)
		}
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating WorkSpaces Pool (%s)", d.Id())
	_, err := conn.UpdateWorkspacesPool(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionUpdating, ResNamePool, d.Id(), err)
	}

	if _, err := waitPoolUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionUpdating, ResNamePool, d.Id(), err)
	}

	// If the pool was stopped, we need to start it again
	if shouldStop && currentState != string(types.WorkspacesPoolStateStopped) {
		log.Printf("[DEBUG] Starting WorkSpaces Pool (%s)", d.Id())
		if err := startPool(ctx, conn, d.Id()); err != nil {
			return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionUpdating, ResNamePool, d.Id(), err)
		}
	}

	return append(diags, resourcePoolRead(ctx, d, meta)...)
}

func resourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	log.Printf("[DEBUG] Deleting WorkSpaces Pool (%s)", d.Id())

	if err := stopPool(ctx, conn, d.Id()); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionDeleting, ResNamePool, d.Id(), err)
	}

	input := &workspaces.TerminateWorkspacesPoolInput{
		PoolId: aws.String(d.Id()),
	}

	if _, err := conn.TerminateWorkspacesPool(ctx, input); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionDeleting, ResNamePool, d.Id(), err)
	}

	_, err := waitPoolDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionDeleting, ResNamePool, d.Id(), err)
	}

	return diags
}

func waitPoolCreated(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*types.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.WorkspacesPoolStateCreating),
		Target:                    enum.Slice(types.WorkspacesPoolStateStopped),
		Refresh:                   statusPool(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.WorkspacesPool); ok {
		return out, err
	}
	return nil, err
}

func waitPoolUpdated(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*types.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.WorkspacesPoolStateUpdating),
		Target:                    enum.Slice(types.WorkspacesPoolStateStopped, types.WorkspacesPoolStateRunning),
		Refresh:                   statusPool(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.WorkspacesPool); ok {
		return out, err
	}
	return nil, err
}

func waitPoolStopped(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*types.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspacesPoolStateStopping),
		Target:  enum.Slice(types.WorkspacesPoolStateStopped),
		Refresh: statusPool(ctx, conn, id),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.WorkspacesPool); ok {
		return out, err
	}
	return nil, err
}

func waitPoolDeleted(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*types.WorkspacesPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspacesPoolStateDeleting),
		Target:  []string{},
		Refresh: statusPool(ctx, conn, id),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.WorkspacesPool); ok {
		return out, err
	}
	return nil, err
}

func statusPool(ctx context.Context, conn *workspaces.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findPoolByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findPoolByID(ctx context.Context, conn *workspaces.Client, id string) (*types.WorkspacesPool, error) {
	input := &workspaces.DescribeWorkspacesPoolsInput{
		PoolIds: []string{id},
	}

	output, err := conn.DescribeWorkspacesPools(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.WorkspacesPools) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.WorkspacesPools[0], nil
}

func expandApplicationSettings(tfList []any) *types.ApplicationSettingsRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	apiObject := &types.ApplicationSettingsRequest{}
	if tfMap[names.AttrStatus] != nil {
		apiObject.Status = types.ApplicationSettingsStatusEnum(tfMap[names.AttrStatus].(string))
	}
	if tfMap["settings_group"] != nil {
		settingsGroup := tfMap["settings_group"].(string)
		apiObject.SettingsGroup = &settingsGroup
	}
	return apiObject
}

func expandCapacity(tfList []any) *types.Capacity {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	apiObject := &types.Capacity{}

	if tfMap["desired_user_sessions"] != nil {
		desiredUserSessions := int32(tfMap["desired_user_sessions"].(int))
		apiObject.DesiredUserSessions = &desiredUserSessions
	}
	return apiObject
}

func expandTimeoutSettings(tfList []any) *types.TimeoutSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	apiObject := &types.TimeoutSettings{}

	if tfMap["max_user_duration_in_seconds"] != nil {
		maxUserDurationInSeconds := int32(tfMap["max_user_duration_in_seconds"].(int))
		apiObject.MaxUserDurationInSeconds = &maxUserDurationInSeconds
	}
	if tfMap["disconnect_timeout_in_seconds"] != nil {
		disconnectTimeoutInSeconds := int32(tfMap["disconnect_timeout_in_seconds"].(int))
		apiObject.DisconnectTimeoutInSeconds = &disconnectTimeoutInSeconds
	}
	if tfMap["idle_disconnect_timeout_in_seconds"] != nil {
		idleDisconnectTimeoutInSeconds := int32(tfMap["idle_disconnect_timeout_in_seconds"].(int))
		apiObject.IdleDisconnectTimeoutInSeconds = &idleDisconnectTimeoutInSeconds
	}
	return apiObject
}

func flattenApplicationSettings(apiObject *types.ApplicationSettingsResponse) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			names.AttrStatus: string(apiObject.Status),
			"settings_group": *apiObject.SettingsGroup,
		},
	}
}

func flattenCapacity(apiObject *types.CapacityStatus) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			"available_user_sessions": apiObject.AvailableUserSessions,
			"actual_user_sessions":    apiObject.ActualUserSessions,
			"active_user_sessions":    apiObject.ActiveUserSessions,
			"desired_user_sessions":   apiObject.DesiredUserSessions,
		},
	}
}

func flattenTimeoutSettings(apiObject *types.TimeoutSettings) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			"max_user_duration_in_seconds":       apiObject.MaxUserDurationInSeconds,
			"disconnect_timeout_in_seconds":      apiObject.DisconnectTimeoutInSeconds,
			"idle_disconnect_timeout_in_seconds": apiObject.IdleDisconnectTimeoutInSeconds,
		},
	}
}

func stopPool(ctx context.Context, conn *workspaces.Client, id string) error {
	input := &workspaces.StopWorkspacesPoolInput{
		PoolId: aws.String(id),
	}

	if _, err := conn.StopWorkspacesPool(ctx, input); err != nil {
		return fmt.Errorf("stopping WorkSpaces Pool (%s): %w", id, err)
	}

	if _, err := waitPoolStopped(ctx, conn, id, 15*time.Minute); err != nil {
		return fmt.Errorf("waiting for WorkSpaces Pool (%s) to stop: %w", id, err)
	}

	return nil
}

func startPool(ctx context.Context, conn *workspaces.Client, id string) error {
	input := &workspaces.StartWorkspacesPoolInput{
		PoolId: aws.String(id),
	}

	if _, err := conn.StartWorkspacesPool(ctx, input); err != nil {
		return fmt.Errorf("starting WorkSpaces Pool (%s): %w", id, err)
	}

	return nil
}
