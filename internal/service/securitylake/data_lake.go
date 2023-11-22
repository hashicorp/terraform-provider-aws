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
	// "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	// "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Data Lake")
// @Tags(identifierAttribute="arn")
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
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"meta_store_manager_role_arn": schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"configurations": schema.ListNestedBlock{
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
						"encryption_configuration": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kms_key_id": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"lifecycle_configuration": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"expiration": schema.SetNestedBlock{
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
						"replication_configuration": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"role_arn": schema.StringAttribute{
										Optional: true,
									},
									"regions": schema.StringAttribute{
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
	conn := r.Meta().SecurityLakeClient(ctx)
	var plan resourceDataLakeData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configurations []dataLakeConfigurationsData
	resp.Diagnostics.Append(plan.Configurations.ElementsAs(ctx, &configurations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &securitylake.CreateDataLakeInput{
		Configurations:          expanddataLakeConfigurations(ctx, configurations),
		MetaStoreManagerRoleArn: aws.String(plan.MetaStoreManagerRoleArn.ValueString()),
		// Tags:                    getTagsIn(ctx),
	}

	out, err := conn.CreateDataLake(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameDataLake, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DataLakes == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameDataLake, plan.ID.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.ARN = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
	plan.ID = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
	state := plan

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitDataLakeCreated(ctx, conn, state.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForCreation, ResNameDataLake, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.Configurations, _ = flattenDataLakeConfigurations(ctx, out.DataLakes)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDataLake) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state resourceDataLakeData
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

	state.ARN = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
	state.ID = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
	state.Configurations, _ = flattenDataLakeConfigurations(ctx, out.DataLakes)

	fmt.Println(state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataLake) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// conn := r.Meta().SecurityLakeClient(ctx)

	// // TIP: -- 2. Fetch the plan
	// var plan, state resourceDataLakeData
	// resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // TIP: -- 3. Populate a modify input structure and check for changes
	// if !plan.Name.Equal(state.Name) ||
	// 	!plan.Description.Equal(state.Description) ||
	// 	!plan.ComplexArgument.Equal(state.ComplexArgument) ||
	// 	!plan.Type.Equal(state.Type) {

	// 	in := &securitylake.UpdateDataLakeInput{
	// 		// TIP: Mandatory or fields that will always be present can be set when
	// 		// you create the Input structure. (Replace these with real fields.)
	// 		DataLakeId:   aws.String(plan.ID.ValueString()),
	// 		DataLakeName: aws.String(plan.Name.ValueString()),
	// 		DataLakeType: aws.String(plan.Type.ValueString()),
	// 	}

	// 	if !plan.Description.IsNull() {
	// 		// TIP: Optional fields should be set based on whether or not they are
	// 		// used.
	// 		in.Description = aws.String(plan.Description.ValueString())
	// 	}
	// 	if !plan.ComplexArgument.IsNull() {
	// 		// TIP: Use an expander to assign a complex argument. The elements must be
	// 		// deserialized into the appropriate struct before being passed to the expander.
	// 		var tfList []complexArgumentData
	// 		resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
	// 		if resp.Diagnostics.HasError() {
	// 			return
	// 		}

	// 		in.ComplexArgument = expandComplexArgument(tfList)
	// 	}

	// 	// TIP: -- 4. Call the AWS modify/update function
	// 	out, err := conn.UpdateDataLake(ctx, in)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, plan.ID.ValueString(), err),
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	if out == nil || out.DataLake == nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, plan.ID.ValueString(), nil),
	// 			errors.New("empty output").Error(),
	// 		)
	// 		return
	// 	}

	// 	// TIP: Using the output from the update function, re-set any computed attributes
	// 	plan.ARN = flex.StringToFramework(ctx, out.DataLake.Arn)
	// 	plan.ID = flex.StringToFramework(ctx, out.DataLake.DataLakeId)
	// }

	// // TIP: -- 5. Use a waiter to wait for update to complete
	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitDataLakeUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForUpdate, ResNameDataLake, plan.ID.ValueString(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// // TIP: -- 6. Save the request plan to response state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataLake) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state resourceDataLakeData
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

func waitDataLakeCreated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.ListDataLakesOutput, error) {
	fmt.Println(id)
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.DataLakeStatusInitialized)},
		Target:                    []string{string(awstypes.DataLakeStatusCompleted)},
		Refresh:                   createStatusDataLake(ctx, conn, id),
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

func waitDataLakeUpdated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.ListDataLakesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.DataLakeStatusPending)},
		Target:                    []string{string(awstypes.DataLakeStatusCompleted)},
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
	fmt.Println(id)
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.DataLakeStatusInitialized), string(awstypes.DataLakeStatusCompleted)},
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
		fmt.Println(id)
		out, err := FindDataLakeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.DataLakes[0].CreateStatus), nil
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
		return out, string(out.DataLakes[0].UpdateStatus.Status), nil
	}
}

