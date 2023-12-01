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
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Data Lake")
// @Tags(identifierAttribute="arn")
func newDataLakeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &dataLakeResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDataLake = "Data Lake"
)

type dataLakeResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *dataLakeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_data_lake"
}

func (r *dataLakeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn":        framework.ARNAttributeComputedOnly(),
			names.AttrID: framework.IDAttribute(),
			"meta_store_manager_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"region": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"encryption_configuration": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
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
									"transition": schema.SetNestedBlock{
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
									"regions": schema.SetAttribute{
										ElementType: types.StringType,
										Optional:    true,
									},
									"role_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
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

func (r *dataLakeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataLakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	var configurations []dataLakeConfigurationModel

	resp.Diagnostics.Append(plan.Configurations.ElementsAs(ctx, &configurations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &securitylake.CreateDataLakeInput{
		Configurations:          expandDataLakeConfigurations(ctx, configurations),
		MetaStoreManagerRoleArn: aws.String(plan.MetaStoreManagerRoleARN.ValueString()),
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

	plan.DataLakeARN = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
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

func (r *dataLakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state dataLakeResourceModel

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

func (r *dataLakeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var plan, state dataLakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Configurations.Equal(state.Configurations) {
		var configurations []dataLakeConfigurationModel
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

func (r *dataLakeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data dataLakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteDataLake(ctx, &securitylake.DeleteDataLakeInput{
		Regions: []string{errs.Must(regionFromARNString(data.ID.ValueString()))},
	})

	// No data lake:
	// "An error occurred (AccessDeniedException) when calling the DeleteDataLake operation: User: ... is not authorized to perform: securitylake:DeleteDataLake"
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform") {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameDataLake, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	if _, err = waitDataLakeDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForDeletion, ResNameDataLake, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *dataLakeResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitDataLakeCreated(ctx context.Context, conn *securitylake.Client, arn string, timeout time.Duration) (*awstypes.DataLakeResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataLakeStatusInitialized),
		Target:  enum.Slice(awstypes.DataLakeStatusCompleted),
		Refresh: statusDataLakeCreate(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataLakeResource); ok {
		return output, err
	}

	return nil, err
}

func waitDataLakeUpdated(ctx context.Context, conn *securitylake.Client, arn string, timeout time.Duration) (*awstypes.DataLakeResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataLakeStatusPending, awstypes.DataLakeStatusInitialized),
		Target:  enum.Slice(awstypes.DataLakeStatusCompleted),
		Refresh: statusDataLakeUpdate(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataLakeResource); ok {
		if v := output.UpdateStatus; v != nil {
			if v := v.Exception; v != nil {
				tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.Code), aws.ToString(v.Reason)))
			}
		}

		return output, err
	}

	return nil, err
}

func waitDataLakeDeleted(ctx context.Context, conn *securitylake.Client, arn string, timeout time.Duration) (*awstypes.DataLakeResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataLakeStatusInitialized, awstypes.DataLakeStatusCompleted),
		Target:  []string{},
		Refresh: statusDataLakeUpdate(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataLakeResource); ok {
		if v := output.UpdateStatus; v != nil {
			if v := v.Exception; v != nil {
				tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(v.Code), aws.ToString(v.Reason)))
			}
		}

		return output, err
	}

	return nil, err
}

func statusDataLakeCreate(ctx context.Context, conn *securitylake.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDataLakeByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.CreateStatus), nil
	}
}

func statusDataLakeUpdate(ctx context.Context, conn *securitylake.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDataLakeByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.UpdateStatus == nil {
			return nil, "", nil
		}

		return output, string(output.UpdateStatus.Status), nil
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
		"expiration": expiration,
		"transition": transitions,
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

func expandDataLakeConfigurations(ctx context.Context, tfList []dataLakeConfigurationModel) []awstypes.DataLakeConfiguration {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []awstypes.DataLakeConfiguration
	var encryptionConfiguration []dataLakeEncryptionConfigurationModel
	var lifecycleConfiguration []dataLakeLifecycleConfigurationModel
	var replicationConfiguration []dataLakeReplicationConfigurationModel

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

func expandEncryptionConfiguration(tfList []dataLakeEncryptionConfigurationModel) *awstypes.DataLakeEncryptionConfiguration {
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

func expandLifecycleConfiguration(ctx context.Context, tfList []dataLakeLifecycleConfigurationModel) (*awstypes.DataLakeLifecycleConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]
	var transitions []dataLakeLifecycleTransitionModel
	diags.Append(tfObj.Transitions.ElementsAs(ctx, &transitions, false)...)
	var expiration []dataLakeLifecycleExpirationModel
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

func expandLifecycleExpiration(tfList []dataLakeLifecycleExpirationModel) *awstypes.DataLakeLifecycleExpiration {
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

func expandLifecycleTransitions(tfList []dataLakeLifecycleTransitionModel) []awstypes.DataLakeLifecycleTransition {
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

func expandReplicationConfiguration(ctx context.Context, tfList []dataLakeReplicationConfigurationModel) *awstypes.DataLakeReplicationConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.DataLakeReplicationConfiguration{}

	if !tfObj.RoleARN.IsNull() {
		apiObject.RoleArn = aws.String(tfObj.RoleARN.ValueString())
	}

	if !tfObj.Regions.IsNull() {
		apiObject.Regions = flex.ExpandFrameworkStringValueSet(ctx, tfObj.Regions)
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
		"expiration": types.ListType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleExpirationTypes}},
		"transition": types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTransitionsTypes}},
	}

	dataLakeConfigurationsReplicationConfigurationTypes = map[string]attr.Type{
		"role_arn": types.StringType,
		"regions":  types.ListType{ElemType: types.StringType},
	}
)

