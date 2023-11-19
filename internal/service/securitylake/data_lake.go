package securitylake

import (
	"context"
	"errors"
	"time"
	"strings"

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
			"metastore_manager_role_arn": schema.StringAttribute{
				Required: true,
			},
			"id": framework.IDAttribute(),
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
						"encryption_configuration": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kms_key_id": schema.StringAttribute{
										Required: true,
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
													Required: true,
												},
											},
										},
									},
									"transitions": schema.SetNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"days": schema.Int64Attribute{
													Required: true,
												},
												"storage_class": schema.StringAttribute{
													Required: true,
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
										Required: true,
									},
									"regions": schema.StringAttribute{
										Required: true,
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
		Configurations: expanddataLakeConfigurations(ctx,configurations),
		MetaStoreManagerRoleArn: aws.String(plan.MetastoreManagerRoleArn.ValueString()),
	}
	
	out, err := conn.CreateDataLake(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameDataLake, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DataLakes == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameDataLake, plan.ARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	
	plan.ARN = flex.StringToFramework(ctx, out.DataLakes[0].DataLakeArn)
	
	id := generateDataLakeID(plan.ARN.String(), *out.DataLakes[0].Region)
	
	plan.ID = types.StringValue(id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDataLakeCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForCreation, ResNameDataLake, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}
	
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDataLake) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)
	
	var state resourceDataLakeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	out, err := findDataLakeByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
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

	
	state.ARN = flex.StringToFramework(ctx, out.DataLakeArn)
	state.ID = flex.StringToFramework(ctx, out.DataLakeArn)
	

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
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, plan.ID.String(), err),
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	if out == nil || out.DataLake == nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, plan.ID.String(), nil),
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
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForUpdate, ResNameDataLake, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	
	// // TIP: -- 6. Save the request plan to response state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataLake) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// conn := r.Meta().SecurityLakeClient(ctx)
	
	// // TIP: -- 2. Fetch the state
	// var state resourceDataLakeData
	// resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
	
	// // TIP: -- 3. Populate a delete input structure
	// in := &securitylake.DeleteDataLakeInput{
	// 	DataLakeId: aws.String(state.ID.ValueString()),
	// }
	
	// // TIP: -- 4. Call the AWS delete function
	// _, err := conn.DeleteDataLake(ctx, in)
	// // TIP: On rare occassions, the API returns a not found error after deleting a
	// // resource. If that happens, we don't want it to show up as an error.
	// if err != nil {
	// 	var nfe *awstypes.ResourceNotFoundException
	// 	if errors.As(err, &nfe) {
	// 		return
	// 	}
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameDataLake, state.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
	
	// // TIP: -- 5. Use a waiter to wait for delete to complete
	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// _, err = waitDataLakeDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForDeletion, ResNameDataLake, state.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
}

func (r *resourceDataLake) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitDataLakeCreated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.ListDataLakesOutput, error) {
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

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
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
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.DataLakeStatusPending), string(awstypes.DataLakeStatusCompleted)},
		Target:                    []string{},
		Refresh:                   updateStatusDataLake(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*securitylake.ListDataLakesOutput); ok {
		return out, err
	}

	return nil, err
}


func createStatusDataLake(ctx context.Context, conn *securitylake.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDataLakeByID(ctx, conn, id)
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
		out, err := findDataLakeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.UpdateStatus.Status), nil
	}
}

