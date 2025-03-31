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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_workspaces_workspace", name="Workspace")
// @Tags(identifierAttribute="id")
func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceCreate,
		ReadWithoutTimeout:   resourceWorkspaceRead,
		UpdateWithoutTimeout: resourceWorkspaceUpdate,
		DeleteWithoutTimeout: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"computer_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrUserName: {
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
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type_name": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ComputeValue,
							ValidateDiagFunc: enum.Validate[types.Compute](),
						},
						"root_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  80,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{80}),
								validation.IntBetween(175, 2000),
							),
						},
						"running_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  types.RunningModeAlwaysOn,
							ValidateFunc: validation.StringInSlice(enum.Slice(
								types.RunningModeAlwaysOn,
								types.RunningModeAutoStop,
							), false),
						},
						"running_mode_auto_stop_timeout_in_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ValidateFunc: func(v any, k string) (ws []string, errors []error) {
								val := v.(int)
								if val%60 != 0 {
									errors = append(errors, fmt.Errorf(
										"%q should be configured in 60-minute intervals, got: %d", k, val))
								}
								return
							},
						},
						"user_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  10,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{10, 50}),
								validation.IntBetween(100, 2000),
							),
						},
					},
				},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	req := types.WorkspaceRequest{
		BundleId:                    aws.String(d.Get("bundle_id").(string)),
		DirectoryId:                 aws.String(d.Get("directory_id").(string)),
		RootVolumeEncryptionEnabled: aws.Bool(d.Get("root_volume_encryption_enabled").(bool)),
		Tags:                        getTagsIn(ctx),
		UserName:                    aws.String(d.Get(names.AttrUserName).(string)),
		UserVolumeEncryptionEnabled: aws.Bool(d.Get("user_volume_encryption_enabled").(bool)),
		WorkspaceProperties:         expandWorkspaceProperties(d.Get("workspace_properties").([]any)),
	}

	if v, ok := d.GetOk("volume_encryption_key"); ok {
		req.VolumeEncryptionKey = aws.String(v.(string))
	}

	input := workspaces.CreateWorkspacesInput{
		Workspaces: []types.WorkspaceRequest{req},
	}
	output, err := conn.CreateWorkspaces(ctx, &input)

	if err == nil && len(output.FailedRequests) > 0 {
		v := output.FailedRequests[0]
		err = fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkSpaces Workspace: %s", err)
	}

	d.SetId(aws.ToString(output.PendingRequests[0].WorkspaceId))

	if _, err := waitWorkspaceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkSpaces Workspace (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	workspace, err := findWorkspaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Workspace (%s): %s", d.Id(), err)
	}

	d.Set("bundle_id", workspace.BundleId)
	d.Set("computer_name", workspace.ComputerName)
	d.Set("directory_id", workspace.DirectoryId)
	d.Set(names.AttrIPAddress, workspace.IpAddress)
	d.Set("root_volume_encryption_enabled", workspace.RootVolumeEncryptionEnabled)
	d.Set(names.AttrState, workspace.State)
	d.Set(names.AttrUserName, workspace.UserName)
	d.Set("user_volume_encryption_enabled", workspace.UserVolumeEncryptionEnabled)
	d.Set("volume_encryption_key", workspace.VolumeEncryptionKey)
	if err := d.Set("workspace_properties", flattenWorkspaceProperties(workspace.WorkspaceProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_properties: %s", err)
	}

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	// IMPORTANT: Only one workspace property could be changed in a time.
	// I've create AWS Support feature request to allow multiple properties modification in a time.
	// https://docs.aws.amazon.com/workspaces/latest/adminguide/modify-workspaces.html

	if key := "workspace_properties.0.compute_type_name"; d.HasChange(key) {
		if err := workspacePropertyUpdate(ctx, conn, d, key); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if key := "workspace_properties.0.root_volume_size_gib"; d.HasChange(key) {
		if err := workspacePropertyUpdate(ctx, conn, d, key); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if key := "workspace_properties.0.running_mode"; d.HasChange(key) {
		if err := workspacePropertyUpdate(ctx, conn, d, key); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if key := "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"; d.HasChange(key) {
		if err := workspacePropertyUpdate(ctx, conn, d, key); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if key := "workspace_properties.0.user_volume_size_gib"; d.HasChange(key) {
		if err := workspacePropertyUpdate(ctx, conn, d, key); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	log.Printf("[DEBUG] Deleting WorkSpaces Workspace: %s", d.Id())
	input := workspaces.TerminateWorkspacesInput{
		TerminateWorkspaceRequests: []types.TerminateRequest{{
			WorkspaceId: aws.String(d.Id()),
		}},
	}
	output, err := conn.TerminateWorkspaces(ctx, &input)

	if err == nil && len(output.FailedRequests) > 0 {
		v := output.FailedRequests[0]
		err = fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WorkSpaces Workspace (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkspaceTerminated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkSpaces Workspace (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func workspacePropertyUpdate(ctx context.Context, conn *workspaces.Client, d *schema.ResourceData, key string) error {
	input := &workspaces.ModifyWorkspacePropertiesInput{
		WorkspaceId: aws.String(d.Id()),
	}

	switch key {
	case "workspace_properties.0.compute_type_name":
		input.WorkspaceProperties = &types.WorkspaceProperties{
			ComputeTypeName: types.Compute(d.Get(key).(string)),
		}
	case "workspace_properties.0.root_volume_size_gib":
		input.WorkspaceProperties = &types.WorkspaceProperties{
			RootVolumeSizeGib: aws.Int32(int32(d.Get(key).(int))),
		}
	case "workspace_properties.0.running_mode":
		input.WorkspaceProperties = &types.WorkspaceProperties{
			RunningMode: types.RunningMode(d.Get(key).(string)),
		}
	case "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes":
		if d.Get("workspace_properties.0.running_mode") != types.RunningModeAutoStop {
			log.Printf("[DEBUG] Property running_mode_auto_stop_timeout_in_minutes makes sense only for AUTO_STOP running mode")
			return nil
		}

		input.WorkspaceProperties = &types.WorkspaceProperties{
			RunningModeAutoStopTimeoutInMinutes: aws.Int32(int32(d.Get(key).(int))),
		}
	case "workspace_properties.0.user_volume_size_gib":
		input.WorkspaceProperties = &types.WorkspaceProperties{
			UserVolumeSizeGib: aws.Int32(int32(d.Get(key).(int))),
		}
	}

	_, err := conn.ModifyWorkspaceProperties(ctx, input)

	if err != nil {
		return fmt.Errorf("updating WorkSpaces Workspace (%s,%s): %w", d.Id(), key, err)
	}

	if _, err := waitWorkspaceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("waiting for WorkSpaces Workspace (%s,%s) update: %w", d.Id(), key, err)
	}

	return nil
}

func findWorkspaceByID(ctx context.Context, conn *workspaces.Client, id string) (*types.Workspace, error) {
	input := &workspaces.DescribeWorkspacesInput{
		WorkspaceIds: []string{id},
	}

	output, err := findWorkspace(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if itypes.IsZero(output) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findWorkspace(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspacesInput) (*types.Workspace, error) {
	output, err := findWorkspaces(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findWorkspaces(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspacesInput) ([]types.Workspace, error) { // nosemgrep:ci.caps0-in-func-name,ci.workspaces-in-func-name
	var output []types.Workspace

	pages := workspaces.NewDescribeWorkspacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Workspaces...)
	}

	return output, nil
}

func statusWorkspace(ctx context.Context, conn *workspaces.Client, workspaceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findWorkspaceByID(ctx, conn, workspaceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitWorkspaceAvailable(ctx context.Context, conn *workspaces.Client, workspaceID string, timeout time.Duration) (*types.Workspace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceStatePending, types.WorkspaceStateStarting),
		Target:  enum.Slice(types.WorkspaceStateAvailable),
		Refresh: statusWorkspace(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Workspace); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceUpdated(ctx context.Context, conn *workspaces.Client, workspaceID string, timeout time.Duration) (*types.Workspace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceStateUpdating),
		Target:  enum.Slice(types.WorkspaceStateAvailable, types.WorkspaceStateStopped),
		Refresh: statusWorkspace(ctx, conn, workspaceID),
		// "OperationInProgressException: The properties of this WorkSpace are currently under modification. Please try again in a moment".
		// AWS Workspaces service doesn't change instance status to "Updating" during property modification.
		// Respective AWS Support feature request has been created. Meanwhile, artificial delay is placed here as a workaround.
		Delay:   1 * time.Minute,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.Workspace); ok {
		return v, err
	}

	return nil, err
}

func waitWorkspaceTerminated(ctx context.Context, conn *workspaces.Client, workspaceID string, timeout time.Duration) (*types.Workspace, error) {
	// https://docs.aws.amazon.com/workspaces/latest/api/API_TerminateWorkspaces.html
	stateConf := &retry.StateChangeConf{
		// You can terminate a WorkSpace that is in any state except SUSPENDED.
		// After a WorkSpace is terminated, the TERMINATED state is returned only briefly before the WorkSpace directory metadata is cleaned up.
		Pending: enum.Slice(tfslices.RemoveAll(enum.EnumValues[types.WorkspaceState](), types.WorkspaceStateSuspended)...),
		Target:  []string{},
		Refresh: statusWorkspace(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Workspace); ok {
		return output, err
	}

	return nil, err
}

func expandWorkspaceProperties(tfList []any) *types.WorkspaceProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &types.WorkspaceProperties{
		ComputeTypeName:   types.Compute(tfMap["compute_type_name"].(string)),
		RootVolumeSizeGib: aws.Int32(int32(tfMap["root_volume_size_gib"].(int))),
		RunningMode:       types.RunningMode(tfMap["running_mode"].(string)),
		UserVolumeSizeGib: aws.Int32(int32(tfMap["user_volume_size_gib"].(int))),
	}

	if tfMap["running_mode"].(string) == string(types.RunningModeAutoStop) {
		apiObject.RunningModeAutoStopTimeoutInMinutes = aws.Int32(int32(tfMap["running_mode_auto_stop_timeout_in_minutes"].(int)))
	}

	return apiObject
}

func flattenWorkspaceProperties(apiObject *types.WorkspaceProperties) []map[string]any {
	if apiObject == nil {
		return []map[string]any{}
	}

	return []map[string]any{{
		"compute_type_name":                         apiObject.ComputeTypeName,
		"root_volume_size_gib":                      aws.ToInt32(apiObject.RootVolumeSizeGib),
		"running_mode":                              apiObject.RunningMode,
		"running_mode_auto_stop_timeout_in_minutes": aws.ToInt32(apiObject.RunningModeAutoStopTimeoutInMinutes),
		"user_volume_size_gib":                      aws.ToInt32(apiObject.UserVolumeSizeGib),
	}}
}
