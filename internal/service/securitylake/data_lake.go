// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Data Lake")
func newResourceDataLake(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDataLake{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDataLake = "Data Lake"
)

type resourceDataLake struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDataLake) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_data_lake"
}

func (r *resourceDataLake) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"datalake_arn": framework.ARNAttributeComputedOnly(),
			"meta_store_manager_role_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags: tftags.TagsAttribute(),
			names.AttrID:   framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"configurations": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"region": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"encryption_configuration": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kms_key_id": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Default:  stringdefault.StaticString("S3_MANAGED_KEY"),
									},
								},
							},
						},
						"lifecycle_configuration": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"expiration": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"days": schema.Int64Attribute{
													Optional: true,
												},
											},
										},
									},
									"transitions": schema.SetNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"days": schema.Int64Attribute{
													Optional: true,
												},
												"storage_class": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"replication_configuration": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"regions": schema.ListAttribute{
										ElementType: types.StringType,
										Optional:    true,
									},
									"role_arn": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDataLake) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datalakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	var configurations []dataLakeConfigurationsData

	resp.Diagnostics.Append(plan.Configurations.ElementsAs(ctx, &configurations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &securitylake.CreateDataLakeInput{
		Configurations:          expandDataLakeConfigurations(ctx, configurations),
		MetaStoreManagerRoleArn: aws.String(plan.MetaStoreManagerRoleArn.ValueString()),
		Tags:                    getTagsIn(ctx),
	}

	out, err := conn.CreateDataLake(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameDataLake, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.DataLakeArn = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
	plan.setID()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitDataLakeCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForCreation, ResNameDataLake, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(plan.refreshFromOutput(ctx, waitOut)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataLake) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state datalakeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindDataLakeByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameDataLake, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataLake) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var plan, state datalakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Configurations.Equal(state.Configurations) {
		var configurations []dataLakeConfigurationsData
		resp.Diagnostics.Append(plan.Configurations.ElementsAs(ctx, &configurations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &securitylake.UpdateDataLakeInput{
			Configurations: expandDataLakeConfigurations(ctx, configurations),
		}

		out, err := conn.UpdateDataLake(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, plan.ID.ValueString(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.DataLakes == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, plan.ID.ValueString(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(state.refreshFromOutput(ctx, &out.DataLakes[0])...)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitDataLakeUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForUpdate, ResNameDataLake, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataLake) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state datalakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, _ := extractRegionFromARN(state.ID.ValueString())

	in := &securitylake.DeleteDataLakeInput{
		Regions: []string{region},
	}

	_, err := conn.DeleteDataLake(ctx, in)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameDataLake, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDataLakeDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForDeletion, ResNameDataLake, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDataLake) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitDataLakeCreated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*awstypes.DataLakeResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DataLakeStatusInitialized),
		Target:                    enum.Slice(awstypes.DataLakeStatusCompleted),
		Refresh:                   createStatusDataLake(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataLakeResource); ok {
		return out, err
	}

	return nil, err
}

func waitDataLakeUpdated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.ListDataLakesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DataLakeStatusPending, awstypes.DataLakeStatusInitialized),
		Target:                    enum.Slice(awstypes.DataLakeStatusCompleted),
		Refresh:                   updateStatusDataLake(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*securitylake.ListDataLakesOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDataLakeDeleted(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.ListDataLakesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataLakeStatusInitialized, awstypes.DataLakeStatusCompleted),
		Target:  []string{},
		Refresh: createStatusDataLake(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*securitylake.ListDataLakesOutput); ok {
		return out, err
	}

	return nil, err
}

func createStatusDataLake(ctx context.Context, conn *securitylake.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDataLakeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return out, string(out.CreateStatus), nil
	}
}

func updateStatusDataLake(ctx context.Context, conn *securitylake.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDataLakeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.UpdateStatus.Status), nil
	}
}

