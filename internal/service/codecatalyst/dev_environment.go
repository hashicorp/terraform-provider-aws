// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecatalyst

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_codecatalyst_dev_environment", name="DevEnvironment")
func ResourceDevEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDevEnvironmentCreate,
		ReadWithoutTimeout:   resourceDevEnvironmentRead,
		UpdateWithoutTimeout: resourceDevEnvironmentUpdate,
		DeleteWithoutTimeout: resourceDevEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ides": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"runtime": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"inactivity_timeout_minutes": {
				Type:     schema.TypeInt,
				Default:  15,
				Optional: true,
			},
			names.AttrInstanceType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.InstanceType](),
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"persistent_storage": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSize: {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"repositories": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrRepositoryName: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	ResNameDevEnvironment = "DevEnvironment"
)

func resourceDevEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)
	storage := expandPersistentStorageConfiguration(d.Get("persistent_storage").([]interface{})[0].(map[string]interface{}))
	instanceType := types.InstanceType(d.Get(names.AttrInstanceType).(string))
	in := &codecatalyst.CreateDevEnvironmentInput{
		ProjectName:       aws.String(d.Get("project_name").(string)),
		SpaceName:         aws.String(d.Get("space_name").(string)),
		PersistentStorage: storage,
		InstanceType:      instanceType,
	}

	if v, ok := d.GetOk("inactivity_timeout_minutes"); ok {
		in.InactivityTimeoutMinutes = int32(v.(int))
	}

	if v, ok := d.GetOk(names.AttrAlias); ok {
		in.Alias = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ides"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Ides = expandIdesConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("repositories"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Repositories = expandRepositorysInput(v.([]interface{}))
	}

	out, err := conn.CreateDevEnvironment(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionCreating, ResNameDevEnvironment, d.Id(), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionCreating, ResNameDevEnvironment, d.Id(), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitDevEnvironmentCreated(ctx, conn, d.Id(), out.SpaceName, out.ProjectName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionWaitingForCreation, ResNameDevEnvironment, d.Id(), err)
	}

	return append(diags, resourceDevEnvironmentRead(ctx, d, meta)...)
}

func resourceDevEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	spaceName := aws.String(d.Get("space_name").(string))
	projectName := aws.String(d.Get("project_name").(string))

	out, err := findDevEnvironmentByID(ctx, conn, d.Id(), spaceName, projectName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Codecatalyst DevEnvironment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionReading, ResNameDevEnvironment, d.Id(), err)
	}

	d.Set(names.AttrAlias, out.Alias)
	d.Set("project_name", out.ProjectName)
	d.Set("space_name", out.SpaceName)
	d.Set(names.AttrInstanceType, out.InstanceType)
	d.Set("inactivity_timeout_minutes", out.InactivityTimeoutMinutes)
	d.Set("persistent_storage", flattenPersistentStorage(out.PersistentStorage))

	if err := d.Set("ides", flattenIdes(out.Ides)); err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionSetting, ResNameDevEnvironment, d.Id(), err)
	}

	if err := d.Set("repositories", flattenRepositories(out.Repositories)); err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionSetting, ResNameDevEnvironment, d.Id(), err)
	}

	return diags
}

func resourceDevEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	update := false

	in := &codecatalyst.UpdateDevEnvironmentInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges(names.AttrAlias) {
		in.Alias = aws.String(d.Get(names.AttrAlias).(string))
		update = true
	}

	if d.HasChanges(names.AttrInstanceType) {
		in.InstanceType = types.InstanceType(d.Get(names.AttrInstanceType).(string))
		update = true
	}
	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating Codecatalyst DevEnvironment (%s): %#v", d.Id(), in)
	out, err := conn.UpdateDevEnvironment(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionUpdating, ResNameDevEnvironment, d.Id(), err)
	}

	if _, err := waitDevEnvironmentUpdated(ctx, conn, aws.ToString(out.Id), out.SpaceName, out.ProjectName, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionWaitingForUpdate, ResNameDevEnvironment, d.Id(), err)
	}

	return append(diags, resourceDevEnvironmentRead(ctx, d, meta)...)
}

func resourceDevEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	log.Printf("[INFO] Deleting Codecatalyst DevEnvironment %s", d.Id())

	_, err := conn.DeleteDevEnvironment(ctx, &codecatalyst.DeleteDevEnvironmentInput{
		Id:          aws.String(d.Id()),
		SpaceName:   aws.String(d.Get("space_name").(string)),
		ProjectName: aws.String(d.Get("project_name").(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionDeleting, ResNameDevEnvironment, d.Id(), err)
	}

	return diags
}

func waitDevEnvironmentCreated(ctx context.Context, conn *codecatalyst.Client, id string, spaceName *string, projectName *string, timeout time.Duration) (*codecatalyst.GetDevEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.DevEnvironmentStatusPending, types.DevEnvironmentStatusStarting),
		Target:                    enum.Slice(types.DevEnvironmentStatusRunning, types.DevEnvironmentStatusStopped, types.DevEnvironmentStatusStopping),
		Refresh:                   statusDevEnvironment(ctx, conn, id, spaceName, projectName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.GetDevEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDevEnvironmentUpdated(ctx context.Context, conn *codecatalyst.Client, id string, spaceName *string, projectName *string, timeout time.Duration) (*codecatalyst.GetDevEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.DevEnvironmentStatusStopping, types.DevEnvironmentStatusPending, types.DevEnvironmentStatusStopped),
		Target:                    enum.Slice(types.DevEnvironmentStatusRunning),
		Refresh:                   statusDevEnvironment(ctx, conn, id, spaceName, projectName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.GetDevEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDevEnvironment(ctx context.Context, conn *codecatalyst.Client, id string, spaceName *string, projectName *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDevEnvironmentByID(ctx, conn, id, spaceName, projectName)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findDevEnvironmentByID(ctx context.Context, conn *codecatalyst.Client, id string, spaceName *string, projectName *string) (*codecatalyst.GetDevEnvironmentOutput, error) {
	in := &codecatalyst.GetDevEnvironmentInput{
		Id:          aws.String(id),
		ProjectName: projectName,
		SpaceName:   spaceName,
	}

	out, err := conn.GetDevEnvironment(ctx, in)
	if errs.IsA[*types.AccessDeniedException](err) || errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenRepositories(apiObjects []types.DevEnvironmentRepositorySummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenRepository(&apiObject))
	}

	return tfList
}

func flattenRepository(apiObject *types.DevEnvironmentRepositorySummary) interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BranchName; v != nil {
		tfMap["branch_name"] = aws.ToString(v)
	}

	if v := apiObject.RepositoryName; v != nil {
		tfMap[names.AttrRepositoryName] = aws.ToString(v)
	}

	return tfMap
}

func flattenIdes(apiObjects []types.Ide) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenIde(&apiObject))
	}

	return tfList
}

func flattenIde(apiObject *types.Ide) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Runtime; v != nil {
		tfMap["runtime"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPersistentStorage(apiObject *types.PersistentStorage) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrSize: aws.ToInt32(apiObject.SizeInGiB),
	}

	return []map[string]interface{}{tfMap}
}

func expandRepositorysInput(tfList []interface{}) []types.RepositoryInput {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.RepositoryInput

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandRepositoryInput(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRepositoryInput(tfMap map[string]interface{}) types.RepositoryInput {
	apiObject := types.RepositoryInput{}

	if v, ok := tfMap["branch_name"].(string); ok && v != "" {
		apiObject.BranchName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrRepositoryName].(string); ok && v != "" {
		apiObject.RepositoryName = aws.String(v)
	}

	return apiObject
}

func expandIdesConfiguration(tfList []interface{}) []types.IdeConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.IdeConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandIdeConfiguration(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIdeConfiguration(tfMap map[string]interface{}) types.IdeConfiguration {
	apiObject := types.IdeConfiguration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["runtime"].(string); ok && v != "" {
		apiObject.Runtime = aws.String(v)
	}

	return apiObject
}

func expandPersistentStorageConfiguration(tfMap map[string]interface{}) *types.PersistentStorageConfiguration {
	apiObject := &types.PersistentStorageConfiguration{}

	if v, ok := tfMap[names.AttrSize].(int); ok && v != 0 {
		apiObject.SizeInGiB = aws.Int32(int32(v))
	}

	return apiObject
}