type dataLakeResourceModel struct {
	Configurations          types.Set      `tfsdk:"configuration"`
	DataLakeARN             types.String   `tfsdk:"arn"`
	ID                      types.String   `tfsdk:"id"`
	MetaStoreManagerRoleARN fwtypes.ARN    `tfsdk:"meta_store_manager_role_arn"`
	Tags                    types.Map      `tfsdk:"tags"`
	TagsAll                 types.Map      `tfsdk:"tags_all"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

func (model *dataLakeResourceModel) setID() {
	model.ID = model.DataLakeARN
}

func (model *dataLakeResourceModel) refreshFromOutput(ctx context.Context, out *awstypes.DataLakeResource) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	model.DataLakeARN = flex.StringToFramework(ctx, out.DataLakeArn)
	model.setID()
	configurations, d := flattenDataLakeConfigurations(ctx, []*awstypes.DataLakeResource{out})
	diags.Append(d...)

	model.Configurations = configurations

	return diags
}

type dataLakeConfigurationModel struct {
	EncryptionConfiguration  fwtypes.ListNestedObjectValueOf[dataLakeEncryptionConfigurationModel]  `tfsdk:"encryption_configuration"`
	LifecycleConfiguration   fwtypes.ListNestedObjectValueOf[dataLakeLifecycleConfigurationModel]   `tfsdk:"lifecycle_configuration"`
	Region                   types.String                                                           `tfsdk:"region"`
	ReplicationConfiguration fwtypes.ListNestedObjectValueOf[dataLakeReplicationConfigurationModel] `tfsdk:"replication_configuration"`
}

type dataLakeEncryptionConfigurationModel struct {
	KmsKeyID types.String `tfsdk:"kms_key_id"`
}

type dataLakeLifecycleConfigurationModel struct {
	Expiration  fwtypes.ListNestedObjectValueOf[dataLakeLifecycleExpirationModel] `tfsdk:"expiration"`
	Transitions fwtypes.SetNestedObjectValueOf[dataLakeLifecycleTransitionModel]  `tfsdk:"transition"`
}

type dataLakeLifecycleExpirationModel struct {
	Days types.Int64 `tfsdk:"days"`
}

type dataLakeLifecycleTransitionModel struct {
	Days         types.Int64  `tfsdk:"days"`
	StorageClass types.String `tfsdk:"storage_class"`
}

type dataLakeReplicationConfigurationModel struct {
	Regions types.Set   `tfsdk:"regions"`
	RoleARN fwtypes.ARN `tfsdk:"role_arn"`
}

func extractRegionFromARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid ARN: %s", arn)
	}
	return parts[3], nil
}

func findDataLakeByARN(ctx context.Context, conn *securitylake.Client, arn string) (*awstypes.DataLakeResource, error) {
	input := &securitylake.ListDataLakesInput{
		Regions: []string{errs.Must(regionFromARNString(arn))},
	}

	return findDataLake(ctx, conn, input, func(v *awstypes.DataLakeResource) bool {
		return aws.ToString(v.DataLakeArn) == arn
	})
}

func findDataLake(ctx context.Context, conn *securitylake.Client, input *securitylake.ListDataLakesInput, filter tfslices.Predicate[*awstypes.DataLakeResource]) (*awstypes.DataLakeResource, error) {
	output, err := findDataLakes(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDataLakes(ctx context.Context, conn *securitylake.Client, input *securitylake.ListDataLakesInput, filter tfslices.Predicate[*awstypes.DataLakeResource]) ([]*awstypes.DataLakeResource, error) {
	var dataLakes []*awstypes.DataLakeResource

	output, err := conn.ListDataLakes(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output.DataLakes {
		v := v
		if v := &v; filter(v) {
			dataLakes = append(dataLakes, v)
		}
	}

	return dataLakes, nil
}

// regionFromARNString return the AWS Region from the specified ARN string.
func regionFromARNString(s string) (string, error) {
	v, err := arn.Parse(s)

	if err != nil {
		return "", err
	}

	return v.Region, nil
}