func FindDataLakeByID(ctx context.Context, conn *securitylake.Client, id string) (*awstypes.DataLakeResource, error) {
	region, err := extractRegionFromARN(id)
	if err != nil {
		return nil, err
	}

	in := &securitylake.ListDataLakesInput{
		Regions: []string{region},
	}

	out, err := conn.ListDataLakes(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || len(out.DataLakes) < 1 {
		return nil, tfresource.NewEmptyResultError(in)
	}
	datalakeResource := out.DataLakes[0]

	return &datalakeResource, nil
}

func flattenDataLakeConfigurations(ctx context.Context, apiObjects []*awstypes.DataLakeResource) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurations}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		encryptionConfiguration, d := flattenEncryptionConfiguration(ctx, apiObject.EncryptionConfiguration)
		diags.Append(d...)
		lifecycleExpiration, d := flattenLifeCycleConfiguration(ctx, apiObject.LifecycleConfiguration)
		diags.Append(d...)
		replicationConfiguration, d := flattenReplicationConfiguration(ctx, apiObject.ReplicationConfiguration)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"encryption_configuration":  encryptionConfiguration,
			"lifecycle_configuration":   lifecycleExpiration,
			"region":                    flex.StringToFramework(ctx, apiObject.Region),
			"replication_configuration": replicationConfiguration,
		}
		objVal, d := types.ObjectValue(dataLakeConfigurations, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenEncryptionConfiguration(ctx context.Context, apiObject *awstypes.DataLakeEncryptionConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsEncryptionTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"kms_key_id": flex.StringToFramework(ctx, apiObject.KmsKeyId),
	}
	objVal, d := types.ObjectValue(dataLakeConfigurationsEncryptionTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenLifeCycleConfiguration(ctx context.Context, apiObject *awstypes.DataLakeLifecycleConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTypes}

	if apiObject == nil || (apiObject.Expiration == nil && len(apiObject.Transitions) == 0) {
		return types.ListNull(elemType), diags
	}

	expiration, d := flattenLifecycleExpiration(ctx, apiObject.Expiration)
	diags.Append(d...)
	transitions, d := flattenLifecycleTransitions(ctx, apiObject.Transitions)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"expiration":  expiration,
		"transitions": transitions,
	}
	objVal, d := types.ObjectValue(dataLakeConfigurationsLifecycleTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenLifecycleExpiration(ctx context.Context, apiObject *awstypes.DataLakeLifecycleExpiration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleExpirationTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"days": flex.Int32ToFramework(ctx, apiObject.Days),
	}

	objVal, d := types.ObjectValue(dataLakeConfigurationsLifecycleExpirationTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenLifecycleTransitions(ctx context.Context, apiObjects []awstypes.DataLakeLifecycleTransition) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTransitionsTypes}

	if len(apiObjects) == 0 {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"days":          flex.Int32ToFramework(ctx, apiObject.Days),
			"storage_class": flex.StringToFramework(ctx, apiObject.StorageClass),
		}
		objVal, d := types.ObjectValue(dataLakeConfigurationsLifecycleTransitionsTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenReplicationConfiguration(ctx context.Context, apiObject *awstypes.DataLakeReplicationConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsReplicationConfigurationTypes}

	if apiObject == nil || (apiObject.Regions == nil && apiObject.RoleArn == nil) {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"role_arn": flex.StringToFramework(ctx, apiObject.RoleArn),
		"regions":  flex.FlattenFrameworkStringValueList(ctx, apiObject.Regions),
	}
	objVal, d := types.ObjectValue(dataLakeConfigurationsReplicationConfigurationTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandDataLakeConfigurations(ctx context.Context, tfList []dataLakeConfigurationsData) []awstypes.DataLakeConfiguration {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []awstypes.DataLakeConfiguration
	var encryptionConfiguration []dataLakeConfigurationsEncryption
	var lifecycleConfiguration []dataLakeConfigurationsLifecycle
	var replicationConfiguration []dataLakeConfigurationsReplicationConfiguration

	for _, tfObj := range tfList {
		diags.Append(tfObj.LifecycleConfiguration.ElementsAs(ctx, &lifecycleConfiguration, false)...)
		diags.Append(tfObj.ReplicationConfiguration.ElementsAs(ctx, &replicationConfiguration, false)...)
		lifecycleConfiguration, d := expandLifecycleConfiguration(ctx, lifecycleConfiguration)
		diags.Append(d...)

		item := awstypes.DataLakeConfiguration{
			Region: aws.String(tfObj.Region.ValueString()),
		}

		if !tfObj.EncryptionConfiguration.IsNull() {
			item.EncryptionConfiguration = expandEncryptionConfiguration(encryptionConfiguration)
		}

		if !tfObj.LifecycleConfiguration.IsNull() {
			item.LifecycleConfiguration = lifecycleConfiguration
		}

		if !tfObj.ReplicationConfiguration.IsNull() {
			item.ReplicationConfiguration = expandReplicationConfiguration(ctx, replicationConfiguration)
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}

func expandEncryptionConfiguration(tfList []dataLakeConfigurationsEncryption) *awstypes.DataLakeEncryptionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.DataLakeEncryptionConfiguration{}
	if !tfObj.KmsKeyID.IsNull() {
		apiObject.KmsKeyId = aws.String(tfObj.KmsKeyID.ValueString())
	}

	return apiObject
}

func expandLifecycleConfiguration(ctx context.Context, tfList []dataLakeConfigurationsLifecycle) (*awstypes.DataLakeLifecycleConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]
	var transitions []dataLakeConfigurationsLifecycleTransitions
	diags.Append(tfObj.Transitions.ElementsAs(ctx, &transitions, false)...)
	var expiration []dataLakeConfigurationsLifecycleExpiration
	diags.Append(tfObj.Expiration.ElementsAs(ctx, &expiration, false)...)
	apiObject := &awstypes.DataLakeLifecycleConfiguration{}

	if !tfObj.Expiration.IsNull() {
		apiObject.Expiration = expandLifecycleExpiration(expiration)
	}

	if !tfObj.Transitions.IsNull() {
		apiObject.Transitions = expandLifecycleTransitions(transitions)
	}

	return apiObject, diags
}

func expandLifecycleExpiration(tfList []dataLakeConfigurationsLifecycleExpiration) *awstypes.DataLakeLifecycleExpiration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.DataLakeLifecycleExpiration{}

	if !tfObj.Days.IsNull() {
		apiObject.Days = aws.Int32(int32(tfObj.Days.ValueInt64()))
	}

	return apiObject
}

func expandLifecycleTransitions(tfList []dataLakeConfigurationsLifecycleTransitions) []awstypes.DataLakeLifecycleTransition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []awstypes.DataLakeLifecycleTransition

	for _, tfObj := range tfList {
		item := awstypes.DataLakeLifecycleTransition{}

		if !tfObj.Days.IsNull() {
			item.Days = aws.Int32(int32(tfObj.Days.ValueInt64()))
		}

		if !tfObj.StorageClass.IsNull() {
			item.StorageClass = aws.String(tfObj.StorageClass.ValueString())
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}

func expandReplicationConfiguration(ctx context.Context, tfList []dataLakeConfigurationsReplicationConfiguration) *awstypes.DataLakeReplicationConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.DataLakeReplicationConfiguration{}

	if !tfObj.RoleArn.IsNull() {
		apiObject.RoleArn = aws.String(tfObj.RoleArn.ValueString())
	}

	if !tfObj.Regions.IsNull() {
		apiObject.Regions = flex.ExpandFrameworkStringValueList(ctx, tfObj.Regions)
	}

	return apiObject
}

var (
	dataLakeConfigurations = map[string]attr.Type{
		"encryption_configuration":  types.ListType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsEncryptionTypes}},
		"lifecycle_configuration":   types.ListType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTypes}},
		"region":                    types.StringType,
		"replication_configuration": types.ListType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsReplicationConfigurationTypes}},
	}

	dataLakeConfigurationsEncryptionTypes = map[string]attr.Type{
		"kms_key_id": types.StringType,
	}

	dataLakeConfigurationsLifecycleExpirationTypes = map[string]attr.Type{
		"days": types.Int64Type,
	}

	dataLakeConfigurationsLifecycleTransitionsTypes = map[string]attr.Type{
		"days":          types.Int64Type,
		"storage_class": types.StringType,
	}

	dataLakeConfigurationsLifecycleTypes = map[string]attr.Type{
		"expiration":  types.ListType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleExpirationTypes}},
		"transitions": types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTransitionsTypes}},
	}

	dataLakeConfigurationsReplicationConfigurationTypes = map[string]attr.Type{
		"role_arn": types.StringType,
		"regions":  types.ListType{ElemType: types.StringType},
	}
)

