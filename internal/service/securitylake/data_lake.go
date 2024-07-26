// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const dataLakeMutexKey = "aws_securitylake_data_lake"

// @FrameworkResource(name="Data Lake")
// @Tags(identifierAttribute="arn")
func newDataLakeResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &dataLakeResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type dataLakeResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *dataLakeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_securitylake_data_lake"
}

func (r *dataLakeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"meta_store_manager_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"s3_bucket_arn":   framework.ARNAttributeComputedOnly(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEncryptionConfiguration: schema.ListAttribute{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeEncryptionConfigurationModel](ctx),
							Optional:   true,
							Computed:   true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									names.AttrKMSKeyID: types.StringType,
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrRegion: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"lifecycle_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeLifecycleConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"expiration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeLifecycleExpirationModel](ctx),
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
										CustomType: fwtypes.NewSetNestedObjectTypeOf[dataLakeLifecycleTransitionModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"days": schema.Int64Attribute{
													Optional: true,
												},
												names.AttrStorageClass: schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"replication_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeReplicationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"regions": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									names.AttrRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *dataLakeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data dataLakeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.CreateDataLakeInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.CreateDataLakeOutput, error) {
		return conn.CreateDataLake(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Data Lake", err.Error())

		return
	}

	// Set values for unknowns.
	dataLake := &output.DataLakes[0]
	data.DataLakeARN = fwflex.StringToFramework(ctx, dataLake.DataLakeArn)
	data.setID()

	dataLake, err = waitDataLakeCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Lake Data Lake (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	var configuration dataLakeConfigurationModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, dataLake, &configuration)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns after creation is complete.
	data.Configurations = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &configuration)
	data.S3BucketARN = fwflex.StringToFramework(ctx, dataLake.S3BucketArn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataLakeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data dataLakeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	dataLake, err := findDataLakeByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Lake Data Lake (%s)", data.ID.ValueString()), err.Error())

		return
	}

	var configuration dataLakeConfigurationModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, dataLake, &configuration)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Configurations = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &configuration)
	data.S3BucketARN = fwflex.StringToFramework(ctx, dataLake.S3BucketArn)

	// Transparent tagging fails with "ResourceNotFoundException: The request failed because the specified resource doesn't exist."
	// if the data lake's AWS Region isn't the configured one.
	if region := configuration.Region.ValueString(); region != r.Meta().Region {
		if tags, err := listTags(ctx, conn, data.ID.ValueString(), func(o *securitylake.Options) { o.Region = region }); err == nil {
			setTagsOut(ctx, Tags(tags))
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataLakeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new dataLakeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	if !new.Configurations.Equal(old.Configurations) {
		input := &securitylake.UpdateDataLakeInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.UpdateDataLakeOutput, error) {
			return conn.UpdateDataLake(ctx, input)
		})

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Security Lake Data Lake (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitDataLakeUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Lake Data Lake (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *dataLakeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data dataLakeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.DeleteDataLakeInput{
		Regions: []string{errs.Must(regionFromARNString(data.ID.ValueString()))},
	}

	_, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.DeleteDataLakeOutput, error) {
		return conn.DeleteDataLake(ctx, input)
	})

	// No data lake:
	// "An error occurred (AccessDeniedException) when calling the DeleteDataLake operation: User: ... is not authorized to perform: securitylake:DeleteDataLake", or
	// "UnauthorizedException: Unauthorized"
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform") ||
		tfawserr.ErrMessageContains(err, errCodeUnauthorizedException, "Unauthorized") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Lake Data Lake (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err = waitDataLakeDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Lake Data Lake (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *dataLakeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
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

func waitDataLakeCreated(ctx context.Context, conn *securitylake.Client, arn string, timeout time.Duration) (*awstypes.DataLakeResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataLakeStatusInitialized),
		Target:  enum.Slice(awstypes.DataLakeStatusCompleted),
		Refresh: statusDataLakeCreate(ctx, conn, arn),
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
		Refresh: statusDataLakeCreate(ctx, conn, arn),
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

// regionFromARNString return the AWS Region from the specified ARN string.
func regionFromARNString(s string) (string, error) {
	v, err := arn.Parse(s)

	if err != nil {
		return "", err
	}

	return v.Region, nil
}

type dataLakeResourceModel struct {
	Configurations          fwtypes.ListNestedObjectValueOf[dataLakeConfigurationModel] `tfsdk:"configuration"`
	DataLakeARN             types.String                                                `tfsdk:"arn"`
	ID                      types.String                                                `tfsdk:"id"`
	MetaStoreManagerRoleARN fwtypes.ARN                                                 `tfsdk:"meta_store_manager_role_arn"`
	S3BucketARN             types.String                                                `tfsdk:"s3_bucket_arn"`
	Tags                    types.Map                                                   `tfsdk:"tags"`
	TagsAll                 types.Map                                                   `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                              `tfsdk:"timeouts"`
}

func (model *dataLakeResourceModel) InitFromID() error {
	model.DataLakeARN = model.ID

	return nil
}

func (model *dataLakeResourceModel) setID() {
	model.ID = model.DataLakeARN
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
	Regions fwtypes.SetValueOf[types.String] `tfsdk:"regions"`
	RoleARN fwtypes.ARN                      `tfsdk:"role_arn"`
}

func retryDataLakeConflictWithMutex[T any](ctx context.Context, f func() (T, error)) (T, error) {
	conns.GlobalMutexKV.Lock(dataLakeMutexKey)
	defer conns.GlobalMutexKV.Unlock(dataLakeMutexKey)

	const dataLakeTimeout = 2 * time.Minute

	raw, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, dataLakeTimeout, func() (any, error) {
		return f()
	})
	if err != nil {
		var zero T
		return zero, err
	}

	return raw.(T), err
}