func findDataLakeByID(ctx context.Context, conn *securitylake.Client, id string) (*awstypes.DataLakeResource, error) {
	region := extractRegionFromID(id)

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

	if out == nil  {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.DataLakes[0], nil
}


// func flattenComplexArgument(ctx context.Context, apiObject *securitylake.ComplexArgument) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

// 	if apiObject == nil {
// 		return types.ListNull(elemType), diags
// 	}

// 	obj := map[string]attr.Value{
// 		"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
// 		"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
// 	}
// 	objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
// 	diags.Append(d...)

// 	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
// 	diags.Append(d...)

// 	return listVal, diags
// }

// func flattenComplexArguments(ctx context.Context, apiObjects []*securitylake.ComplexArgument) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

// 	if len(apiObjects) == 0 {
// 		return types.ListNull(elemType), diags
// 	}

// 	elems := []attr.Value{}
// 	for _, apiObject := range apiObjects {
// 		if apiObject == nil {
// 			continue
// 		}

// 		obj := map[string]attr.Value{
// 			"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
// 			"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
// 		}
// 		objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
// 		diags.Append(d...)

// 		elems = append(elems, objVal)
// 	}

// 	listVal, d := types.ListValue(elemType, elems)
// 	diags.Append(d...)

// 	return listVal, diags
// }


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
			item.LifecycleConfiguration,_ = expandLifecycleConfiguration(ctx, lifecycleConfiguration)
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


func expandLifecycleConfiguration(ctx context.Context, tfList []dataLakeConfigurationsLifecycle) (*awstypes.DataLakeLifecycleConfiguration,diag.Diagnostics) {
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
	dataLakeConfigurationsEncryptionTypes = map[string]attr.Type{
		"kms_key_id":      types.StringType,
	}

	dataLakeConfigurationsLifecycleExpirationTypes = map[string]attr.Type{
		"days":  types.Int64Type,
	}

	dataLakeConfigurationsLifecycleTransitionsTypes = map[string]attr.Type{
		"days":  types.Int64Type,
		"storage_class": types.StringType,
	}

	dataLakeConfigurationsLifecycleTypes = map[string]attr.Type{
		"expiration": types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleExpirationTypes}},
		"transitions": types.SetType{ElemType: types.ObjectType{AttrTypes: dataLakeConfigurationsLifecycleTransitionsTypes}},
	}

	dataLakeConfigurationsReplicationConfigurationTypes = map[string]attr.Type{
		"role_arn":  types.StringType,
		"regions": types.ListType{ElemType: types.StringType},
	}
)

type resourceDataLakeData struct {
	ARN             			types.String   `tfsdk:"arn"`
	MetastoreManagerRoleArn     types.String   `tfsdk:"metastore_manager_role_arn"`
	ID              			types.String   `tfsdk:"id"`
	Configurations 				types.List     `tfsdk:"configurations"`
	Tags            			types.Map      `tfsdk:"tags"`
	TagsAll         			types.Map      `tfsdk:"tags_all"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

type dataLakeConfigurationsData struct {
	EncryptionConfiguration types.Set `tfsdk:"encryption_configuration"`
	LifecycleConfiguration types.Set `tfsdk:"lifecycle_configuration"`
	Region types.String `tfsdk:"nested_optional"`
	ReplicationConfiguration types.Set `tfsdk:"replication_configuration"`
}

type dataLakeConfigurationsEncryption struct {
	KmsKeyID types.String `tfsdk:"kms_key_id"`
}

type dataLakeConfigurationsLifecycle struct {
	Expiration types.Set	`tfsdk:"expiration"`
	Transitions types.Set 	`tfsdk:"transitions"`
}

type dataLakeConfigurationsLifecycleExpiration struct {
	Days types.Int64	`tfsdk:"days"`
}

type dataLakeConfigurationsLifecycleTransitions struct {
	Days types.Int64			`tfsdk:"days"`
	StorageClass types.String	`tfsdk:"storage_class"`
}

type dataLakeConfigurationsReplicationConfiguration struct {
	RoleArn types.String `tfsdk:"role_arn"`
	Regions types.List `tfsdk:"regions"`
}

func generateDataLakeID(arn, region string) string {
    return arn + "|" + region
}

func extractRegionFromID(id string) string {
    parts := strings.Split(id, "|")
    if len(parts) < 2 {
        // Handle error or return a default value
        return ""
    }
    return parts[1]
}