func FindDataLakeByID(ctx context.Context, conn *securitylake.Client, id string) (*securitylake.ListDataLakesOutput, error) {
	region, err := extractRegionFromARN(id)
	if err != nil {
		return nil, err
	}
	fmt.Printf("The region is %s\n", region)
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

	if out == nil || out.DataLakes == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenDataLakeConfigurations(ctx context.Context, apiObjects []awstypes.DataLakeResource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurations}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
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

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenLifeCycleConfiguration(ctx context.Context, apiObject *awstypes.DataLakeLifecycleConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTypes}

	if apiObject == nil {
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

func flattenLifecycleTransitions(ctx context.Context, apiObjects []awstypes.DataLakeLifecycleTransition) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTransitionsTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
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

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenReplicationConfiguration(ctx context.Context, apiObject *awstypes.DataLakeReplicationConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataLakeConfigurationsReplicationConfigurationTypes}

	if apiObject == nil {
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

func expanddataLakeConfigurations(ctx context.Context, tfList []dataLakeConfigurationsData) []awstypes.DataLakeConfiguration {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil
	}

	var apiObject []awstypes.DataLakeConfiguration
	var encryptionConfiguration []dataLakeConfigurationsEncryption
	var lifecycleConfiguration []dataLakeConfigurationsLifecycle
	var replicationConfiguration []dataLakeConfigurationsReplicationConfiguration

	for _, tfObj := range tfList {
		diags.Append(tfObj.EncryptionConfiguration.ElementsAs(ctx, &encryptionConfiguration, false)...)
		diags.Append(tfObj.LifecycleConfiguration.ElementsAs(ctx, &lifecycleConfiguration, false)...)
		diags.Append(tfObj.ReplicationConfiguration.ElementsAs(ctx, &replicationConfiguration, false)...)

		item := awstypes.DataLakeConfiguration{
			Region: aws.String(tfObj.Region.ValueString()),
		}

		if !tfObj.EncryptionConfiguration.IsNull() {
			item.EncryptionConfiguration = expandEncryptionConfiguration(encryptionConfiguration)
		}

		if !tfObj.LifecycleConfiguration.IsNull() {
			item.LifecycleConfiguration, _ = expandLifecycleConfiguration(ctx, lifecycleConfiguration)
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
		int32Days := int32(tfObj.Days.ValueInt64())
		apiObject.Days = aws.Int32(int32Days)
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
			int32Days := int32(tfObj.Days.ValueInt64())
			item.Days = aws.Int32(int32Days)
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
		"encryption_configuration":  types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsEncryptionTypes}},
		"lifecycle_configuration":   types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTypes}},
		"region":                    types.StringType,
		"replication_configuration": types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsReplicationConfigurationTypes}},
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
		"expiration":  types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleExpirationTypes}},
		"transitions": types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTransitionsTypes}},
	}

	dataLakeConfigurationsReplicationConfigurationTypes = map[string]attr.Type{
		"role_arn": types.StringType,
		"regions":  types.ListType{ElemType: types.StringType},
	}
)

type resourceDataLakeData struct {
	ARN                     types.String   `tfsdk:"arn"`
	ID                      types.String   `tfsdk:"id"`
	MetaStoreManagerRoleArn types.String   `tfsdk:"meta_store_manager_role_arn"`
	Configurations          types.List     `tfsdk:"configurations"`
	Tags                    types.Map      `tfsdk:"tags"`
	TagsAll                 types.Map      `tfsdk:"tags_all"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

type dataLakeConfigurationsData struct {
	EncryptionConfiguration  types.Set    `tfsdk:"encryption_configuration"`
	LifecycleConfiguration   types.Set    `tfsdk:"lifecycle_configuration"`
	Region                   types.String `tfsdk:"region"`
	ReplicationConfiguration types.Set    `tfsdk:"replication_configuration"`
}

type dataLakeConfigurationsEncryption struct {
	KmsKeyID types.String `tfsdk:"kms_key_id"`
}

type dataLakeConfigurationsLifecycle struct {
	Expiration  types.Set `tfsdk:"expiration"`
	Transitions types.Set `tfsdk:"transitions"`
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
		return "", fmt.Errorf("invalid ARN format")
	}
	return parts[3], nil
}

// refreshFromOutput writes state data from an AWS response object
// func (rd *resourceDataLakeData) refreshFromOutput(ctx context.Context, out *awstypes.DataLakeResource) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	if out == nil  {
// 		return diags
// 	}

// 	rd.ARN = flex.StringToFramework(ctx, out.DataLakeArn)
// 	rd.Configurations, d = flattenDataLakeConfigurations(ctx, out)
// 	if out.Framework != nil {
// 		rd.FrameworkID = flex.StringToFramework(ctx, out.Framework.Id)
// 	}
// 	rd.ID = flex.StringToFramework(ctx, metadata.Id)
// 	rd.Name = flex.StringToFramework(ctx, metadata.Name)
// 	rd.Status = flex.StringValueToFramework(ctx, metadata.Status)

// 	reportsDestination, d := flattenAssessmentReportsDestination(ctx, metadata.AssessmentReportsDestination)
// 	diags.Append(d...)
// 	rd.AssessmentReportsDestination = reportsDestination
// 	roles, d := flattenAssessmentRoles(ctx, metadata.Roles)
// 	diags.Append(d...)
// 	rd.RolesAll = roles
// 	scope, d := flattenAssessmentScope(ctx, metadata.Scope)
// 	diags.Append(d...)
// 	rd.Scope = scope

// 	setTagsOut(ctx, out.Tags)

// 	return diags
// }
