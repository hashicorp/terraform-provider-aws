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
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeConfigurationModel](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeEncryptionConfigurationModel](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeReplicationConfigurationModel](ctx),
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
	var data dataLakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.CreateDataLakeInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDataLake(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameDataLake, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	data.DataLakeARN = flex.StringToFramework(ctx, output.DataLakes[0].DataLakeArn)
	data.setID()

	if _, err := waitDataLakeCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForCreation, ResNameDataLake, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dataLakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data dataLakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataLake, err := findDataLakeByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameDataLake, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, dataLake, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dataLakeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new dataLakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	if !new.Configurations.Equal(old.Configurations) {
		input := &securitylake.UpdateDataLakeInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateDataLake(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameDataLake, new.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		if _, err := waitDataLakeUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForUpdate, ResNameDataLake, new.ID.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
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

type dataLakeResourceModel struct {
	Configurations          fwtypes.ListNestedObjectValueOf[dataLakeConfigurationModel] `tfsdk:"configuration"`
	DataLakeARN             types.String                                                `tfsdk:"arn"`
	ID                      types.String                                                `tfsdk:"id"`
	MetaStoreManagerRoleARN fwtypes.ARN                                                 `tfsdk:"meta_store_manager_role_arn"`
	Tags                    types.Map                                                   `tfsdk:"tags"`
	TagsAll                 types.Map                                                   `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                              `tfsdk:"timeouts"`
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
	Regions types.Set   `tfsdk:"regions"`
	RoleARN fwtypes.ARN `tfsdk:"role_arn"`
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