type datalakeResourceModel struct {
	DataLakeArn             types.String   `tfsdk:"datalake_arn"`
	ID                      types.String   `tfsdk:"id"`
	MetaStoreManagerRoleArn types.String   `tfsdk:"meta_store_manager_role_arn"`
	Configurations          types.Set      `tfsdk:"configurations"`
	Tags                    types.Map      `tfsdk:"tags"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

func (model *datalakeResourceModel) setID() {
	model.ID = model.DataLakeArn
}

func (model *datalakeResourceModel) refreshFromOutput(ctx context.Context, out *awstypes.DataLakeResource) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	model.DataLakeArn = flex.StringToFramework(ctx, out.DataLakeArn)
	model.setID()
	configurations, d := flattenDataLakeConfigurations(ctx, []*awstypes.DataLakeResource{out})
	diags.Append(d...)

	model.Configurations = configurations

	return diags
}

type dataLakeConfigurationsData struct {
	EncryptionConfiguration  types.List   `tfsdk:"encryption_configuration"`
	LifecycleConfiguration   types.List   `tfsdk:"lifecycle_configuration"`
	Region                   types.String `tfsdk:"region"`
	ReplicationConfiguration types.List   `tfsdk:"replication_configuration"`
}

type dataLakeConfigurationsEncryption struct {
	KmsKeyID types.String `tfsdk:"kms_key_id"`
}

type dataLakeConfigurationsLifecycle struct {
	Expiration  types.List `tfsdk:"expiration"`
	Transitions types.Set  `tfsdk:"transitions"`
}

type dataLakeConfigurationsLifecycleExpiration struct {
	Days types.Int64 `tfsdk:"days"`
}

type dataLakeConfigurationsLifecycleTransitions struct {
	Days         types.Int64  `tfsdk:"days"`
	StorageClass types.String `tfsdk:"storage_class"`
}

type dataLakeConfigurationsReplicationConfiguration struct {
	RoleArn types.String `tfsdk:"role_arn"`
	Regions types.List   `tfsdk:"regions"`
}

func extractRegionFromARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid ARN: %s", arn)
	}
	return parts[3], nil
}
